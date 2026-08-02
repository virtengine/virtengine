#!/usr/bin/env python3

import tempfile
import unittest
import json
from pathlib import Path

import yaml

from validate_slurm_chart_semantics import DIGEST_IMAGE, HELM_RANDOM, load_rendered, result, validate, validate_source


ROOT = Path(__file__).resolve().parents[1]
TESTDATA = ROOT / "scripts" / "testdata" / "slurm-chart-semantics"


def load_documents(path: Path) -> list[dict]:
    documents = []
    for value in yaml.safe_load_all(path.read_text(encoding="utf-8")):
        if isinstance(value, dict):
            documents.append(value)
    return documents


class SlurmChartSemanticsTest(unittest.TestCase):
    def setUp(self) -> None:
        self.hardened = load_documents(TESTDATA / "hardened.yaml")

    def test_hardened_fixture_passes(self) -> None:
        self.assertEqual(validate(self.hardened), [])

    def assert_fixture_rejects(self, invariant: str) -> None:
        overlay = load_documents(TESTDATA / f"negative-{invariant}.yaml")
        failed = {finding.invariant for finding in validate([*self.hardened, *overlay])}
        self.assertIn(invariant, failed)

    def test_rejects_generated_secrets(self) -> None:
        self.assert_fixture_rejects("stable-secrets")

    def test_rejects_replica_partition_mismatch(self) -> None:
        self.assert_fixture_rejects("replica-capacity-equality")

    def test_rejects_mutable_images(self) -> None:
        self.assert_fixture_rejects("immutable-images")

    def test_rejects_privileged_root_containers(self) -> None:
        self.assert_fixture_rejects("least-privilege")

    def test_rejects_non_pvc_durable_state(self) -> None:
        self.assert_fixture_rejects("durable-state")

    def test_zero_document_input_fails(self) -> None:
        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as handle:
            path = Path(handle.name)
        try:
            with self.assertRaisesRegex(ValueError, "zero Kubernetes documents"):
                load_rendered(path)
        finally:
            path.unlink()

    def test_recursively_flattens_nested_lists(self) -> None:
        nested = {"apiVersion": "v1", "kind": "List", "items": [{"kind": "List", "items": self.hardened}]}
        self.assertEqual(validate([nested]), [])

    def test_rejects_annotation_with_inline_secret(self) -> None:
        for field, value in (("data", {"password": "aW5saW5l"}), ("stringData", {"password": "inline"})):
            with self.subTest(field=field):
                secret = {
                    "kind": "Secret",
                    "metadata": {"name": "inline", "annotations": {"virtengine.com/external-secret": "true", "virtengine.com/external-secret-ref": "vault/slurm"}},
                    field: value,
                }
                self.assertIn("stable-secrets", {finding.invariant for finding in validate([*self.hardened, secret])})

    def test_detects_all_forbidden_helm_generators(self) -> None:
        for function in ("randBytes", "genCA", "genPrivateKey", "genSelfSignedCert", "genSignedCert", "derivePassword", "htpasswd", "uuidv4"):
            with self.subTest(function=function):
                self.assertIsNotNone(HELM_RANDOM.search(f"{{{{ {function} 32 }}}}"))

    def test_rejects_sidecar_pvc_masking_primary_empty_dir(self) -> None:
        candidate = yaml.safe_load(yaml.safe_dump(self.hardened[1]))
        candidate["metadata"]["labels"]["app.kubernetes.io/component"] = "controller"
        candidate["metadata"]["name"] = "masked-controller"
        candidate["spec"]["template"]["spec"]["volumes"] = [
            {"name": "state", "emptyDir": {}},
            {"name": "sidecar-state", "persistentVolumeClaim": {"claimName": "logs"}},
        ]
        candidate["spec"]["template"]["spec"]["containers"][0]["name"] = "slurmctld"
        candidate["spec"]["template"]["spec"]["containers"][0]["volumeMounts"] = [{"name": "state", "mountPath": "/var/spool/slurm"}]
        sidecar = yaml.safe_load(yaml.safe_dump(candidate["spec"]["template"]["spec"]["containers"][0]))
        sidecar["name"] = "metrics"
        sidecar["volumeMounts"] = [{"name": "sidecar-state", "mountPath": "/var/spool/slurm"}]
        candidate["spec"]["template"]["spec"]["containers"].append(sidecar)
        self.assertIn("durable-state", {finding.invariant for finding in validate([*self.hardened, candidate])})

    def test_rejects_any_added_capability(self) -> None:
        candidate = yaml.safe_load(yaml.safe_dump(self.hardened[1]))
        candidate["spec"]["template"]["spec"]["containers"][0]["securityContext"]["capabilities"]["add"] = ["CHOWN"]
        self.assertIn("least-privilege", {finding.invariant for finding in validate([*self.hardened, candidate])})

    def test_accepts_pod_level_non_root_and_seccomp_inheritance(self) -> None:
        candidate = yaml.safe_load(yaml.safe_dump(self.hardened[1]))
        pod_security = candidate["spec"]["template"]["spec"]["securityContext"]
        pod_security.update({"runAsNonRoot": True, "runAsUser": 1000})
        container_security = candidate["spec"]["template"]["spec"]["containers"][0]["securityContext"]
        container_security.pop("runAsNonRoot")
        container_security.pop("runAsUser")
        self.assertNotIn("least-privilege", {finding.invariant for finding in validate([*self.hardened, candidate])})

    def test_accepts_preexisting_pvc(self) -> None:
        candidate = yaml.safe_load(yaml.safe_dump(self.hardened[2]))
        candidate["spec"].pop("volumeClaimTemplates")
        candidate["spec"]["template"]["spec"]["volumes"] = [{"name": "state", "persistentVolumeClaim": {"claimName": "slurmdbd-state"}}]
        findings = validate([*self.hardened, candidate])
        self.assertFalse(any(finding.invariant == "durable-state" and "slurm-database" in finding.location for finding in findings))

    def test_accepts_compressed_node_name(self) -> None:
        candidate = yaml.safe_load(yaml.safe_dump(self.hardened[0]))
        candidate["data"]["slurm.conf"] = "NodeName=slurm-compute-[0,1-1] CPUs=4\nPartitionName=normal Nodes=slurm-compute-[0,1-1] MaxNodes=2\n"
        self.assertNotIn("replica-capacity-equality", {finding.invariant for finding in validate([candidate, *self.hardened[1:]])})

    def test_counts_deployment_and_replication_controller_compute_replicas(self) -> None:
        config = yaml.safe_load(yaml.safe_dump(self.hardened[0]))
        config["data"]["slurm.conf"] = "NodeName=compute-[0-2]\nPartitionName=normal MaxNodes=3\n"
        deployment = yaml.safe_load(yaml.safe_dump(self.hardened[1]))
        deployment["kind"] = "Deployment"
        deployment["metadata"]["name"] = "compute-deployment"
        deployment["spec"]["replicas"] = 2
        replication_controller = yaml.safe_load(yaml.safe_dump(deployment))
        replication_controller["apiVersion"] = "v1"
        replication_controller["kind"] = "ReplicationController"
        replication_controller["metadata"]["name"] = "compute-rc"
        replication_controller["spec"]["replicas"] = 1
        self.assertNotIn("replica-capacity-equality", {finding.invariant for finding in validate([config, deployment, replication_controller, *self.hardened[2:]])})

    def test_strict_image_reference(self) -> None:
        digest = "a" * 64
        self.assertIsNotNone(DIGEST_IMAGE.fullmatch(f"ghcr.io/virtengine/slurm@sha256:{digest}"))
        for image in (f"@sha256:{digest}", f"repo@sha256:{digest.upper()}", f"repo@sha256:{digest[:-1]}", f"repo:{digest}"):
            with self.subTest(image=image):
                self.assertIsNone(DIGEST_IMAGE.fullmatch(image))

    def test_rejects_unknown_workload_shape(self) -> None:
        unknown = {"apiVersion": "example/v1", "kind": "MysteryWorkload", "metadata": {"name": "unknown"}, "spec": {"template": {"spec": {"containers": []}}}}
        self.assertIn("least-privilege", {finding.invariant for finding in validate([*self.hardened, unknown])})

    def test_canonical_source_reports_known_current_violations(self) -> None:
        findings = validate_source(ROOT / "deploy" / "slurm" / "slurm-cluster")
        failed = {finding.invariant for finding in findings}
        self.assertNotIn("stable-secrets", failed)
        self.assertTrue({"replica-capacity-equality", "immutable-images", "least-privilege"}.issubset(failed))
        self.assertNotIn("durable-state", failed)

    def test_source_rejects_missing_fail_closed_secret_guard(self) -> None:
        source = ROOT / "deploy" / "slurm" / "slurm-cluster"
        with tempfile.TemporaryDirectory() as directory:
            chart = Path(directory)
            (chart / "templates").mkdir()
            (chart / "values.yaml").write_text((source / "values.yaml").read_text(encoding="utf-8"), encoding="utf-8")
            guards = (source / "templates" / "secrets.yaml").read_text(encoding="utf-8")
            guards = guards.replace(".Values.munge.existingSecret", ".Values.munge.missingSecret")
            (chart / "templates" / "secrets.yaml").write_text(guards, encoding="utf-8")
            findings = validate_source(chart)
        self.assertTrue(any(finding.invariant == "stable-secrets" and "munge" in finding.message for finding in findings))

    def test_source_rejects_node_agent_tls_projection_mismatch(self) -> None:
        source = ROOT / "deploy" / "slurm" / "slurm-cluster"
        with tempfile.TemporaryDirectory() as directory:
            chart = Path(directory)
            (chart / "templates").mkdir()
            (chart / "values.yaml").write_text((source / "values.yaml").read_text(encoding="utf-8"), encoding="utf-8")
            for template in (source / "templates").glob("*.yaml"):
                content = template.read_text(encoding="utf-8").replace("path: client.crt", "path: tls.crt")
                (chart / "templates" / template.name).write_text(content, encoding="utf-8")
            findings = validate_source(chart)
        self.assertTrue(any(finding.invariant == "stable-secrets" and "projection mismatch" in finding.message for finding in findings))

    def test_shared_secret_schema_requirements_are_unconditional(self) -> None:
        schema = json.loads((ROOT / "deploy" / "slurm" / "slurm-cluster" / "values.schema.json").read_text(encoding="utf-8"))
        self.assertEqual(schema["definitions"]["secretComponent"]["required"], ["existingSecret", "secretKeyName"])
        self.assertEqual(schema["properties"]["database"]["required"], ["config"])
        self.assertEqual(schema["definitions"]["databaseSecret"]["required"], ["existingSecret", "secretPasswordKey"])
        self.assertEqual(schema["definitions"]["mariadbSecret"]["required"], ["existingSecret", "secretRootPasswordKey"])

    def test_source_diagnostic_only_promotes_stable_secrets(self) -> None:
        statuses = result([], diagnostic=True)["invariants"]
        self.assertEqual(statuses["stable-secrets"], "passed")
        self.assertEqual(
            {status for invariant, status in statuses.items() if invariant != "stable-secrets"},
            {"unverified"},
        )


if __name__ == "__main__":
    unittest.main()
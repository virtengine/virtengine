#!/usr/bin/env python3

import hashlib
import json
import shutil
import tempfile
import unittest
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

    def test_rejects_writable_root_filesystem(self) -> None:
        candidate = yaml.safe_load(yaml.safe_dump(self.hardened[1]))
        candidate["spec"]["template"]["spec"]["containers"][0]["securityContext"]["readOnlyRootFilesystem"] = False
        self.assertIn("least-privilege", {finding.invariant for finding in validate([*self.hardened, candidate])})

    def test_rejects_host_path_volume(self) -> None:
        candidate = yaml.safe_load(yaml.safe_dump(self.hardened[1]))
        candidate["spec"]["template"]["spec"]["volumes"] = [{"name": "cgroup", "hostPath": {"path": "/sys/fs/cgroup"}}]
        self.assertIn("least-privilege", {finding.invariant for finding in validate([*self.hardened, candidate])})

    def test_rejects_host_namespaces(self) -> None:
        for field in ("hostIPC", "hostNetwork", "hostPID"):
            with self.subTest(field=field):
                candidate = yaml.safe_load(yaml.safe_dump(self.hardened[1]))
                candidate["spec"]["template"]["spec"][field] = True
                self.assertIn("least-privilege", {finding.invariant for finding in validate([*self.hardened, candidate])})

    def test_rejects_unmasked_proc_and_host_process(self) -> None:
        for field, value in (("procMount", "Unmasked"), ("windowsOptions", {"hostProcess": True})):
            with self.subTest(field=field):
                candidate = yaml.safe_load(yaml.safe_dump(self.hardened[1]))
                candidate["spec"]["template"]["spec"]["containers"][0]["securityContext"][field] = value
                self.assertIn("least-privilege", {finding.invariant for finding in validate([*self.hardened, candidate])})

    def test_accepts_pod_level_non_root_and_seccomp_inheritance(self) -> None:
        candidate = yaml.safe_load(yaml.safe_dump(self.hardened[1]))
        pod_security = candidate["spec"]["template"]["spec"]["securityContext"]
        pod_security.update({"runAsNonRoot": True, "runAsUser": 1002, "runAsGroup": 1002})
        container_security = candidate["spec"]["template"]["spec"]["containers"][0]["securityContext"]
        container_security.pop("runAsNonRoot")
        container_security.pop("runAsUser")
        container_security.pop("runAsGroup")
        self.assertNotIn("least-privilege", {finding.invariant for finding in validate([*self.hardened, candidate])})

    def test_rejects_missing_or_inconsistent_slurm_daemon_users(self) -> None:
        for label, mutate in (
            ("missing SlurmUser", lambda config: config.replace("SlurmUser=slurm\n", "", 1)),
            ("missing SlurmdUser", lambda config: config.replace("SlurmdUser=slurm\n", "", 1)),
            ("different slurmd user", lambda config: config.replace("SlurmdUser=slurm", "SlurmdUser=slurmd", 1)),
        ):
            with self.subTest(label=label):
                candidate = yaml.safe_load(yaml.safe_dump(self.hardened[0]))
                candidate["data"]["slurm.conf"] = mutate(candidate["data"]["slurm.conf"])
                findings = validate([candidate, *self.hardened[1:]])
                self.assertTrue(any(finding.invariant == "least-privilege" and "user" in finding.location.lower() for finding in findings))

    def test_rejects_divergent_rendered_slurm_daemon_ids(self) -> None:
        candidate = yaml.safe_load(yaml.safe_dump(self.hardened[1]))
        security = candidate["spec"]["template"]["spec"]["containers"][0]["securityContext"]
        security.update(runAsUser=2002, runAsGroup=2002)
        findings = validate([self.hardened[0], candidate, *self.hardened[2:]])
        self.assertTrue(any(finding.invariant == "least-privilege" and "daemon identities" in finding.location for finding in findings))

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

    def test_accepts_multiple_enabled_node_pools(self) -> None:
        documents = load_documents(TESTDATA / "hardened-multiple-pools.yaml")
        self.assertNotIn("replica-capacity-equality", {finding.invariant for finding in validate(documents)})

    def test_accepts_partition_specific_membership(self) -> None:
        documents = load_documents(TESTDATA / "hardened-multiple-pools.yaml")
        config = documents[0]
        config["data"]["slurm.conf"] = "\n".join([
            "NodeName=slurm-compute-[0-1] CPUs=4",
            "NodeName=slurm-gpu-0 CPUs=8",
            "NodeName=slurm-memory-[0-1] CPUs=4",
            "PartitionName=normal Nodes=slurm-compute-[0-1],slurm-memory-[0-1] MaxNodes=4",
            "PartitionName=gpu Nodes=slurm-gpu-0 MaxNodes=1",
        ])
        self.assertNotIn("replica-capacity-equality", {finding.invariant for finding in validate(documents)})

    def test_rejects_partition_unknown_nodes_and_wrong_max(self) -> None:
        candidate = yaml.safe_load(yaml.safe_dump(self.hardened[0]))
        candidate["data"]["slurm.conf"] = "NodeName=slurm-compute-[0-1]\nPartitionName=normal Nodes=slurm-compute-0,missing-0 MaxNodes=1\n"
        self.assertIn("replica-capacity-equality", {finding.invariant for finding in validate([candidate, *self.hardened[1:]])})

    def test_rejects_unpartitioned_rendered_nodes(self) -> None:
        candidate = yaml.safe_load(yaml.safe_dump(self.hardened[0]))
        candidate["data"]["slurm.conf"] = "NodeName=slurm-compute-[0-1]\nPartitionName=normal Nodes=slurm-compute-0 MaxNodes=1\n"
        self.assertIn("replica-capacity-equality", {finding.invariant for finding in validate([candidate, *self.hardened[1:]])})

    def test_counts_deployment_and_replication_controller_compute_replicas(self) -> None:
        config = yaml.safe_load(yaml.safe_dump(self.hardened[0]))
        config["data"]["slurm.conf"] = "NodeName=compute-[0-2]\nPartitionName=normal Nodes=compute-[0-2] MaxNodes=3\n"
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

    def test_canonical_source_has_no_known_current_violations(self) -> None:
        findings = validate_source(ROOT / "deploy" / "slurm" / "slurm-cluster")
        failed = {finding.invariant for finding in findings}
        self.assertNotIn("stable-secrets", failed)
        self.assertNotIn("replica-capacity-equality", failed)
        self.assertNotIn("immutable-images", failed)
        self.assertNotIn("least-privilege", failed)
        self.assertNotIn("durable-state", failed)

    def test_source_rejects_missing_container_security_helper(self) -> None:
        source = ROOT / "deploy" / "slurm" / "slurm-cluster"
        with tempfile.TemporaryDirectory() as directory:
            chart = Path(directory) / "chart"
            shutil.copytree(source, chart)
            template = chart / "templates" / "compute-statefulset.yaml"
            content = template.read_text(encoding="utf-8")
            content = content.replace('{{- include "slurm-cluster.containerSecurityContext" (list . "munge") | nindent 12 }}', "{}", 1)
            template.write_text(content, encoding="utf-8")
            findings = validate_source(chart)
        self.assertTrue(any(finding.invariant == "least-privilege" and "security helper uses" in finding.message for finding in findings))

    def test_source_rejects_missing_or_zero_component_identity(self) -> None:
        source = ROOT / "deploy" / "slurm" / "slurm-cluster"
        for label, mutate in (
            ("missing", lambda values: values["securityIdentities"].pop("slurm")),
            ("zero uid", lambda values: values["securityIdentities"]["slurm"].update(uid=0)),
            ("zero gid", lambda values: values["securityIdentities"]["slurm"].update(gid=0)),
        ):
            with self.subTest(label=label), tempfile.TemporaryDirectory() as directory:
                chart = Path(directory) / "chart"
                shutil.copytree(source, chart)
                values = yaml.safe_load((chart / "values.yaml").read_text(encoding="utf-8"))
                mutate(values)
                (chart / "values.yaml").write_text(yaml.safe_dump(values), encoding="utf-8")
                findings = validate_source(chart)
            self.assertTrue(any(finding.invariant == "least-privilege" and ("securityIdentities.slurm" in finding.message or "shared slurm identity" in finding.message) for finding in findings))

    def test_source_rejects_missing_daemon_user_directives_and_divergent_ids(self) -> None:
        source = ROOT / "deploy" / "slurm" / "slurm-cluster"
        mutations = (
            ("missing SlurmUser", "templates/configmap.yaml", lambda text: text.replace('SlurmUser={{ include "slurm-cluster.slurmUser" . }}', "", 1)),
            ("missing SlurmdUser", "templates/configmap.yaml", lambda text: text.replace('SlurmdUser={{ include "slurm-cluster.slurmUser" . }}', "", 1)),
            ("divergent compute IDs", "values.yaml", lambda text: text.replace("  mariadb: { uid: 1004, gid: 1004 }", "  compute: { uid: 2002, gid: 2002 }\n  mariadb: { uid: 1004, gid: 1004 }", 1)),
        )
        for label, relative, mutate in mutations:
            with self.subTest(label=label), tempfile.TemporaryDirectory() as directory:
                chart = Path(directory) / "chart"
                shutil.copytree(source, chart)
                path = chart / relative
                path.write_text(mutate(path.read_text(encoding="utf-8")), encoding="utf-8")
                findings = validate_source(chart)
            self.assertTrue(any(finding.invariant == "least-privilege" for finding in findings))

    def test_source_rejects_mismatched_container_identity_binding(self) -> None:
        source = ROOT / "deploy" / "slurm" / "slurm-cluster"
        with tempfile.TemporaryDirectory() as directory:
            chart = Path(directory) / "chart"
            shutil.copytree(source, chart)
            template = chart / "templates" / "controller-statefulset.yaml"
            content = template.read_text(encoding="utf-8").replace(
                '(list . "controller")', '(list . "munge")', 1
            )
            template.write_text(content, encoding="utf-8")
            findings = validate_source(chart)
        self.assertTrue(any(finding.invariant == "least-privilege" and "identity contract for controller" in finding.message for finding in findings))

    def test_source_rejects_malicious_values_at_any_depth(self) -> None:
        source = ROOT / "deploy" / "slurm" / "slurm-cluster"
        mutations = (
            ("root sidecars", lambda values: values.update(extraContainers=[{"securityContext": {"privileged": True}}])),
            ("component context", lambda values: values["controller"].update(podSecurityContext={"hostPID": True})),
            ("node pool init", lambda values: values.update(nodePools=[{"name": "evil", "replicas": 1, "initContainers": []}])),
            ("nested host path", lambda values: values.update(extraVolumes=[{"name": "escape", "source": {"hostPath": {"path": "/"}}}])),
        )
        for label, mutate in mutations:
            with self.subTest(label=label), tempfile.TemporaryDirectory() as directory:
                chart = Path(directory) / "chart"
                shutil.copytree(source, chart)
                values = yaml.safe_load((chart / "values.yaml").read_text(encoding="utf-8"))
                mutate(values)
                (chart / "values.yaml").write_text(yaml.safe_dump(values), encoding="utf-8")
                findings = validate_source(chart)
            self.assertTrue(any(finding.invariant == "least-privilege" and "security override is forbidden" in finding.message for finding in findings))

    def test_source_rejects_cgroup_plugin_override(self) -> None:
        source = ROOT / "deploy" / "slurm" / "slurm-cluster"
        with tempfile.TemporaryDirectory() as directory:
            chart = Path(directory) / "chart"
            shutil.copytree(source, chart)
            values = yaml.safe_load((chart / "values.yaml").read_text(encoding="utf-8"))
            values["controller"]["config"]["taskPlugin"] = "task/cgroup,task/affinity"
            (chart / "values.yaml").write_text(yaml.safe_dump(values), encoding="utf-8")
            findings = validate_source(chart)
        self.assertTrue(any(finding.invariant == "least-privilege" and "rootless prototype requires" in finding.message for finding in findings))

    def test_source_rejects_mutable_or_bypassed_image_contracts(self) -> None:
        source = ROOT / "deploy" / "slurm" / "slurm-cluster"
        mutations = {
            "tagged default": ("values.yaml", 'reference: ""', 'reference: "busybox:latest"'),
            "uppercase digest": ("values.schema.json", "[a-f0-9]{64}", "[A-F0-9]{64}"),
            "template helper bypass": ("templates/mariadb-statefulset.yaml", 'include "slurm-cluster.mariadb.image" .', ".Values.mariadb.image.reference"),
            "helper source bypass": ("templates/_helpers.tpl", '.Values.mariadb.image.reference', '.Values.compute.image.reference'),
            "bad pull policy enum": ("values.schema.json", '"Never"]', '"Sometimes"]'),
            "bad pull policy source": ("templates/mariadb-statefulset.yaml", '.Values.mariadb.image.pullPolicy', '.Values.compute.image.pullPolicy'),
        }
        for label, (relative, old, new) in mutations.items():
            with self.subTest(label=label), tempfile.TemporaryDirectory() as directory:
                chart = Path(directory)
                (chart / "templates").mkdir()
                for path in source.iterdir():
                    if path.is_file():
                        content = path.read_text(encoding="utf-8")
                        if path.name == relative:
                            content = content.replace(old, new, 1)
                        (chart / path.name).write_text(content, encoding="utf-8")
                for template in (source / "templates").glob("*"):
                    if template.is_file():
                        content = template.read_text(encoding="utf-8")
                        if f"templates/{template.name}" == relative:
                            content = content.replace(old, new, 1)
                        (chart / "templates" / template.name).write_text(content, encoding="utf-8")
                self.assertIn("immutable-images", {finding.invariant for finding in validate_source(chart)})

    def test_source_rejects_partition_capacity_override(self) -> None:
        source = ROOT / "deploy" / "slurm" / "slurm-cluster"
        with tempfile.TemporaryDirectory() as directory:
            chart = Path(directory)
            (chart / "templates").mkdir()
            for template in (source / "templates").glob("*"):
                if template.is_file():
                    (chart / "templates" / template.name).write_text(template.read_text(encoding="utf-8"), encoding="utf-8")
            values = yaml.safe_load((source / "values.yaml").read_text(encoding="utf-8"))
            values["partitions"][0]["maxNodes"] = 32
            (chart / "values.yaml").write_text(yaml.safe_dump(values), encoding="utf-8")
            findings = validate_source(chart)
        self.assertTrue(any(finding.invariant == "replica-capacity-equality" and "overrides derived maxNodes" in finding.message for finding in findings))

    def test_source_rejects_missing_zero_capacity_guard(self) -> None:
        source = ROOT / "deploy" / "slurm" / "slurm-cluster"
        with tempfile.TemporaryDirectory() as directory:
            chart = Path(directory)
            (chart / "templates").mkdir()
            (chart / "values.yaml").write_text((source / "values.yaml").read_text(encoding="utf-8"), encoding="utf-8")
            for template in (source / "templates").glob("*"):
                if template.is_file():
                    content = template.read_text(encoding="utf-8").replace('fail "at least one compute replica must be enabled"', 'printf "no capacity"')
                    (chart / "templates" / template.name).write_text(content, encoding="utf-8")
            findings = validate_source(chart)
        self.assertTrue(any(finding.invariant == "replica-capacity-equality" and "at least one compute replica" in finding.message for finding in findings))

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

    def test_image_schema_requirements_are_workload_conditional(self) -> None:
        schema = json.loads((ROOT / "deploy" / "slurm" / "slurm-cluster" / "values.schema.json").read_text(encoding="utf-8"))
        self.assertEqual(schema["definitions"]["secretComponent"]["required"], ["existingSecret", "secretKeyName"])
        self.assertEqual(schema["properties"]["database"]["required"], ["config"])
        self.assertEqual(schema["definitions"]["databaseSecret"]["required"], ["existingSecret", "secretPasswordKey"])
        self.assertEqual(schema["definitions"]["mariadbSecret"]["required"], ["existingSecret", "secretRootPasswordKey"])
        conditional_requirements = json.dumps(schema["allOf"], sort_keys=True)
        for image_owner in ("controller", "database", "mariadb", "munge", "compute", "nodeAgent", "utilityImage"):
            self.assertIn(image_owner, conditional_requirements)

    def test_pull_policy_schema_is_required_and_enumerated(self) -> None:
        schema = json.loads((ROOT / "deploy" / "slurm" / "slurm-cluster" / "values.schema.json").read_text(encoding="utf-8"))
        pinned_image = schema["definitions"]["pinnedImage"]
        self.assertEqual(pinned_image["properties"]["pullPolicy"], {"enum": ["Always", "IfNotPresent", "Never"]})
        self.assertEqual(pinned_image["required"], ["reference", "pullPolicy"])

    def test_capacity_schema_forbids_partition_overrides(self) -> None:
        schema = json.loads((ROOT / "deploy" / "slurm" / "slurm-cluster" / "values.schema.json").read_text(encoding="utf-8"))
        self.assertEqual(schema["properties"]["nodePools"]["items"]["required"], ["name", "replicas"])
        pool_condition = schema["properties"]["nodePools"]["items"]["allOf"][0]
        self.assertEqual(pool_condition["if"]["properties"]["enabled"]["const"], False)
        self.assertEqual(pool_condition["else"]["properties"]["replicas"]["minimum"], 1)
        compute_condition = schema["properties"]["compute"]["allOf"][0]
        self.assertEqual(compute_condition["then"]["properties"]["replicas"]["minimum"], 1)
        forbidden = schema["properties"]["partitions"]["items"]["not"]["anyOf"]
        self.assertEqual(forbidden, [{"required": ["nodes"]}, {"required": ["maxNodes"]}])

    def test_capacity_schema_reserves_resource_names_and_partition_selectors(self) -> None:
        schema = json.loads((ROOT / "deploy" / "slurm" / "slurm-cluster" / "values.schema.json").read_text(encoding="utf-8"))
        pool_name = schema["properties"]["nodePools"]["items"]["properties"]["name"]
        self.assertTrue({"controller", "slurmdbd", "database", "mariadb", "compute", "munge", "node-agent"}.issubset(pool_name["not"]["enum"]))
        partition = schema["properties"]["partitions"]["items"]
        self.assertEqual(partition["properties"]["nodePools"]["minItems"], 1)
        self.assertTrue(partition["properties"]["nodePools"]["uniqueItems"])
        self.assertEqual(partition["allOf"][0]["else"]["required"], ["nodePools"])

    def test_security_identity_and_recursive_override_schema_contracts(self) -> None:
        schema = json.loads((ROOT / "deploy" / "slurm" / "slurm-cluster" / "values.schema.json").read_text(encoding="utf-8"))
        components = ["munge", "slurm", "mariadb", "nodeAgent", "utility"]
        self.assertEqual(schema["required"], ["securityIdentities"])
        self.assertEqual(schema["definitions"]["securityIdentities"]["required"], components)
        identity = schema["definitions"]["securityIdentity"]
        self.assertEqual(identity["required"], ["uid", "gid"])
        self.assertEqual(identity["properties"]["uid"], {"type": "integer", "minimum": 1})
        self.assertEqual(identity["properties"]["gid"], {"type": "integer", "minimum": 1})
        slurm_identity = schema["definitions"]["slurmSecurityIdentity"]
        self.assertEqual(slurm_identity["required"], ["username", "uid", "gid"])
        self.assertIn("pattern", slurm_identity["properties"]["username"])
        self.assertNotIn("controller", schema["definitions"]["securityIdentities"]["properties"])
        self.assertNotIn("database", schema["definitions"]["securityIdentities"]["properties"])
        self.assertNotIn("compute", schema["definitions"]["securityIdentities"]["properties"])
        recursive = json.dumps(schema["definitions"]["noSecurityOverrides"], sort_keys=True)
        for key in ("securityContext", "podSecurityContext", "containerSecurityContext", "extraContainers", "initContainers", "hostPID", "hostIPC", "hostNetwork", "hostUsers", "hostPath", "hostPaths"):
            self.assertIn(f'"{key}"', recursive)
            self.assertIs(schema["properties"][key], False)
        self.assertEqual(schema["properties"]["extraVolumes"]["items"], {"$ref": "#/definitions/noSecurityOverrides"})
        profile = schema["properties"]["controller"]["properties"]["config"]
        self.assertEqual(profile["properties"]["jobAcctGatherType"], {"const": "jobacct_gather/linux"})
        self.assertEqual(profile["properties"]["proctrackType"], {"const": "proctrack/linuxproc"})
        self.assertEqual(profile["properties"]["taskPlugin"], {"const": "task/affinity"})

    def test_source_requires_hash_stable_dns_and_selector_guards(self) -> None:
        source = ROOT / "deploy" / "slurm" / "slurm-cluster"
        with tempfile.TemporaryDirectory() as directory:
            chart = Path(directory)
            (chart / "templates").mkdir()
            (chart / "values.yaml").write_text((source / "values.yaml").read_text(encoding="utf-8"), encoding="utf-8")
            (chart / "values.schema.json").write_text((source / "values.schema.json").read_text(encoding="utf-8"), encoding="utf-8")
            for template in (source / "templates").glob("*"):
                if template.is_file():
                    content = template.read_text(encoding="utf-8").replace("sha256sum $raw | trunc 8", "trunc 8 $raw")
                    (chart / "templates" / template.name).write_text(content, encoding="utf-8")
            findings = validate_source(chart)
        self.assertTrue(any(finding.invariant == "replica-capacity-equality" and "sha256sum" in finding.message for finding in findings))

    def test_long_dns_names_are_bounded_for_ordinals_and_collision_resistant(self) -> None:
        def bounded(raw: str, replicas: int) -> str:
            ordinal_budget = 1 + len(str(max(replicas - 1, 0))) if replicas > 0 else 0
            limit = 63 - ordinal_budget
            if len(raw) <= limit:
                return raw.rstrip("-")
            digest = hashlib.sha256(raw.encode()).hexdigest()[:8]
            return f"{raw[:limit - 9].rstrip('-')}-{digest}"

        common = "release-with-a-deliberately-long-name-slurm-cluster-pool-prefix-"
        first = bounded(common + "alpha", 10_000)
        second = bounded(common + "bravo", 10_000)
        self.assertNotEqual(first, second)
        self.assertLessEqual(len(f"{first}-9999"), 63)
        self.assertLessEqual(len(f"{second}-9999"), 63)
        self.assertRegex(first, r"-[a-f0-9]{8}$")

    def test_source_diagnostic_promotes_source_verifiable_invariants(self) -> None:
        statuses = result([], diagnostic=True)["invariants"]
        self.assertEqual(statuses["stable-secrets"], "passed")
        self.assertEqual(statuses["immutable-images"], "passed")
        self.assertEqual(statuses["replica-capacity-equality"], "unverified")
        self.assertEqual(
            {status for invariant, status in statuses.items() if invariant not in {"stable-secrets", "immutable-images"}},
            {"unverified"},
        )


if __name__ == "__main__":
    unittest.main()
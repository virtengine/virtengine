from __future__ import annotations

import importlib.util
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
SCRIPT_PATH = REPO_ROOT / ".github" / "scripts" / "validate_inference_deployment_policy.py"


def load_validator_module():
    spec = importlib.util.spec_from_file_location("validate_inference_deployment_policy", SCRIPT_PATH)
    module = importlib.util.module_from_spec(spec)
    assert spec is not None and spec.loader is not None
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module


class InferenceDeploymentPolicyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.validator = load_validator_module()

    def write_file(self, root: Path, relative_path: str, content: str) -> Path:
        target = root / relative_path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(content, encoding="utf-8")
        return target

    def test_validator_rejects_stub_flag_in_deployment_surface(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            deploy_root = root / "deploy"
            self.write_file(
                deploy_root,
                "kubernetes/base/inference-sidecar.yaml",
                """
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
        - name: inference-sidecar
          args:
            - --grpc-addr=:50051
            - --allow-fallback-to-stub
""".strip(),
            )

            errors = self.validator.validate_surfaces(
                [self.validator.PolicySurface(deploy_root, recursive=True)]
            )

            self.assertTrue(any("--allow-fallback-to-stub" in error for error in errors))

    def test_validator_rejects_local_stub_marker_in_workflow(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            workflows_root = root / ".github" / "workflows"
            self.write_file(
                workflows_root,
                "deploy.yaml",
                """
name: Deploy
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - run: echo "local_stub accepted"
""".strip(),
            )

            errors = self.validator.validate_surfaces(
                [self.validator.PolicySurface(workflows_root, recursive=True)]
            )

            self.assertTrue(any("local_stub" in error for error in errors))

    def test_validator_ignores_dev_only_docs_outside_production_surfaces(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            internal_docs = root / "_docs"
            self.write_file(
                internal_docs,
                "inference-fallback-behavior.md",
                "Use --allow-fallback-to-stub only for local testing; local_stub indicates fallback mode.",
            )

            errors = self.validator.validate_surfaces(
                [self.validator.PolicySurface(root / "deploy", recursive=True)]
            )

            self.assertEqual(errors, [])

    def test_validator_cli_passes_on_repo_surfaces(self) -> None:
        result = subprocess.run(
            [sys.executable, str(SCRIPT_PATH)],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=False,
        )

        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertIn("PASS inference deployment policy", result.stdout)

    def test_validator_cli_fails_for_invalid_fixture(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            self.write_file(
                root,
                "docs/inference-sidecar-deployment.md",
                "Production example: --allow-fallback-to-stub",
            )

            result = subprocess.run(
                [sys.executable, str(SCRIPT_PATH), "--paths", str(root / "docs")],
                cwd=REPO_ROOT,
                capture_output=True,
                text=True,
                check=False,
            )

            self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertIn("--allow-fallback-to-stub", result.stdout)

    def test_render_policy_accepts_hardened_inference_sidecar_fixture(self) -> None:
        errors = self.validator.validate_rendered_yaml(valid_inference_render())

        self.assertEqual(errors, [])

    def test_render_policy_rejects_missing_sidecar(self) -> None:
        rendered = valid_inference_render().replace("name: inference-sidecar", "name: other-sidecar")

        errors = self.validator.validate_rendered_yaml(rendered)

        self.assertTrue(any("missing StatefulSet/inference-sidecar" in error for error in errors))

    def test_render_policy_rejects_mutable_sidecar_image(self) -> None:
        rendered = valid_inference_render().replace(
            "ghcr.io/virtengine/inference-sidecar@sha256:" + "a" * 64,
            "ghcr.io/virtengine/inference-sidecar:latest",
        )

        errors = self.validator.validate_rendered_yaml(rendered)

        self.assertTrue(any("immutable digest" in error for error in errors))

    def test_render_policy_rejects_fallback_and_plaintext_tokens(self) -> None:
        rendered = valid_inference_render().replace(
            "- --require-mtls=true",
            "- --require-mtls=false\n            - --allow-fallback-to-stub",
        )

        errors = self.validator.validate_rendered_yaml(rendered)

        self.assertTrue(any("require-mtls=true" in error for error in errors))
        self.assertTrue(any("--allow-fallback-to-stub" in error for error in errors))

    def test_render_policy_rejects_missing_absolute_secret_mounts(self) -> None:
        rendered = valid_inference_render().replace(
            "mountPath: /var/run/secrets/virtengine/inference-sidecar/server",
            "mountPath: var/run/secrets/virtengine/inference-sidecar/server",
        )

        errors = self.validator.validate_rendered_yaml(rendered)

        self.assertTrue(any("absolute" in error and "secret mount" in error for error in errors))

    def test_render_policy_rejects_sidecar_replica_drift(self) -> None:
        rendered = valid_inference_render().replace("replicas: 1", "replicas: 2", 1)

        errors = self.validator.validate_rendered_yaml(rendered)

        self.assertTrue(any("exactly one replica" in error for error in errors))

    def test_render_policy_rejects_missing_security_readiness_network_pdb_or_bundle(self) -> None:
        rendered = valid_inference_render()
        rendered = rendered.replace("readOnlyRootFilesystem: true", "readOnlyRootFilesystem: false")
        rendered = rendered.replace("path: /health", "path: /ready")
        rendered = rendered.replace("kind: PodDisruptionBudget", "kind: ConfigMap", 1)
        rendered = rendered.replace("name: inference-sidecar-model-bundle", "name: other-model-bundle")
        rendered = rendered.replace("kind: NetworkPolicy", "kind: ConfigMap", 1)

        errors = self.validator.validate_rendered_yaml(rendered)

        self.assertTrue(any("read-only root filesystem" in error for error in errors))
        self.assertTrue(any("/health" in error for error in errors))
        self.assertTrue(any("PodDisruptionBudget/inference-sidecar-pdb" in error for error in errors))
        self.assertTrue(any("model bundle" in error for error in errors))
        self.assertTrue(any("NetworkPolicy/inference-sidecar" in error for error in errors))

    def test_render_policy_rejects_wrong_ports_public_service_and_raw_keys(self) -> None:
        rendered = valid_inference_render()
        rendered = rendered.replace("containerPort: 50051", "containerPort: 50052")
        rendered = rendered.replace("port: 50051", "port: 50052")
        rendered = rendered.replace("type: ClusterIP", "type: LoadBalancer")
        rendered = rendered.replace("property: server-cert", "property: consensus-private-key")

        errors = self.validator.validate_rendered_yaml(rendered)

        self.assertTrue(any("50051" in error for error in errors))
        self.assertTrue(any("public Service" in error for error in errors))
        self.assertTrue(any("raw signing/consensus" in error for error in errors))

    def test_render_policy_rejects_missing_model_metadata_external_secret_mount_and_reloader(self) -> None:
        rendered = valid_inference_render()
        rendered = rendered.replace('reloader.stakater.com/auto: "true"', 'reloader.stakater.com/auto: "false"')
        rendered = rendered.replace("name: inference-sidecar-model-metadata", "name: missing-model-metadata")
        rendered = rendered.replace(
            "mountPath: /var/run/secrets/virtengine/inference-sidecar/model-metadata",
            "mountPath: /var/run/secrets/virtengine/inference-sidecar/other-metadata",
        )

        errors = self.validator.validate_rendered_yaml(rendered)

        self.assertTrue(any("reloader.stakater.com/auto" in error for error in errors))
        self.assertTrue(any("model metadata ExternalSecret" in error for error in errors))
        self.assertTrue(any("ExternalSecret/inference-sidecar-model-metadata" in error for error in errors))

    def test_render_policy_rejects_broad_sidecar_network_policy_and_public_routes(self) -> None:
        rendered = valid_inference_render().replace(
            "from:\n        - podSelector:",
            "from: []\n      # blocked broad sidecar ingress\n      ports:\n        - protocol: TCP\n          port: 50051\n    - from:\n        - podSelector:",
            1,
        )
        rendered += """
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: inference-sidecar-public
spec:
  rules:
    - http:
        paths:
          - backend:
              service:
                name: inference-sidecar
                port:
                  number: 50051
""".rstrip()

        errors = self.validator.validate_rendered_yaml(rendered)

        self.assertTrue(any("broad rule from: []" in error for error in errors))
        self.assertTrue(any("Ingress/inference-sidecar-public" in error for error in errors))

    def test_render_policy_rejects_hpa_and_bad_external_secret_keys(self) -> None:
        rendered = valid_inference_render().replace("secretKey: model-digest", "secretKey: consensus-private-key", 1)
        rendered += """
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: inference-sidecar
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: StatefulSet
    name: inference-sidecar
""".rstrip()

        errors = self.validator.validate_rendered_yaml(rendered)

        self.assertTrue(any("must not have an HPA" in error for error in errors))
        self.assertTrue(any("ExternalSecret/inference-sidecar-model-metadata" in error for error in errors))
        self.assertTrue(any("raw signing/consensus" in error for error in errors))

    def test_render_policy_detects_canonical_infra_drift(self) -> None:
        errors = self.validator.compare_rendered_yaml("kind: Service\nmetadata:\n  name: a\n", "kind: Service\nmetadata:\n  name: b\n")

        self.assertTrue(any("render differs" in error for error in errors))

    def test_render_policy_covers_staging_compatibility_render(self) -> None:
        self.assertIn(
            ("deploy/kubernetes/overlays/staging", "infra/kubernetes/overlays/staging"),
            self.validator.RENDER_PAIRS,
        )

    def test_render_policy_rejects_validator_plaintext_sidecar_client(self) -> None:
        rendered = valid_inference_render().replace(
            "name: VEID_INFERENCE_SIDECAR_TLS\n              value: \"true\"",
            "name: VEID_INFERENCE_SIDECAR_TLS\n              value: \"false\"",
        )

        errors = self.validator.validate_rendered_yaml(rendered)

        self.assertTrue(any("VEID_INFERENCE_SIDECAR_TLS=true" in error for error in errors))

    def test_render_policy_rejects_missing_serving_endpoint(self) -> None:
        rendered = valid_inference_render().replace(
            "            - --serving-url=http://tf-serving.virtengine-runtime.svc.cluster.local:8501\n",
            "",
        )

        errors = self.validator.validate_rendered_yaml(rendered)

        self.assertTrue(any("--serving-url=" in error for error in errors))

    def test_render_policy_rejects_duplicate_named_resource(self) -> None:
        rendered = valid_inference_render() + """
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: inference-sidecar
spec:
  replicas: 1
""".rstrip()

        errors = self.validator.validate_rendered_yaml(rendered)

        self.assertTrue(any("duplicate StatefulSet/inference-sidecar" in error for error in errors))

    def test_environment_policy_rejects_cross_environment_inference_secret_refs(self) -> None:
        prod_errors = self.validator.validate_environment_inference_secret_refs(
            valid_inference_render().replace("secret/virtengine/prod/", "secret/virtengine/dev/"),
            "prod",
        )
        staging_render = valid_inference_render().replace(
            "secret/virtengine/prod/", "secret/virtengine/staging/"
        )
        staging_errors = self.validator.validate_environment_inference_secret_refs(
            staging_render,
            "staging",
        )

        self.assertTrue(any("prod" in error and "dev" in error for error in prod_errors))
        self.assertEqual(staging_errors, [])


def valid_inference_render() -> str:
    digest = "a" * 64
    return f"""
apiVersion: v1
kind: ServiceAccount
metadata:
  name: inference-sidecar
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: inference-sidecar
  annotations:
    reloader.stakater.com/auto: "true"
    virtengine.com/model-digest: {"b" * 64}
    virtengine.com/runtime-digest: {"c" * 64}
    virtengine.com/manifest-digest: {"d" * 64}
spec:
  serviceName: inference-sidecar
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: inference-sidecar
  template:
    metadata:
      labels:
        app.kubernetes.io/name: inference-sidecar
    spec:
      serviceAccountName: inference-sidecar
      securityContext:
        runAsNonRoot: true
        seccompProfile:
          type: RuntimeDefault
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution: []
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: topology.kubernetes.io/zone
          whenUnsatisfiable: ScheduleAnyway
      containers:
        - name: inference-sidecar
          image: ghcr.io/virtengine/inference-sidecar@sha256:{digest}
          args:
            - --grpc-addr=:50051
            - --metrics-addr=:9092
            - --require-mtls=true
            - --tls-cert-file=/var/run/secrets/virtengine/inference-sidecar/server/tls.crt
            - --tls-key-file=/var/run/secrets/virtengine/inference-sidecar/server/tls.key
            - --tls-client-ca-file=/var/run/secrets/virtengine/inference-sidecar/client-ca/ca.crt
            - --force-cpu=true
            - --random-seed=42
            - --model-path=/models/trust-score/v1/model
            - --model-version=v1.0.0
            - --expected-hash={"b" * 64}
            - --manifest-path=/models/trust-score/v1/release_manifest.json
            - --serving-url=http://tf-serving.virtengine-runtime.svc.cluster.local:8501
          ports:
            - name: grpc
              containerPort: 50051
            - name: metrics
              containerPort: 9092
          readinessProbe:
            httpGet:
              path: /health
              port: metrics
          livenessProbe:
            httpGet:
              path: /health
              port: metrics
          startupProbe:
            httpGet:
              path: /health
              port: metrics
          resources:
            requests:
              cpu: "1000m"
              memory: "2Gi"
            limits:
              cpu: "2000m"
              memory: "4Gi"
          volumeMounts:
            - name: server-mtls
              mountPath: /var/run/secrets/virtengine/inference-sidecar/server
              readOnly: true
            - name: client-ca
              mountPath: /var/run/secrets/virtengine/inference-sidecar/client-ca
              readOnly: true
            - name: model-bundle
              mountPath: /models/trust-score/v1
              readOnly: true
            - name: model-metadata
              mountPath: /var/run/secrets/virtengine/inference-sidecar/model-metadata
              readOnly: true
            - name: tmp
              mountPath: /tmp
          securityContext:
            runAsNonRoot: true
            readOnlyRootFilesystem: true
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
      volumes:
        - name: server-mtls
          secret:
            secretName: inference-sidecar-server-mtls
        - name: client-ca
          secret:
            secretName: inference-sidecar-client-ca
        - name: model-bundle
          persistentVolumeClaim:
            claimName: inference-sidecar-model-bundle
            readOnly: true
        - name: model-metadata
          secret:
            secretName: inference-sidecar-model-metadata
        - name: tmp
          emptyDir:
            sizeLimit: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: inference-sidecar
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: inference-sidecar
  ports:
    - name: grpc
      port: 50051
    - name: metrics
      port: 9092
---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: inference-sidecar-pdb
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: inference-sidecar
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: inference-sidecar-model-bundle
  annotations:
    virtengine.com/model-digest: {"b" * 64}
    virtengine.com/runtime-digest: {"c" * 64}
    virtengine.com/manifest-digest: {"d" * 64}
spec: {{}}
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: inference-sidecar
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: inference-sidecar
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: virtengine-validator
      ports:
        - protocol: TCP
          port: 50051
    - from:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: monitoring
      ports:
        - protocol: TCP
          port: 9092
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: kube-system
      ports:
        - protocol: UDP
          port: 53
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: virtengine-runtime
          podSelector:
            matchLabels:
              app.kubernetes.io/name: tf-serving
      ports:
        - protocol: TCP
          port: 8501
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: inference-sidecar-server-mtls
spec:
  data:
    - secretKey: tls.crt
      remoteRef:
        key: secret/virtengine/prod/inference-sidecar/server-mtls
        property: server-cert
    - secretKey: tls.key
      remoteRef:
        key: secret/virtengine/prod/inference-sidecar/server-mtls
        property: server-key
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: inference-sidecar-client-ca
spec:
  data:
    - secretKey: ca.crt
      remoteRef:
        key: secret/virtengine/prod/inference-sidecar/client-ca
        property: client-ca
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: inference-sidecar-client-mtls
spec:
  data:
    - secretKey: tls.crt
      remoteRef:
        key: secret/virtengine/prod/inference-sidecar/validator-client-mtls
        property: client-cert
    - secretKey: tls.key
      remoteRef:
        key: secret/virtengine/prod/inference-sidecar/validator-client-mtls
        property: client-key
    - secretKey: ca.crt
      remoteRef:
        key: secret/virtengine/prod/inference-sidecar/server-ca
        property: server-ca
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: inference-sidecar-model-metadata
spec:
  data:
    - secretKey: model-version
      remoteRef:
        key: secret/virtengine/prod/inference-sidecar/model-metadata
        property: model-version
    - secretKey: model-digest
      remoteRef:
        key: secret/virtengine/prod/inference-sidecar/model-metadata
        property: model-digest
    - secretKey: runtime-digest
      remoteRef:
        key: secret/virtengine/prod/inference-sidecar/model-metadata
        property: runtime-digest
    - secretKey: manifest-digest
      remoteRef:
        key: secret/virtengine/prod/inference-sidecar/model-metadata
        property: manifest-digest
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: virtengine-validator
spec:
  template:
    spec:
      containers:
        - name: virtengine-validator
          env:
            - name: VEID_INFERENCE_SIDECAR_ADDR
              value: inference-sidecar.virtengine.svc.cluster.local:50051
            - name: VEID_INFERENCE_SIDECAR_TLS
              value: "true"
            - name: VEID_INFERENCE_SIDECAR_TLS_CERT_FILE
              value: /var/run/secrets/virtengine/inference-sidecar-client/tls.crt
            - name: VEID_INFERENCE_SIDECAR_TLS_KEY_FILE
              value: /var/run/secrets/virtengine/inference-sidecar-client/tls.key
            - name: VEID_INFERENCE_SIDECAR_TLS_CA_FILE
              value: /var/run/secrets/virtengine/inference-sidecar-client/ca.crt
          volumeMounts:
            - name: inference-sidecar-client-mtls
              mountPath: /var/run/secrets/virtengine/inference-sidecar-client
              readOnly: true
      volumes:
        - name: inference-sidecar-client-mtls
          secret:
            secretName: inference-sidecar-client-mtls
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: virtengine-validator
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: virtengine-validator
  policyTypes:
    - Egress
  egress:
    - to:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: inference-sidecar
      ports:
        - protocol: TCP
          port: 50051
""".strip()


if __name__ == "__main__":
    unittest.main()

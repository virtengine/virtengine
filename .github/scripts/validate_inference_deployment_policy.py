#!/usr/bin/env python3

from __future__ import annotations

import argparse
import re
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]


FORBIDDEN_TOKENS = (
    "--allow-fallback-to-stub",
    "--require-mtls=false",
    "--serving-fallback-url",
    "local_stub",
)

RAW_KEY_PATTERNS = (
    "consensus-private-key",
    "validator-private-key",
    "receipt-signing-private-key",
    "inference-signing-private-key",
)


@dataclass(frozen=True)
class PolicySurface:
    path: Path
    recursive: bool = False


@dataclass(frozen=True)
class RenderedResource:
    kind: str
    name: str
    text: str


DEFAULT_SURFACES = (
    PolicySurface(ROOT / "deploy", recursive=True),
    PolicySurface(ROOT / "_build", recursive=True),
    PolicySurface(ROOT / ".github" / "workflows", recursive=True),
    PolicySurface(ROOT / "docs", recursive=True),
)

RENDER_PAIRS = (
    ("deploy/kubernetes/base", "infra/kubernetes/base"),
    ("deploy/kubernetes/overlays/staging", "infra/kubernetes/overlays/staging"),
    ("deploy/kubernetes/overlays/prod", "infra/kubernetes/overlays/prod"),
)


def repo_path(path: Path) -> str:
    try:
        return str(path.relative_to(ROOT)).replace("\\", "/")
    except ValueError:
        return str(path)


def iter_surface_files(surfaces: list[PolicySurface]) -> list[Path]:
    files: list[Path] = []
    for surface in surfaces:
        if not surface.path.exists():
            continue
        if surface.path.is_file():
            files.append(surface.path)
            continue

        pattern = "**/*" if surface.recursive else "*"
        for candidate in sorted(surface.path.glob(pattern)):
            if candidate.is_file():
                files.append(candidate)
    return files


def markdown_line_is_deployable_example(line: str) -> bool:
    lowered = line.lower()
    if "non-production" in lowered or "development" in lowered or "local testing" in lowered:
        return False
    return any(marker in lowered for marker in ("production", "deploy/kubernetes", "kubectl", "args:"))


def find_violations(path: Path) -> list[str]:
    try:
        raw = path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        return []

    errors: list[str] = []
    is_markdown = path.suffix.lower() == ".md"
    for line_number, line in enumerate(raw.splitlines(), start=1):
        if is_markdown and not markdown_line_is_deployable_example(line):
            continue
        for token in FORBIDDEN_TOKENS:
            if token in line:
                errors.append(
                    f"{repo_path(path)}:{line_number} contains forbidden deployable inference fallback token {token}"
                )
    return errors


def validate_surfaces(surfaces: list[PolicySurface]) -> list[str]:
    errors: list[str] = []
    for file_path in iter_surface_files(surfaces):
        errors.extend(find_violations(file_path))
    return errors


def split_rendered_yaml(rendered_yaml: str) -> list[RenderedResource]:
    resources: list[RenderedResource] = []
    for document in re.split(r"^---\s*$", rendered_yaml, flags=re.MULTILINE):
        text = document.strip()
        if not text:
            continue
        kind = re.search(r"^kind:\s*([^\s#]+)", text, flags=re.MULTILINE)
        metadata = re.search(r"^metadata:\s*$([\s\S]*?)(?=^(?:spec|data|rules|subjects|roleRef):|\Z)", text, flags=re.MULTILINE)
        name = None
        if metadata:
            name = re.search(r"^\s{2}name:\s*\"?([^\"\n]+)\"?", metadata.group(1), flags=re.MULTILINE)
        resources.append(
            RenderedResource(
                kind=kind.group(1) if kind else "",
                name=name.group(1).strip() if name else "",
                text=text,
            )
        )
    return resources


def find_resource(resources: list[RenderedResource], kind: str, name: str) -> RenderedResource | None:
    return next((resource for resource in resources if resource.kind == kind and resource.name == name), None)


def rendered_replicas(resource: RenderedResource) -> int | None:
    match = re.search(r"^\s{2}replicas:\s*(\d+)", resource.text, flags=re.MULTILINE)
    return int(match.group(1)) if match else None


def resource_secret_keys(resource: RenderedResource) -> set[str]:
    return set(re.findall(r"^\s*(?:-\s*)?secretKey:\s*([^\s#]+)", resource.text, flags=re.MULTILINE))


def contains_64_hex_annotation(resource: RenderedResource, annotation: str) -> bool:
    return (
        re.search(
            rf"{re.escape(annotation)}:\s*\"?[a-f0-9]{{64}}\"?",
            resource.text,
            flags=re.IGNORECASE,
        )
        is not None
    )


def normalize_render(rendered_yaml: str) -> str:
    return "\n".join(line.rstrip() for line in rendered_yaml.replace("\r\n", "\n").strip().splitlines())


def compare_rendered_yaml(canonical_yaml: str, infra_yaml: str) -> list[str]:
    if normalize_render(canonical_yaml) == normalize_render(infra_yaml):
        return []
    return ["infra compatibility render differs from canonical deploy/kubernetes render"]


def validate_environment_inference_secret_refs(rendered_yaml: str, environment: str) -> list[str]:
    errors: list[str] = []
    expected_prefix = f"secret/virtengine/{environment}/inference-sidecar/"
    refs = re.findall(
        r"^\s*key:\s*(secret/virtengine/([^/\s]+)/inference-sidecar/[^\s#]+)",
        rendered_yaml,
        flags=re.MULTILINE,
    )
    for ref, actual_environment in refs:
        if not ref.startswith(expected_prefix):
            errors.append(
                f"{environment} inference ExternalSecret references {actual_environment} environment path {ref}"
            )
    return errors


def validate_rendered_yaml(rendered_yaml: str) -> list[str]:
    errors: list[str] = []
    resources = split_rendered_yaml(rendered_yaml)
    rendered_lower = rendered_yaml.lower()
    allowed_inference_external_secret_keys = {
        "inference-sidecar-server-mtls": {"tls.crt", "tls.key"},
        "inference-sidecar-client-ca": {"ca.crt"},
        "inference-sidecar-client-mtls": {"tls.crt", "tls.key", "ca.crt"},
        "inference-sidecar-model-metadata": {
            "model-version",
            "model-digest",
            "runtime-digest",
            "manifest-digest",
        },
    }

    seen_resources: set[tuple[str, str]] = set()
    for resource in resources:
        if not resource.kind or not resource.name:
            continue
        identity = (resource.kind, resource.name)
        if identity in seen_resources:
            errors.append(f"duplicate {resource.kind}/{resource.name} in rendered deployment")
        seen_resources.add(identity)

    for token in FORBIDDEN_TOKENS:
        if token.lower() in rendered_lower:
            errors.append(f"deployable runtime contains forbidden fallback token {token}")
    for token in RAW_KEY_PATTERNS:
        if token in rendered_lower:
            errors.append(f"deployable runtime references raw signing/consensus key material: {token}")

    sidecar = find_resource(resources, "StatefulSet", "inference-sidecar")
    if sidecar is None:
        errors.append("missing StatefulSet/inference-sidecar")
    else:
        if rendered_replicas(sidecar) != 1:
            errors.append("StatefulSet/inference-sidecar must have exactly one replica")
        if "serviceAccountName: inference-sidecar" not in sidecar.text:
            errors.append("StatefulSet/inference-sidecar must use ServiceAccount/inference-sidecar")
        if 'reloader.stakater.com/auto: "true"' not in sidecar.text:
            errors.append("StatefulSet/inference-sidecar must enable reloader.stakater.com/auto for secret rotation")
        image_match = re.search(r"^\s*image:\s*([^\s]+)", sidecar.text, flags=re.MULTILINE)
        if not image_match or not re.search(r"@sha256:[a-f0-9]{64}$", image_match.group(1), flags=re.IGNORECASE):
            errors.append("StatefulSet/inference-sidecar image must use an immutable digest")
        for required in (
            "--require-mtls=true",
            "--tls-cert-file=/",
            "--tls-key-file=/",
            "--tls-client-ca-file=/",
            "--force-cpu=true",
            "--random-seed=42",
            "--model-path=/",
            "--model-version=",
            "--expected-hash=",
            "--manifest-path=/",
        ):
            if required not in sidecar.text:
                errors.append(f"StatefulSet/inference-sidecar missing required argument {required}")
        if "--require-mtls=false" in sidecar.text:
            errors.append("StatefulSet/inference-sidecar must require --require-mtls=true")
        if not re.search(r"--serving-url=\S+", sidecar.text):
            errors.append("StatefulSet/inference-sidecar missing required argument --serving-url=")
        if "--metrics-addr=:9092" not in sidecar.text:
            errors.append("StatefulSet/inference-sidecar metrics/readiness must use port 9092")
        if "containerPort: 50051" not in sidecar.text:
            errors.append("StatefulSet/inference-sidecar must expose gRPC containerPort 50051")
        if "containerPort: 9092" not in sidecar.text:
            errors.append("StatefulSet/inference-sidecar must expose metrics containerPort 9092")
        for security in (
            "runAsNonRoot: true",
            "readOnlyRootFilesystem: true",
            "allowPrivilegeEscalation: false",
            "seccompProfile:",
            "drop:",
            "- ALL",
            "emptyDir:",
            "sizeLimit:",
            "podAntiAffinity:",
            "topologySpreadConstraints:",
        ):
            if security not in sidecar.text:
                readable = security.rstrip(":")
                if security == "readOnlyRootFilesystem: true":
                    readable = "read-only root filesystem"
                errors.append(f"StatefulSet/inference-sidecar missing strict security control {readable}")
        for probe in ("readinessProbe:", "livenessProbe:", "startupProbe:"):
            if probe not in sidecar.text or "path: /health" not in sidecar.text:
                errors.append(f"StatefulSet/inference-sidecar {probe.rstrip(':')} must use /health")
        for mount_match in re.finditer(r"mountPath:\s*([^\s]+)", sidecar.text):
            mount_path = mount_match.group(1)
            if "inference-sidecar" in mount_path and not mount_path.startswith("/"):
                errors.append(f"StatefulSet/inference-sidecar secret mount must be absolute: {mount_path}")
        for secret_name in ("inference-sidecar-server-mtls", "inference-sidecar-client-ca"):
            if secret_name not in sidecar.text:
                errors.append(f"StatefulSet/inference-sidecar missing strict Secret mount {secret_name}")
        if "claimName: inference-sidecar-model-bundle" not in sidecar.text or "readOnly: true" not in sidecar.text:
            errors.append("StatefulSet/inference-sidecar must mount the model bundle read-only")
        if (
            "secretName: inference-sidecar-model-metadata" not in sidecar.text
            or "mountPath: /var/run/secrets/virtengine/inference-sidecar/model-metadata" not in sidecar.text
            or not re.search(
                r"(?:name:\s*model-metadata[\s\S]*?mountPath:\s*/var/run/secrets/virtengine/inference-sidecar/model-metadata|mountPath:\s*/var/run/secrets/virtengine/inference-sidecar/model-metadata[\s\S]*?name:\s*model-metadata)[\s\S]*?readOnly:\s*true",
                sidecar.text,
            )
        ):
            errors.append("StatefulSet/inference-sidecar must mount model metadata ExternalSecret read-only")
        if not re.search(r"--expected-hash=[a-f0-9]{64}", sidecar.text, flags=re.IGNORECASE):
            errors.append("StatefulSet/inference-sidecar must pin model hash as a 64 hex digest")
        for annotation in ("virtengine.com/model-digest", "virtengine.com/runtime-digest", "virtengine.com/manifest-digest"):
            if not contains_64_hex_annotation(sidecar, annotation):
                errors.append(f"StatefulSet/inference-sidecar missing 64-hex commitment {annotation}")

    service = find_resource(resources, "Service", "inference-sidecar")
    if service is None:
        errors.append("missing Service/inference-sidecar")
    else:
        if re.search(r"^\s{2}type:\s*(LoadBalancer|NodePort)", service.text, flags=re.MULTILINE):
            errors.append("Service/inference-sidecar must not be a public Service")
        if "port: 50051" not in service.text:
            errors.append("Service/inference-sidecar must expose port 50051")
        if "port: 9092" not in service.text:
            errors.append("Service/inference-sidecar must expose metrics/readiness port 9092")

    service_account = find_resource(resources, "ServiceAccount", "inference-sidecar")
    if service_account is None:
        errors.append("missing ServiceAccount/inference-sidecar")

    pdb = find_resource(resources, "PodDisruptionBudget", "inference-sidecar-pdb")
    if pdb is None:
        errors.append("missing PodDisruptionBudget/inference-sidecar-pdb")
    else:
        if not re.search(r"^\s{2}minAvailable:\s*1", pdb.text, flags=re.MULTILINE):
            errors.append("PodDisruptionBudget/inference-sidecar-pdb must set minAvailable=1")
        if "selector:" not in pdb.text or "app.kubernetes.io/name: inference-sidecar" not in pdb.text:
            errors.append("PodDisruptionBudget/inference-sidecar-pdb must select inference-sidecar pods")

    pvc = find_resource(resources, "PersistentVolumeClaim", "inference-sidecar-model-bundle")
    if pvc is None:
        errors.append("missing model bundle PersistentVolumeClaim/inference-sidecar-model-bundle")
    else:
        for annotation in ("virtengine.com/model-digest", "virtengine.com/runtime-digest", "virtengine.com/manifest-digest"):
            if not contains_64_hex_annotation(pvc, annotation):
                errors.append(f"model bundle PVC missing 64-hex commitment {annotation}")

    network_policy = find_resource(resources, "NetworkPolicy", "inference-sidecar")
    if network_policy is None:
        errors.append("missing NetworkPolicy/inference-sidecar")
    else:
        for required in (
            "app.kubernetes.io/name: virtengine-validator",
            "port: 50051",
            "kubernetes.io/metadata.name: monitoring",
            "port: 9092",
            "kubernetes.io/metadata.name: kube-system",
            "port: 53",
        ):
            if required not in network_policy.text:
                errors.append(f"NetworkPolicy/inference-sidecar missing {required}")
        if "tf-serving" in rendered_yaml and "port: 8501" not in network_policy.text:
            errors.append("NetworkPolicy/inference-sidecar must include TF Serving egress port 8501 when serving-url is configured")
        for broad in ("from: []", "to: []", "namespaceSelector: {}", "podSelector: {}"):
            if broad in network_policy.text:
                errors.append(f"NetworkPolicy/inference-sidecar must not contain broad rule {broad}")

    validator = find_resource(resources, "StatefulSet", "virtengine-validator")
    if validator is None:
        errors.append("missing StatefulSet/virtengine-validator")
    else:
        for required in (
            "VEID_INFERENCE_SIDECAR_ADDR",
            "VEID_INFERENCE_SIDECAR_TLS_CERT_FILE",
            "VEID_INFERENCE_SIDECAR_TLS_KEY_FILE",
            "VEID_INFERENCE_SIDECAR_TLS_CA_FILE",
            "inference-sidecar-client-mtls",
            "mountPath: /var/run/secrets/virtengine/inference-sidecar-client",
            "readOnly: true",
        ):
            if required not in validator.text:
                errors.append(f"StatefulSet/virtengine-validator missing inference transport metadata {required}")
        if not re.search(
            r"name:\s*VEID_INFERENCE_SIDECAR_TLS\s*\n\s*value:\s*[\"']?true[\"']?\s*$",
            validator.text,
            flags=re.MULTILINE,
        ):
            errors.append("StatefulSet/virtengine-validator must set VEID_INFERENCE_SIDECAR_TLS=true")

    validator_policy = find_resource(resources, "NetworkPolicy", "virtengine-validator")
    if validator_policy is None:
        errors.append("missing NetworkPolicy/virtengine-validator")
    elif "app.kubernetes.io/name: inference-sidecar" not in validator_policy.text or "port: 50051" not in validator_policy.text:
        errors.append("NetworkPolicy/virtengine-validator must allow sidecar egress on TCP 50051")

    for name in allowed_inference_external_secret_keys:
        if find_resource(resources, "ExternalSecret", name) is None:
            errors.append(f"missing ExternalSecret/{name}")

    for resource in resources:
        if resource.kind == "HorizontalPodAutoscaler" and "inference-sidecar" in resource.text:
            errors.append("inference-sidecar must not have an HPA")
        if resource.kind in ("Ingress", "HTTPRoute", "Gateway") and "inference-sidecar" in resource.text:
            errors.append(f"{resource.kind}/{resource.name} must not expose inference-sidecar")
        if resource.kind == "Service" and resource.name == "inference-sidecar" and "LoadBalancer" in resource.text:
            errors.append("Service/inference-sidecar must not be public")
        if resource.kind == "ExternalSecret" and resource.name.startswith("inference-sidecar"):
            expected_keys = allowed_inference_external_secret_keys.get(resource.name)
            if expected_keys is None:
                errors.append(f"ExternalSecret/{resource.name} is not an allowlisted inference ExternalSecret")
            else:
                actual_keys = resource_secret_keys(resource)
                if actual_keys != expected_keys:
                    errors.append(
                        f"ExternalSecret/{resource.name} must contain exactly {sorted(expected_keys)}, got {sorted(actual_keys)}"
                    )
            for forbidden in RAW_KEY_PATTERNS:
                if forbidden in resource.text.lower():
                    errors.append(f"ExternalSecret/{resource.name} references raw signing/consensus key material")

    return errors


def render_kustomize(path: str) -> tuple[str, list[str]]:
    result = subprocess.run(
        ["kubectl", "kustomize", "--load-restrictor=LoadRestrictionsNone", path],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        return "", [f"kubectl kustomize {path} failed: {result.stderr or result.stdout}".strip()]
    if result.stderr.strip():
        return result.stdout, [f"kubectl kustomize {path} emitted stderr: {result.stderr.strip()}"]
    return result.stdout, []


def validate_repo_renders() -> list[str]:
    errors: list[str] = []
    rendered: dict[str, str] = {}
    for canonical, infra in RENDER_PAIRS:
        for root in (canonical, infra):
            output, render_errors = render_kustomize(root)
            errors.extend(render_errors)
            rendered[root] = output
    if errors:
        return errors

    for canonical, infra in RENDER_PAIRS:
        errors.extend(compare_rendered_yaml(rendered[canonical], rendered[infra]))
        errors.extend([f"{canonical}: {error}" for error in validate_rendered_yaml(rendered[canonical])])
        errors.extend([f"{infra}: {error}" for error in validate_rendered_yaml(rendered[infra])])
        environment = Path(canonical).name
        if environment in ("staging", "prod"):
            errors.extend(
                [
                    f"{canonical}: {error}"
                    for error in validate_environment_inference_secret_refs(
                        rendered[canonical], environment
                    )
                ]
            )
            errors.extend(
                [
                    f"{infra}: {error}"
                    for error in validate_environment_inference_secret_refs(
                        rendered[infra], environment
                    )
                ]
            )
    return errors


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Validate production inference sidecar deployment policy."
    )
    parser.add_argument(
        "--paths",
        nargs="*",
        help="Optional explicit files or directories to scan for deployable fallback tokens.",
    )
    parser.add_argument(
        "--skip-render",
        action="store_true",
        help="Skip kubectl kustomize render checks; intended only for narrow fixture scans.",
    )
    return parser.parse_args()


def resolve_surfaces(args: argparse.Namespace) -> list[PolicySurface]:
    if not args.paths:
        return list(DEFAULT_SURFACES)

    surfaces: list[PolicySurface] = []
    for item in args.paths:
        path = Path(item)
        resolved = path if path.is_absolute() else ROOT / path
        surfaces.append(PolicySurface(resolved, recursive=resolved.is_dir()))
    return surfaces


def main() -> int:
    args = parse_args()
    errors = validate_surfaces(resolve_surfaces(args))
    if not args.paths and not args.skip_render:
        errors.extend(validate_repo_renders())

    if errors:
        for error in errors:
            print(f"FAIL {error}")
        return 1

    print("PASS inference deployment policy")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

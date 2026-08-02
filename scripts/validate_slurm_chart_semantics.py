#!/usr/bin/env python3

"""Validate security, capacity, and durability semantics of a rendered SLURM chart."""

from __future__ import annotations

import argparse
import json
import re
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Iterable

import yaml


INVARIANTS = (
    "stable-secrets",
    "replica-capacity-equality",
    "immutable-images",
    "least-privilege",
    "durable-state",
)
SOURCE_VERIFIABLE_INVARIANTS = {"stable-secrets"}
WORKLOAD_KINDS = {"CronJob", "DaemonSet", "Deployment", "Job", "Pod", "ReplicaSet", "ReplicationController", "StatefulSet"}
DURABLE_COMPONENTS = {"controller", "database", "mariadb"}
DURABLE_PATHS = {
    "controller": ("/var/spool/slurm",),
    "database": ("/var/spool/slurm",),
    "mariadb": ("/var/lib/mysql",),
}
DIGEST_IMAGE = re.compile(
    r"^[a-z0-9]+(?:[._-][a-z0-9]+)*(?::[0-9]+)?"
    r"(?:/[a-z0-9]+(?:[._-][a-z0-9]+)*)*@sha256:[a-f0-9]{64}$"
)
HELM_RANDOM = re.compile(
    r"{{[^}]*\b(?:randAlpha|randAlphaNum|randAscii|randBytes|randNumeric|genCA|genPrivateKey|"
    r"genSelfSignedCert|genSignedCert|derivePassword|htpasswd|uuidv4)\b",
    re.IGNORECASE,
)
PRIVILEGED_SOURCE = re.compile(r"^\s*(?:privileged:\s*true|runAsUser:\s*0|runAsNonRoot:\s*false|allowPrivilegeEscalation:\s*true)\b", re.MULTILINE)
NODE_LINE = re.compile(r"^\s*NodeName=(\S+)", re.MULTILINE)
PARTITION_LINE = re.compile(r"^\s*PartitionName=(\S+)(.*)$", re.MULTILINE)


@dataclass(frozen=True)
class Finding:
    invariant: str
    location: str
    message: str

    def as_dict(self) -> dict[str, str]:
        return {"invariant": self.invariant, "location": self.location, "message": self.message}


def _documents(value: Any) -> list[dict[str, Any]]:
    if not isinstance(value, dict):
        return []
    if value.get("kind") == "List" and isinstance(value.get("items"), list):
        documents: list[dict[str, Any]] = []
        for item in value["items"]:
            documents.extend(_documents(item))
        return documents
    return [value]


def load_rendered(path: Path) -> list[dict[str, Any]]:
    text = sys.stdin.read() if str(path) == "-" else path.read_text(encoding="utf-8")
    documents: list[dict[str, Any]] = []
    for value in yaml.safe_load_all(text):
        documents.extend(_documents(value))
    if not documents:
        raise ValueError("rendered input contains zero Kubernetes documents")
    return documents


def _metadata(document: dict[str, Any]) -> tuple[str, str, dict[str, Any]]:
    metadata = document.get("metadata") if isinstance(document.get("metadata"), dict) else {}
    return str(document.get("kind", "<unknown>")), str(metadata.get("name", "<unnamed>")), metadata


def _pod_spec(document: dict[str, Any]) -> dict[str, Any] | None:
    kind = document.get("kind")
    if kind == "Pod":
        return document.get("spec")
    spec = document.get("spec", {})
    if kind == "CronJob":
        return spec.get("jobTemplate", {}).get("spec", {}).get("template", {}).get("spec")
    if kind == "Job":
        return spec.get("template", {}).get("spec")
    if kind in WORKLOAD_KINDS:
        return spec.get("template", {}).get("spec")
    return None


def _contains_pod_structure(value: Any) -> bool:
    if isinstance(value, dict):
        if any(field in value for field in ("containers", "initContainers", "ephemeralContainers")):
            return True
        return any(_contains_pod_structure(child) for child in value.values())
    if isinstance(value, list):
        return any(_contains_pod_structure(child) for child in value)
    return False


def _containers(pod_spec: dict[str, Any]) -> Iterable[tuple[str, dict[str, Any]]]:
    for field in ("initContainers", "containers", "ephemeralContainers"):
        for container in pod_spec.get(field, []) or []:
            if isinstance(container, dict):
                yield field, container


def _component(document: dict[str, Any]) -> str:
    _, name, metadata = _metadata(document)
    labels = metadata.get("labels") if isinstance(metadata.get("labels"), dict) else {}
    component = str(labels.get("app.kubernetes.io/component", "")).lower()
    if component:
        return component
    lowered_name = name.lower()
    for candidate in (*DURABLE_COMPONENTS, "compute"):
        if candidate in lowered_name:
            return candidate
    return ""


def _primary_containers(pod_spec: dict[str, Any], component: str) -> list[dict[str, Any]]:
    containers = [item for item in pod_spec.get("containers", []) or [] if isinstance(item, dict)]
    aliases = {
        "controller": ("controller", "slurmctld"),
        "database": ("database", "slurmdbd"),
        "mariadb": ("mariadb", "mysql"),
    }.get(component, (component,))
    matching = [container for container in containers if any(alias in str(container.get("name", "")).lower() for alias in aliases)]
    return matching or containers[:1]


def _expand_hostlist(expression: str) -> set[str]:
    items = re.split(r",(?![^\[]*\])", expression)
    nodes: set[str] = set()
    for item in items:
        match = re.fullmatch(r"([^\[]*)\[([^]]+)\]", item)
        if not match:
            nodes.add(item)
            continue
        prefix, ranges = match.groups()
        for range_item in ranges.split(","):
            range_match = re.fullmatch(r"(\d+)-(\d+)(?::(\d+))?", range_item)
            if not range_match:
                nodes.add(f"{prefix}{range_item}")
                continue
            first, last, step = range_match.groups()
            width = max(len(first), len(last))
            nodes.update(f"{prefix}{index:0{width}d}" for index in range(int(first), int(last) + 1, int(step or 1)))
    return nodes


def _slurm_config(documents: list[dict[str, Any]]) -> str:
    chunks = []
    for document in documents:
        if document.get("kind") != "ConfigMap":
            continue
        data = document.get("data") if isinstance(document.get("data"), dict) else {}
        if isinstance(data.get("slurm.conf"), str):
            chunks.append(data["slurm.conf"])
    return "\n".join(chunks)


def _static_findings(chart_dir: Path) -> list[Finding]:
    findings: list[Finding] = []
    values_path = chart_dir / "values.yaml"
    if not values_path.is_file():
        return [Finding("stable-secrets", str(values_path), "canonical values.yaml is missing")]

    values = yaml.safe_load(values_path.read_text(encoding="utf-8")) or {}
    for path in chart_dir.rglob("*.yaml"):
        relative = path.relative_to(chart_dir).as_posix()
        content = path.read_text(encoding="utf-8")
        if HELM_RANDOM.search(content):
            findings.append(Finding("stable-secrets", relative, "render-time random secret generation is forbidden"))
        if path.name != "values.yaml" and PRIVILEGED_SOURCE.search(content):
            findings.append(Finding("least-privilege", relative, "template contains privileged or root container defaults"))

    template_source = "\n".join(
        path.read_text(encoding="utf-8")
        for path in (chart_dir / "templates").rglob("*")
        if path.is_file()
    )
    secret_contracts = (
        ("munge", values.get("munge", {}), ("key",), "munge.existingSecret", ("munge.secretKeyName",)),
        ("database", values.get("database", {}).get("config", {}), ("password",), "database.config.existingSecret", ("database.config.secretPasswordKey",)),
        ("mariadb", values.get("mariadb", {}), ("rootPassword",), "mariadb.existingSecret", ("mariadb.secretRootPasswordKey",)),
        (
            "nodeAgent TLS",
            values.get("nodeAgent", {}).get("tls", {}),
            ("caCert", "clientCert", "clientKey"),
            "nodeAgent.tls.existingSecret",
            ("nodeAgent.tls.caCertKey", "nodeAgent.tls.clientCertKey", "nodeAgent.tls.clientKeyKey"),
        ),
    )
    for name, settings, inline_keys, secret_path, key_paths in secret_contracts:
        required_secret = re.search(rf"\brequired\b[^}}]*\.Values\.{re.escape(secret_path)}\b", template_source)
        if not isinstance(settings, dict) or (not settings.get("existingSecret") and not required_secret):
            findings.append(Finding("stable-secrets", "values.yaml", f"{name} has no fail-closed pre-provisioned secret reference"))
        for key_path in key_paths:
            required_key = re.search(rf"\brequired\b[^}}]*\.Values\.{re.escape(key_path)}\b", template_source)
            if not isinstance(settings, dict) or not required_key:
                findings.append(Finding("stable-secrets", "values.yaml", f"{name} has no required {key_path.rsplit('.', 1)[-1]} reference"))
        if isinstance(settings, dict) and any(settings.get(inline_key) for inline_key in inline_keys):
            findings.append(Finding("stable-secrets", "values.yaml", f"{name} contains a forbidden inline secret value"))

    tls_projection_contract = (
        ("path: ca.crt", 'ca_cert: "/etc/virtengine/virtengine-agent/tls/ca.crt"'),
        ("path: client.crt", 'client_cert: "/etc/virtengine/virtengine-agent/tls/client.crt"'),
        ("path: client.key", 'client_key: "/etc/virtengine/virtengine-agent/tls/client.key"'),
    )
    for projected_path, configured_path in tls_projection_contract:
        if projected_path not in template_source or configured_path not in template_source:
            findings.append(Finding("stable-secrets", "templates", f"nodeAgent TLS projection mismatch for {projected_path.removeprefix('path: ')}"))

    capacity_contract = {
        "templates/_helpers.tpl": (
            'define "slurm-cluster.dnsName"',
            'define "slurm-cluster.resourceName"',
            "sha256sum $raw | trunc 8",
            "ordinalBudget",
            'define "slurm-cluster.compute.capacity"',
            'define "slurm-cluster.partition.capacity"',
            "compute.replicas must be at least 1 when compute.enabled is true",
            "replicas must be at least 1 when enabled",
            "is reserved by an existing chart resource or component",
            "is duplicated or conflicts with the default compute pool",
            "collides at rendered StatefulSet name",
            "has duplicate node pool selector",
            "selects unknown node pool",
            "selects disabled node pool",
            "selects zero compute nodes",
            'fail "at least one compute replica must be enabled"',
            'dict "replicas" $total "nodes" (join "," $nodes) "pools" $pools',
        ),
        "templates/configmap.yaml": (
            'include "slurm-cluster.compute.capacity" . | fromJson',
            'include "slurm-cluster.partition.capacity"',
            'NodeName={{ include "slurm-cluster.compute.serviceName"',
            'NodeName={{ include "slurm-cluster.nodePool.serviceName"',
            "Nodes={{ $partitionCapacity.nodes }}",
            "MaxNodes={{ $partitionCapacity.replicas }}",
        ),
        "templates/compute-nodepools-statefulset.yaml": (
            'include "slurm-cluster.compute.capacity" .',
            'include "slurm-cluster.nodePool.enabled"',
        ),
    }
    for relative, required_fragments in capacity_contract.items():
        path = chart_dir / relative
        content = path.read_text(encoding="utf-8") if path.is_file() else ""
        for fragment in required_fragments:
            if fragment not in content:
                findings.append(Finding("replica-capacity-equality", relative, f"derived capacity contract is missing {fragment}"))

    schema_path = chart_dir / "values.schema.json"
    schema = json.loads(schema_path.read_text(encoding="utf-8")) if schema_path.is_file() else {}
    properties = schema.get("properties", {}) if isinstance(schema, dict) else {}
    pool_schema = properties.get("nodePools", {}).get("items", {}) if isinstance(properties.get("nodePools"), dict) else {}
    partition_schema = properties.get("partitions", {}).get("items", {}) if isinstance(properties.get("partitions"), dict) else {}
    if pool_schema.get("required") != ["name", "replicas"] or not pool_schema.get("allOf"):
        findings.append(Finding("replica-capacity-equality", "values.schema.json", "enabled node pools do not require names, replicas, and conditional capacity validation"))
    forbidden_capacity = partition_schema.get("not", {}).get("anyOf", []) if isinstance(partition_schema, dict) else []
    if forbidden_capacity != [{"required": ["nodes"]}, {"required": ["maxNodes"]}]:
        findings.append(Finding("replica-capacity-equality", "values.schema.json", "partition nodes and maxNodes overrides are not forbidden"))
    reserved_names = pool_schema.get("properties", {}).get("name", {}).get("not", {}).get("enum", [])
    required_reserved = {"controller", "slurmdbd", "database", "mariadb", "compute", "munge", "node-agent"}
    if not required_reserved.issubset(set(reserved_names)):
        findings.append(Finding("replica-capacity-equality", "values.schema.json", "node pool names do not reserve existing chart resources and components"))
    selector_schema = partition_schema.get("properties", {}).get("nodePools", {})
    if selector_schema.get("minItems") != 1 or selector_schema.get("uniqueItems") is not True or not partition_schema.get("allOf"):
        findings.append(Finding("replica-capacity-equality", "values.schema.json", "partition nodePools selectors are not nonempty, unique, and required outside the default partition"))

    for partition in values.get("partitions", []) or []:
        if not isinstance(partition, dict):
            continue
        for forbidden in ("nodes", "maxNodes"):
            if forbidden in partition:
                findings.append(Finding("replica-capacity-equality", "values.yaml", f"partition {partition.get('name', '<unnamed>')} overrides derived {forbidden}"))

    for component in ("munge", "controller", "database", "mariadb", "compute", "nodeAgent"):
        settings = values.get(component, {})
        image = settings.get("image", {}) if isinstance(settings, dict) else {}
        if not isinstance(image, dict) or not image:
            continue
        repository = str(image.get("repository", ""))
        tag = str(image.get("tag", ""))
        if not DIGEST_IMAGE.search(repository) and not DIGEST_IMAGE.search(tag):
            findings.append(Finding("immutable-images", "values.yaml", f"{component} image is tag-based rather than digest-pinned"))

    for component in DURABLE_COMPONENTS:
        settings = values.get(component, {})
        persistence = settings.get("persistence", {}) if isinstance(settings, dict) else {}
        if settings.get("enabled", True) and not persistence.get("enabled", False):
            findings.append(Finding("durable-state", "values.yaml", f"{component} persistence is not enabled"))
    return findings


def validate_source(chart_dir: Path) -> list[Finding]:
    return _static_findings(chart_dir)


def validate(documents: list[dict[str, Any]], chart_dir: Path | None = None) -> list[Finding]:
    documents = [document for value in documents for document in _documents(value)]
    findings = _static_findings(chart_dir) if chart_dir else []
    compute_replicas = 0
    stateful_compute_nodes: set[str] = set()
    durable_seen: set[str] = set()

    for document in documents:
        kind, name, metadata = _metadata(document)
        location = f"{kind}/{name}"
        component = _component(document)
        annotations = metadata.get("annotations") if isinstance(metadata.get("annotations"), dict) else {}
        if kind == "Secret":
            has_inline_data = bool(document.get("data")) or bool(document.get("stringData"))
            external_reference = annotations.get("virtengine.com/external-secret-ref")
            if has_inline_data or annotations.get("virtengine.com/external-secret") != "true" or not external_reference:
                findings.append(Finding("stable-secrets", location, "Secret must be empty and declare an explicit external-secret-ref contract"))

        pod_spec = _pod_spec(document)
        if pod_spec is None and kind not in WORKLOAD_KINDS and _contains_pod_structure(document):
            findings.append(Finding("least-privilege", location, "unsupported document contains an unvalidated pod or container structure"))
        if pod_spec is not None:
            pod_security = pod_spec.get("securityContext") if isinstance(pod_spec.get("securityContext"), dict) else {}
            for field, container in _containers(pod_spec):
                container_location = f"{location}:{field}/{container.get('name', '<unnamed>')}"
                image = str(container.get("image", ""))
                if not DIGEST_IMAGE.search(image):
                    findings.append(Finding("immutable-images", container_location, f"image is not pinned by sha256 digest: {image or '<missing>'}"))
                security = container.get("securityContext") if isinstance(container.get("securityContext"), dict) else {}
                capabilities = security.get("capabilities") if isinstance(security.get("capabilities"), dict) else {}
                added = {str(item).upper() for item in capabilities.get("add", []) or []}
                dropped = {str(item).upper() for item in capabilities.get("drop", []) or []}
                run_as_non_root = security.get("runAsNonRoot", pod_security.get("runAsNonRoot"))
                run_as_user = security.get("runAsUser", pod_security.get("runAsUser"))
                seccomp = security.get("seccompProfile", pod_security.get("seccompProfile", {}))
                seccomp_type = seccomp.get("type") if isinstance(seccomp, dict) else None
                violations = []
                if security.get("privileged") is not False:
                    violations.append("privileged must be false")
                if run_as_non_root is not True or run_as_user == 0:
                    violations.append("container must run as non-root")
                if security.get("allowPrivilegeEscalation") is not False:
                    violations.append("allowPrivilegeEscalation must be false")
                if "ALL" not in dropped:
                    violations.append("capabilities.drop must include ALL")
                if added:
                    violations.append(f"capabilities.add must be empty: {', '.join(sorted(added))}")
                if seccomp_type not in {"RuntimeDefault", "Localhost"}:
                    violations.append("seccompProfile.type must be RuntimeDefault or Localhost")
                if violations:
                    findings.append(Finding("least-privilege", container_location, "; ".join(violations)))

        if kind in {"Deployment", "ReplicaSet", "ReplicationController", "StatefulSet"} and component == "compute":
            replicas = int(document.get("spec", {}).get("replicas", 1))
            compute_replicas += replicas
            if kind == "StatefulSet":
                stateful_compute_nodes.update(f"{name}-{ordinal}" for ordinal in range(replicas))
        if pod_spec is not None and component in DURABLE_COMPONENTS:
            durable_seen.add(component)
            spec = document.get("spec", {})
            claims = {claim.get("metadata", {}).get("name") for claim in spec.get("volumeClaimTemplates", []) or [] if isinstance(claim, dict)}
            volumes = {volume.get("name"): volume for volume in pod_spec.get("volumes", []) or [] if isinstance(volume, dict)}
            for container in _primary_containers(pod_spec, component):
                mounted_paths = {mount.get("mountPath"): mount.get("name") for mount in container.get("volumeMounts", []) or [] if isinstance(mount, dict)}
                for required_path in DURABLE_PATHS[component]:
                    volume_name = mounted_paths.get(required_path)
                    volume = volumes.get(volume_name, {})
                    if volume_name not in claims and not isinstance(volume.get("persistentVolumeClaim"), dict):
                        findings.append(Finding("durable-state", f"{location}:containers/{container.get('name', '<unnamed>')}", f"{required_path} is not backed by a persistent volume claim"))

    config = _slurm_config(documents)
    declared_nodes = set().union(*(_expand_hostlist(value) for value in NODE_LINE.findall(config))) if NODE_LINE.search(config) else set()
    partitions = []
    for partition_name, settings in PARTITION_LINE.findall(config):
        fields = dict(re.findall(r"\b(\w+)=(\S+)", settings))
        nodes = _expand_hostlist(fields["Nodes"]) if "Nodes" in fields else set()
        max_nodes = int(fields["MaxNodes"]) if fields.get("MaxNodes", "").isdigit() else None
        partitions.append((partition_name, nodes, max_nodes))
    if not config:
        findings.append(Finding("replica-capacity-equality", "render", "no ConfigMap contains data.slurm.conf"))
    elif len(declared_nodes) != compute_replicas:
        findings.append(Finding("replica-capacity-equality", "ConfigMap/slurm.conf", f"render has {compute_replicas} compute replicas and {len(declared_nodes)} declared nodes"))
    elif stateful_compute_nodes and stateful_compute_nodes != declared_nodes:
        findings.append(Finding("replica-capacity-equality", "ConfigMap/slurm.conf", "declared NodeName values do not exactly match rendered StatefulSet pod names"))
    covered_nodes: set[str] = set()
    for partition_name, partition_nodes, max_nodes in partitions:
        covered_nodes.update(partition_nodes)
        if not partition_nodes or not partition_nodes.issubset(declared_nodes) or max_nodes != len(partition_nodes):
            findings.append(Finding("replica-capacity-equality", "ConfigMap/slurm.conf", f"partition {partition_name} does not exactly advertise its selected rendered compute replicas"))
    if config and (not partitions or covered_nodes != declared_nodes):
        findings.append(Finding("replica-capacity-equality", "ConfigMap/slurm.conf", "partition membership does not cover every rendered compute replica"))

    for missing in sorted(DURABLE_COMPONENTS - durable_seen):
        findings.append(Finding("durable-state", "render", f"required durable component {missing} has no StatefulSet"))
    return findings


def result(findings: list[Finding], diagnostic: bool) -> dict[str, Any]:
    failed = {finding.invariant for finding in findings}
    statuses = {}
    for invariant in INVARIANTS:
        if invariant in failed:
            statuses[invariant] = "failed"
        elif not diagnostic or invariant in SOURCE_VERIFIABLE_INVARIANTS:
            statuses[invariant] = "passed"
        else:
            statuses[invariant] = "unverified"
    return {
        "schema_version": "virtengine.slurm-semantic-validation/v1",
        "mode": "diagnostic" if diagnostic else "enforcing",
        "passed": not findings and not diagnostic,
        "invariants": statuses,
        "findings": [finding.as_dict() for finding in findings],
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--rendered", type=Path, help="rendered multi-document Kubernetes YAML, or - for stdin")
    parser.add_argument("--chart", type=Path, help="canonical chart directory for static source and values checks")
    parser.add_argument("--diagnostic", action="store_true", help="report source blockers without allowing success")
    parser.add_argument("--json", action="store_true", help="emit machine-readable JSON")
    args = parser.parse_args()
    try:
        if args.rendered:
            documents = load_rendered(args.rendered)
        elif args.diagnostic and args.chart:
            documents = []
        else:
            parser.error("--rendered is required outside source-only diagnostic mode")
        findings = validate_source(args.chart) if args.diagnostic and not args.rendered else validate(documents, args.chart)
        report = result(findings, args.diagnostic)
    except (OSError, ValueError, yaml.YAMLError) as error:
        report = result([Finding("input", "render", str(error))], args.diagnostic)

    if args.json:
        print(json.dumps(report, indent=2, sort_keys=True))
    else:
        for finding in report["findings"]:
            print(f"FAIL [{finding['invariant']}] {finding['location']}: {finding['message']}")
        print("SLURM semantic validation: " + ("passed" if report["passed"] else "blocked"))
    return 0 if report["passed"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
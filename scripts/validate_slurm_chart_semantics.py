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
SOURCE_VERIFIABLE_INVARIANTS = {"stable-secrets", "immutable-images"}
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
HELM_DIGEST_PATTERN = (
    r"^[a-z0-9]+([._-][a-z0-9]+)*(:[0-9]+)?"
    r"(/[a-z0-9]+([._-][a-z0-9]+)*)*@sha256:[a-f0-9]{64}$"
)
HELM_RANDOM = re.compile(
    r"{{[^}]*\b(?:randAlpha|randAlphaNum|randAscii|randBytes|randNumeric|genCA|genPrivateKey|"
    r"genSelfSignedCert|genSignedCert|derivePassword|htpasswd|uuidv4)\b",
    re.IGNORECASE,
)
PRIVILEGED_SOURCE = re.compile(r"^\s*(?:privileged:\s*true|runAsUser:\s*0|runAsNonRoot:\s*false|allowPrivilegeEscalation:\s*true)\b", re.MULTILINE)
HOST_INTEGRATION_SOURCE = re.compile(r"^\s*(?:hostPath:|hostPID:\s*true|hostIPC:\s*true|hostNetwork:\s*true|hostUsers:\s*false|procMount:\s*Unmasked)\b", re.MULTILINE)
CAPABILITY_ADD_SOURCE = re.compile(r"^\s*add:\s*(?:\[|$)", re.MULTILINE)
NODE_LINE = re.compile(r"^\s*NodeName=(\S+)", re.MULTILINE)
PARTITION_LINE = re.compile(r"^\s*PartitionName=(\S+)(.*)$", re.MULTILINE)
SLURM_USER_LINE = re.compile(r"^\s*SlurmUser=(\S+)", re.MULTILINE)
SLURMD_USER_LINE = re.compile(r"^\s*SlurmdUser=(\S+)", re.MULTILINE)
HELPER_DEFINITION = re.compile(r'{{-?\s*define\s+"([^"]+)"\s*-?}}(.*?){{-?\s*end\s*}}', re.DOTALL)
IMAGE_HELPERS = {
    "munge": ("slurm-cluster.munge.image", "munge", "munge.image"),
    "controller": ("slurm-cluster.controller.image", "controller", "controller.image"),
    "database": ("slurm-cluster.database.image", "database", "database.image"),
    "mariadb": ("slurm-cluster.mariadb.image", "mariadb", "mariadb.image"),
    "compute": ("slurm-cluster.compute.image", "compute", "compute.image"),
    "nodeAgent": ("slurm-cluster.nodeAgent.image", "nodeAgent", "nodeAgent.image"),
    "utility": ("slurm-cluster.utility.image", "utility", "utilityImage"),
}
IDENTITY_COMPONENTS = ("munge", "slurm", "mariadb", "nodeAgent", "utility")
SLURM_DAEMON_COMPONENTS = ("controller", "database", "compute")
FORBIDDEN_SECURITY_KEYS = {
    "securityContext",
    "podSecurityContext",
    "containerSecurityContext",
    "extraContainers",
    "initContainers",
    "hostPID",
    "hostIPC",
    "hostNetwork",
    "hostUsers",
    "hostPath",
    "hostPaths",
}


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
        "compute": ("compute", "slurmd"),
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


def _slurmdbd_config(documents: list[dict[str, Any]]) -> str:
    chunks = []
    for document in documents:
        if document.get("kind") != "ConfigMap":
            continue
        data = document.get("data") if isinstance(document.get("data"), dict) else {}
        if isinstance(data.get("slurmdbd.conf"), str):
            chunks.append(data["slurmdbd.conf"])
    return "\n".join(chunks)


def _forbidden_value_paths(value: Any, path: str = "") -> list[str]:
    found: list[str] = []
    if isinstance(value, dict):
        for key, child in value.items():
            child_path = f"{path}.{key}" if path else str(key)
            if key in FORBIDDEN_SECURITY_KEYS:
                found.append(child_path)
            found.extend(_forbidden_value_paths(child, child_path))
    elif isinstance(value, list):
        for index, child in enumerate(value):
            found.extend(_forbidden_value_paths(child, f"{path}[{index}]"))
    return found


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
        if path.name != "values.yaml" and HOST_INTEGRATION_SOURCE.search(content):
            findings.append(Finding("least-privilege", relative, "template contains forbidden host integration"))
        if path.name != "values.yaml" and CAPABILITY_ADD_SOURCE.search(content):
            findings.append(Finding("least-privilege", relative, "template adds Linux capabilities"))

    template_source = "\n".join(
        path.read_text(encoding="utf-8")
        for path in (chart_dir / "templates").rglob("*")
        if path.is_file()
    )
    least_privilege_contract = {
        "templates/_helpers.tpl": (
            'define "slurm-cluster.podSecurityContext"',
            "fsGroupChangePolicy: OnRootMismatch",
            'define "slurm-cluster.containerSecurityContext"',
            "allowPrivilegeEscalation: false",
            "privileged: false",
            "readOnlyRootFilesystem: true",
            "runAsNonRoot: true",
            "$identity.gid",
            "$identity.uid",
            '$identityKey = "slurm"',
            'define "slurm-cluster.slurmUser"',
            "securityIdentities.slurm.username is required",
            "type: RuntimeDefault",
        ),
        "templates/controller-statefulset.yaml": ('include "slurm-cluster.podSecurityContext"', 'include "slurm-cluster.containerSecurityContext"'),
        "templates/database-statefulset.yaml": ('include "slurm-cluster.podSecurityContext"', 'include "slurm-cluster.containerSecurityContext"'),
        "templates/mariadb-statefulset.yaml": ('include "slurm-cluster.podSecurityContext"', 'include "slurm-cluster.containerSecurityContext"'),
        "templates/compute-statefulset.yaml": ('include "slurm-cluster.podSecurityContext"', 'include "slurm-cluster.containerSecurityContext"'),
        "templates/compute-nodepools-statefulset.yaml": ('include "slurm-cluster.podSecurityContext"', 'include "slurm-cluster.containerSecurityContext"'),
        "templates/configmap.yaml": (
            'SlurmUser={{ include "slurm-cluster.slurmUser" . }}',
            'SlurmdUser={{ include "slurm-cluster.slurmUser" . }}',
            "JobAcctGatherType={{ .Values.controller.config.jobAcctGatherType | default \"jobacct_gather/linux\" }}",
            "ProctrackType={{ .Values.controller.config.proctrackType | default \"proctrack/linuxproc\" }}",
            "TaskPlugin={{ .Values.controller.config.taskPlugin | default \"task/affinity\" }}",
            "CgroupAutomount=no",
            "ConstrainCores=no",
            "ConstrainDevices=no",
            "ConstrainRAMSpace=no",
            "ConstrainSwapSpace=no",
        ),
    }
    for relative, required_fragments in least_privilege_contract.items():
        path = chart_dir / relative
        content = path.read_text(encoding="utf-8") if path.is_file() else ""
        for fragment in required_fragments:
            if fragment not in content:
                findings.append(Finding("least-privilege", relative, f"least-privilege contract is missing {fragment}"))
    expected_security_helper_counts = {
        "templates/controller-statefulset.yaml": 3,
        "templates/database-statefulset.yaml": 4,
        "templates/mariadb-statefulset.yaml": 1,
        "templates/compute-statefulset.yaml": 5,
        "templates/compute-nodepools-statefulset.yaml": 5,
    }
    for relative, expected_count in expected_security_helper_counts.items():
        path = chart_dir / relative
        content = path.read_text(encoding="utf-8") if path.is_file() else ""
        pod_count = content.count('include "slurm-cluster.podSecurityContext"')
        container_count = content.count('include "slurm-cluster.containerSecurityContext"')
        image_count = len(re.findall(r"^\s*image:\s*", content, re.MULTILINE))
        if pod_count != 1 or container_count != expected_count or container_count != image_count:
            findings.append(Finding("least-privilege", relative, f"expected 1 pod and {expected_count} container security helper uses matching every image; found {pod_count}, {container_count}, and {image_count} images"))
    identity_bindings = {
        "templates/controller-statefulset.yaml": {"munge": 2, "controller": 1},
        "templates/database-statefulset.yaml": {"munge": 2, "utility": 1, "database": 1},
        "templates/mariadb-statefulset.yaml": {"mariadb": 1},
        "templates/compute-statefulset.yaml": {"munge": 2, "utility": 1, "compute": 1, "nodeAgent": 1},
        "templates/compute-nodepools-statefulset.yaml": {"munge": 2, "utility": 1, "compute": 1, "nodeAgent": 1},
    }
    for relative, bindings in identity_bindings.items():
        path = chart_dir / relative
        content = path.read_text(encoding="utf-8") if path.is_file() else ""
        context = "$root" if "nodepools" in relative else "."
        for component, expected in bindings.items():
            binding = f'include "slurm-cluster.containerSecurityContext" (list {context} "{component}")'
            if content.count(binding) != expected:
                findings.append(Finding("least-privilege", relative, f"identity contract for {component} must be used {expected} time(s)"))
    for forbidden in (".Values.podSecurityContext", ".Values.containerSecurityContext", "chown munge:munge", "/sys/fs/cgroup", "CgroupPlugin=", "ConstrainCores=yes", "ConstrainDevices=yes", "ConstrainRAMSpace=yes", "ConstrainSwapSpace=yes"):
        if forbidden in template_source:
            findings.append(Finding("least-privilege", "templates", f"least-privilege bypass is forbidden: {forbidden}"))
    for forbidden_path in _forbidden_value_paths(values):
        findings.append(Finding("least-privilege", "values.yaml", f"configurable security override is forbidden: {forbidden_path}"))
    identities = values.get("securityIdentities") if isinstance(values, dict) else None
    if not isinstance(identities, dict) or set(identities) != set(IDENTITY_COMPONENTS):
        findings.append(Finding("least-privilege", "values.yaml", "securityIdentities must contain only the shared slurm identity and approved component-specific identities"))
    for component in IDENTITY_COMPONENTS:
        identity = identities.get(component) if isinstance(identities, dict) else None
        required_fields = {"username", "uid", "gid"} if component == "slurm" else {"uid", "gid"}
        if not isinstance(identity, dict) or set(identity) != required_fields:
            findings.append(Finding("least-privilege", "values.yaml", f"securityIdentities.{component} must define exactly {', '.join(sorted(required_fields))}"))
            continue
        if not all(isinstance(identity[field], int) and not isinstance(identity[field], bool) and identity[field] > 0 for field in ("uid", "gid")):
            findings.append(Finding("least-privilege", "values.yaml", f"securityIdentities.{component} uid and gid must be positive integers"))
        if component == "slurm" and not isinstance(identity.get("username"), str):
            findings.append(Finding("least-privilege", "values.yaml", "securityIdentities.slurm.username must be a string"))
    configmap_source = (chart_dir / "templates" / "configmap.yaml").read_text(encoding="utf-8") if (chart_dir / "templates" / "configmap.yaml").is_file() else ""
    slurm_user_directive = 'SlurmUser={{ include "slurm-cluster.slurmUser" . }}'
    slurmd_user_directive = 'SlurmdUser={{ include "slurm-cluster.slurmUser" . }}'
    if configmap_source.count(slurm_user_directive) != 2 or configmap_source.count(slurmd_user_directive) != 1:
        findings.append(Finding("least-privilege", "templates/configmap.yaml", "slurm.conf and slurmdbd.conf must render the shared SlurmUser and slurm.conf must render SlurmdUser"))
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
    compute_schema = properties.get("compute", {}) if isinstance(properties.get("compute"), dict) else {}
    extra_volume_items = properties.get("extraVolumes", {}).get("items", {}) if isinstance(properties.get("extraVolumes"), dict) else {}
    if any(properties.get(key) is not False for key in FORBIDDEN_SECURITY_KEYS):
        findings.append(Finding("least-privilege", "values.schema.json", "security context overrides must be forbidden"))
    if compute_schema.get("properties", {}).get("hostCgroup") is not False:
        findings.append(Finding("least-privilege", "values.schema.json", "compute.hostCgroup must be forbidden in the production chart"))
    if extra_volume_items != {"$ref": "#/definitions/noSecurityOverrides"}:
        findings.append(Finding("least-privilege", "values.schema.json", "extraVolumes must deeply reject hostPath and security overrides"))
    rootless_profile = {
        "jobAcctGatherType": "jobacct_gather/linux",
        "proctrackType": "proctrack/linuxproc",
        "taskPlugin": "task/affinity",
    }
    controller_config = values.get("controller", {}).get("config", {})
    controller_config_schema = properties.get("controller", {}).get("properties", {}).get("config", {})
    for key, expected in rootless_profile.items():
        if controller_config.get(key) != expected:
            findings.append(Finding("least-privilege", "values.yaml", f"rootless prototype requires controller.config.{key}={expected}"))
        if controller_config_schema.get("properties", {}).get(key) != {"const": expected}:
            findings.append(Finding("least-privilege", "values.schema.json", f"rootless prototype schema must lock controller.config.{key}"))
    if controller_config_schema.get("required") != list(rootless_profile):
        findings.append(Finding("least-privilege", "values.schema.json", "rootless prototype plugin settings must be required"))
    identity_schema = schema.get("definitions", {}).get("securityIdentities", {})
    identity_definition = schema.get("definitions", {}).get("securityIdentity", {})
    slurm_identity_definition = schema.get("definitions", {}).get("slurmSecurityIdentity", {})
    if schema.get("required") != ["securityIdentities"] or identity_schema.get("required") != list(IDENTITY_COMPONENTS):
        findings.append(Finding("least-privilege", "values.schema.json", "all component security identities must be required"))
    fields = identity_definition.get("properties", {})
    if identity_definition.get("required") != ["uid", "gid"] or any(fields.get(field) != {"type": "integer", "minimum": 1} for field in ("uid", "gid")):
        findings.append(Finding("least-privilege", "values.schema.json", "identity uid and gid must be required positive integers"))
    slurm_fields = slurm_identity_definition.get("properties", {})
    if slurm_identity_definition.get("required") != ["username", "uid", "gid"] or any(slurm_fields.get(field) != {"type": "integer", "minimum": 1} for field in ("uid", "gid")) or "pattern" not in slurm_fields.get("username", {}):
        findings.append(Finding("least-privilege", "values.schema.json", "shared SLURM username, uid, and gid must be required"))
    forbidden_definition = json.dumps(schema.get("definitions", {}).get("noSecurityOverrides", {}), sort_keys=True)
    for key in FORBIDDEN_SECURITY_KEYS:
        if f'"{key}"' not in forbidden_definition:
            findings.append(Finding("least-privilege", "values.schema.json", f"recursive schema does not forbid {key}"))
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

    helpers_path = chart_dir / "templates" / "_helpers.tpl"
    helpers = helpers_path.read_text(encoding="utf-8") if helpers_path.is_file() else ""
    pinned_image = schema.get("definitions", {}).get("pinnedImage", {})
    image_pattern = pinned_image.get("properties", {}).get("reference", {}).get("pattern")
    if image_pattern != DIGEST_IMAGE.pattern:
        findings.append(Finding("immutable-images", "values.schema.json", "pinned image schema does not require the exact lowercase sha256 reference grammar"))
    pull_policy = pinned_image.get("properties", {}).get("pullPolicy", {})
    if pull_policy != {"enum": ["Always", "IfNotPresent", "Never"]} or pinned_image.get("required") != ["reference", "pullPolicy"]:
        findings.append(Finding("immutable-images", "values.schema.json", "pinned image schema must require pullPolicy with the Kubernetes policy enum"))
    helper_contract = (
        'define "slurm-cluster.immutableImage"',
        "required (printf",
        f'regexMatch "{HELM_DIGEST_PATTERN}"',
    )
    for fragment in helper_contract:
        if fragment not in helpers:
            findings.append(Finding("immutable-images", "templates/_helpers.tpl", f"fail-closed image helper is missing {fragment}"))
    helper_bodies = {name: body for name, body in HELPER_DEFINITION.findall(helpers)}
    approved_helpers = {helper: values_path for helper, _, values_path in IMAGE_HELPERS.values()}
    for component, (helper, label, values_path) in IMAGE_HELPERS.items():
        image = values.get("utilityImage", {}) if component == "utility" else values.get(component, {}).get("image", {})
        if not isinstance(image, dict) or set(image) != {"reference", "pullPolicy"}:
            findings.append(Finding("immutable-images", "values.yaml", f"{component} image must expose only reference and pullPolicy"))
        elif image["reference"] and not DIGEST_IMAGE.fullmatch(str(image["reference"])):
            findings.append(Finding("immutable-images", "values.yaml", f"{component} default image reference is not an exact lowercase sha256 reference"))
        body = helper_bodies.get(helper, "")
        expected_body = re.compile(
            rf'^\s*{{{{-?\s*include\s+"slurm-cluster\.immutableImage"\s+\(list\s+"{re.escape(label)}"\s+\.Values\.{re.escape(values_path)}\.reference\)\s*-?}}}}\s*$'
        )
        if not expected_body.fullmatch(body):
            findings.append(Finding("immutable-images", "templates/_helpers.tpl", f"{component} image helper must call immutableImage with .Values.{values_path}.reference"))
        if f'include "{helper}"' not in template_source:
            findings.append(Finding("immutable-images", "templates", f"{component} image helper is not used by a workload"))
    for path in (chart_dir / "templates").glob("*.yaml"):
        lines = path.read_text(encoding="utf-8").splitlines()
        paired_pull_policy_lines: set[int] = set()
        for line_number, line in enumerate(lines, 1):
            if not re.match(r"^\s*image:\s*", line):
                continue
            match = re.fullmatch(r'\s*image:\s*{{\s*include\s+"([^"]+)"\s+(\$root|\.)\s*}}\s*', line)
            if not match or match.group(1) not in approved_helpers:
                findings.append(Finding("immutable-images", f"templates/{path.name}:{line_number}", "container image bypasses an approved immutable image helper"))
                continue
            helper, context = match.groups()
            values_path = approved_helpers[helper]
            values_root = ".Values" if context == "." else f"{context}.Values"
            expected_source = f"{values_root}.{values_path}.pullPolicy"
            policy_line = lines[line_number] if line_number < len(lines) else ""
            expected_policy = re.compile(rf"\s*imagePullPolicy:\s*{{{{\s*{re.escape(expected_source)}\s*}}}}\s*")
            if not expected_policy.fullmatch(policy_line):
                findings.append(Finding("immutable-images", f"templates/{path.name}:{line_number + 1}", f"{helper} must use imagePullPolicy from {expected_source}"))
            else:
                paired_pull_policy_lines.add(line_number + 1)
        for line_number, line in enumerate(lines, 1):
            if re.match(r"^\s*imagePullPolicy:\s*", line) and line_number not in paired_pull_policy_lines:
                findings.append(Finding("immutable-images", f"templates/{path.name}:{line_number}", "imagePullPolicy is not paired with its approved image values path"))

    for component in DURABLE_COMPONENTS:
        settings = values.get(component, {})
        persistence = settings.get("persistence", {}) if isinstance(settings, dict) else {}
        if settings.get("enabled", True) and not persistence.get("enabled", False):
            findings.append(Finding("durable-state", "values.yaml", f"{component} persistence is not enabled"))
        if settings.get("enabled", True) and set(persistence) != {"enabled", "existingClaim", "size", "storageClass", "accessMode"}:
            findings.append(Finding("durable-state", "values.yaml", f"{component} persistence must expose only enabled, existingClaim, size, storageClass, and accessMode"))

    durable_persistence = schema.get("definitions", {}).get("durablePersistence", {})
    if durable_persistence.get("required") != ["enabled", "existingClaim", "size", "storageClass", "accessMode"] or durable_persistence.get("additionalProperties") is not False:
        findings.append(Finding("durable-state", "values.schema.json", "durablePersistence must require the closed retained-claim contract"))
    if durable_persistence.get("properties", {}).get("enabled") != {"const": True}:
        findings.append(Finding("durable-state", "values.schema.json", "durable persistence cannot be disabled"))
    claim_schema = durable_persistence.get("properties", {}).get("existingClaim", {})
    if claim_schema.get("maxLength") != 253 or not claim_schema.get("pattern", "").startswith("^$|"):
        findings.append(Finding("durable-state", "values.schema.json", "existingClaim must allow empty or a bounded Kubernetes PVC name"))
    persistence_refs = {
        "controller": properties.get("controller", {}).get("properties", {}).get("persistence"),
        "database": properties.get("database", {}).get("properties", {}).get("persistence"),
        "mariadb": schema.get("definitions", {}).get("mariadbComponent", {}).get("properties", {}).get("persistence"),
    }
    for component, reference in persistence_refs.items():
        if reference != {"$ref": "#/definitions/durablePersistence"}:
            findings.append(Finding("durable-state", "values.schema.json", f"{component} must use durablePersistence"))
    replica_safety = schema.get("definitions", {}).get("existingClaimReplicaSafety", {})
    if replica_safety.get("then") != {"properties": {"replicas": {"const": 1}}, "required": ["replicas"]}:
        findings.append(Finding("durable-state", "values.schema.json", "existing claims must require exactly one replica"))
    for component in ("controller", "database"):
        component_schema = properties.get(component, {})
        if component_schema.get("allOf") != [{"$ref": "#/definitions/existingClaimReplicaSafety"}]:
            findings.append(Finding("durable-state", "values.schema.json", f"{component} must apply existing-claim replica safety"))
    conditional_contract = json.dumps(schema.get("allOf", []), sort_keys=True)
    for component in DURABLE_COMPONENTS:
        if not re.search(rf'"{component}".*"required": \["image", "persistence"\]', conditional_contract):
            findings.append(Finding("durable-state", "values.schema.json", f"enabled {component} must require persistence"))

    durable_templates = {
        "controller": ("templates/controller-statefulset.yaml", "slurm-spool", "/var/spool/slurm"),
        "database": ("templates/database-statefulset.yaml", "slurmdbd-spool", "/var/spool/slurm"),
        "mariadb": ("templates/mariadb-statefulset.yaml", "mariadb-data", "/var/lib/mysql"),
    }
    for component, (relative, volume_name, mount_path) in durable_templates.items():
        content = (chart_dir / relative).read_text(encoding="utf-8") if (chart_dir / relative).is_file() else ""
        required_fragments = (
            f".Values.{component}.persistence.existingClaim",
            'include "slurm-cluster.requireSafePersistenceReplicas"',
            "persistentVolumeClaimRetentionPolicy:",
            "whenDeleted: Retain",
            "whenScaled: Retain",
            "persistentVolumeClaim:",
            "claimName:",
            "volumeClaimTemplates:",
            f"name: {volume_name}",
            f"mountPath: {mount_path}",
        )
        for fragment in required_fragments:
            if fragment not in content:
                findings.append(Finding("durable-state", relative, f"durable-state contract is missing {fragment}"))
        if re.search(rf"name:\s*{re.escape(volume_name)}\s*\n\s*emptyDir:", content):
            findings.append(Finding("durable-state", relative, f"authoritative volume {volume_name} must not use emptyDir"))
    return findings


def validate_source(chart_dir: Path) -> list[Finding]:
    return _static_findings(chart_dir)


def validate(documents: list[dict[str, Any]], chart_dir: Path | None = None) -> list[Finding]:
    documents = [document for value in documents for document in _documents(value)]
    findings = _static_findings(chart_dir) if chart_dir else []
    compute_replicas = 0
    stateful_compute_nodes: set[str] = set()
    durable_seen: set[str] = set()
    daemon_identities: dict[str, set[tuple[Any, Any]]] = {component: set() for component in SLURM_DAEMON_COMPONENTS}

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
            host_features = [field for field in ("hostIPC", "hostNetwork", "hostPID") if pod_spec.get(field) is True]
            if pod_spec.get("hostUsers") is False:
                host_features.append("hostUsers=false")
            host_paths = [str(volume.get("name", "<unnamed>")) for volume in pod_spec.get("volumes", []) or [] if isinstance(volume, dict) and "hostPath" in volume]
            if host_features or host_paths:
                details = []
                if host_features:
                    details.append(f"host features are forbidden: {', '.join(host_features)}")
                if host_paths:
                    details.append(f"hostPath volumes are forbidden: {', '.join(host_paths)}")
                findings.append(Finding("least-privilege", location, "; ".join(details)))
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
                if security.get("readOnlyRootFilesystem") is not True:
                    violations.append("readOnlyRootFilesystem must be true")
                if "ALL" not in dropped:
                    violations.append("capabilities.drop must include ALL")
                if added:
                    violations.append(f"capabilities.add must be empty: {', '.join(sorted(added))}")
                if seccomp_type not in {"RuntimeDefault", "Localhost"}:
                    violations.append("seccompProfile.type must be RuntimeDefault or Localhost")
                if security.get("procMount") == "Unmasked":
                    violations.append("procMount Unmasked is forbidden")
                windows_options = security.get("windowsOptions") if isinstance(security.get("windowsOptions"), dict) else {}
                if windows_options.get("hostProcess") is True:
                    violations.append("windowsOptions.hostProcess is forbidden")
                if violations:
                    findings.append(Finding("least-privilege", container_location, "; ".join(violations)))
            if component in SLURM_DAEMON_COMPONENTS:
                for container in _primary_containers(pod_spec, component):
                    security = container.get("securityContext") if isinstance(container.get("securityContext"), dict) else {}
                    daemon_identities[component].add((security.get("runAsUser", pod_security.get("runAsUser")), security.get("runAsGroup", pod_security.get("runAsGroup"))))

        if kind in {"Deployment", "ReplicaSet", "ReplicationController", "StatefulSet"} and component == "compute":
            replicas = int(document.get("spec", {}).get("replicas", 1))
            compute_replicas += replicas
            if kind == "StatefulSet":
                stateful_compute_nodes.update(f"{name}-{ordinal}" for ordinal in range(replicas))
        if pod_spec is not None and component in DURABLE_COMPONENTS:
            durable_seen.add(component)
            spec = document.get("spec", {})
            claims = {claim.get("metadata", {}).get("name") for claim in spec.get("volumeClaimTemplates", []) or [] if isinstance(claim, dict)}
            if kind != "StatefulSet":
                findings.append(Finding("durable-state", location, f"required durable component {component} must be a StatefulSet"))
            if claims:
                retention = spec.get("persistentVolumeClaimRetentionPolicy", {})
                if retention != {"whenDeleted": "Retain", "whenScaled": "Retain"}:
                    findings.append(Finding("durable-state", location, "generated claims must retain PVCs when deleted and scaled"))
            elif int(spec.get("replicas", 1)) != 1:
                findings.append(Finding("durable-state", location, "an existing durable claim requires replicas=1; HA must use generated per-replica claims"))
            volumes = {volume.get("name"): volume for volume in pod_spec.get("volumes", []) or [] if isinstance(volume, dict)}
            for container in _primary_containers(pod_spec, component):
                mounted_paths = {mount.get("mountPath"): mount.get("name") for mount in container.get("volumeMounts", []) or [] if isinstance(mount, dict)}
                for required_path in DURABLE_PATHS[component]:
                    volume_name = mounted_paths.get(required_path)
                    volume = volumes.get(volume_name, {})
                    if volume_name not in claims and not isinstance(volume.get("persistentVolumeClaim"), dict):
                        findings.append(Finding("durable-state", f"{location}:containers/{container.get('name', '<unnamed>')}", f"{required_path} is not backed by a persistent volume claim"))

    config = _slurm_config(documents)
    slurmdbd_config = _slurmdbd_config(documents)
    slurm_users = SLURM_USER_LINE.findall(config)
    slurmd_users = SLURMD_USER_LINE.findall(config)
    slurmdbd_users = SLURM_USER_LINE.findall(slurmdbd_config)
    if len(slurm_users) != 1 or len(slurmd_users) != 1 or len(slurmdbd_users) != 1:
        findings.append(Finding("least-privilege", "ConfigMap/SLURM user contracts", "slurm.conf must define one SlurmUser and SlurmdUser and slurmdbd.conf must define one SlurmUser"))
    elif len({slurm_users[0], slurmd_users[0], slurmdbd_users[0]}) != 1:
        findings.append(Finding("least-privilege", "ConfigMap/SLURM user contracts", "SlurmUser and SlurmdUser must use one shared daemon username"))
    observed_daemon_ids = set().union(*daemon_identities.values())
    if any(len(identities) != 1 or (None, None) in identities or any(None in identity for identity in identities) for identities in daemon_identities.values()) or len(observed_daemon_ids) != 1:
        findings.append(Finding("least-privilege", "render/SLURM daemon identities", "controller, database, and compute daemons must use one explicit shared runAsUser/runAsGroup identity"))
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
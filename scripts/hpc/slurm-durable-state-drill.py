#!/usr/bin/env python3

"""Exercise checksum-gated SLURM durable-state backup and restore offline."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import shutil
import tarfile
import tempfile
from pathlib import Path, PurePosixPath, PureWindowsPath


RESTORE_ORDER = ("mariadb", "slurmdbd", "slurmctld")
MANIFEST_NAME = "manifest.json"


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def backup(source: Path, bundle: Path) -> dict[str, object]:
    bundle.mkdir(parents=True, mode=0o700, exist_ok=False)
    os.chmod(bundle, 0o700)
    archives: dict[str, dict[str, str]] = {}
    for component in RESTORE_ORDER:
        component_source = source / component
        if not component_source.is_dir():
            raise ValueError(f"missing authoritative state directory: {component_source}")
        archive = bundle / f"{component}.tar.gz"
        with tarfile.open(archive, "w:gz") as handle:
            handle.add(component_source, arcname=component)
        _validated_members(component, archive)
        os.chmod(archive, 0o600)
        archives[component] = {"file": archive.name, "sha256": sha256(archive)}
    manifest: dict[str, object] = {
        "schema_version": "virtengine.slurm-durable-state-backup/v1",
        "restore_order": list(RESTORE_ORDER),
        "archives": archives,
    }
    manifest_path = bundle / MANIFEST_NAME
    manifest_path.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    os.chmod(manifest_path, 0o600)
    return manifest


def _manifest_archive(bundle: Path, component: str, value: object) -> Path:
    if not isinstance(value, str) or not value or value in {".", ".."}:
        raise ValueError(f"archive filename for {component} must be a simple basename")
    if Path(value).is_absolute() or PureWindowsPath(value).is_absolute() or Path(value).name != value or "/" in value or "\\" in value:
        raise ValueError(f"archive filename for {component} must be a simple basename")
    return bundle / value


def _validated_members(component: str, archive: Path) -> list[tarfile.TarInfo]:
    with tarfile.open(archive, "r:gz") as handle:
        members = handle.getmembers()
    if not members:
        raise ValueError(f"archive for {component} is empty")
    root_entries = 0
    seen: set[str] = set()
    for member in members:
        path = PurePosixPath(member.name)
        if path.is_absolute() or PureWindowsPath(member.name).is_absolute() or "\\" in member.name or ".." in path.parts:
            raise ValueError(f"unsafe archive path for {component}")
        if not (member.isfile() or member.isdir()):
            raise ValueError(f"archive for {component} contains a non-regular entry")
        if not path.parts or path.parts[0] != component:
            raise ValueError(f"archive entry escapes component root for {component}")
        normalized = path.as_posix().rstrip("/")
        if normalized in seen:
            raise ValueError(f"archive for {component} contains duplicate entries")
        seen.add(normalized)
        if normalized == component:
            root_entries += 1
            if not member.isdir():
                raise ValueError(f"archive component root for {component} must be a directory")
    if root_entries != 1:
        raise ValueError(f"archive for {component} must contain one component root directory")
    return members


def _validated_archives(bundle: Path) -> list[tuple[str, Path]]:
    manifest = json.loads((bundle / MANIFEST_NAME).read_text(encoding="utf-8"))
    if manifest.get("schema_version") != "virtengine.slurm-durable-state-backup/v1":
        raise ValueError("unsupported backup manifest")
    if manifest.get("restore_order") != list(RESTORE_ORDER):
        raise ValueError("restore order must be mariadb, slurmdbd, slurmctld")
    archives = manifest.get("archives")
    if not isinstance(archives, dict) or set(archives) != set(RESTORE_ORDER):
        raise ValueError("backup manifest must contain exactly three authoritative archives")

    validated = []
    for component in RESTORE_ORDER:
        record = archives[component]
        if not isinstance(record, dict) or set(record) != {"file", "sha256"}:
            raise ValueError(f"invalid manifest record for {component}")
        archive = _manifest_archive(bundle, component, record["file"])
        if not archive.is_file() or sha256(archive) != record["sha256"]:
            raise ValueError(f"checksum verification failed for {component}")
        _validated_members(component, archive)
        validated.append((component, archive))
    return validated


def restore(bundle: Path, destination: Path) -> list[str]:
    validated = _validated_archives(bundle)
    if destination.exists():
        raise ValueError(f"restore destination must not exist: {destination}")

    staging = destination.with_name(f".{destination.name}.restore")
    if staging.exists():
        raise ValueError(f"restore staging directory must not exist: {staging}")
    staging.mkdir(parents=True, mode=0o700)
    os.chmod(staging, 0o700)
    restored = []
    try:
        for component, archive in validated:
            with tarfile.open(archive, "r:gz") as handle:
                members = _validated_members(component, archive)
                handle.extractall(staging, members=members, filter="data")
            restored.append(component)
        staging.replace(destination)
    except Exception:
        shutil.rmtree(staging, ignore_errors=True)
        raise
    return restored


def drill(workdir: Path) -> None:
    source = workdir / "source"
    bundle = workdir / "backup"
    destination = workdir / "restored"
    fixtures = {
        "mariadb": ("slurm.sql", "CREATE TABLE job_table (id INT);\n"),
        "slurmdbd": ("archive_events", "event=42\n"),
        "slurmctld": ("state", "next_job_id=43\n"),
    }
    for component, (name, contents) in fixtures.items():
        path = source / component
        path.mkdir(parents=True)
        (path / name).write_text(contents, encoding="utf-8")

    backup(source, bundle)
    if restore(bundle, destination) != list(RESTORE_ORDER):
        raise RuntimeError("restore did not follow the required order")
    for component, (name, _) in fixtures.items():
        if (source / component / name).read_bytes() != (destination / component / name).read_bytes():
            raise RuntimeError(f"restored state mismatch for {component}")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--workdir", type=Path, help="directory for retained drill artifacts")
    args = parser.parse_args()
    if args.workdir:
        args.workdir.mkdir(parents=True, exist_ok=True)
        drill(args.workdir)
    else:
        with tempfile.TemporaryDirectory(prefix="slurm-durable-state-") as directory:
            drill(Path(directory))
    print("offline SLURM durable-state drill: passed (not live storage evidence)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
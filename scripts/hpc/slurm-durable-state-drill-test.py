#!/usr/bin/env python3

import importlib.util
import io
import json
import os
import stat
import tarfile
import tempfile
import unittest
from pathlib import Path
from unittest import mock


SCRIPT = Path(__file__).with_name("slurm-durable-state-drill.py")
SPEC = importlib.util.spec_from_file_location("slurm_durable_state_drill", SCRIPT)
assert SPEC and SPEC.loader
DRILL = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(DRILL)


class SlurmDurableStateDrillTest(unittest.TestCase):
    def fixture(self, root: Path) -> Path:
        source = root / "source"
        for component in DRILL.RESTORE_ORDER:
            path = source / component
            path.mkdir(parents=True)
            (path / "state").write_text(f"{component}\n", encoding="utf-8")
        return source

    def replace_archive(self, bundle: Path, component: str, members: list[tuple[tarfile.TarInfo, bytes]]) -> None:
        archive = bundle / f"{component}.tar.gz"
        with tarfile.open(archive, "w:gz") as handle:
            for member, contents in members:
                handle.addfile(member, io.BytesIO(contents) if member.isfile() else None)
        manifest_path = bundle / DRILL.MANIFEST_NAME
        manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
        manifest["archives"][component]["sha256"] = DRILL.sha256(archive)
        manifest_path.write_text(json.dumps(manifest), encoding="utf-8")

    def test_round_trip_uses_required_order(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            bundle = root / "backup"
            source = self.fixture(root)
            manifest = DRILL.backup(source, bundle)
            restored = root / "restored"
            self.assertEqual(manifest["restore_order"], ["mariadb", "slurmdbd", "slurmctld"])
            self.assertEqual(DRILL.restore(bundle, restored), ["mariadb", "slurmdbd", "slurmctld"])
            for component in DRILL.RESTORE_ORDER:
                self.assertEqual((source / component / "state").read_bytes(), (restored / component / "state").read_bytes())

    def test_tamper_fails_before_destination_is_created(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            bundle = root / "backup"
            DRILL.backup(self.fixture(root), bundle)
            with (bundle / "slurmdbd.tar.gz").open("ab") as handle:
                handle.write(b"tamper")
            destination = root / "restored"
            with self.assertRaisesRegex(ValueError, "checksum verification failed for slurmdbd"):
                DRILL.restore(bundle, destination)
            self.assertFalse(destination.exists())

    def test_manifest_order_change_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            bundle = root / "backup"
            DRILL.backup(self.fixture(root), bundle)
            manifest_path = bundle / DRILL.MANIFEST_NAME
            manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
            manifest["restore_order"].reverse()
            manifest_path.write_text(json.dumps(manifest), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "restore order"):
                DRILL.restore(bundle, root / "restored")

    def test_manifest_archive_filename_must_be_simple_basename(self) -> None:
        for filename in ("../outside.tar.gz", "/tmp/outside.tar.gz", "nested/archive.tar.gz", "C:\\outside.tar.gz"):
            with self.subTest(filename=filename), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                bundle = root / "backup"
                DRILL.backup(self.fixture(root), bundle)
                manifest_path = bundle / DRILL.MANIFEST_NAME
                manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
                manifest["archives"]["mariadb"]["file"] = filename
                manifest_path.write_text(json.dumps(manifest), encoding="utf-8")
                with self.assertRaisesRegex(ValueError, "simple basename"):
                    DRILL.restore(bundle, root / "restored")

    def test_rejects_traversal_and_cross_component_entries(self) -> None:
        for name, message in (("mariadb/../../outside", "unsafe archive path"), ("slurmdbd/state", "escapes component root")):
            with self.subTest(name=name), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                bundle = root / "backup"
                DRILL.backup(self.fixture(root), bundle)
                component_root = tarfile.TarInfo("mariadb")
                component_root.type = tarfile.DIRTYPE
                member = tarfile.TarInfo(name)
                member.size = 4
                self.replace_archive(bundle, "mariadb", [(component_root, b""), (member, b"data")])
                with self.assertRaisesRegex(ValueError, message):
                    DRILL.restore(bundle, root / "restored")
                self.assertFalse((root / "outside").exists())

    def test_rejects_links_and_special_archive_entries(self) -> None:
        entries = ((tarfile.SYMTYPE, "link"), (tarfile.LNKTYPE, "hardlink"), (tarfile.FIFOTYPE, "fifo"))
        for entry_type, label in entries:
            with self.subTest(label=label), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                bundle = root / "backup"
                DRILL.backup(self.fixture(root), bundle)
                component_root = tarfile.TarInfo("mariadb")
                component_root.type = tarfile.DIRTYPE
                member = tarfile.TarInfo("mariadb/unsafe")
                member.type = entry_type
                member.linkname = "mariadb/state"
                self.replace_archive(bundle, "mariadb", [(component_root, b""), (member, b"")])
                with self.assertRaisesRegex(ValueError, "non-regular entry"):
                    DRILL.restore(bundle, root / "restored")

    @unittest.skipIf(os.name == "nt", "creating symlinks may require Windows developer mode")
    def test_backup_rejects_source_symlink(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = self.fixture(root)
            (source / "mariadb" / "link").symlink_to(source / "mariadb" / "state")
            with self.assertRaisesRegex(ValueError, "non-regular entry"):
                DRILL.backup(source, root / "backup")

    @unittest.skipIf(os.name == "nt", "Windows does not expose POSIX mode bits")
    def test_backup_permissions_are_independent_of_umask(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            bundle = root / "backup"
            previous = os.umask(0)
            try:
                DRILL.backup(self.fixture(root), bundle)
            finally:
                os.umask(previous)
            self.assertEqual(stat.S_IMODE(bundle.stat().st_mode), 0o700)
            for path in [bundle / DRILL.MANIFEST_NAME, *(bundle / f"{component}.tar.gz" for component in DRILL.RESTORE_ORDER)]:
                self.assertEqual(stat.S_IMODE(path.stat().st_mode), 0o600)

    def test_backup_explicitly_enforces_restrictive_modes(self) -> None:
        with tempfile.TemporaryDirectory() as directory, mock.patch.object(DRILL.os, "chmod", wraps=DRILL.os.chmod) as chmod:
            root = Path(directory)
            bundle = root / "backup"
            DRILL.backup(self.fixture(root), bundle)
            chmod.assert_any_call(bundle, 0o700)
            chmod.assert_any_call(bundle / DRILL.MANIFEST_NAME, 0o600)
            for component in DRILL.RESTORE_ORDER:
                chmod.assert_any_call(bundle / f"{component}.tar.gz", 0o600)

    def test_preexisting_staging_and_target_are_not_mutated(self) -> None:
        for existing_name in (".restored.restore", "restored"):
            with self.subTest(existing_name=existing_name), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                bundle = root / "backup"
                DRILL.backup(self.fixture(root), bundle)
                existing = root / existing_name
                existing.mkdir()
                sentinel = existing / "sentinel"
                sentinel.write_text("retain\n", encoding="utf-8")
                with self.assertRaisesRegex(ValueError, "must not exist"):
                    DRILL.restore(bundle, root / "restored")
                self.assertEqual(sentinel.read_text(encoding="utf-8"), "retain\n")


if __name__ == "__main__":
    unittest.main()
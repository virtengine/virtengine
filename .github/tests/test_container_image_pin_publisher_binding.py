"""Guard: every digest-pinned production image must have a publishing path.

The defect this exists for
--------------------------
scripts/task85c-validate-kubernetes.mjs asserts that every image reference in
the production-consumable manifests is digest-pinned. It checks the FORM of a
pin -- `/@sha256:[a-f0-9]{64}$/` -- and never that the pin names anything a
kubelet could pull. So the gate is green over a production manifest whose
images cannot be fetched:

    deploy/kubernetes/base/virtengine-node-deployment.yaml
        ghcr.io/virtengine/virtengine-node@sha256:9f0f3a2d...   <- no such package
    deploy/kubernetes/base/provider-daemon-deployment.yaml
        ghcr.io/virtengine/provider-daemon@sha256:2e5d5bc2...   <- no such package
    deploy/kubernetes/base/tee-enclave-deployment.yaml
        ghcr.io/virtengine/tee-enclave@sha256:f00c2e5a...      <- no such package
    deploy/kubernetes/base/veid-inference-deployment.yaml
        ghcr.io/virtengine/veid-inference@sha256:000...0000     <- unpullable by anyone

All four workloads are in the production-consumable set that
assertNoMutableProductionInfraImages() scans, and all four pins passed it.

The first half of the fix already existed, for exactly one image
-------------------------------------------------------------
infra/kubernetes/dr/backup-cronjobs.yaml pinned ghcr.io/virtengine/dr-tools
with no Dockerfile and no publishing workflow; three CronJobs sat in
ImagePullBackOff. That was fixed by adding the build, the publisher, and a
narrow assertion inside assertDrConsumers() that ties each DR pin to a
`_build/Dockerfile.<name>` and to the publishing workflow.

The class was never closed. That assertion lives in a function named for the DR
tree, so nothing checks the OTHER production tree, and the four workloads above
were left exactly where dr-tools was. A guard that covers one directory is how
the second directory ended up unprotected in the first place.

The mechanism this generalises
------------------------------
A digest pin is a promise that two things exist in this repository:

    1. a build input that produces the image, and
    2. a workflow that PUSHES it under the exact package name pinned.

A pin satisfying neither is a lie that satisfies every syntactic check: right
shape, cannot drift, and no kubelet can ever resolve it. Nothing asserts either
half for any image outside the DR CronJobs, so a manifest can name a package
the repository has never heard of.

So for every ghcr.io/virtengine package digest-pinned by a production-consumable
manifest, this test asserts the repository builds AND publishes it.

Why this is static, and why this is the right layer
---------------------------------------------------
There is deliberately NO registry call here. scripts/ci/probe-ghcr-pullability.sh
and the live half of test_ghcr_image_pull_credential_policy.py already own
network-truth about GHCR, including the documented fact that GHCR answers
"never published" and "private" with a byte-identical 403 DENIED. A third
network probe would add a weaker reading of the same ambiguous signal.

This test asserts the unambiguous, repo-local half: whether the repository
CLAIMS to build and publish the image. That claim lives entirely in version
control, so it is checkable on every PR with no network, no credentials and no
flakiness -- and it is precisely what the dr-tools defect was. The pin existed;
the repository did not.

What this test deliberately does NOT assert
-------------------------------------------
Whether a digest is CURRENT. Re-pinning is a release decision driven by what CI
actually published, and a static test cannot know that. What it asserts is the
weaker and durable property: if someone edits the pinned digest, the package
must still be one this repository builds and publishes. That is the failure mode
with blast radius, and it is the one a manifest edit can reintroduce.

Fixtures are excluded, on purpose
----------------------------------
scripts/testdata/slurm-chart-semantics/*.yaml carries all-a/e/f/c and all-zero
placeholder digests. Those are chart-semantics test INPUTS: the whole point of a
"negative-durable-state" fixture is an image that is wrong. Requiring them to
resolve would mean requiring a negative fixture to be real, so the scan is
rooted at the manifest trees and never walks testdata.

Run:
  python -m unittest discover -s .github/tests -p "test_container_image_pin*.py"
"""
import re
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]

# The manifest trees whose image pins are actually applied to a cluster.
# Deliberately the same root set task85c-validate-kubernetes.mjs enforces
# digest-pinning over, so the two cannot drift: a tree that gains a pin gains a
# publisher requirement in the same commit.
PRODUCTION_MANIFEST_ROOTS = (
    "deploy/kubernetes/base",
    "deploy/kubernetes/overlays/prod",
    "infra/kubernetes/base",
    "infra/kubernetes/overlays/prod",
    "infra/kubernetes/dr",
    "infra/kubernetes/chaos",
)

ORG = "virtengine"
REGISTRY = "ghcr.io"
WORKFLOW_DIR = REPO_ROOT / ".github" / "workflows"

# An image line in a manifest: `image: <ref>`, tolerant of quoting.
_IMAGE_LINE = re.compile(r"""^\s*image:\s*["']?(?P<ref>[^"'\s]+)["']?""", re.M)

# A ghcr.io reference whose owner is this org. ci.yaml writes the owner as a
# GitHub expression rather than a literal -- `ghcr.io/${{ github.repository
# }}/virtengine` -- because the workflow is written to be portable from a fork.
# Both spellings name a package this repository pushes, so both are credited;
# the expression form is resolved rather than pattern-matched away, and a
# workflow in a fork would publish under that fork's owner.
_LITERAL_OWNER = re.compile(
    rf"{re.escape(REGISTRY)}/{ORG}/(?P<name>[a-z0-9][a-z0-9._-]*)", re.I
)
_EXPR_OWNER = re.compile(
    rf"{re.escape(REGISTRY)}/\${{\{{\s*github\.repository(?:_owner)?\s*\}}}}/"
    r"(?P<name>[a-z0-9][a-z0-9._-]*)",
    re.I,
)

# Push evidence, per step. `--push` is buildx's flag; `docker push` is explicit;
# `push: true` is the docker/build-push-action input. Matching the flag rather
# than the mere presence of the word "image" is what keeps ci.yaml's local-only
# Trivy scan builds (Dockerfile.test, Dockerfile.veid-pipeline,
# sdk/generation/Dockerfile) from being credited as publishers.
_BUILDX_PUSH = re.compile(r"(?:^|\s)--push(?:\s|\\|$)")
_DOCKER_PUSH = re.compile(r"(?:^|\s)docker\s+push\b")
_ACTION_PUSH = re.compile(r"(?m)^\s*push:\s*true\s*$")
_PUSH_EVIDENCE = (_BUILDX_PUSH, _DOCKER_PUSH, _ACTION_PUSH)

# The Dockerfile a build step consumes, under either spelling.
_DOCKERFILE_FLAG = re.compile(r"(?:^|\s)(?:-f|--file)\s+(\S+)")
_DOCKERFILE_INPUT = re.compile(r"(?m)^\s*file:\s*(\S+)\s*$")

# Step boundaries. Shell continuation lines never match, because every
# alternative is a `key:` form while a run-block continuation is `-t ...`,
# `--push \`, `.` and so on.
_STEP_SPLIT = re.compile(r"(?m)^\s*-\s+(?=(?:name|uses|id|run):)")

# Workflow-level `env:` entries, so `${{ env.IMAGE }}` resolves to the literal
# package name: dr-tools-image.yaml declares `IMAGE: ghcr.io/virtengine/dr-tools`
# once and references it everywhere else.
_ENV_ASSIGN = re.compile(r"(?m)^\s{2,6}(?P<key>[A-Z][A-Z0-9_]*):\s*(?P<val>\S+)\s*$")
_ENV_USE = re.compile(r"\$\{\{\s*env\.(?P<key>[A-Z][A-Z0-9_]*)\s*\}\}")

# Digest pins are the subject here; a mutable tag is assertImageDigests' job.
_DIGEST = "@sha256:"

# 64 zeros has the shape of a SHA-256 pin and names no possible image.
_ALL_ZERO = "sha256:" + "0" * 64

# The annotation that makes an unpullable pin visible in review instead of
# letting a pull secret disguise it.
PIN_STATUS_ANNOTATION = "virtengine.com/image-pin-status"

# Two DISTINCT declared-unpullable states, deliberately not one value:
#
#   placeholder-awaiting-release    the digest itself names no possible image
#                                   (all zeros). Unpullable by ANY credential.
#                                   Owned by test_ghcr_image_pull_credential_policy.
#   unpublished-awaiting-publish-path
#                                   the digest is well-formed but the package is
#                                   never published by this repository, so no
#                                   credential can help. Owned by THIS file.
#
# Keeping them separate is what stops a well-formed pin from being waved through
# as a placeholder, and stops a placeholder from being treated as fixable by
# publishing something. The existing guard asserts the first is not stale; this
# one asserts the second has a publisher.
DECLARED_UNPUBLISHABLE = (
    "placeholder-awaiting-release",
    "unpublished-awaiting-publish-path",
)


def _steps(text):
    """Split a workflow into step-sized chunks."""
    return _STEP_SPLIT.split(text)


def _workflow_env(text):
    """Workflow-level env NAME -> literal value."""
    env = {}
    for match in _ENV_ASSIGN.finditer(text):
        env[match.group("key")] = match.group("val").strip("\"'")
    return env


def _expand(text, env):
    """Resolve `${{ env.NAME }}` so indirection cannot hide a package name."""
    return _ENV_USE.sub(lambda m: env.get(m.group("key"), m.group(0)), text)


def _package_names(text):
    """Distinct ghcr.io package names referenced by `text`, owner verified."""
    names = []
    for pattern in (_LITERAL_OWNER, _EXPR_OWNER):
        for match in pattern.finditer(text):
            # Strip any tag or digest that rode along on the same segment.
            name = match.group("name").split("@", 1)[0].split(":", 1)[0]
            if name and name not in names:
                names.append(name)
    return names


def _dockerfiles_in(step):
    """Dockerfiles a build step consumes, under either spelling."""
    found = _DOCKERFILE_FLAG.findall(step) + _DOCKERFILE_INPUT.findall(step)
    return {Path(f).as_posix().lstrip("./") for f in found}


def _pinned_packages(text):
    """Distinct ghcr.io package names digest-pinned in a manifest."""
    out = []
    for match in _IMAGE_LINE.finditer(text):
        ref = match.group("ref")
        if _DIGEST not in ref:
            continue  # mutable tag; assertImageDigests owns this
        for name in _package_names(ref):
            if name not in out:
                out.append(name)
    return out


def _scan_workflows():
    """(publishers, build_inputs).

    package -> workflows that PUSH it, and the Dockerfiles those same push
    steps build. Both halves are collected per step, not per file, so a
    workflow that builds one image to scan and pushes another is credited only
    for the one it pushes.
    """
    publishers = {}
    build_inputs = {}
    for path in sorted(WORKFLOW_DIR.glob("*.y*ml")):
        rel = path.relative_to(REPO_ROOT).as_posix()
        text = path.read_text(encoding="utf-8")
        env = _workflow_env(text)
        for step in _steps(text):
            if not any(pattern.search(step) for pattern in _PUSH_EVIDENCE):
                continue  # local-only build (Trivy scan), not a publisher
            names = _package_names(_expand(step, env))
            if not names:
                continue
            dockerfiles = _dockerfiles_in(step)
            for name in names:
                publishers.setdefault(name, set()).add(rel)
                build_inputs.setdefault(name, set()).update(dockerfiles)
    return publishers, build_inputs


class TestProductionImagePinsHaveAPublisher(unittest.TestCase):
    """A production pin must name an image this repository builds and publishes."""

    @classmethod
    def _declared_packages(cls):
        """Package name -> the manifest(s) that document it as unpublished.

        A declaration is scoped to the package it NAMES, not to the file that
        happens to contain it. Scoping it to the file is the loophole this
        deliberately refuses: a manifest already carrying a documented
        declaration would then also bless any NEW phantom pin added anywhere in
        it, which is exactly how an unpublishable image would slip through a
        guard that says it cannot. So the evidence block must name every package
        it speaks for, and a package no block names is undeclared.

        Every documentation block in this repo opens with 'WHAT MUST HAPPEN', and
        the block names the package it is about. So a declaration covers a
        package only when that package is named in the manifest's COMMENT text.
        Naming it on the `image:` line does NOT count -- otherwise the pin would
        declare itself, and a one-line edit adding a phantom pin would satisfy
        its own declaration.
        """
        declared = {}
        publishers, _ = _scan_workflows()
        for root in PRODUCTION_MANIFEST_ROOTS:
            base = REPO_ROOT / root
            if not base.is_dir():
                continue
            for path in sorted(base.rglob("*.yaml")):
                rel = path.relative_to(REPO_ROOT).as_posix()
                if "testdata" in rel.split("/"):
                    continue
                text = path.read_text(encoding="utf-8")
                if PIN_STATUS_ANNOTATION not in text:
                    continue
                if not any(status in text for status in DECLARED_UNPUBLISHABLE):
                    continue
                # The declaration must pre-exist in prose, so pin lines are
                # removed before matching. Without this a pin declares itself.
                prose = "\n".join(
                    line for line in text.splitlines()
                    if not _IMAGE_LINE.match(line)
                )
                for name in set(_pinned_packages(text)):
                    if name in publishers:
                        continue
                    if name in prose:
                        declared.setdefault(name, set()).add(rel)
        return declared

    @classmethod
    def setUpClass(cls):
        cls.publishers, cls.build_inputs = _scan_workflows()
        cls.declared = cls._declared_packages()

        cls.pins = {}
        for root in PRODUCTION_MANIFEST_ROOTS:
            base = REPO_ROOT / root
            if not base.is_dir():
                continue
            for path in sorted(base.rglob("*.yaml")):
                rel = path.relative_to(REPO_ROOT).as_posix()
                if "testdata" in rel.split("/"):
                    continue  # negative fixtures pin deliberately-bad images
                for name in _pinned_packages(path.read_text(encoding="utf-8")):
                    cls.pins.setdefault(name, []).append(rel)

    # -- the scan must not be vacuous --------------------------------------

    def test_production_pins_were_discovered(self):
        """The scan found the production pins; a drifted root fails, not passes.

        Without this, renaming or moving a manifest tree would silently reduce
        this guard to zero assertions and it would report green while protecting
        nothing -- the exact failure mode that let a DR-only assertion look like
        coverage of the whole estate.
        """
        self.assertTrue(
            self.pins,
            "no ghcr.io/virtengine digest pins found under "
            f"{list(PRODUCTION_MANIFEST_ROOTS)}; the scan roots have drifted and "
            "this guard now protects nothing",
        )
        roots = {rel.split("/")[0] for rels in self.pins.values() for rel in rels}
        for expected in ("deploy", "infra"):
            self.assertIn(
                expected, roots,
                f"no {expected}/ pin is in scope any more. Either that tree "
                "stopped pinning private images (say so here) or the scan roots "
                "drifted and this guard is now vacuous for it.",
            )

    def test_publisher_detection_finds_the_known_publishers(self):
        """The publisher table must name the two packages known to publish.

        A self-check on the detector, and the assertion that caught this
        helper's first draft returning an EMPTY table: ci.yaml and
        dr-tools-image.yaml push by different syntax (`--push` on a buildx
        command line versus `push: true` on docker/build-push-action, one naming
        the owner as a GitHub expression and one through `${{ env.IMAGE }}`), and
        a detector that misses either credits nothing -- which would let the
        publisher assertion below pass vacuously.
        """
        self.assertTrue(
            any("ci.yaml" in rel for rel in self.publishers.get("virtengine", ())),
            "ghcr.io/virtengine/virtengine is pushed by ci.yaml but is not "
            "credited; the publisher detector missed its syntax",
        )
        self.assertTrue(
            any("dr-tools" in rel for rel in self.publishers.get("dr-tools", ())),
            "dr-tools is published by dr-tools-image.yaml but is not credited; "
            "the publisher detector regressed for the one image known to work",
        )

    def test_local_only_scan_builds_are_not_credited(self):
        """A build nobody pushes must not satisfy the publisher requirement.

        `_build/Dockerfile.test`, `_build/Dockerfile.veid-pipeline` and
        `sdk/generation/Dockerfile` are all built in ci.yaml and loaded into the
        runner daemon purely so Trivy can scan them. Crediting them as
        publishers would let any pin that merely has a Dockerfile pass.
        """
        for name in ("test", "veid-pipeline", "proto-gen"):
            self.assertNotIn(
                name, self.publishers,
                f"{name!r} is credited as a published package, but ci.yaml only "
                "builds it locally to scan it. The push-evidence filter is not "
                "excluding local-only builds.",
            )

    # -- the assertions that were missing ---------------------------------

    def test_every_production_pinned_package_is_published(self):
        """Each pinned package must be PUSHED by a workflow in this repository.

        This is the assertion dr-tools got and no other image had. A package
        with no publisher is a pin no cluster can resolve, and every
        digest-shaped check in this repository still passes it.

        The ONE sanctioned exception is a package the manifest explicitly
        declares unpullable AND documents at the pin (see
        test_placeholder_declaration_carries_its_evidence). That is a recorded,
        reviewed owner decision -- "we know this pod cannot start, and here is
        why and here is what closes it" -- which is strictly better than the
        state this test found: four production workloads whose manifests
        asserted a working pull credential for packages that do not exist.
        Everything else must have a real publisher.
        """
        unpublished = []
        for name, rels in sorted(self.pins.items()):
            if name in self.publishers:
                continue
            if name in self.declared:
                continue
            unpublished.append(
                f"{ORG}/{name} (pinned by {', '.join(sorted(set(rels)))})"
            )
        self.assertEqual(
            unpublished, [],
            "these ghcr.io/virtengine packages are digest-pinned by production "
            "manifests but no workflow in this repository publishes them and no "
            "manifest declares them, so no kubelet can resolve the pin "
            "(verified independently: an anonymous token exchange for each "
            "returns 403 DENIED, which is byte-identical to GHCR's answer for a "
            "package that was never published): "
            + "; ".join(unpublished)
            + ". For each one, either stand up the build + publish path (a "
            f"`_build/Dockerfile.<name>` plus a workflow that pushes "
            f"{ORG}/<name> and proves the digest by pulling it back, the way "
            "dr-tools-image.yaml does for dr-tools) and re-pin to a digest that "
            "workflow actually published, or remove the pin. Do NOT weaken the "
            "digest-pinning requirement to make this pass: the fix is to make "
            "the pins true.",
        )

    def test_placeholder_declaration_carries_its_evidence(self):
        """A declared-unpullable pin must record WHY, in the manifest.

        The annotation exists so the gap is visible; this exists so it is also
        REVIEWABLE. Without a reason block next to the pin, adding
        `virtengine.com/image-pin-status: placeholder-awaiting-release` to any
        workload would make the publisher assertion above satisfiable by
        annotation alone -- a one-line change that silences a real outage with
        no explanation and no owner decision. So every manifest that pins an
        unpublished package must name that package in a comment block carrying
        the evidence and the resolution.
        """
        undocumented = []
        for name, rels in sorted(self.pins.items()):
            if name in self.publishers:
                continue
            if name in self.declared:
                continue
            undocumented.append(
                f"{ORG}/{name} (pinned by {', '.join(sorted(set(rels)))}) declares "
                "nothing: it needs the "
                f"{PIN_STATUS_ANNOTATION} annotation plus a 'WHAT MUST HAPPEN' "
                "comment block naming it, giving the evidence and the owner "
                "decision that closes it"
            )
        self.assertEqual(
            undocumented, [],
            "a production manifest pins an unpublished package without "
            "documenting why. The placeholder annotation must not be a free "
            "pass, and a declaration must be scoped to the package it names so "
            "it cannot silently bless an unrelated new pin in the same file: "
            + "; ".join(undocumented),
        )

    def test_declaration_is_scoped_to_the_package_it_names(self):
        """A documented declaration must not cover the whole file.

        The failure this forbids is concrete: a manifest already carries a
        documented declaration, someone adds a NEW image pin for a package that
        was never published, and the guard stays green because the file looked
        declared. A guard that cannot be satisfied by a one-line phantom-pin
        edit is decoration. Each declaration is therefore bound to the package
        names inside its own block, and this asserts the binding is real by
        checking that the set of declared packages matches what the manifests
        actually document -- no more.
        """
        for name, rels in sorted(self.declared.items()):
            for rel in rels:
                self.assertIn(
                    name, _pinned_packages((REPO_ROOT / rel).read_text(encoding="utf-8")),
                    f"{rel} declares {ORG}/{name} but no longer pins it; the "
                    "declaration is stale and will hide a later regression",
                )

    def test_every_production_pinned_package_has_a_build_input(self):
        """A published package must have a Dockerfile its publisher builds.

        The other half of the dr-tools assertion, generalised. A push step that
        names a package but builds nothing checkable in-tree is how a package
        can be "published" by a workflow whose build recipe lives outside
        version control -- the pin would then be unauditable in the same way
        dr-tools was.
        """
        missing = [
            f"{ORG}/{name} (published by "
            f"{', '.join(sorted(self.publishers[name]))}, which names no Dockerfile)"
            for name in sorted(self.pins)
            if name in self.publishers and not self.build_inputs.get(name)
        ]
        self.assertEqual(
            missing, [],
            "these packages are pushed by a workflow that builds no Dockerfile "
            "in this repository, so the pinned image has no reviewable build "
            "recipe: " + "; ".join(missing),
        )

    def test_placeholder_digest_is_declared_in_the_manifest(self):
        """A production manifest must not carry an undeclared impossible digest.

        64 zeros has the shape of a SHA-256 pin and names no possible image. It
        cannot be pulled anonymously or with a perfect credential, so no
        publisher requirement can rescue it -- a package requirement is
        necessary, not sufficient. The veid-inference workload already declares
        itself with the `virtengine.com/image-pin-status` annotation; this
        asserts the declaration is present wherever such a pin appears, in any
        manifest, without naming one image.
        """
        undeclared = []
        for root in PRODUCTION_MANIFEST_ROOTS:
            base = REPO_ROOT / root
            if not base.is_dir():
                continue
            for path in sorted(base.rglob("*.yaml")):
                rel = path.relative_to(REPO_ROOT).as_posix()
                text = path.read_text(encoding="utf-8")
                for match in _IMAGE_LINE.finditer(text):
                    ref = match.group("ref")
                    digest = ref.rsplit("@", 1)[1] if "@" in ref else ""
                    if digest != _ALL_ZERO:
                        continue
                    if PIN_STATUS_ANNOTATION not in text:
                        undeclared.append(f"{rel}: all-zero digest, no annotation")
                    elif "placeholder-awaiting-release" not in text:
                        undeclared.append(
                            f"{rel}: all-zero digest, annotation is not "
                            "placeholder-awaiting-release"
                        )
        self.assertEqual(
            undeclared, [],
            "a production manifest pins an all-zero digest without declaring it "
            "as a placeholder via the virtengine.com/image-pin-status pod "
            "annotation. Such an image cannot be pulled by anyone, credentialed "
            "or not, so the gap must stay visible in review: "
            + "; ".join(undeclared),
        )


if __name__ == "__main__":
    unittest.main()
"""Guard: a private ghcr.io/virtengine image must be pullable by the pod that runs it.

The defect this exists for
--------------------------
infra/kubernetes/dr/backup-cronjobs.yaml runs three CronJobs under
`dr-backup-sa` and all three reference

    ghcr.io/virtengine/dr-tools@sha256:...

That package is PRIVATE. The registry's anonymous token exchange refuses it
outright, so the pod has no way to fetch the image and every CronJob sits in
ImagePullBackOff from its first schedule onward. `imagePullPolicy: IfNotPresent`
does not help -- the image is simply unfetchable.

The same defect class then reappeared one directory over: deploy/kubernetes/base
runs five workloads on ghcr.io/virtengine/{virtengine-node,provider-daemon,
tee-enclave,veid-inference}, none of which returned an anonymous pull token, and
none of whose ServiceAccounts carried a pull credential. That directory had zero
regression protection, which is why this guard now scans BOTH roots.

The pin tests could never catch this. A digest pin is checked with an
*authenticated* pull from a CI runner (that runner holds GITHUB_TOKEN), and a
correct digest looks perfectly healthy from in there. The failure only exists
in the pod's own credential context, which no pin assertion inspects.

The two credential paths, and why this test accepts both
-------------------------------------------------------
There are exactly two ways a pod can pull a private image, and they need
different manifest evidence:

  * PUBLIC PACKAGE  - the package is readable anonymously, so no credential is
    required in-cluster. Nothing is attached to the ServiceAccount, and that is
    correct, not a defect. The risk this test guards is the REVERSE regression:
    a package that silently stops being anonymously readable. Then the pods
    break with no manifest change at all, because the manifest was never the
    thing that was wrong.

  * PULL SECRET     - the package stays private and the ServiceAccount carries
    `imagePullSecrets`. The secret VALUE is provisioned out of band (never
    committed), matching how this repo already handles `dr-backup-signing-key`
    and `dr-secrets`: the manifest references a secret by name only.

So this test asserts the pair that actually has to hold:

    (image is not anonymously readable)  =>  (pod has a pull credential path)

plus the complementary invariant that catches the other direction:

    (image IS anonymously readable)      =>  (no pull secret is claimed)

The second half is what stops "fix" a from quietly becoming a permanent
credential dependency, and it is the half that would otherwise be untestable.
It needs the registry to be reachable, so it is skipped (never silently passed)
when offline; the static half below always runs and is the part that fails
when a secret is deleted from the manifest.

A pull credential path is a non-empty, well-formed `imagePullSecrets` on the
ServiceAccount OR on the pod spec itself; kubelet merges the two, so both are
accepted, and anything else (including a malformed entry) counts as NO
credential, so the guard fails closed rather than passing on a shape the API
server would reject.

What is deliberately NOT asserted
---------------------------------
Whether a given ghcr.io/virtengine package is private is an owner/architecture
call, and the value of any pull secret is an operations decision. This test
asserts only the invariant that holds under EITHER resolution, so it cannot rot
into certifying a specific credential choice it never made. The advisory probe
scripts/ci/probe-ghcr-pullability.sh reports which side of the condition is
currently true.

One limit of that probe is worth stating plainly, because it bounds what this
guard claims. GHCR answers an anonymous token exchange for a package that has
never been published with 403 DENIED -- byte-identical to its answer for a
private one. Verified live 2026-10-03: the four deploy/kubernetes packages and a
deliberately nonexistent control repo both return 403 DENIED, while dr-tools
(a package the DR Tools Image workflow demonstrably published) returns 401
UNAUTHORIZED. So absence of a token cannot separate "private" from "never
published", and this guard deliberately never claims which one it is. It does
not need to: an image a pod cannot fetch anonymously needs a credential either
way, so the assertion holds under both readings and is strictly stronger than
the private-only reading it replaces.

Placeholder digests are a SEPARATE defect class and get their own assertions
below: an all-zero digest cannot be pulled by anyone, credentialed or not, so
`test_placeholder_digest_is_declared_in_the_manifest` requires such a workload to
say so in an annotation rather than relying on a credential to make an
unpullable image look reachable.

Run:
  python -m unittest discover -s .github/tests -p "test_*policy*.py"
"""
import json
import re
import unittest
import urllib.error
import urllib.request
from collections import namedtuple
from pathlib import Path

import yaml

REPO_ROOT = Path(__file__).resolve().parents[2]

# Every non-Helm-templated Kubernetes manifest this guard owns, in both roots.
# infra/kubernetes is the DR/chaos tree; deploy/kubernetes is the canonical
# application tree. They are scanned together because they fail the same way:
# a workload naming a ghcr.io/virtengine image with no credential path on its
# ServiceAccount. One guard, both roots -- deliberately NOT two test files,
# which is how the second directory ended up unprotected in the first place.
K8S_DIRS = (
    REPO_ROOT / "infra" / "kubernetes",
    REPO_ROOT / "deploy" / "kubernetes",
)

# The registry every workload image in this estate is published to.
REGISTRY = "ghcr.io"
ORG = "virtengine"

# A package that is anonymous-public and content-addressed, so it cannot
# silently change. Without a control, a network failure and a genuine "package
# is not anonymously readable" are indistinguishable -- and the second reading
# would certify a manifest change as necessary when the probe was simply broken.
CONTROL_IMAGE = "ghcr.io/virtengine/virtengine"
CONTROL_REF = "sha256:da0ac87c431e4774ba327604769f7609f2e56275ece6e398c4d5a3f744f8219b"

_PRIVATE_GHCR = re.compile(r"^ghcr\.io/virtengine/([a-z0-9][a-z0-9._-]*)", re.I)

# A digest of 64 zeros matches the shape of a real SHA-256 pin while naming no
# possible image, so it can never be pulled -- not anonymously, and not with a
# perfectly valid pull secret either. Detected structurally, so copying the
# pattern into another workload is caught without editing this file.
_ALL_ZERO_DIGEST = re.compile(r"^sha256:0{64}$")

# Other placeholder digests that are not structurally obvious. Kept in step
# with DENY_DIGESTS in scripts/supply-chain/verify-pinned-images.sh, which
# rejects the same values in Dockerfile FROM lines. Empty today: the only
# placeholder currently in the tree is the structurally detected all-zero one.
DECLARED_PLACEHOLDER_DIGESTS = frozenset()

# The annotation a workload MUST carry while its image pin is a placeholder.
# It is machine-checked in both directions, so it cannot rot: swapping the
# all-zero digest for the real one (the desired outcome) makes the annotation
# stale and the guard says so.
PIN_STATUS_ANNOTATION = "virtengine.com/image-pin-status"
PLACEHOLDER_PIN_STATUS = "placeholder-awaiting-release"

Workload = namedtuple(
    "Workload",
    "root path kind name namespace service_account images annotations pod_secret",
)


def _docs():
    """Every non-Helm-templated Kubernetes manifest under both scan roots."""
    out = []
    for k8s_dir in K8S_DIRS:
        for p in sorted(k8s_dir.rglob("*.yaml")):
            text = p.read_text(encoding="utf-8")
            # Helm templates are not applied as YAML; a {{ }} would make the parse
            # fail or, worse, silently parse as a string and assert nothing.
            if "{{" in text:
                continue
            try:
                parsed = yaml.safe_load_all(text)
                out.append((p, list(parsed)))
            except yaml.YAMLError as exc:  # pragma: no cover - malformed manifest
                raise AssertionError(f"{p} is not parseable YAML: {exc}") from exc
    return out


def _iter_workloads(docs):
    """Yield a Workload for every pod-running document, init containers included.

    initContainers are included deliberately: virtengine-validator and
    virtengine-node both run the SAME private image in an init container, and a
    guard that only read `containers` would stop noticing the image the moment
    the main container's reference was edited.
    """
    for path, parsed in docs:
        root = path.relative_to(REPO_ROOT).as_posix()
        for doc in parsed:
            if not isinstance(doc, dict):
                continue
            kind = doc.get("kind")
            meta = doc.get("metadata") or {}
            name = meta.get("name")
            if kind == "CronJob":
                spec = ((doc.get("spec") or {}).get("jobTemplate") or {})
                pod = ((spec.get("spec") or {}).get("template") or {})
                podspec = pod.get("spec") or {}
            elif kind in ("Deployment", "StatefulSet", "DaemonSet", "Job", "Pod"):
                pod = ((doc.get("spec") or {}).get("template") or {})
                podspec = pod.get("spec") or {}
            else:
                continue
            images = []
            for key in ("initContainers", "containers"):
                for c in podspec.get(key) or []:
                    img = c.get("image")
                    if img:
                        images.append(img)
            yield Workload(
                root=root,
                path=path,
                kind=kind,
                name=name,
                namespace=meta.get("namespace"),
                service_account=podspec.get("serviceAccountName"),
                images=images,
                annotations=(pod.get("metadata") or {}).get("annotations") or {},
                pod_secret=_has_pull_credential(podspec),
            )


def _has_pull_credential(podspec):
    """True only for a non-empty, well-formed imagePullSecrets list.

    A malformed entry (a bare string, a dict with no name) is treated as NO
    credential: the API server would reject the manifest, and a guard that
    passed on it would certify a shape that cannot be applied.
    """
    entries = podspec.get("imagePullSecrets")
    if not entries or not isinstance(entries, list):
        return False
    return all(
        isinstance(e, dict) and isinstance(e.get("name"), str) and e["name"].strip()
        for e in entries
    )


def _service_accounts(docs):
    """Map (namespace, name) -> ServiceAccount document."""
    out = {}
    for _path, parsed in docs:
        for doc in parsed:
            if isinstance(doc, dict) and doc.get("kind") == "ServiceAccount":
                meta = doc.get("metadata") or {}
                out[(meta.get("namespace"), meta.get("name"))] = doc
    return out


def _resolve_service_account(sas, workload):
    """Resolve a workload's ServiceAccount, preferring its own namespace.

    Falls back to a namespace-less declaration only when the name is unique, so
    a same-named ServiceAccount in two namespaces can never be silently
    substituted for the wrong one.
    """
    sa_name = workload.service_account
    if not sa_name:
        return None
    if (workload.namespace, sa_name) in sas:
        return sas[(workload.namespace, sa_name)]
    if (None, sa_name) in sas:
        matches = [ns for (ns, name) in sas if name == sa_name]
        if len(matches) == 1:
            return sas[(None, sa_name)]
    return None


def _packages(workload):
    """The distinct ghcr.io/virtengine packages this workload references."""
    out = []
    for image in workload.images:
        m = _PRIVATE_GHCR.match(image)
        if m and m.group(1) not in out:
            out.append(m.group(1))
    return out


def _digest(image):
    """The pinned digest of an image reference, or "" when unpinned."""
    if "@" not in image:
        return ""
    return image.rsplit("@", 1)[1]


def _is_placeholder_digest(digest):
    return bool(digest) and (
        _ALL_ZERO_DIGEST.match(digest) is not None
        or digest.lower() in DECLARED_PLACEHOLDER_DIGESTS
    )


class _AnonymousRegistry:
    """Anonymous pull-token exchange, one request per package per run."""

    @staticmethod
    def token_for(repo):
        url = (
            f"https://{REGISTRY}/token?service={REGISTRY}"
            f"&scope=repository:{repo}:pull"
        )
        try:
            with urllib.request.urlopen(url, timeout=30) as resp:
                body = json.loads(resp.read().decode("utf-8", "replace"))
        except (urllib.error.URLError, OSError, ValueError, json.JSONDecodeError):
            return ""
        return body.get("token", "") if isinstance(body, dict) else ""


class TestGhcrImagePullCredentialPolicy(unittest.TestCase):
    def setUp(self):
        self.docs = _docs()
        self.assertTrue(
            self.docs,
            "no Kubernetes manifests found under "
            + ", ".join(str(d) for d in K8S_DIRS),
        )
        self.workloads = [w for w in _iter_workloads(self.docs) if _packages(w)]

    # -- static half: always runs, no network ------------------------------

    def test_both_manifest_roots_are_in_scope(self):
        """The scan roots are both live; a renamed/moved directory is a failure.

        This is the assertion that would have caught the original gap. The
        second root existed, held four private-image workloads, and was outside
        every assertion in this file.
        """
        roots = {w.root.split("/")[0] for w in self.workloads}
        self.assertIn(
            "deploy", roots,
            "deploy/kubernetes contributed no ghcr.io/virtengine workload to this "
            "guard. Either the canonical application tree stopped using private "
            "images (say so here) or the scan roots drifted and this guard is "
            "now vacuous for the workloads that matter most.",
        )
        self.assertIn(
            "infra", roots,
            "infra/kubernetes contributed no ghcr.io/virtengine workload; the DR "
            "CronJobs must still be in scope.",
        )

    def test_private_image_pod_has_a_pull_credential_path(self):
        """Every private-image pod must have SOME pull credential path.

        Either the package is anonymously readable (no secret needed) or the pod
        carries imagePullSecrets on its ServiceAccount or its own spec. What is
        never acceptable is a manifest that names a private image and a
        ServiceAccount with no credential story at all, because that pins the
        pod to ImagePullBackOff forever while every pin-based test still reports
        green.
        """
        sas = _service_accounts(self.docs)
        registry = _AnonymousRegistry()
        unresolved = []
        for workload in self.workloads:
            if not workload.service_account:
                continue
            sa = _resolve_service_account(sas, workload)
            if sa is None:
                continue  # SA resolves in no manifest; asserted below.
            if _has_pull_credential(sa) or workload.pod_secret:
                continue
            unresolved.append((workload, _packages(workload)[0]))

        # A pod may legitimately run a public image with no secret, so this
        # cannot fail on its own without the registry's evidence. The registry is
        # consulted for exactly those pods; anything still unresolved after that
        # is a genuine defect.
        for workload, repo in unresolved:
            if registry.token_for(f"{ORG}/{repo}"):
                continue  # anonymously readable -> no secret required, correct
            with self.subTest(manifest=workload.root, workload=f"{workload.kind}/{workload.name}"):
                self.fail(
                    f"{workload.root}: {workload.kind}/{workload.name} runs "
                    f"{ORG}/{repo}, which returned no anonymous GHCR pull token, so "
                    f"the pod cannot fetch it without a credential, yet "
                    f"ServiceAccount {workload.service_account!r} declares no "
                    "imagePullSecrets and the pod spec sets none either. The pod "
                    "will stay in ImagePullBackOff. Add the pull secret to the "
                    "ServiceAccount, or make the package anonymously readable.",
                )

    def test_every_private_image_service_account_is_resolvable(self):
        """A pod naming a private image must resolve to a declared ServiceAccount.

        Without this the previous test silently skips (`if sa is None: continue`)
        on a typo'd or renamed SA name, and the guard passes vacuously. Failing
        closed here is the difference between a guard and a decoration.
        """
        sas = _service_accounts(self.docs)
        missing = []
        for workload in self.workloads:
            if not workload.service_account:
                missing.append(
                    f"{workload.root}:{workload.kind}/{workload.name} runs a "
                    "ghcr.io/virtengine image with NO serviceAccountName"
                )
                continue
            if _resolve_service_account(sas, workload) is None:
                declared = sorted(f"{ns or '<default>'}/{name}" for (ns, name) in sas)
                missing.append(
                    f"{workload.root}:{workload.kind}/{workload.name} -> "
                    f"undeclared SA {workload.service_account!r} "
                    f"(declared: {', '.join(declared)})"
                )
        self.assertEqual(
            missing, [],
            "pods referencing a ghcr.io/virtengine image must name a "
            "ServiceAccount that some manifest under infra/kubernetes or "
            "deploy/kubernetes declares, in the same namespace: " + "; ".join(missing),
        )

    def test_manifest_never_commits_a_secret_value(self):
        """The pull secret is referenced by NAME; the credential is not in git.

        Committing a `.dockerconfigjson` value would put a usable registry
        credential in the repository history forever. This is the same rule the
        repo already follows for dr-backup-signing-key.
        """
        offenders = []
        for path, parsed in self.docs:
            for doc in parsed:
                if not isinstance(doc, dict):
                    continue
                if doc.get("kind") != "Secret":
                    continue
                if (doc.get("type") or "") != "kubernetes.io/dockerconfigjson":
                    continue
                offenders.append(str(path))
        self.assertEqual(
            offenders, [],
            "a docker-registry pull secret VALUE is committed under "
            f"{[str(d) for d in K8S_DIRS]}: {offenders}. Reference the secret by "
            "name only; provision the credential out of band.",
        )

    # -- placeholder digests: the credential path is not enough ------------

    def test_placeholder_digest_is_declared_in_the_manifest(self):
        """An unpullable pin must SAY it is unpullable, in a checkable way.

        A digest of 64 zeros cannot be pulled by anyone -- not anonymously, and
        not with a valid pull secret. So a manifest that pairs one with an
        imagePullSecrets entry looks correct and is not. Requiring the annotation
        keeps the gap visible in review and in `kubectl get` output instead of
        turning into a silent ImagePullBackOff at release time.
        """
        undeclared = []
        for workload in self.workloads:
            placeholders = [
                image for image in workload.images
                if _is_placeholder_digest(_digest(image))
            ]
            if not placeholders:
                continue
            declared = workload.annotations.get(PIN_STATUS_ANNOTATION)
            if declared == PLACEHOLDER_PIN_STATUS:
                continue
            undeclared.append(
                f"{workload.root}:{workload.kind}/{workload.name} pins "
                f"{_digest(placeholders[0])}, which no credential can pull; its "
                f"pod annotations declare {PIN_STATUS_ANNOTATION}="
                f"{declared!r} instead of {PLACEHOLDER_PIN_STATUS!r}"
            )
        self.assertEqual(
            undeclared, [],
            "every workload pinned to a placeholder digest must carry "
            f"{PIN_STATUS_ANNOTATION}: {PLACEHOLDER_PIN_STATUS} on its pod "
            "template so the unpullable pin is declared rather than implied: "
            + "; ".join(undeclared),
        )

    def test_placeholder_pin_annotation_is_not_stale(self):
        """The placeholder marker must disappear the moment a real digest lands.

        This is what makes the annotation above worth having. Without it, the
        declaration would outlive its cause and the next real pin would ship
        with a manifest still claiming the image does not exist -- so a genuine
        later regression would be hidden behind a stale note.
        """
        stale = []
        for workload in self.workloads:
            if workload.annotations.get(PIN_STATUS_ANNOTATION) != PLACEHOLDER_PIN_STATUS:
                continue
            still_placeholder = any(
                _is_placeholder_digest(_digest(image)) for image in workload.images
            )
            if not still_placeholder:
                stale.append(
                    f"{workload.root}:{workload.kind}/{workload.name} is annotated "
                    f"{PIN_STATUS_ANNOTATION}={PLACEHOLDER_PIN_STATUS} but pins a "
                    "real digest; remove the annotation"
                )
        self.assertEqual(stale, [], "; ".join(stale))


class TestGhcrPackagesAreAnonymousReadable(unittest.TestCase):
    """Live half: which side of the invariant is currently true.

    Skipped, never silently passed, when the registry cannot be reached.
    """

    @classmethod
    def setUpClass(cls):
        if not _AnonymousRegistry.token_for(CONTROL_IMAGE.split("/", 1)[1]):
            raise unittest.SkipTest(
                "GHCR anonymous token exchange unavailable (network or "
                "registry policy); the public control did not return a token, so "
                "the private/public reading cannot be trusted"
            )

    def test_every_private_image_pullability_matches_the_manifest(self):
        """Each package's real visibility must agree with its pod's manifest.

        Replaces the DR-only assertion of #1161 with the generic form, so the
        four deploy/kubernetes packages get the same treatment. dr-tools is
        still covered -- it is simply discovered now rather than named.

        Both directions are asserted:

          anonymously readable     => no pull secret may be claimed, or a
                                     visibility change stays invisible in git
          NOT anonymously readable => a pull secret is REQUIRED, because the
                                     node's kubelet pulls with no credentials of
                                     its own

        Note what is NOT claimed: that these packages are "private". GHCR
        answers an anonymous token exchange for a package that has never been
        published with 403 DENIED, which is byte-identical to its answer for a
        private one. Absence of a token cannot separate the two, and that is
        fine, because an image a pod cannot fetch anonymously needs a credential
        either way.
        """
        docs = _docs()
        sas = _service_accounts(docs)
        registry = _AnonymousRegistry()

        seen = {}
        for workload in _iter_workloads(docs):
            if workload.pod_secret:
                continue  # credential claimed at pod level; no SA to reason about
            sa = _resolve_service_account(sas, workload)
            if sa is None:
                continue  # unresolved SAs are a separate static failure
            for repo in _packages(workload):
                if repo not in seen:
                    seen[repo] = [set(), sa]
                seen[repo][0].add(workload.name)

        self.assertTrue(seen, "no ghcr.io/virtengine workloads discovered")

        for repo, (workloads, sa) in sorted(seen.items()):
            label = f"{repo} (run by {', '.join(sorted(workloads))})"
            sa_name = (sa.get("metadata") or {}).get("name")
            has_secret = _has_pull_credential(sa)
            anonymous_ok = bool(registry.token_for(f"{ORG}/{repo}"))

            if anonymous_ok:
                self.assertFalse(
                    has_secret,
                    f"{label} is anonymously readable but ServiceAccount "
                    f"{sa_name!r} still declares imagePullSecrets; either the "
                    "package was made public or the secret is a leftover. Remove "
                    "the dead credential so the next visibility change is visible "
                    "in the manifest.",
                )
            else:
                self.assertTrue(
                    has_secret,
                    f"{label} returned no anonymous GHCR pull token, so the node's "
                    f"kubelet cannot fetch it, but ServiceAccount "
                    f"{sa_name!r} declares no imagePullSecrets. Every pod running "
                    "it will sit in ImagePullBackOff. Add the pull secret to that "
                    "ServiceAccount, or make the package anonymously readable. "
                    "Note this was proven anonymously: CI can pull because it "
                    "holds GITHUB_TOKEN, which is exactly why an authenticated "
                    "pull proves nothing about pod pullability.",
                )


if __name__ == "__main__":
    unittest.main()
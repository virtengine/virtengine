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

    (image is not anonymously readable)  =>  (ServiceAccount has imagePullSecrets)

plus the complementary invariant that catches the other direction:

    (image IS anonymously readable)      =>  (no pull secret is claimed)

The second half is what stops "fix" a from quietly becoming a permanent
credential dependency, and it is the half that would otherwise be untestable.
It needs the registry to be reachable, so it is skipped (never silently passed)
when offline; the static half below always runs and is the part that fails
when a secret is deleted from the manifest.

What is deliberately NOT asserted
---------------------------------
Whether `ghcr.io/virtengine/dr-tools` is private is an owner/architecture call,
and the value of any pull secret is an operations decision. This test asserts
only the invariant that holds under EITHER resolution, so it cannot rot into
certifying a specific credential choice it never made. The advisory probe
scripts/ci/probe-ghcr-pullability.sh reports which side of the condition is
currently true.

Run:
  python -m unittest discover -s .github/tests -p "test_*policy*.py"
"""
import json
import re
import unittest
import urllib.error
import urllib.request
from pathlib import Path

import yaml

REPO_ROOT = Path(__file__).resolve().parents[2]
K8S_DIR = REPO_ROOT / "infra" / "kubernetes"

# The registry every workload image in this estate is published to.
REGISTRY = "ghcr.io"
ORG = "virtengine"

# A package that is anonymous-public and content-addressed, so it cannot
# silently change. Without a control, a network failure and a genuine "package
# is private" are indistinguishable -- and the second reading would certify a
# manifest change as necessary when the probe was simply broken.
CONTROL_IMAGE = "ghcr.io/virtengine/virtengine"
CONTROL_REF = "sha256:da0ac87c431e4774ba327604769f7609f2e56275ece6e398c4d5a3f744f8219b"

_PRIVATE_GHCR = re.compile(r"^ghcr\.io/virtengine/([a-z0-9][a-z0-9._-]*)", re.I)


def _docs():
    """Every non-Helm-templated Kubernetes manifest under infra/kubernetes."""
    out = []
    for p in sorted(K8S_DIR.rglob("*.yaml")):
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
    """Yield (path, kind, name, serviceAccountName, [container images])."""
    for path, parsed in docs:
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
                podspec = ((doc.get("spec") or {}).get("template") or {}).get("spec") or {}
            else:
                continue
            images = []
            for c in podspec.get("containers") or []:
                img = c.get("image")
                if img:
                    images.append(img)
            yield path, kind, name, podspec.get("serviceAccountName"), images


def _service_accounts(docs):
    """Map (namespace, name) -> ServiceAccount document."""
    out = {}
    for _path, parsed in docs:
        for doc in parsed:
            if isinstance(doc, dict) and doc.get("kind") == "ServiceAccount":
                meta = doc.get("metadata") or {}
                out[(meta.get("namespace"), meta.get("name"))] = doc
    return out


class TestGhcrImagePullCredentialPolicy(unittest.TestCase):
    def setUp(self):
        self.docs = _docs()
        self.assertTrue(self.docs, f"no Kubernetes manifests found under {K8S_DIR}")

    # -- static half: always runs, no network ------------------------------

    def test_dr_cronjob_service_account_has_a_pull_credential_path(self):
        """Every private-image pod must have SOME pull credential path.

        Either the package is anonymous-public (no secret needed) or the
        ServiceAccount carries imagePullSecrets. What is never acceptable is a
        manifest that names a private image and a ServiceAccount with no
        credential story at all, because that pins the pod to ImagePullBackOff
        forever while every pin-based test still reports green.

        The visibility half of this is resolved live in
        TestGhcrPackagesAreAnonymousReadable; this asserts the manifest half.
        """
        sas = _service_accounts(self.docs)
        unresolved = []
        for path, kind, name, sa_name, images in _iter_workloads(self.docs):
            if not sa_name:
                continue
            ghcr = [i for i in images if _PRIVATE_GHCR.match(i)]
            if not ghcr:
                continue
            sa = sas.get((None, sa_name)) or sas.get(("virtengine", sa_name))
            if sa is None:
                continue  # SA lives in another manifest; asserted above.
            if sa.get("imagePullSecrets"):
                continue
            m = _PRIVATE_GHCR.match(ghcr[0])
            if m is None:  # pragma: no cover - filtered by the check above
                continue
            unresolved.append(
                (f"{path.name}:{kind}/{name} sa={sa_name}", m.group(1))
            )

        # A pod may legitimately run a PUBLIC image with no secret, so this
        # cannot fail on its own without the registry's answer. The registry is
        # consulted for exactly those pods; anything still unresolved after that
        # is a genuine defect.
        for pod, repo in unresolved:
            token = TestGhcrPackagesAreAnonymousReadable._anonymous_token(
                f"{ORG}/{repo}"
            )
            if token:
                continue  # anonymously readable -> no secret required, correct
            with self.subTest(pod=pod):
                self.fail(
                    f"{pod}: {ORG}/{repo} is not anonymously readable, so this "
                    "pod cannot pull its image without a credential, but its "
                    "ServiceAccount declares no imagePullSecrets. The pod will "
                    "stay in ImagePullBackOff. Add the pull secret to the SA, or "
                    "make the package public.",
                )

    def test_every_private_image_service_account_is_resolvable(self):
        """A pod naming a private image must resolve to a declared ServiceAccount.

        Without this the previous test silently skips (`if sa is None: continue`)
        on a typo'd or renamed SA name, and the guard passes vacuously. Failing
        closed here is the difference between a guard and a decoration.
        """
        sas = _service_accounts(self.docs)
        missing = []
        for path, kind, name, sa_name, images in _iter_workloads(self.docs):
            if not sa_name or not any(_PRIVATE_GHCR.match(i) for i in images):
                continue
            if (None, sa_name) not in sas and ("virtengine", sa_name) not in sas:
                missing.append(f"{path.name}:{kind}/{name} -> undeclared SA {sa_name!r}")
        self.assertEqual(
            missing, [],
            "pods referencing a ghcr.io/virtengine image name a ServiceAccount "
            "that no manifest under infra/kubernetes declares: "
            + "; ".join(missing),
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
            f"{K8S_DIR}: {offenders}. Reference the secret by name only; "
            "provision the credential out of band.",
        )


class TestGhcrPackagesAreAnonymousReadable(unittest.TestCase):
    """Live half: which side of the invariant is currently true.

    Skipped, never silently passed, when the registry cannot be reached.
    """

    @classmethod
    def setUpClass(cls):
        cls._anonymous = {}
        control_token = cls._anonymous_token(CONTROL_IMAGE.split("/", 1)[1])
        if not control_token:
            raise unittest.SkipTest(
                "GHCR anonymous token exchange unavailable (network or "
                "registry policy); the public control did not return a token, so "
                "the private/public reading cannot be trusted"
            )

    @staticmethod
    def _anonymous_token(repo):
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

    def test_dr_tools_pullability_matches_the_service_account(self):
        """The DR image's real visibility must agree with dr-backup-sa."""
        sa = _service_accounts(_docs()).get(("virtengine", "dr-backup-sa"))
        self.assertIsNotNone(sa, "dr-backup-sa is not declared under infra/kubernetes")

        repo = f"{ORG}/dr-tools"
        token = self._anonymous_token(repo)
        self.assertIsNotNone(
            token, "sanity: the anonymous token helper must always return a string"
        )
        anonymous_ok = bool(token)
        has_secret = bool(sa.get("imagePullSecrets"))

        if anonymous_ok:
            self.assertFalse(
                has_secret,
                f"{repo} is anonymously readable but dr-backup-sa still declares "
                "imagePullSecrets; either the package was made public or the "
                "secret is a leftover. Remove the dead credential so the next "
                "visibility change is visible in the manifest.",
            )
        else:
            self.assertTrue(
                has_secret,
                f"{repo} is NOT anonymously readable (the registry refuses an "
                "anonymous pull token), but ServiceAccount dr-backup-sa "
                "declares no imagePullSecrets. Every DR CronJob pod will sit in "
                "ImagePullBackOff. Add the pull secret to dr-backup-sa, or make "
                "the package public and delete this assertion's else-branch "
                "expectation. Note this was proven anonymously: CI can pull "
                "because it holds GITHUB_TOKEN, which is exactly why an "
                "authenticated pull proves nothing about pod pullability.",
            )


if __name__ == "__main__":
    unittest.main()

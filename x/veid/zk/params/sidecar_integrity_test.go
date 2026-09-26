package params

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var sidecarDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type sidecarRecord struct {
	digest string
	label  string
}

// parseSidecar accepts both the text-mode (`<digest>  <name>`) and binary-mode
// (`<digest> *<name>`) digest record forms.
func parseSidecar(data []byte) (sidecarRecord, error) {
	fields := strings.Fields(string(data))
	if len(fields) != 2 {
		return sidecarRecord{}, fmt.Errorf("invalid checksum record %q", strings.TrimSpace(string(data)))
	}
	digest := fields[0]
	label := strings.TrimPrefix(fields[1], "*")
	if !sidecarDigestPattern.MatchString(digest) {
		return sidecarRecord{}, fmt.Errorf("digest %q is not a 64-character lowercase SHA-256 hex digest", digest)
	}
	if label == "" {
		return sidecarRecord{}, fmt.Errorf("checksum record %q has no artifact label", strings.TrimSpace(string(data)))
	}
	return sidecarRecord{digest: digest, label: label}, nil
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		_, goModErr := os.Stat(filepath.Join(dir, "go.mod"))
		_, gitErr := os.Stat(filepath.Join(dir, ".git"))
		if goModErr == nil && gitErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skipf("repository root not found above %s", dir)
		}
		dir = parent
	}
}

// TestEmbeddedParamsMetadataDigestMatchesSidecar pins the invariant that
// params_metadata.json.sha256 records the digest of the bytes go:embed actually
// embeds.
//
// Regression guard: the sidecar has twice recorded the digest of the metadata file
// *as checked out on a Windows working tree* (CRLF, 21261041...) instead of the
// committed blob (LF, a4e05914...). `.gitattributes` pins
// `/x/veid/zk/params/*.json text eol=lf` precisely because git stores, and go:embed
// embeds, the LF form. When the two disagree, loadArtifactSet fails digest
// verification for the embedded bundle and keeper.NewKeeper panics with
// "failed to initialize VEID ZK proof system" — which reddens the Go Tests job and
// every package that constructs a VEID keeper.
func TestEmbeddedParamsMetadataDigestMatchesSidecar(t *testing.T) {
	require.NotEmpty(t, metadataBytes, "embedded params_metadata.json is empty")
	require.NotContains(t, string(metadataBytes), "\r",
		"embedded params_metadata.json contains CR bytes; the eol=lf .gitattributes rule must not be removed")

	recorded, err := parseSidecar(metadataChecksum)
	require.NoError(t, err, "params_metadata.json.sha256 is malformed")
	require.Equal(t, metadataFileName, recorded.label,
		"params_metadata.json.sha256 must label the artifact it describes")
	require.Equal(t, hashBytes(metadataBytes), recorded.digest,
		"params_metadata.json.sha256 does not match the embedded params_metadata.json bytes; "+
			"regenerate it from the committed blob, not from a working-tree copy: "+
			"git cat-file -p HEAD:x/veid/zk/params/params_metadata.json | sha256sum")
}

// TestEmbeddedParamsArtifactsKeepLineEndingPolicy asserts the .gitattributes rules
// that keep the embedded artifacts byte-stable across platforms. Without them a
// Windows checkout rewrites the JSON to CRLF, the embedded digest stops matching the
// recorded one, and the failure only reproduces on Windows.
func TestEmbeddedParamsArtifactsKeepLineEndingPolicy(t *testing.T) {
	root := repositoryRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, ".gitattributes"))
	require.NoError(t, err)
	attributes := string(raw)

	for _, rule := range []string{
		"/x/veid/zk/params/*.json text eol=lf",
		"/x/veid/zk/params/*.sha256 text eol=lf",
	} {
		require.True(t, strings.Contains(attributes, rule),
			".gitattributes is missing %q; without it a Windows checkout rewrites these files to CRLF and the embedded digests stop matching", rule)
	}
}

// TestTrackedSidecarsMatchOnDiskArtifacts verifies every sha256 sidecar staged
// beside the embedded artifacts against the artifact bytes that ship with it. It
// fails both when a sidecar is stale and when the checkout rewrote line endings away
// from the bytes the binary embeds.
func TestTrackedSidecarsMatchOnDiskArtifacts(t *testing.T) {
	required := []string{
		"age_vk.bin",
		"residency_vk.bin",
		"score_vk.bin",
		metadataFileName,
	}

	entries, err := os.ReadDir(".")
	require.NoError(t, err)

	verified := make(map[string]struct{}, len(required))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".sha256") {
			continue
		}
		artifact := strings.TrimSuffix(name, ".sha256")

		data, err := os.ReadFile(artifact)
		require.NoErrorf(t, err, "%s names a missing artifact", name)

		recorded, err := parseSidecar(mustReadSidecar(t, name))
		require.NoErrorf(t, err, "%s is malformed", name)
		require.Equalf(t, artifact, recorded.label,
			"%s must label the artifact it describes", name)
		require.Equalf(t, hashBytes(data), recorded.digest,
			"%s records a digest that does not match %s on disk", name, artifact)

		verified[artifact] = struct{}{}
	}

	for _, artifact := range required {
		_, ok := verified[artifact]
		require.Truef(t, ok, "expected a %s.sha256 sidecar for %s", artifact, artifact)
	}
}

func mustReadSidecar(t *testing.T, name string) []byte {
	t.Helper()

	data, err := os.ReadFile(name)
	require.NoErrorf(t, err, "read %s", name)
	return data
}

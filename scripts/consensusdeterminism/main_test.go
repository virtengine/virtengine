package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanRepositoryFindsConsensusNondeterminism(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "x", "demo", "keeper", "msg_server.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	source := `package keeper
import (
	"crypto/rand"
	"net/http"
	"os"
	"time"
)
func bad() {
	_ = time.Now()
	_, _ = rand.Read(make([]byte, 8))
	_, _ = http.Get("https://validator-local.invalid")
	_, _ = os.ReadFile("/host/local")
	var values map[string]string
	for range values {}
	if float64(1) > 0.5 {}
}
`
	require.NoError(t, os.WriteFile(path, []byte(source), 0o600))

	findings, err := scanRepository(root, nil)
	require.NoError(t, err)
	rules := make(map[string]bool)
	for _, item := range findings {
		rules[item.Rule] = true
	}
	require.True(t, rules[ruleWallClock])
	require.True(t, rules[ruleRandomness])
	require.True(t, rules[ruleExternalNetwork])
	require.True(t, rules[ruleFilesystem])
	require.True(t, rules[ruleMapIteration])
	require.True(t, rules[ruleFloatingDecision])
}

func TestScanRepositoryRequiresExactAllowlist(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "app", "proposal_handler.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte(`package app
import "time"
func proposal() { _ = time.Now() }
`), 0o600))

	findings, err := scanRepository(root, []allowance{{
		Rule: ruleWallClock, Path: "app/proposal_handler.go", Function: "proposal", Reason: "test-only exact allowance",
	}})
	require.NoError(t, err)
	require.Len(t, findings, 1)
	require.True(t, findings[0].Allowed)
	require.Equal(t, "test-only exact allowance", findings[0].Reason)

	findings, err = scanRepository(root, []allowance{{
		Rule: ruleWallClock, Path: "app/proposal_handler.go", Function: "different", Reason: "must not match",
	}})
	require.NoError(t, err)
	require.Len(t, findings, 1)
	require.False(t, findings[0].Allowed)
}

func TestRepositoryHasNoUnapprovedConsensusFindings(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)

	findings, err := scanRepository(root, defaultAllowlist)
	require.NoError(t, err)
	var unapproved []finding
	for _, item := range findings {
		if !item.Allowed {
			unapproved = append(unapproved, item)
		}
	}
	require.Empty(t, unapproved)
}

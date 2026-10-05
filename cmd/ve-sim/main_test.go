package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/virtengine/virtengine/sim/core"
)

// The ve-sim JSON helpers are what an operator's `ve-sim suite` and
// `ve-sim check` runs depend on to hand results to the dashboard and to
// resume from a saved config. They had no tests at all (the package measured
// 0.0%), so a regression in the defaults below would only ever surface as a
// silently wrong simulation, never as a failed test.

// TestLoadConfigDefaults pins the two defaults loadConfig fills in. They are
// load-bearing: a config with no TimeStep or EndTime would otherwise simulate
// a zero-length window and report no results, with no error anywhere.
func TestLoadConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	path := filepath.Join(dir, "config.json")

	body, err := json.Marshal(map[string]any{
		"ScenarioName": "custom",
		"StartTime":    start,
		"Seed":         42,
		"NumUsers":     10,
	})
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	if cfg.ScenarioName != "custom" {
		t.Errorf("ScenarioName = %q, want %q", cfg.ScenarioName, "custom")
	}
	if cfg.Seed != 42 {
		t.Errorf("Seed = %d, want 42", cfg.Seed)
	}
	if cfg.NumUsers != 10 {
		t.Errorf("NumUsers = %d, want 10", cfg.NumUsers)
	}

	// TimeStep defaults to 24h.
	if cfg.TimeStep != 24*time.Hour {
		t.Errorf("TimeStep = %v, want %v", cfg.TimeStep, 24*time.Hour)
	}

	// EndTime defaults to StartTime + 365d.
	wantEnd := start.Add(365 * 24 * time.Hour)
	if !cfg.EndTime.Equal(wantEnd) {
		t.Errorf("EndTime = %v, want %v", cfg.EndTime, wantEnd)
	}

	// The window must be non-empty, or the simulation has nothing to do.
	if !cfg.EndTime.After(cfg.StartTime) {
		t.Errorf("defaulted window is empty: Start=%v End=%v", cfg.StartTime, cfg.EndTime)
	}
}

// TestLoadConfigPreservesExplicitValues pins that the defaults above only
// apply to ZERO fields. Overwriting a supplied EndTime or TimeStep with the
// defaults would silently run the wrong simulation.
func TestLoadConfigPreservesExplicitValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(48 * time.Hour)

	body, err := json.Marshal(map[string]any{
		"ScenarioName": "explicit",
		"StartTime":    start,
		"EndTime":      end,
		"TimeStep":     time.Hour,
	})
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	if cfg.TimeStep != time.Hour {
		t.Errorf("TimeStep = %v, want %v (an explicit value was overwritten by the default)", cfg.TimeStep, time.Hour)
	}
	if !cfg.EndTime.Equal(end) {
		t.Errorf("EndTime = %v, want %v (an explicit value was overwritten by the default)", cfg.EndTime, end)
	}
}

// TestLoadConfigErrors covers both failure paths. A malformed or absent config
// must return an error rather than a zero-value Config, which would simulate a
// chain with no validators and report success.
func TestLoadConfigErrors(t *testing.T) {
	dir := t.TempDir()

	t.Run("missing file", func(t *testing.T) {
		if _, err := loadConfig(filepath.Join(dir, "does-not-exist.json")); err == nil {
			t.Error("loadConfig on a missing file = nil error; want a failure")
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		path := filepath.Join(dir, "bad.json")
		if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
		if _, err := loadConfig(path); err == nil {
			t.Error("loadConfig on malformed JSON = nil error; want a failure")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		// An empty file decodes as io.EOF. It must NOT come back as a
		// zero Config with a filled-in default window.
		path := filepath.Join(dir, "empty.json")
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
		if _, err := loadConfig(path); err == nil {
			t.Error("loadConfig on an empty file = nil error; want a failure")
		}
	})
}

// TestJSONRoundTrip covers writeJSON and readJSON together, since they are
// only ever used as a pair (write results, read them back). It also pins that
// writeJSON produces indented output, which is what makes a saved
// metrics.json reviewable by hand.
func TestJSONRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics.json")

	want := core.Config{
		ScenarioName: "round-trip",
		StartTime:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		EndTime:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		TimeStep:     24 * time.Hour,
		Seed:         7,
	}

	if err := writeJSON(path, want); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	// SetIndent("", "  ") means every line after the first is indented.
	if len(raw) > 2 && !bytesContainNewlineIndent(raw) {
		t.Error("writeJSON output is not indented; the operator-facing " +
			"metrics.json is meant to be human-readable")
	}

	var got core.Config
	if err := readJSON(path, &got); err != nil {
		t.Fatalf("readJSON: %v", err)
	}

	if got.ScenarioName != want.ScenarioName {
		t.Errorf("ScenarioName = %q, want %q", got.ScenarioName, want.ScenarioName)
	}
	if got.Seed != want.Seed {
		t.Errorf("Seed = %d, want %d", got.Seed, want.Seed)
	}
	if got.TimeStep != want.TimeStep {
		t.Errorf("TimeStep = %v, want %v", got.TimeStep, want.TimeStep)
	}
	if !got.StartTime.Equal(want.StartTime) {
		t.Errorf("StartTime = %v, want %v", got.StartTime, want.StartTime)
	}
	if !got.EndTime.Equal(want.EndTime) {
		t.Errorf("EndTime = %v, want %v", got.EndTime, want.EndTime)
	}
}

// TestJSONWriteReadErrors covers the error branches of both helpers.
func TestJSONWriteReadErrors(t *testing.T) {
	dir := t.TempDir()

	t.Run("write to a missing directory", func(t *testing.T) {
		path := filepath.Join(dir, "no-such-dir", "out.json")
		if err := writeJSON(path, core.Config{}); err == nil {
			t.Error("writeJSON into a missing directory = nil error; want a failure")
		}
	})

	t.Run("read a missing file", func(t *testing.T) {
		var cfg core.Config
		if err := readJSON(filepath.Join(dir, "absent.json"), &cfg); err == nil {
			t.Error("readJSON on a missing file = nil error; want a failure")
		}
	})

	t.Run("read malformed json", func(t *testing.T) {
		path := filepath.Join(dir, "malformed.json")
		if err := os.WriteFile(path, []byte("{oops"), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
		var cfg core.Config
		if err := readJSON(path, &cfg); err == nil {
			t.Error("readJSON on malformed JSON = nil error; want a failure")
		}
	})
}

// TestResolveConfig covers the scenario-name switch. An unknown scenario must
// error rather than return a zero Config, which would simulate an empty chain
// and exit 0.
func TestResolveConfig(t *testing.T) {
	known := []string{"baseline", "bull", "bear", "attack", "black_swan"}
	for _, name := range known {
		cfg, err := resolveConfig(name, "")
		if err != nil {
			t.Errorf("resolveConfig(%q) = error %v; want a config", name, err)
			continue
		}
		// A named scenario must actually populate the chain, or it is not
		// doing anything.
		if cfg.NumUsers == 0 && cfg.NumValidators == 0 {
			t.Errorf("resolveConfig(%q) returned an empty chain "+
				"(NumUsers=%d NumValidators=%d)", name, cfg.NumUsers, cfg.NumValidators)
		}
	}

	if _, err := resolveConfig("no-such-scenario", ""); err == nil {
		t.Error("resolveConfig with an unknown scenario = nil error; want a failure")
	}
}

// TestResolveConfigPrefersExplicitPath pins precedence: a path on the command
// line must win over the scenario name. If the switch were consulted first,
// an operator's explicit config would be silently ignored.
func TestResolveConfigPrefersExplicitPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "explicit.json")

	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	body, err := json.Marshal(map[string]any{
		"ScenarioName": "from-file",
		"StartTime":    start,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	// The scenario name is "no-such-scenario", which would error if consulted.
	cfg, err := resolveConfig("no-such-scenario", path)
	if err != nil {
		t.Fatalf("resolveConfig with an explicit path = error %v; the path must "+
			"take precedence over the (invalid) scenario name", err)
	}
	if cfg.ScenarioName != "from-file" {
		t.Errorf("ScenarioName = %q, want %q", cfg.ScenarioName, "from-file")
	}
}

// TestResolveConfigBadPath checks a non-empty path that cannot be read still
// surfaces the error instead of silently falling back to a named scenario.
func TestResolveConfigBadPath(t *testing.T) {
	if _, err := resolveConfig("baseline", filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("resolveConfig with an unreadable explicit path = nil error; " +
			"want the load failure, not a silent fallback to the named scenario")
	}
}

func bytesContainNewlineIndent(b []byte) bool {
	for i := 1; i < len(b); i++ {
		if b[i-1] == '\n' && (b[i] == ' ' || b[i] == '\t') {
			return true
		}
	}
	return false
}

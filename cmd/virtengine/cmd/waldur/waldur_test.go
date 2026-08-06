package waldur

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := GetCmd()
	if cmd.Name() != "waldur" {
		t.Fatalf("GetCmd().Name() = %q, want %q", cmd.Name(), "waldur")
	}

	names := make([]string, 0, len(cmd.Commands()))
	for _, subcmd := range cmd.Commands() {
		names = append(names, subcmd.Name())
	}

	if !containsString(names, "init-categories") || !containsString(names, "list-categories") {
		t.Fatalf("GetCmd() subcommands = %v", names)
	}
}

func TestListCategoriesCmd_UsesEnvAndOutputsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/me/":
			http.NotFound(w, r)
		case "/api/users/me/":
			w.Header().Set("Content-Type", "application/json")
			username := "cli"
			uuid := "550e8400-e29b-41d4-a716-446655440096"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"uuid":     &uuid,
				"username": &username,
			})
		case "/api/marketplace-categories/":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"count": 1,
				"results": []map[string]any{
					{"uuid": "cat-1", "title": "Compute", "description": "Compute category"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("WALDUR_URL", server.URL)
	t.Setenv("WALDUR_TOKEN", "test-token")

	cmd := getListCategoriesCmd()
	cmd.SetArgs([]string{"--output", "json", "--timeout", "5"})
	cmd.SetContext(context.Background())

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() unexpected error: %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = originalStdout
	}()

	runErr := cmd.Execute()
	_ = writer.Close()
	var output bytes.Buffer
	_, _ = io.Copy(&output, reader)
	_ = reader.Close()

	if runErr != nil {
		t.Fatalf("Execute() unexpected error: %v", runErr)
	}
	if !strings.Contains(output.String(), "\"title\": \"Compute\"") {
		t.Fatalf("Execute() output = %s", output.String())
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

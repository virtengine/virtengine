package enclave_runtime

import (
	"strings"
	"testing"
)

func TestProductionServiceRejectsUnsafeConfigurationBeforeFactoryCreation(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if _, err := CreateProductionServiceWithConfig(nil); err == nil || !strings.Contains(err.Error(), "cannot be nil") {
			t.Fatalf("expected nil configuration rejection, got %v", err)
		}
	})
	t.Run("development mode", func(t *testing.T) {
		cfg := DefaultDevelopmentConfig()
		if _, err := CreateProductionServiceWithConfig(&cfg); err == nil || !strings.Contains(err.Error(), "requires production TEE mode") {
			t.Fatalf("expected development-mode rejection, got %v", err)
		}
	})
	t.Run("missing measurement allowlist", func(t *testing.T) {
		cfg := DefaultProductionConfig()
		if _, err := CreateProductionServiceWithConfig(&cfg); err == nil || !strings.Contains(err.Error(), "no measurement allowlist") {
			t.Fatalf("expected readiness rejection, got %v", err)
		}
	})
}

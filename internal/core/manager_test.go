package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestControllerSecretPersistsAcrossManagers(t *testing.T) {
	dir := t.TempDir()
	first := New(dir, "missing")
	_, firstSecret := first.Controller()
	if len(firstSecret) != 48 {
		t.Fatalf("generated secret length = %d", len(firstSecret))
	}

	second := New(dir, "missing")
	_, secondSecret := second.Controller()
	if secondSecret != firstSecret {
		t.Fatalf("controller secret changed across managers")
	}
}

func TestControllerSecretMigratesFromExistingConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("secret: existing-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}

	manager := New(dir, "missing")
	_, secret := manager.Controller()
	if secret != "existing-secret" {
		t.Fatalf("controller secret = %q", secret)
	}
	persisted, err := os.ReadFile(filepath.Join(dir, "controller-secret"))
	if err != nil {
		t.Fatal(err)
	}
	if string(persisted) != secret {
		t.Fatalf("persisted secret = %q", persisted)
	}
}

package model

import "testing"

func TestDefaultStateAllowsLAN(t *testing.T) {
	settings := DefaultState().Settings
	if !settings.AllowLAN {
		t.Fatal("LAN access should be enabled by default")
	}
	if settings.BindAddress != "*" {
		t.Fatalf("default bind address = %q, want *", settings.BindAddress)
	}
}

package model

import (
	"encoding/json"
	"testing"
)

func TestDefaultStateAllowsLAN(t *testing.T) {
	settings := DefaultState().Settings
	if !settings.AllowLAN {
		t.Fatal("LAN access should be enabled by default")
	}
	if settings.BindAddress != "*" {
		t.Fatalf("default bind address = %q, want *", settings.BindAddress)
	}
}

func TestEntryGroupJSONAppliesFallbackDefaults(t *testing.T) {
	var group EntryGroup
	if err := json.Unmarshal([]byte(`{"id":"g1","name":"auto","type":"fallback","nodeIds":["n1"]}`), &group); err != nil {
		t.Fatal(err)
	}
	if group.TestURL != DefaultEntryGroupTestURL || group.Interval != DefaultEntryGroupInterval {
		t.Fatalf("fallback defaults = %q, %d", group.TestURL, group.Interval)
	}
}

func TestEntryGroupJSONSelectsFirstMemberByDefault(t *testing.T) {
	var group EntryGroup
	if err := json.Unmarshal([]byte(`{"id":"g1","name":"manual","type":"select","nodeIds":["n1","n2"]}`), &group); err != nil {
		t.Fatal(err)
	}
	if group.SelectedNodeID != "n1" {
		t.Fatalf("selected node = %q, want n1", group.SelectedNodeID)
	}
}

func TestNewIDIsRandomAndNonEmpty(t *testing.T) {
	first, err := NewID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewID()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 32 || len(second) != 32 || first == second {
		t.Fatalf("unexpected IDs %q and %q", first, second)
	}
}

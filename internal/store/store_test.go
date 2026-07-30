package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/foliageSea/docker-clash/internal/model"
)

func TestOpenLegacyStateWithoutEntryGroups(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	legacy := []byte(`{"settings":{"listen":"127.0.0.1:9080","mixedPort":7890,"allowLan":true,"bindAddress":"*"},"nodes":[]}`)
	if err := os.WriteFile(path, legacy, 0600); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Get().EntryGroups) != 0 {
		t.Fatalf("legacy state loaded unexpected entry groups: %+v", s.Get().EntryGroups)
	}
}

func TestOpenNormalizesLegacyDuplicateNodeIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	legacy := []byte(`{"settings":{"mixedPort":7890},"nodes":[{"id":"same","name":"one"},{"id":"same","name":"two"},{"name":"three"}]}`)
	if err := os.WriteFile(path, legacy, 0600); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	nodes := s.Get().Nodes
	if nodes[0].ID != "same" || nodes[1].ID == "" || nodes[2].ID == "" || nodes[1].ID == nodes[2].ID {
		t.Fatalf("node IDs were not normalized: %+v", nodes)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Get().Nodes[1].ID != nodes[1].ID || reopened.Get().Nodes[2].ID != nodes[2].ID {
		t.Fatal("normalized IDs were not persisted")
	}
}

func TestUpdateRollsBackNestedMutations(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Update(func(state *model.State) error {
		state.Nodes = []model.Node{{ID: "n1", Name: "original"}}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("reject update")
	if err := s.Update(func(state *model.State) error {
		state.Nodes[0].Name = "mutated"
		return wantErr
	}); !errors.Is(err, wantErr) {
		t.Fatalf("Update error = %v, want %v", err, wantErr)
	}
	if got := s.Get().Nodes[0].Name; got != "original" {
		t.Fatalf("failed update leaked nested mutation: %q", got)
	}
}

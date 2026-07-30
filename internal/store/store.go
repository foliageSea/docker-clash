package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/foliageSea/docker-clash/internal/model"
)

type Store struct {
	mu    sync.RWMutex
	path  string
	state model.State
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, state: model.DefaultState()}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, s.saveLocked()
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &s.state); err != nil {
		return nil, err
	}
	changed, err := normalizeNodeIDs(&s.state)
	if err != nil {
		return nil, err
	}
	if changed {
		if err := s.saveLocked(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) Get() model.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, _ := json.Marshal(s.state)
	var out model.State
	_ = json.Unmarshal(b, &out)
	return out
}

func (s *Store) Update(fn func(*model.State) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	before, err := cloneState(s.state)
	if err != nil {
		return err
	}
	if err := fn(&s.state); err != nil {
		s.state = before
		return err
	}
	if err := s.saveLocked(); err != nil {
		s.state = before
		return err
	}
	return nil
}

func cloneState(state model.State) (model.State, error) {
	b, err := json.Marshal(state)
	if err != nil {
		return model.State{}, err
	}
	var cloned model.State
	if err := json.Unmarshal(b, &cloned); err != nil {
		return model.State{}, err
	}
	return cloned, nil
}

func normalizeNodeIDs(state *model.State) (bool, error) {
	seen := make(map[string]bool, len(state.Nodes))
	changed := false
	for i := range state.Nodes {
		id := state.Nodes[i].ID
		if id != "" && !seen[id] {
			seen[id] = true
			continue
		}
		newID, err := model.NewID()
		if err != nil {
			return false, err
		}
		state.Nodes[i].ID = newID
		seen[newID] = true
		changed = true
	}
	return changed, nil
}

func (s *Store) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

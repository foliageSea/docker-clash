package config

import (
	"github.com/foliageSea/nexus-proxy-ui/internal/model"
	"strings"
	"testing"
)

func TestValidateRejectsCycle(t *testing.T) {
	s := model.DefaultState()
	s.Nodes = []model.Node{{Name: "a", DialerProxy: "b"}, {Name: "b", DialerProxy: "a"}}
	if err := Validate(s); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error, got %v", err)
	}
}

func TestRender(t *testing.T) {
	s := model.DefaultState()
	s.Nodes = []model.Node{{Name: "edge", Type: "ss", Server: "host", Port: 443, Options: map[string]any{"cipher": "aes-128-gcm", "password": "x"}}}
	b, err := Render(s, "127.0.0.1:9090", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "name: edge") || !strings.Contains(string(b), "MATCH,NEXUS") {
		t.Fatalf("unexpected config: %s", b)
	}
}

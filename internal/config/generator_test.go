package config

import (
	"github.com/foliageSea/docker-clash/internal/model"
	"gopkg.in/yaml.v3"
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

func TestValidateRejectsMixedNodeGroupCycle(t *testing.T) {
	s := model.DefaultState()
	s.Nodes = []model.Node{{ID: "n1", Name: "edge", DialerProxy: "entry"}}
	s.EntryGroups = []model.EntryGroup{{ID: "g1", Name: "entry", Type: "select", NodeIDs: []string{"n1"}}}
	if err := Validate(s); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected mixed cycle error, got %v", err)
	}
}

func TestValidateEntryGroupConstraints(t *testing.T) {
	tests := []struct {
		name  string
		group model.EntryGroup
		want  string
	}{
		{name: "reserved name", group: model.EntryGroup{Name: "DIRECT", Type: "select", NodeIDs: []string{"n1"}}, want: "reserved"},
		{name: "missing member", group: model.EntryGroup{Name: "entry", Type: "select", NodeIDs: []string{"missing"}}, want: "missing node"},
		{name: "duplicate member", group: model.EntryGroup{Name: "entry", Type: "select", NodeIDs: []string{"n1", "n1"}}, want: "duplicate node"},
		{name: "invalid selection", group: model.EntryGroup{Name: "entry", Type: "select", NodeIDs: []string{"n1"}, SelectedNodeID: "n2"}, want: "not a member"},
		{name: "invalid fallback URL", group: model.EntryGroup{Name: "entry", Type: "fallback", NodeIDs: []string{"n1"}, TestURL: "file:///tmp/test", Interval: 60}, want: "invalid test URL"},
		{name: "invalid fallback interval", group: model.EntryGroup{Name: "entry", Type: "fallback", NodeIDs: []string{"n1"}, TestURL: model.DefaultEntryGroupTestURL}, want: "interval"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := model.DefaultState()
			s.Nodes = []model.Node{{ID: "n1", Name: "edge"}, {ID: "n2", Name: "other"}}
			s.EntryGroups = []model.EntryGroup{tt.group}
			if err := Validate(s); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
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

func TestRenderEntryGroupsBeforeNexus(t *testing.T) {
	s := model.DefaultState()
	s.Nodes = []model.Node{
		{ID: "n1", Name: "first", Type: "ss", Server: "one", Port: 1},
		{ID: "n2", Name: "second", Type: "ss", Server: "two", Port: 2},
	}
	s.EntryGroups = []model.EntryGroup{
		{ID: "g1", Name: "manual", Type: "select", NodeIDs: []string{"n1", "n2"}, SelectedNodeID: "n2"},
		{ID: "g2", Name: "auto", Type: "fallback", NodeIDs: []string{"n1", "n2"}, TestURL: model.DefaultEntryGroupTestURL, Interval: 60},
	}
	b, err := Render(s, "127.0.0.1:9090", "secret")
	if err != nil {
		t.Fatal(err)
	}
	var rendered struct {
		Groups []struct {
			Name     string   `yaml:"name"`
			Type     string   `yaml:"type"`
			Proxies  []string `yaml:"proxies"`
			URL      string   `yaml:"url"`
			Interval int      `yaml:"interval"`
		} `yaml:"proxy-groups"`
	}
	if err := yaml.Unmarshal(b, &rendered); err != nil {
		t.Fatal(err)
	}
	if len(rendered.Groups) != 3 || rendered.Groups[0].Name != "manual" || rendered.Groups[1].Name != "auto" || rendered.Groups[2].Name != "NEXUS" {
		t.Fatalf("unexpected group order: %+v", rendered.Groups)
	}
	if got := rendered.Groups[0].Proxies; len(got) != 2 || got[0] != "second" || got[1] != "first" {
		t.Fatalf("selected member was not rendered first: %v", got)
	}
	if rendered.Groups[1].URL != model.DefaultEntryGroupTestURL || rendered.Groups[1].Interval != 60 {
		t.Fatalf("fallback options missing: %+v", rendered.Groups[1])
	}
	if slicesContain(rendered.Groups[2].Proxies, "manual") || slicesContain(rendered.Groups[2].Proxies, "auto") {
		t.Fatalf("entry groups leaked into NEXUS: %v", rendered.Groups[2].Proxies)
	}
}

func slicesContain(values []string, value string) bool {
	for _, current := range values {
		if current == value {
			return true
		}
	}
	return false
}

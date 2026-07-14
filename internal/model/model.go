package model

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

const (
	DefaultEntryGroupTestURL  = "https://www.gstatic.com/generate_204"
	DefaultEntryGroupInterval = 60
)

type Node struct {
	ID          string         `json:"id" yaml:"-"`
	Name        string         `json:"name" yaml:"name"`
	Type        string         `json:"type" yaml:"type"`
	Server      string         `json:"server" yaml:"server"`
	Port        int            `json:"port" yaml:"port"`
	DialerProxy string         `json:"dialerProxy,omitempty" yaml:"dialer-proxy,omitempty"`
	Options     map[string]any `json:"options,omitempty" yaml:",inline"`
	CreatedAt   time.Time      `json:"createdAt" yaml:"-"`
}

type Settings struct {
	Listen       string `json:"listen"`
	MixedPort    int    `json:"mixedPort"`
	AllowLAN     bool   `json:"allowLan"`
	BindAddress  string `json:"bindAddress"`
	SelectedNode string `json:"selectedNode,omitempty"`
}

type EntryGroup struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	NodeIDs        []string `json:"nodeIds"`
	SelectedNodeID string   `json:"selectedNodeId,omitempty"`
	TestURL        string   `json:"testUrl,omitempty"`
	Interval       int      `json:"interval,omitempty"`
}

func (g *EntryGroup) UnmarshalJSON(data []byte) error {
	type plain EntryGroup
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*g = EntryGroup(decoded)
	g.SetDefaults()
	return nil
}

func (g *EntryGroup) SetDefaults() {
	if g.Type == "select" {
		g.TestURL = ""
		g.Interval = 0
		selectedExists := false
		for _, id := range g.NodeIDs {
			if id == g.SelectedNodeID {
				selectedExists = true
				break
			}
		}
		if !selectedExists && len(g.NodeIDs) > 0 {
			g.SelectedNodeID = g.NodeIDs[0]
		}
		return
	}
	if g.Type != "fallback" {
		return
	}
	g.SelectedNodeID = ""
	if g.TestURL == "" {
		g.TestURL = DefaultEntryGroupTestURL
	}
	if g.Interval == 0 {
		g.Interval = DefaultEntryGroupInterval
	}
}

func NewID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

type State struct {
	Settings    Settings     `json:"settings"`
	Nodes       []Node       `json:"nodes"`
	EntryGroups []EntryGroup `json:"entryGroups"`
}

func DefaultState() State {
	return State{Settings: Settings{Listen: "127.0.0.1:9080", MixedPort: 7890, AllowLAN: true, BindAddress: "*"}, Nodes: []Node{}, EntryGroups: []EntryGroup{}}
}

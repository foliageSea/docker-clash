package model

import "time"

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

type State struct {
	Settings Settings `json:"settings"`
	Nodes    []Node   `json:"nodes"`
}

func DefaultState() State {
	return State{Settings: Settings{Listen: "127.0.0.1:9080", MixedPort: 7890, BindAddress: "*"}, Nodes: []Node{}}
}

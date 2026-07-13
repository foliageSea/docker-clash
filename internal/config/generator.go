package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/foliageSea/nexus-proxy-ui/internal/model"
	"gopkg.in/yaml.v3"
)

func Validate(state model.State) error {
	if state.Settings.MixedPort < 1 || state.Settings.MixedPort > 65535 {
		return fmt.Errorf("mixed port must be between 1 and 65535")
	}
	names := map[string]bool{}
	for _, n := range state.Nodes {
		if n.Name == "" {
			return fmt.Errorf("node name is required")
		}
		if names[n.Name] {
			return fmt.Errorf("duplicate node name %q", n.Name)
		}
		names[n.Name] = true
	}
	for _, n := range state.Nodes {
		if n.DialerProxy != "" && !names[n.DialerProxy] {
			return fmt.Errorf("dialer proxy %q does not exist", n.DialerProxy)
		}
	}
	visiting, visited := map[string]bool{}, map[string]bool{}
	byName := map[string]model.Node{}
	for _, n := range state.Nodes {
		byName[n.Name] = n
	}
	var visit func(string) error
	visit = func(name string) error {
		if visiting[name] {
			return fmt.Errorf("proxy chain contains a cycle at %q", name)
		}
		if visited[name] {
			return nil
		}
		visiting[name] = true
		if next := byName[name].DialerProxy; next != "" {
			if err := visit(next); err != nil {
				return err
			}
		}
		visiting[name] = false
		visited[name] = true
		return nil
	}
	for name := range byName {
		if err := visit(name); err != nil {
			return err
		}
	}
	return nil
}

func Render(state model.State, controller, secret string) ([]byte, error) {
	if err := Validate(state); err != nil {
		return nil, err
	}
	proxies := make([]map[string]any, 0, len(state.Nodes))
	names := make([]string, 0, len(state.Nodes))
	for _, n := range state.Nodes {
		p := map[string]any{"name": n.Name, "type": n.Type, "server": n.Server, "port": n.Port}
		for k, v := range n.Options {
			p[k] = v
		}
		if n.DialerProxy != "" {
			p["dialer-proxy"] = n.DialerProxy
		}
		proxies = append(proxies, p)
		names = append(names, n.Name)
	}
	selected := state.Settings.SelectedNode
	if selected == "" && len(names) > 0 {
		selected = names[0]
	}
	groups := []map[string]any{{"name": "NEXUS", "type": "select", "proxies": append(names, "DIRECT")}}
	cfg := map[string]any{"mixed-port": state.Settings.MixedPort, "allow-lan": state.Settings.AllowLAN, "bind-address": state.Settings.BindAddress, "mode": "rule", "log-level": "info", "external-controller": controller, "secret": secret, "profile": map[string]any{"store-selected": true}, "proxies": proxies, "proxy-groups": groups, "rules": []string{"MATCH,NEXUS"}}
	_ = selected
	return yaml.Marshal(cfg)
}

func Write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

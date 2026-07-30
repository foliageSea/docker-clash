package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"

	"github.com/foliageSea/docker-clash/internal/model"
	"gopkg.in/yaml.v3"
)

func Validate(state model.State) error {
	if state.Settings.MixedPort < 1 || state.Settings.MixedPort > 65535 {
		return fmt.Errorf("mixed port must be between 1 and 65535")
	}
	reserved := map[string]bool{"DOCKER_CLASH": true, "DIRECT": true, "REJECT": true, "PASS": true, "GLOBAL": true, "COMPATIBLE": true}
	names := map[string]string{}
	nodeNamesByID := map[string]string{}
	for _, n := range state.Nodes {
		if n.Name == "" {
			return fmt.Errorf("node name is required")
		}
		if reserved[n.Name] {
			return fmt.Errorf("node name %q is reserved", n.Name)
		}
		if kind := names[n.Name]; kind != "" {
			return fmt.Errorf("node name %q conflicts with existing %s", n.Name, kind)
		}
		names[n.Name] = "node"
		if n.ID != "" {
			if existing := nodeNamesByID[n.ID]; existing != "" {
				return fmt.Errorf("nodes %q and %q have duplicate ID %q", existing, n.Name, n.ID)
			}
			nodeNamesByID[n.ID] = n.Name
		}
	}
	for _, g := range state.EntryGroups {
		if g.Name == "" {
			return fmt.Errorf("entry group name is required")
		}
		if reserved[g.Name] {
			return fmt.Errorf("entry group name %q is reserved", g.Name)
		}
		if kind := names[g.Name]; kind != "" {
			return fmt.Errorf("entry group name %q conflicts with existing %s", g.Name, kind)
		}
		names[g.Name] = "entry group"
		if g.Type != "select" && g.Type != "fallback" {
			return fmt.Errorf("entry group %q has invalid type %q", g.Name, g.Type)
		}
		if len(g.NodeIDs) == 0 {
			return fmt.Errorf("entry group %q must contain at least one node", g.Name)
		}
		members := map[string]bool{}
		for _, id := range g.NodeIDs {
			if members[id] {
				return fmt.Errorf("entry group %q contains duplicate node ID %q", g.Name, id)
			}
			members[id] = true
			if nodeNamesByID[id] == "" {
				return fmt.Errorf("entry group %q references missing node ID %q", g.Name, id)
			}
		}
		if g.Type == "select" && g.SelectedNodeID != "" && !members[g.SelectedNodeID] {
			return fmt.Errorf("entry group %q selected node ID %q is not a member", g.Name, g.SelectedNodeID)
		}
		if g.TestURL != "" {
			u, err := url.ParseRequestURI(g.TestURL)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
				return fmt.Errorf("entry group %q has invalid test URL %q", g.Name, g.TestURL)
			}
		}
		if g.Interval < 0 || (g.Type == "fallback" && g.Interval == 0) {
			return fmt.Errorf("entry group %q interval must be greater than zero", g.Name)
		}
	}
	for _, n := range state.Nodes {
		if n.DialerProxy != "" && names[n.DialerProxy] == "" {
			return fmt.Errorf("dialer proxy %q does not exist", n.DialerProxy)
		}
	}
	visiting, visited := map[string]bool{}, map[string]bool{}
	edges := map[string][]string{}
	for _, n := range state.Nodes {
		if n.DialerProxy != "" {
			edges[n.Name] = []string{n.DialerProxy}
		}
	}
	for _, g := range state.EntryGroups {
		for _, id := range g.NodeIDs {
			edges[g.Name] = append(edges[g.Name], nodeNamesByID[id])
		}
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
		for _, next := range edges[name] {
			if err := visit(next); err != nil {
				return err
			}
		}
		visiting[name] = false
		visited[name] = true
		return nil
	}
	for name := range names {
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
	groups := make([]map[string]any, 0, len(state.EntryGroups)+1)
	nodeNamesByID := make(map[string]string, len(state.Nodes))
	for _, n := range state.Nodes {
		nodeNamesByID[n.ID] = n.Name
	}
	for _, entry := range state.EntryGroups {
		entry.SetDefaults()
		members := make([]string, 0, len(entry.NodeIDs))
		for _, id := range entry.NodeIDs {
			members = append(members, nodeNamesByID[id])
		}
		if entry.Type == "select" && entry.SelectedNodeID != "" {
			selected := nodeNamesByID[entry.SelectedNodeID]
			idx := slices.Index(members, selected)
			members[0], members[idx] = members[idx], members[0]
		}
		group := map[string]any{"name": entry.Name, "type": entry.Type, "proxies": members}
		if entry.Type == "fallback" {
			group["url"] = entry.TestURL
			group["interval"] = entry.Interval
		}
		groups = append(groups, group)
	}
	dockerClashMembers := append([]string(nil), names...)
	if idx := slices.Index(dockerClashMembers, state.Settings.SelectedNode); idx > 0 {
		dockerClashMembers[0], dockerClashMembers[idx] = dockerClashMembers[idx], dockerClashMembers[0]
	}
	groups = append(groups, map[string]any{"name": "DOCKER_CLASH", "type": "select", "proxies": append(dockerClashMembers, "DIRECT")})
	cfg := map[string]any{"mixed-port": state.Settings.MixedPort, "allow-lan": state.Settings.AllowLAN, "bind-address": state.Settings.BindAddress, "mode": "rule", "log-level": "info", "external-controller": controller, "secret": secret, "profile": map[string]any{"store-selected": true}, "proxies": proxies, "proxy-groups": groups, "rules": []string{"MATCH,DOCKER_CLASH"}}
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

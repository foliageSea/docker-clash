package node

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/foliageSea/nexus-proxy-ui/internal/model"
)

func Parse(raw string) (model.Node, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "vmess://") {
		return parseVMess(strings.TrimPrefix(raw, "vmess://"))
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" {
		return model.Node{}, fmt.Errorf("invalid node URI")
	}
	typeName := strings.ToLower(u.Scheme)
	if typeName == "hy2" {
		typeName = "hysteria2"
	}
	supported := map[string]bool{"ss": true, "socks5": true, "vless": true, "trojan": true, "hysteria2": true, "tuic": true}
	if !supported[typeName] {
		return model.Node{}, fmt.Errorf("unsupported protocol %q", u.Scheme)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil || port < 1 || port > 65535 {
		return model.Node{}, fmt.Errorf("invalid port")
	}
	name, _ := url.QueryUnescape(u.Fragment)
	if name == "" {
		name = typeName + "-" + u.Hostname()
	}
	n := model.Node{ID: newID(), Name: name, Type: typeName, Server: u.Hostname(), Port: port, CreatedAt: time.Now().UTC(), Options: map[string]any{}}
	q := u.Query()
	switch typeName {
	case "ss":
		if err := parseSSUser(u, &n); err != nil {
			return model.Node{}, err
		}
	case "socks5":
		if username := u.User.Username(); username != "" {
			n.Options["username"] = username
		}
		if password, ok := u.User.Password(); ok {
			n.Options["password"] = password
		}
		copyQuery(q, n.Options, "sni", "ip-version")
		copyBoolQuery(q, n.Options, "tls", "udp", "skip-cert-verify")
	case "vless":
		n.Options["uuid"] = u.User.Username()
		copyQuery(q, n.Options, "network", "flow", "servername", "client-fingerprint", "skip-cert-verify", "udp")
		if q.Get("security") == "tls" || q.Get("security") == "reality" {
			n.Options["tls"] = true
		}
		if q.Get("sni") != "" {
			n.Options["servername"] = q.Get("sni")
		}
		if q.Get("pbk") != "" {
			n.Options["reality-opts"] = map[string]any{"public-key": q.Get("pbk"), "short-id": q.Get("sid")}
		}
	case "trojan":
		n.Options["password"] = u.User.Username()
		copyQuery(q, n.Options, "network", "sni", "skip-cert-verify", "udp")
	case "hysteria2":
		password, _ := u.User.Password()
		if password == "" {
			password = u.User.Username()
		}
		n.Options["password"] = password
		copyQuery(q, n.Options, "sni", "obfs", "obfs-password", "skip-cert-verify")
	case "tuic":
		n.Options["uuid"] = u.User.Username()
		password, _ := u.User.Password()
		n.Options["password"] = password
		copyQuery(q, n.Options, "sni", "congestion-controller", "udp-relay-mode", "skip-cert-verify")
	}
	return n, nil
}

func parseVMess(payload string) (model.Node, error) {
	b, err := decodeBase64(payload)
	if err != nil {
		return model.Node{}, fmt.Errorf("invalid vmess payload: %w", err)
	}
	var v map[string]any
	if err := json.Unmarshal(b, &v); err != nil {
		return model.Node{}, fmt.Errorf("invalid vmess JSON: %w", err)
	}
	port, err := strconv.Atoi(fmt.Sprint(v["port"]))
	if err != nil {
		return model.Node{}, fmt.Errorf("invalid port")
	}
	n := model.Node{ID: newID(), Name: stringValue(v, "ps", "vmess-node"), Type: "vmess", Server: fmt.Sprint(v["add"]), Port: port, CreatedAt: time.Now().UTC(), Options: map[string]any{"uuid": v["id"], "alterId": intValue(v["aid"]), "cipher": "auto"}}
	if net := fmt.Sprint(v["net"]); net != "" {
		n.Options["network"] = net
	}
	if fmt.Sprint(v["tls"]) != "" {
		n.Options["tls"] = true
	}
	if sni := fmt.Sprint(v["sni"]); sni != "" {
		n.Options["servername"] = sni
	}
	return n, nil
}

func parseSSUser(u *url.URL, n *model.Node) error {
	user := u.User.String()
	if !strings.Contains(user, ":") {
		b, err := decodeBase64(user)
		if err != nil {
			return fmt.Errorf("invalid ss credentials")
		}
		user = string(b)
	}
	parts := strings.SplitN(user, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid ss credentials")
	}
	password, _ := url.QueryUnescape(parts[1])
	n.Options["cipher"] = parts[0]
	n.Options["password"] = password
	return nil
}

func copyQuery(q url.Values, out map[string]any, keys ...string) {
	for _, k := range keys {
		if v := q.Get(k); v != "" {
			out[k] = v
		}
	}
}
func copyBoolQuery(q url.Values, out map[string]any, keys ...string) {
	for _, k := range keys {
		if v := q.Get(k); v != "" {
			parsed, err := strconv.ParseBool(v)
			if err == nil {
				out[k] = parsed
			}
		}
	}
}
func decodeBase64(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if b, e := base64.RawURLEncoding.DecodeString(s); e == nil {
		return b, nil
	}
	return base64.StdEncoding.DecodeString(s)
}
func newID() string {
	id, err := model.NewID()
	if err == nil {
		return id
	}
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
func stringValue(m map[string]any, key, fallback string) string {
	if v := fmt.Sprint(m[key]); v != "" && v != "<nil>" {
		return v
	}
	return fallback
}
func intValue(v any) int { n, _ := strconv.Atoi(fmt.Sprint(v)); return n }

package node

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/foliageSea/docker-clash/internal/model"
	"gopkg.in/yaml.v3"
)

const maxSubscriptionSize = 8 << 20

var subscriptionClient = &http.Client{Timeout: 20 * time.Second}

func Import(source string) ([]model.Node, error) {
	source = strings.TrimSpace(source)
	u, err := url.Parse(source)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		return importURL(source)
	}
	n, err := Parse(source)
	if err != nil {
		return nil, err
	}
	return []model.Node{n}, nil
}

func importURL(source string) ([]model.Node, error) {
	req, err := http.NewRequest(http.MethodGet, source, nil)
	if err != nil {
		return nil, err
	}
		req.Header.Set("User-Agent", "DockerClash/1.0 mihomo")
	resp, err := subscriptionClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download subscription: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("subscription server returned %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSubscriptionSize+1))
	if err != nil {
		return nil, fmt.Errorf("read subscription: %w", err)
	}
	if len(body) > maxSubscriptionSize {
		return nil, fmt.Errorf("subscription exceeds 8 MiB")
	}
	return ParseSubscription(body)
}

func ParseSubscription(body []byte) ([]model.Node, error) {
	if nodes, err := parseYAMLSubscription(body); err == nil && len(nodes) > 0 {
		return nodes, nil
	}
	text := strings.TrimSpace(string(body))
	if decoded, err := decodeSubscriptionBase64(text); err == nil {
		text = string(decoded)
	}
	var nodes []model.Node
	var firstError error
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		n, err := Parse(line)
		if err != nil {
			if firstError == nil {
				firstError = err
			}
			continue
		}
		nodes = append(nodes, n)
	}
	if len(nodes) == 0 {
		if firstError != nil {
			return nil, fmt.Errorf("subscription contains no supported nodes: %w", firstError)
		}
		return nil, fmt.Errorf("subscription is empty or unsupported")
	}
	return nodes, nil
}

func parseYAMLSubscription(body []byte) ([]model.Node, error) {
	var document struct {
		Proxies []map[string]any `yaml:"proxies"`
	}
	if err := yaml.NewDecoder(bytes.NewReader(body)).Decode(&document); err != nil || len(document.Proxies) == 0 {
		return nil, fmt.Errorf("not a proxy YAML subscription")
	}
	nodes := make([]model.Node, 0, len(document.Proxies))
	for _, proxy := range document.Proxies {
		name, _ := proxy["name"].(string)
		typeName, _ := proxy["type"].(string)
		server, _ := proxy["server"].(string)
		port, ok := numberToInt(proxy["port"])
		if name == "" || typeName == "" || server == "" || !ok {
			continue
		}
		delete(proxy, "name")
		delete(proxy, "type")
		delete(proxy, "server")
		delete(proxy, "port")
		dialer, _ := proxy["dialer-proxy"].(string)
		delete(proxy, "dialer-proxy")
		nodes = append(nodes, model.Node{ID: newID(), Name: name, Type: strings.ToLower(typeName), Server: server, Port: port, DialerProxy: dialer, Options: proxy, CreatedAt: time.Now().UTC()})
	}
	return nodes, nil
}

func decodeSubscriptionBase64(value string) ([]byte, error) {
	value = strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' || r == ' ' || r == '\t' {
			return -1
		}
		return r
	}, value)
	if decoded, err := base64.RawStdEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}
	if decoded, err := base64.RawURLEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}
	return base64.StdEncoding.DecodeString(value)
}

func numberToInt(value any) (int, bool) {
	switch number := value.(type) {
	case int:
		return number, number > 0 && number <= 65535
	case uint64:
		return int(number), number > 0 && number <= 65535
	case float64:
		return int(number), number > 0 && number <= 65535
	default:
		return 0, false
	}
}

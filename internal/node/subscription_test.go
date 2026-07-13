package node

import (
	"encoding/base64"
	"testing"
)

func TestParseBase64Subscription(t *testing.T) {
	raw := "vless://abc@example.com:443?security=tls#Tokyo\nss://YWVzLTEyOC1nY206cGFzcw@example.org:8388#Edge\n"
	nodes, err := ParseSubscription([]byte(base64.StdEncoding.EncodeToString([]byte(raw))))
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 2 || nodes[0].Name != "Tokyo" || nodes[1].Name != "Edge" {
		t.Fatalf("unexpected nodes: %#v", nodes)
	}
}

func TestParseYAMLSubscription(t *testing.T) {
	raw := `proxies:
  - name: Tokyo
    type: ss
    server: example.com
    port: 8388
    cipher: aes-128-gcm
    password: secret
`
	nodes, err := ParseSubscription([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || nodes[0].Options["cipher"] != "aes-128-gcm" {
		t.Fatalf("unexpected nodes: %#v", nodes)
	}
}

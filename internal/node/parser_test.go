package node

import "testing"

func TestParseVLESS(t *testing.T) {
	n, err := Parse("vless://abc@example.com:443?security=tls&sni=example.com#Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	if n.Name != "Tokyo" || n.Type != "vless" || n.Options["uuid"] != "abc" {
		t.Fatalf("unexpected node: %#v", n)
	}
}

func TestParseSS(t *testing.T) {
	n, err := Parse("ss://YWVzLTEyOC1nY206cGFzcw@example.com:8388#Edge")
	if err != nil {
		t.Fatal(err)
	}
	if n.Options["cipher"] != "aes-128-gcm" {
		t.Fatalf("unexpected node: %#v", n)
	}
}

func TestParseSOCKS5(t *testing.T) {
	n, err := Parse("socks5://alice:secret@example.com:1080?udp=true#Office")
	if err != nil {
		t.Fatal(err)
	}
	if n.Type != "socks5" || n.Name != "Office" || n.Options["username"] != "alice" || n.Options["password"] != "secret" || n.Options["udp"] != true {
		t.Fatalf("unexpected node: %#v", n)
	}
}

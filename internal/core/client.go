package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	base, secret string
	http         *http.Client
}

func NewClient(address, secret string) *Client {
	return &Client{base: "http://" + address, secret: secret, http: &http.Client{Timeout: 12 * time.Second}}
}
func (c *Client) do(method, path string, payload any, out any) error {
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		b, e := json.Marshal(payload)
		if e != nil {
			return e
		}
		body = bytes.NewReader(b)
	}
	req, e := http.NewRequest(method, c.base+path, body)
	if e != nil {
		return e
	}
	req.Header.Set("Authorization", "Bearer "+c.secret)
	req.Header.Set("Content-Type", "application/json")
	resp, e := c.http.Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("controller returned %s", resp.Status)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
func (c *Client) Select(group, name string) error {
	return c.do(http.MethodPut, "/proxies/"+url.PathEscape(group), map[string]string{"name": name}, nil)
}
func (c *Client) Proxies() (map[string]any, error) {
	var out map[string]any
	err := c.do(http.MethodGet, "/proxies", nil, &out)
	return out, err
}
func (c *Client) Delay(name, target string, timeout int) (map[string]any, error) {
	q := url.Values{"url": {target}, "timeout": {fmt.Sprint(timeout)}}
	var out map[string]any
	err := c.do(http.MethodGet, "/proxies/"+url.PathEscape(name)+"/delay?"+q.Encode(), nil, &out)
	return out, err
}

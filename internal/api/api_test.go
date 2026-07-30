package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/foliageSea/docker-clash/internal/core"
	"github.com/foliageSea/docker-clash/internal/model"
	"github.com/foliageSea/docker-clash/internal/store"
	"github.com/gin-gonic/gin"
)

func testAPI(t *testing.T, state model.State) (*gin.Engine, *store.Store) {
	t.Helper()
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Update(func(current *model.State) error {
		*current = state
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	m := core.New(dir, filepath.Join(dir, "missing-mihomo"))
	r := gin.New()
	New(s, m).Register(r.Group("/api"))
	return r, s
}

func request(t *testing.T, r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCreateEntryGroupAppliesDefaultsAndStatusCountsIt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	state := model.DefaultState()
	state.Nodes = []model.Node{{ID: "n1", Name: "edge", Type: "ss", Server: "host", Port: 443}}
	r, _ := testAPI(t, state)
	w := request(t, r, http.MethodPost, "/api/entry-groups", `{"name":"auto","type":"fallback","nodeIds":["n1"]}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", w.Code, w.Body.String())
	}
	var group model.EntryGroup
	if err := json.Unmarshal(w.Body.Bytes(), &group); err != nil {
		t.Fatal(err)
	}
	if len(group.ID) != 32 || group.TestURL != model.DefaultEntryGroupTestURL || group.Interval != 60 {
		t.Fatalf("unexpected created group: %+v", group)
	}
	w = request(t, r, http.MethodGet, "/api/status", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"entryGroupCount":1`) {
		t.Fatalf("status response = %d, %s", w.Code, w.Body.String())
	}
}

func TestEntryGroupReferencesRenameSelectAndProtectDeletion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	state := model.DefaultState()
	state.Nodes = []model.Node{
		{ID: "n1", Name: "edge", Type: "ss", Server: "one", Port: 1},
		{ID: "n2", Name: "relay", Type: "ss", Server: "two", Port: 2, DialerProxy: "edge"},
		{ID: "n3", Name: "chained", Type: "ss", Server: "three", Port: 3, DialerProxy: "pick"},
	}
	state.EntryGroups = []model.EntryGroup{{ID: "g1", Name: "pick", Type: "select", NodeIDs: []string{"n1", "n2"}}}
	r, s := testAPI(t, state)

	w := request(t, r, http.MethodPut, "/api/nodes/n1", `{"name":"edge-new","type":"ss","server":"one","port":1}`)
	if w.Code != http.StatusOK || s.Get().Nodes[1].DialerProxy != "edge-new" {
		t.Fatalf("node rename response = %d, state = %+v", w.Code, s.Get().Nodes)
	}
	w = request(t, r, http.MethodPut, "/api/entry-groups/g1", `{"name":"pick-new","type":"select","nodeIds":["n1","n2"]}`)
	if w.Code != http.StatusOK || s.Get().Nodes[2].DialerProxy != "pick-new" || s.Get().EntryGroups[0].ID != "g1" {
		t.Fatalf("group rename response = %d, state = %+v", w.Code, s.Get())
	}
	w = request(t, r, http.MethodPost, "/api/entry-groups/g1/select", `{"nodeId":"n2"}`)
	if w.Code != http.StatusNoContent || s.Get().EntryGroups[0].SelectedNodeID != "n2" {
		t.Fatalf("stopped selection response = %d, state = %+v", w.Code, s.Get().EntryGroups[0])
	}
	w = request(t, r, http.MethodDelete, "/api/nodes/n1", "")
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "relay") {
		t.Fatalf("node deletion response = %d, %s", w.Code, w.Body.String())
	}
	w = request(t, r, http.MethodDelete, "/api/entry-groups/g1", "")
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "chained") {
		t.Fatalf("group deletion response = %d, %s", w.Code, w.Body.String())
	}
	w = request(t, r, http.MethodDelete, "/api/nodes", "")
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "entry groups") {
		t.Fatalf("clear nodes response = %d, %s", w.Code, w.Body.String())
	}
}

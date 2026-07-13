package api

import (
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/foliageSea/nexus-proxy-ui/internal/config"
	"github.com/foliageSea/nexus-proxy-ui/internal/core"
	"github.com/foliageSea/nexus-proxy-ui/internal/model"
	nodeparser "github.com/foliageSea/nexus-proxy-ui/internal/node"
	"github.com/foliageSea/nexus-proxy-ui/internal/store"
	"github.com/gin-gonic/gin"
)

type API struct {
	store  *store.Store
	core   *core.Manager
	client *core.Client
}

func New(s *store.Store, m *core.Manager) *API {
	address, secret := m.Controller()
	return &API{store: s, core: m, client: core.NewClient(address, secret)}
}
func (a *API) Register(r *gin.RouterGroup) {
	r.GET("/status", a.status)
	r.GET("/nodes", a.nodes)
	r.POST("/nodes/import", a.importNode)
	r.PUT("/nodes/:id", a.updateNode)
	r.DELETE("/nodes/:id", a.deleteNode)
	r.POST("/nodes/:id/select", a.selectNode)
	r.POST("/nodes/:id/delay", a.delay)
	r.GET("/settings", a.settings)
	r.PUT("/settings", a.updateSettings)
	r.POST("/core/:action", a.coreAction)
	r.GET("/core/log", a.log)
}
func fail(c *gin.Context, status int, err error) { c.JSON(status, gin.H{"error": err.Error()}) }
func (a *API) apply(c *gin.Context, fn func(*model.State) error) bool {
	var updated model.State
	err := a.store.Update(func(s *model.State) error {
		if err := fn(s); err != nil {
			return err
		}
		if err := config.Validate(*s); err != nil {
			return err
		}
		updated = *s
		return nil
	})
	if err != nil {
		fail(c, 400, err)
		return false
	}
	if err := a.core.Apply(updated); err != nil {
		fail(c, 500, err)
		return false
	}
	return true
}
func (a *API) status(c *gin.Context) {
	s := a.store.Get()
	c.JSON(200, gin.H{"core": a.core.Status(), "settings": s.Settings, "nodeCount": len(s.Nodes)})
}
func (a *API) nodes(c *gin.Context) { c.JSON(200, a.store.Get().Nodes) }
func (a *API) importNode(c *gin.Context) {
	var req struct {
		URI string `json:"uri" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, err)
		return
	}
	nodes, err := nodeparser.Import(req.URI)
	if err != nil {
		fail(c, 400, err)
		return
	}
	if !a.apply(c, func(s *model.State) error {
		existing := make(map[string]bool, len(s.Nodes))
		for _, current := range s.Nodes {
			existing[current.Name] = true
		}
		for _, imported := range nodes {
			if existing[imported.Name] {
				return errors.New("a node with the name " + imported.Name + " already exists")
			}
			existing[imported.Name] = true
		}
		s.Nodes = append(s.Nodes, nodes...)
		return nil
	}) {
		return
	}
	c.JSON(201, gin.H{"count": len(nodes), "nodes": nodes})
}
func (a *API) updateNode(c *gin.Context) {
	var req model.Node
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, err)
		return
	}
	id := c.Param("id")
	if !a.apply(c, func(s *model.State) error {
		for i := range s.Nodes {
			if s.Nodes[i].ID == id {
				req.ID = id
				req.CreatedAt = s.Nodes[i].CreatedAt
				s.Nodes[i] = req
				return nil
			}
		}
		return errors.New("node not found")
	}) {
		return
	}
	c.JSON(200, req)
}
func (a *API) deleteNode(c *gin.Context) {
	id := c.Param("id")
	if !a.apply(c, func(s *model.State) error {
		idx := slices.IndexFunc(s.Nodes, func(n model.Node) bool { return n.ID == id })
		if idx < 0 {
			return errors.New("node not found")
		}
		name := s.Nodes[idx].Name
		for _, n := range s.Nodes {
			if n.DialerProxy == name {
				return errors.New("node is used by a proxy chain")
			}
		}
		s.Nodes = append(s.Nodes[:idx], s.Nodes[idx+1:]...)
		if s.Settings.SelectedNode == name {
			s.Settings.SelectedNode = ""
		}
		return nil
	}) {
		return
	}
	c.Status(204)
}
func (a *API) selectNode(c *gin.Context) {
	id := c.Param("id")
	s := a.store.Get()
	idx := slices.IndexFunc(s.Nodes, func(n model.Node) bool { return n.ID == id })
	if idx < 0 {
		fail(c, 404, errors.New("node not found"))
		return
	}
	name := s.Nodes[idx].Name
	if a.core.Status().Running {
		if err := a.client.Select("NEXUS", name); err != nil {
			fail(c, 502, err)
			return
		}
	}
	_ = a.store.Update(func(st *model.State) error { st.Settings.SelectedNode = name; return nil })
	c.Status(204)
}
func (a *API) delay(c *gin.Context) {
	id := c.Param("id")
	s := a.store.Get()
	idx := slices.IndexFunc(s.Nodes, func(n model.Node) bool { return n.ID == id })
	if idx < 0 {
		fail(c, 404, errors.New("node not found"))
		return
	}
	out, err := a.client.Delay(s.Nodes[idx].Name, "https://www.gstatic.com/generate_204", 5000)
	if err != nil {
		fail(c, 502, err)
		return
	}
	c.JSON(200, out)
}
func (a *API) settings(c *gin.Context) { c.JSON(200, a.store.Get().Settings) }
func (a *API) updateSettings(c *gin.Context) {
	var req model.Settings
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, err)
		return
	}
	if !a.apply(c, func(s *model.State) error { req.SelectedNode = s.Settings.SelectedNode; s.Settings = req; return nil }) {
		return
	}
	c.JSON(200, req)
}
func (a *API) coreAction(c *gin.Context) {
	var err error
	switch c.Param("action") {
	case "start":
		err = a.core.Start()
	case "stop":
		err = a.core.Stop()
	case "restart":
		err = a.core.Restart()
	default:
		fail(c, 404, errors.New("unknown action"))
		return
	}
	if err != nil {
		fail(c, 500, err)
		return
	}
	time.Sleep(100 * time.Millisecond)
	c.JSON(200, a.core.Status())
}
func (a *API) log(c *gin.Context) {
	log, err := a.core.TailLog(64 << 10)
	if err != nil {
		fail(c, 500, err)
		return
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(log))
}

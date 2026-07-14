package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	coreconfig "github.com/foliageSea/nexus-proxy-ui/internal/config"
	"github.com/foliageSea/nexus-proxy-ui/internal/model"
	"gopkg.in/yaml.v3"
)

type Status struct {
	Running bool   `json:"running"`
	PID     int    `json:"pid,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Manager struct {
	mu                                  sync.Mutex
	dataDir, binary, controller, secret string
	cmd                                 *exec.Cmd
	lastError                           string
}

func New(dataDir, binary string) *Manager {
	return &Manager{dataDir: dataDir, binary: binary, controller: "127.0.0.1:19090", secret: loadControllerSecret(dataDir)}
}

func loadControllerSecret(dataDir string) string {
	secretPath := filepath.Join(dataDir, "controller-secret")
	if b, err := os.ReadFile(secretPath); err == nil {
		if secret := strings.TrimSpace(string(b)); secret != "" {
			return secret
		}
	}

	var existing struct {
		Secret string `yaml:"secret"`
	}
	if b, err := os.ReadFile(filepath.Join(dataDir, "config.yaml")); err == nil {
		_ = yaml.Unmarshal(b, &existing)
	}
	secret := strings.TrimSpace(existing.Secret)
	if secret == "" {
		b := make([]byte, 24)
		_, _ = rand.Read(b)
		secret = hex.EncodeToString(b)
	}
	_ = coreconfig.Write(secretPath, []byte(secret))
	return secret
}
func (m *Manager) Controller() (string, string) { return m.controller, m.secret }

func (m *Manager) Apply(state model.State) error {
	b, err := coreconfig.Render(state, m.controller, m.secret)
	if err != nil {
		return err
	}
	path := filepath.Join(m.dataDir, "config.yaml")
	if err := coreconfig.Write(path, b); err != nil {
		return err
	}
	if _, err := os.Stat(m.binary); err != nil {
		return nil
	}
	return m.Restart()
}

func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cmd != nil && m.cmd.Process != nil {
		return nil
	}
	if _, err := os.Stat(m.binary); err != nil {
		m.lastError = "mihomo binary not found at " + m.binary
		return fmt.Errorf("%s", m.lastError)
	}
	args := []string{"-d", m.dataDir, "-f", filepath.Join(m.dataDir, "config.yaml")}
	cmd := exec.Command(m.binary, args...)
	logFile, err := os.OpenFile(filepath.Join(m.dataDir, "core.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	cmd.Stdout, cmd.Stderr = logFile, logFile
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		m.lastError = err.Error()
		return err
	}
	m.cmd, m.lastError = cmd, ""
	go func() {
		err := cmd.Wait()
		_ = logFile.Close()
		m.mu.Lock()
		if m.cmd == cmd {
			m.cmd = nil
			if err != nil {
				m.lastError = err.Error()
			}
		}
		m.mu.Unlock()
	}()
	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	cmd := m.cmd
	m.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	if runtime.GOOS != "windows" {
		_ = cmd.Process.Signal(os.Interrupt)
	} else {
		_ = cmd.Process.Kill()
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		m.mu.Lock()
		stopped := m.cmd != cmd
		m.mu.Unlock()
		if stopped {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	_ = cmd.Process.Kill()
	return nil
}

func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil {
		return err
	}
	return m.Start()
}
func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := Status{Running: m.cmd != nil, Error: m.lastError}
	if m.cmd != nil && m.cmd.Process != nil {
		s.PID = m.cmd.Process.Pid
	}
	return s
}

func (m *Manager) TailLog(max int64) (string, error) {
	f, err := os.Open(filepath.Join(m.dataDir, "core.log"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return "", err
	}
	start := st.Size() - max
	if start < 0 {
		start = 0
	}
	_, _ = f.Seek(start, io.SeekStart)
	b, err := io.ReadAll(f)
	return string(b), err
}

func (m *Manager) Shutdown(ctx context.Context) error {
	done := make(chan error, 1)
	go func() { done <- m.Stop() }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

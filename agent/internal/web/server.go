package web

import (
	"encoding/json"
	"net/http"

	"github.com/aschorle/kiosk-client/agent/internal/config"
	"github.com/aschorle/kiosk-client/agent/internal/status"
)

// Server wraps the local kiosk-agent HTTP server.
type Server struct {
	addr     string
	provider status.Provider
}

// Route describes one registered HTTP route.
type Route struct {
	Method string
	Path   string
}

// NewServer creates the HTTP server for the local kiosk-agent API.
func NewServer(addr string, provider status.Provider) Server {
	return Server{
		addr:     addr,
		provider: provider,
	}
}

// Routes returns the routes registered by the HTTP server.
func (s Server) Routes() []Route {
	return []Route{
		{Method: http.MethodGet, Path: "/"},
		{Method: http.MethodGet, Path: "/api/config"},
		{Method: http.MethodGet, Path: "/api/status"},
	}
}

// ListenAndServe starts the HTTP server.
func (s Server) ListenAndServe() error {
	server := &http.Server{
		Addr:    s.addr,
		Handler: s.mux(),
	}

	return server.ListenAndServe()
}

func (s Server) mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/status", s.handleStatus)

	return mux
}

func (s Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("kiosk-agent running"))
}

func (s Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	cfg, err := config.Current()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	if err := json.NewEncoder(w).Encode(cfg); err != nil {
		http.Error(w, "failed to encode config", http.StatusInternalServerError)
		return
	}
}

func (s Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(s.provider.Current()); err != nil {
		http.Error(w, "failed to encode status", http.StatusInternalServerError)
		return
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

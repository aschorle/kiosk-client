package web

import (
	"encoding/json"
	"net/http"

	"github.com/aschorle/kiosk-client/agent/internal/status"
)

// Server wraps the local kiosk-agent HTTP server.
type Server struct {
	addr     string
	provider status.Provider
}

// NewServer creates the HTTP server for the local kiosk-agent API.
func NewServer(addr string, provider status.Provider) Server {
	return Server{
		addr:     addr,
		provider: provider,
	}
}

// ListenAndServe starts the HTTP server.
func (s Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleRoot)
	mux.HandleFunc("GET /api/status", s.handleStatus)

	server := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	return server.ListenAndServe()
}

func (s Server) handleRoot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("kiosk-agent running"))
}

func (s Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(s.provider.Current()); err != nil {
		http.Error(w, "failed to encode status", http.StatusInternalServerError)
		return
	}
}

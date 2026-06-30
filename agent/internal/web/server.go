package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/aschorle/kiosk-client/agent/internal/browser"
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
		{Method: http.MethodPut, Path: "/api/config"},
		{Method: http.MethodGet, Path: "/api/health"},
		{Method: http.MethodGet, Path: "/api/info"},
		{Method: http.MethodGet, Path: "/api/metrics"},
		{Method: http.MethodGet, Path: "/api/status"},
		{Method: http.MethodPost, Path: "/api/browser/reload"},
		{Method: http.MethodPost, Path: "/api/browser/restart"},
	}
}

// ListenAndServe starts the HTTP server.
func (s Server) ListenAndServe() error {
	server := &http.Server{
		Addr:    s.addr,
		Handler: countRequests(authenticateWrites(s.mux())),
	}

	return server.ListenAndServe()
}

func (s Server) mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/info", s.handleInfo)
	mux.HandleFunc("/api/metrics", s.handleMetrics)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/browser/reload", s.handleBrowserReload)
	mux.HandleFunc("/api/browser/restart", s.handleBrowserRestart)

	return mux
}

func countRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status.IncrementHTTPRequest()
		next.ServeHTTP(w, r)
	})
}

func authenticateWrites(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		token := config.AuthToken()
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		if r.Header.Get("Authorization") != "Bearer "+token {
			writeError(w, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}

		next.ServeHTTP(w, r)
	})
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
	switch r.Method {
	case http.MethodGet:
		s.handleConfigGet(w, r)
	case http.MethodPut:
		s.handleConfigPut(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s Server) handleConfigGet(w http.ResponseWriter, r *http.Request) {
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

func (s Server) handleConfigPut(w http.ResponseWriter, r *http.Request) {
	var cfg config.Config
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, errors.New("request body must contain a single JSON object"))
		return
	}

	if err := config.Update(cfg); err != nil {
		var validationError config.ValidationError
		if errors.As(err, &validationError) {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := browser.Restart(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeOK(w)
}

func (s Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(s.provider.Health()); err != nil {
		http.Error(w, "failed to encode health", http.StatusInternalServerError)
		return
	}
}

func (s Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(s.provider.Info()); err != nil {
		http.Error(w, "failed to encode info", http.StatusInternalServerError)
		return
	}
}

func (s Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(s.provider.Metrics()); err != nil {
		http.Error(w, "failed to encode metrics", http.StatusInternalServerError)
		return
	}
}

func (s Server) handleBrowserRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := browser.RestartService(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeOK(w)
}

func (s Server) handleBrowserReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := browser.ReloadService(); err != nil {
		if errors.Is(err, browser.ErrReloadNotSupported) {
			writeError(w, http.StatusNotImplemented, err)
			return
		}

		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeOK(w)
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

type statusResponse struct {
	Status string `json:"status"`
}

func writeOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(statusResponse{Status: "ok"})
}

func writeError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
}

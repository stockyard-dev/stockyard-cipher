package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/stockyard-dev/stockyard-cipher/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	port   int
	limits Limits
}

func New(db *store.DB, port int, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), port: port, limits: limits}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Projects
	s.mux.HandleFunc("POST /api/projects", s.handleCreateProject)
	s.mux.HandleFunc("GET /api/projects", s.handleListProjects)
	s.mux.HandleFunc("GET /api/projects/{id}", s.handleGetProject)
	s.mux.HandleFunc("DELETE /api/projects/{id}", s.handleDeleteProject)

	// Secrets
	s.mux.HandleFunc("PUT /api/projects/{id}/secrets/{key}", s.handleSetSecret)
	s.mux.HandleFunc("GET /api/projects/{id}/secrets/{key}", s.handleGetSecret)
	s.mux.HandleFunc("GET /api/projects/{id}/secrets", s.handleListSecrets)
	s.mux.HandleFunc("DELETE /api/projects/{id}/secrets/{key}", s.handleDeleteSecret)

	// Tokens
	s.mux.HandleFunc("POST /api/projects/{id}/tokens", s.handleCreateToken)
	s.mux.HandleFunc("GET /api/projects/{id}/tokens", s.handleListTokens)
	s.mux.HandleFunc("DELETE /api/tokens/{id}", s.handleRevokeToken)

	// Token-based access (the CI/CD hot path)
	s.mux.HandleFunc("GET /api/secrets", s.handleTokenSecrets)

	// Audit log
	s.mux.HandleFunc("GET /api/projects/{id}/audit", s.handleAuditLog)

	// Status
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /ui", s.handleUI)

	s.mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"product": "stockyard-cipher", "version": "0.1.0"})
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("[cipher] listening on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.Split(fwd, ",")[0]
	}
	return r.RemoteAddr
}

// --- Project handlers ---

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "name is required"})
		return
	}
	if s.limits.MaxProjects > 0 {
		projects, _ := s.db.ListProjects()
		if LimitReached(s.limits.MaxProjects, len(projects)) {
			writeJSON(w, 402, map[string]string{
				"error":   fmt.Sprintf("free tier limit: %d projects max — upgrade to Pro", s.limits.MaxProjects),
				"upgrade": "https://stockyard.dev/cipher/",
			})
			return
		}
	}
	p, err := s.db.CreateProject(req.Name, req.Description)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"project": p})
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.db.ListProjects()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if projects == nil {
		projects = []store.Project{}
	}
	writeJSON(w, 200, map[string]any{"projects": projects, "count": len(projects)})
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	p, err := s.db.GetProject(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "project not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"project": p})
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.db.GetProject(id); err != nil {
		writeJSON(w, 404, map[string]string{"error": "project not found"})
		return
	}
	s.db.DeleteProject(id)
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// --- Secret handlers ---

func (s *Server) handleSetSecret(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	key := r.PathValue("key")

	if _, err := s.db.GetProject(projectID); err != nil {
		writeJSON(w, 404, map[string]string{"error": "project not found"})
		return
	}

	var req struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}

	if s.limits.MaxSecrets > 0 {
		total := s.db.TotalSecrets()
		if LimitReached(s.limits.MaxSecrets, total) {
			// Check if this is an update (doesn't count against limit)
			if _, err := s.db.GetSecret(projectID, key); err != nil {
				writeJSON(w, 402, map[string]string{
					"error":   fmt.Sprintf("free tier limit: %d secrets max — upgrade to Pro", s.limits.MaxSecrets),
					"upgrade": "https://stockyard.dev/cipher/",
				})
				return
			}
		}
	}

	sec, err := s.db.SetSecret(projectID, key, req.Value)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	s.db.LogAccess(projectID, "", "set", key, clientIP(r))
	writeJSON(w, 200, map[string]any{"secret": sec})
}

func (s *Server) handleGetSecret(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	key := r.PathValue("key")

	sec, err := s.db.GetSecret(projectID, key)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "secret not found"})
		return
	}

	s.db.LogAccess(projectID, "", "get", key, clientIP(r))
	writeJSON(w, 200, map[string]any{"key": sec.Key, "value": sec.Value, "version": sec.Version})
}

func (s *Server) handleListSecrets(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	secrets, err := s.db.ListSecrets(projectID)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if secrets == nil {
		secrets = []store.Secret{}
	}
	// List returns keys without values
	writeJSON(w, 200, map[string]any{"secrets": secrets, "count": len(secrets)})
}

func (s *Server) handleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	key := r.PathValue("key")
	s.db.DeleteSecret(projectID, key)
	s.db.LogAccess(projectID, "", "delete", key, clientIP(r))
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// --- Token handlers ---

func (s *Server) handleCreateToken(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	if _, err := s.db.GetProject(projectID); err != nil {
		writeJSON(w, 404, map[string]string{"error": "project not found"})
		return
	}

	var req struct {
		Name       string `json:"name"`
		Scopes     string `json:"scopes"`
		TTLMinutes int    `json:"ttl_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.TTLMinutes <= 0 {
		req.TTLMinutes = 60 // 1 hour default
	}

	tok, rawToken, err := s.db.CreateToken(projectID, req.Name, req.Scopes, req.TTLMinutes)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	s.db.LogAccess(projectID, tok.ID, "token_created", "", clientIP(r))
	writeJSON(w, 201, map[string]any{
		"token":   tok,
		"raw_token": rawToken,
		"warning": "Save this token now — it cannot be retrieved again",
	})
}

func (s *Server) handleListTokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := s.db.ListTokens(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if tokens == nil {
		tokens = []store.Token{}
	}
	writeJSON(w, 200, map[string]any{"tokens": tokens, "count": len(tokens)})
}

func (s *Server) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	s.db.RevokeToken(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"status": "revoked"})
}

// --- Token-based secret access (CI/CD hot path) ---

func (s *Server) handleTokenSecrets(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization: Bearer <token>
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		writeJSON(w, 401, map[string]string{"error": "Authorization: Bearer <token> required"})
		return
	}
	rawToken := strings.TrimPrefix(auth, "Bearer ")

	tok, err := s.db.ValidateToken(rawToken)
	if err != nil {
		writeJSON(w, 401, map[string]string{"error": err.Error()})
		return
	}

	secrets, err := s.db.GetAllSecrets(tok.ProjectID)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	s.db.LogAccess(tok.ProjectID, tok.ID, "token_read_all", "", clientIP(r))

	// Check if they want env format
	if r.URL.Query().Get("format") == "env" {
		w.Header().Set("Content-Type", "text/plain")
		for k, v := range secrets {
			fmt.Fprintf(w, "%s=%s\n", k, v)
		}
		return
	}

	// Check if they want a specific key
	if key := r.URL.Query().Get("key"); key != "" {
		if val, ok := secrets[key]; ok {
			s.db.LogAccess(tok.ProjectID, tok.ID, "token_read", key, clientIP(r))
			writeJSON(w, 200, map[string]any{"key": key, "value": val})
			return
		}
		writeJSON(w, 404, map[string]string{"error": "secret not found"})
		return
	}

	writeJSON(w, 200, map[string]any{"secrets": secrets, "project_id": tok.ProjectID})
}

// --- Audit log ---

func (s *Server) handleAuditLog(w http.ResponseWriter, r *http.Request) {
	if !s.limits.AuditLog {
		writeJSON(w, 402, map[string]string{
			"error":   "audit log requires Pro — upgrade at https://stockyard.dev/cipher/",
			"upgrade": "https://stockyard.dev/cipher/",
		})
		return
	}
	entries, err := s.db.ListAccessLog(r.PathValue("id"), 200)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if entries == nil {
		entries = []store.AccessEntry{}
	}
	writeJSON(w, 200, map[string]any{"log": entries, "count": len(entries)})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.db.Stats())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

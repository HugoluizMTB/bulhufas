package mcp

import (
	"encoding/json"
	"net/http"

	"github.com/HugoluizMTB/bulhufas/internal/domain"
	"github.com/HugoluizMTB/bulhufas/internal/store"
)

type Server struct {
	handler *Handler
	store   store.Store
	mux     *http.ServeMux
}

func NewServer(h *Handler, s store.Store) *Server {
	srv := &Server{handler: h, store: s, mux: http.NewServeMux()}
	srv.routes()
	return srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("POST /api/conversations", s.handleSaveConversation)
	s.mux.HandleFunc("POST /api/search", s.handleSearch)
	s.mux.HandleFunc("GET /api/chunks", s.handleListChunks)
	s.mux.HandleFunc("PATCH /api/chunks/{id}/status", s.handleUpdateStatus)
	s.mux.HandleFunc("DELETE /api/chunks/{id}", s.handleDeleteChunk)
	s.mux.HandleFunc("GET /api/actions", s.handleListActions)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

type saveConversationRequest struct {
	Source       string         `json:"source"`
	RawText      string         `json:"raw_text"`
	Summary      string         `json:"summary"`
	Participants []string       `json:"participants"`
	Date         string         `json:"date"`
	Chunks       []domain.Chunk `json:"chunks"`
}

func (s *Server) handleSaveConversation(w http.ResponseWriter, r *http.Request) {
	var req saveConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conv := &domain.Conversation{
		Source:       req.Source,
		RawText:      req.RawText,
		Summary:      req.Summary,
		Participants: req.Participants,
	}

	if err := s.handler.SaveConversation(r.Context(), conv, req.Chunks); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"conversation_id": conv.ID,
		"chunks_saved":    len(req.Chunks),
	})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	var query domain.SearchQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results, err := s.handler.Search(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func (s *Server) handleListChunks(w http.ResponseWriter, r *http.Request) {
	filters := store.ChunkFilters{
		Type:   domain.ChunkType(r.URL.Query().Get("type")),
		Status: domain.Status(r.URL.Query().Get("status")),
		Limit:  50,
	}

	chunks, err := s.store.ListChunks(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, chunks)
}

type updateStatusRequest struct {
	Status domain.Status `json:"status"`
}

func (s *Server) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.handler.UpdateChunkStatus(r.Context(), id, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) handleDeleteChunk(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := s.handler.DeleteChunk(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleListActions(w http.ResponseWriter, r *http.Request) {
	allChunks, err := s.store.ListChunks(r.Context(), store.ChunkFilters{
		Status: domain.StatusPending,
		Limit:  100,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var actions []domain.Chunk
	for _, c := range allChunks {
		if c.ActionItem != "" {
			actions = append(actions, c)
		}
	}

	writeJSON(w, http.StatusOK, actions)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

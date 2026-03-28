package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/HugoluizMTB/bulhufas/internal/domain"
	"github.com/HugoluizMTB/bulhufas/internal/store"
)

type mockEmbedder struct{}

func (m *mockEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	vec := make([]float32, 8)
	for i := range vec {
		if i < len(text) {
			vec[i] = float32(text[i]) / 255.0
		}
	}
	return vec, nil
}

func (m *mockEmbedder) EmbedBatch(_ context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, t := range texts {
		v, _ := m.Embed(context.Background(), t)
		results[i] = v
	}
	return results, nil
}

type mockVectorStore struct {
	indexed map[string][]float32
}

func newMockVectorStore() *mockVectorStore {
	return &mockVectorStore{indexed: make(map[string][]float32)}
}

func (m *mockVectorStore) Index(_ context.Context, id string, emb []float32, _ map[string]any) error {
	m.indexed[id] = emb
	return nil
}

func (m *mockVectorStore) Search(_ context.Context, _ []float32, limit int) ([]domain.SearchResult, error) {
	return nil, nil
}

func (m *mockVectorStore) Delete(_ context.Context, id string) error {
	delete(m.indexed, id)
	return nil
}

type mockStore struct {
	conversations map[string]*domain.Conversation
	chunks        map[string]*domain.Chunk
}

func newMockStore() *mockStore {
	return &mockStore{
		conversations: make(map[string]*domain.Conversation),
		chunks:        make(map[string]*domain.Chunk),
	}
}

func (m *mockStore) SaveConversation(_ context.Context, conv *domain.Conversation) error {
	m.conversations[conv.ID] = conv
	return nil
}

func (m *mockStore) GetConversation(_ context.Context, id string) (*domain.Conversation, error) {
	return m.conversations[id], nil
}

func (m *mockStore) ListConversations(_ context.Context, _, _ int) ([]domain.Conversation, error) {
	return nil, nil
}

func (m *mockStore) SaveChunk(_ context.Context, chunk *domain.Chunk) error {
	m.chunks[chunk.ID] = chunk
	return nil
}

func (m *mockStore) GetChunk(_ context.Context, id string) (*domain.Chunk, error) {
	return m.chunks[id], nil
}

func (m *mockStore) UpdateChunk(_ context.Context, chunk *domain.Chunk) error {
	m.chunks[chunk.ID] = chunk
	return nil
}

func (m *mockStore) DeleteChunk(_ context.Context, id string) error {
	delete(m.chunks, id)
	return nil
}

func (m *mockStore) ListChunks(_ context.Context, _ store.ChunkFilters) ([]domain.Chunk, error) {
	return nil, nil
}

func (m *mockStore) SaveWorkItem(_ context.Context, _ *domain.WorkItem) error   { return nil }
func (m *mockStore) GetWorkItem(_ context.Context, _ string) (*domain.WorkItem, error) {
	return nil, nil
}
func (m *mockStore) UpdateWorkItem(_ context.Context, _ *domain.WorkItem) error { return nil }
func (m *mockStore) DeleteWorkItem(_ context.Context, _ string) error           { return nil }
func (m *mockStore) ListWorkItems(_ context.Context, _ store.WorkItemFilters) ([]domain.WorkItem, error) {
	return nil, nil
}
func (m *mockStore) SaveRelation(_ context.Context, _ *domain.Relation) error { return nil }
func (m *mockStore) DeleteRelation(_ context.Context, _ string) error         { return nil }
func (m *mockStore) GetRelations(_ context.Context, _ string) ([]domain.Relation, error) {
	return nil, nil
}

func TestSaveConversation(t *testing.T) {
	s := newMockStore()
	v := newMockVectorStore()
	e := &mockEmbedder{}
	h := NewHandler(s, v, e)

	conv := &domain.Conversation{
		ID:           "conv-1",
		Source:       "whatsapp",
		RawText:      "test conversation",
		Participants: []string{"hugo", "renan"},
		Date:         time.Now(),
		CreatedAt:    time.Now(),
	}

	chunks := []domain.Chunk{
		{
			ID:      "chunk-1",
			Content: "Renan quer acesso read-only ao Postgres",
			Type:    domain.ChunkDecision,
			Status:  domain.StatusPending,
			People:  []string{"renan"},
			Tags:    []string{"infra", "postgres"},
		},
		{
			ID:      "chunk-2",
			Content: "Criar credencial read-only para Renan",
			Type:    domain.ChunkActionItem,
			Status:  domain.StatusPending,
			People:  []string{"hugo"},
			Tags:    []string{"infra"},
		},
	}

	err := h.SaveConversation(context.Background(), conv, chunks)
	if err != nil {
		t.Fatalf("SaveConversation() error = %v", err)
	}

	if len(s.conversations) != 1 {
		t.Errorf("expected 1 conversation, got %d", len(s.conversations))
	}

	if len(s.chunks) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(s.chunks))
	}

	if len(v.indexed) != 2 {
		t.Errorf("expected 2 vectors indexed, got %d", len(v.indexed))
	}

	for _, c := range s.chunks {
		if c.ConversationID != "conv-1" {
			t.Errorf("chunk.ConversationID = %q, want %q", c.ConversationID, "conv-1")
		}
	}
}

func TestDeleteChunk(t *testing.T) {
	s := newMockStore()
	v := newMockVectorStore()
	e := &mockEmbedder{}
	h := NewHandler(s, v, e)

	s.chunks["chunk-1"] = &domain.Chunk{ID: "chunk-1"}
	v.indexed["chunk-1"] = []float32{1, 2, 3}

	err := h.DeleteChunk(context.Background(), "chunk-1")
	if err != nil {
		t.Fatalf("DeleteChunk() error = %v", err)
	}

	if _, exists := s.chunks["chunk-1"]; exists {
		t.Error("chunk should be deleted from store")
	}

	if _, exists := v.indexed["chunk-1"]; exists {
		t.Error("chunk should be deleted from vector store")
	}
}

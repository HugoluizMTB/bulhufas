package mcp

import (
	"context"

	"github.com/HugoluizMTB/bulhufas/internal/domain"
	"github.com/HugoluizMTB/bulhufas/internal/embedder"
	"github.com/HugoluizMTB/bulhufas/internal/store"
	"github.com/HugoluizMTB/bulhufas/internal/vectorstore"
)

type Handler struct {
	store       store.Store
	vectors     vectorstore.VectorStore
	embedder    embedder.Embedder
}

func NewHandler(s store.Store, v vectorstore.VectorStore, e embedder.Embedder) *Handler {
	return &Handler{store: s, vectors: v, embedder: e}
}

func (h *Handler) SaveConversation(ctx context.Context, conv *domain.Conversation, chunks []domain.Chunk) error {
	if err := h.store.SaveConversation(ctx, conv); err != nil {
		return err
	}

	for i := range chunks {
		chunks[i].ConversationID = conv.ID

		emb, err := h.embedder.Embed(ctx, chunks[i].Content)
		if err != nil {
			return err
		}
		chunks[i].Embedding = emb

		if err := h.store.SaveChunk(ctx, &chunks[i]); err != nil {
			return err
		}

		metadata := map[string]any{
			"type":   string(chunks[i].Type),
			"status": string(chunks[i].Status),
		}
		if err := h.vectors.Index(ctx, chunks[i].ID, emb, metadata); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) Search(ctx context.Context, query domain.SearchQuery) ([]domain.SearchResult, error) {
	if query.Limit == 0 {
		query.Limit = 10
	}

	emb, err := h.embedder.Embed(ctx, query.Text)
	if err != nil {
		return nil, err
	}

	vectorResults, err := h.vectors.Search(ctx, emb, query.Limit)
	if err != nil {
		return nil, err
	}

	var results []domain.SearchResult
	for _, vr := range vectorResults {
		chunk, err := h.store.GetChunk(ctx, vr.Chunk.ID)
		if err != nil {
			continue
		}
		results = append(results, domain.SearchResult{
			Chunk: *chunk,
			Score: vr.Score,
		})
	}

	return results, nil
}

func (h *Handler) UpdateChunkStatus(ctx context.Context, id string, status domain.Status) error {
	chunk, err := h.store.GetChunk(ctx, id)
	if err != nil {
		return err
	}
	chunk.Status = status
	return h.store.UpdateChunk(ctx, chunk)
}

func (h *Handler) DeleteChunk(ctx context.Context, id string) error {
	if err := h.vectors.Delete(ctx, id); err != nil {
		return err
	}
	return h.store.DeleteChunk(ctx, id)
}

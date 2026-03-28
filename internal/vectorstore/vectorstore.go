package vectorstore

import (
	"context"

	"github.com/HugoluizMTB/bulhufas/internal/domain"
)

type VectorStore interface {
	Index(ctx context.Context, id string, embedding []float32, metadata map[string]any) error
	Search(ctx context.Context, query []float32, limit int) ([]domain.SearchResult, error)
	Delete(ctx context.Context, id string) error
}

package vectorstore

import (
	"context"
	"fmt"

	"github.com/HugoluizMTB/bulhufas/internal/domain"
	"github.com/HugoluizMTB/bulhufas/internal/embedder"
	chromem "github.com/philippgille/chromem-go"
)

type ChromemStore struct {
	db         *chromem.DB
	collection *chromem.Collection
	emb        embedder.Embedder
}

func NewChromem(persistDir string, emb embedder.Embedder) (*ChromemStore, error) {
	db, err := chromem.NewPersistentDB(persistDir, false)
	if err != nil {
		return nil, fmt.Errorf("creating chromem db: %w", err)
	}

	embFunc := chromem.EmbeddingFunc(func(ctx context.Context, text string) ([]float32, error) {
		return emb.Embed(ctx, text)
	})

	col, err := db.GetOrCreateCollection("chunks", nil, embFunc)
	if err != nil {
		return nil, fmt.Errorf("creating collection: %w", err)
	}

	return &ChromemStore{db: db, collection: col, emb: emb}, nil
}

func (c *ChromemStore) Index(ctx context.Context, id string, embedding []float32, metadata map[string]any) error {
	strMeta := make(map[string]string)
	for k, v := range metadata {
		strMeta[k] = fmt.Sprintf("%v", v)
	}

	return c.collection.AddDocument(ctx, chromem.Document{
		ID:        id,
		Embedding: embedding,
		Metadata:  strMeta,
		Content:   strMeta["content"],
	})
}

func (c *ChromemStore) Search(ctx context.Context, query []float32, limit int) ([]domain.SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	count := c.collection.Count()
	if count == 0 {
		return nil, nil
	}
	if limit > count {
		limit = count
	}

	results, err := c.collection.QueryEmbedding(ctx, query, limit, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("querying chromem: %w", err)
	}

	var searchResults []domain.SearchResult
	for _, r := range results {
		searchResults = append(searchResults, domain.SearchResult{
			Chunk: domain.Chunk{
				ID:      r.ID,
				Content: r.Content,
			},
			Score: r.Similarity,
		})
	}

	return searchResults, nil
}

func (c *ChromemStore) Delete(_ context.Context, id string) error {
	return c.collection.Delete(context.Background(), nil, nil, id)
}

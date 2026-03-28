package store

import (
	"context"

	"github.com/HugoluizMTB/bulhufas/internal/domain"
)

// Store defines the persistence interface for all entities.
type Store interface {
	// Conversations
	SaveConversation(ctx context.Context, conv *domain.Conversation) error
	GetConversation(ctx context.Context, id string) (*domain.Conversation, error)
	ListConversations(ctx context.Context, limit, offset int) ([]domain.Conversation, error)

	// Chunks
	SaveChunk(ctx context.Context, chunk *domain.Chunk) error
	GetChunk(ctx context.Context, id string) (*domain.Chunk, error)
	UpdateChunk(ctx context.Context, chunk *domain.Chunk) error
	DeleteChunk(ctx context.Context, id string) error
	ListChunks(ctx context.Context, filters ChunkFilters) ([]domain.Chunk, error)

	// Work Items
	SaveWorkItem(ctx context.Context, item *domain.WorkItem) error
	GetWorkItem(ctx context.Context, id string) (*domain.WorkItem, error)
	UpdateWorkItem(ctx context.Context, item *domain.WorkItem) error
	DeleteWorkItem(ctx context.Context, id string) error
	ListWorkItems(ctx context.Context, filters WorkItemFilters) ([]domain.WorkItem, error)

	// Relations
	SaveRelation(ctx context.Context, rel *domain.Relation) error
	DeleteRelation(ctx context.Context, id string) error
	GetRelations(ctx context.Context, entityID string) ([]domain.Relation, error)
}

// ChunkFilters holds optional filters for chunk queries.
type ChunkFilters struct {
	Type           domain.ChunkType
	Status         domain.Status
	ConversationID string
	People         []string
	Tags           []string
	Limit          int
	Offset         int
}

// WorkItemFilters holds optional filters for work item queries.
type WorkItemFilters struct {
	Type     string
	Status   domain.Status
	Labels   []string
	ParentID string
	Limit    int
	Offset   int
}

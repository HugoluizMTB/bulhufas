package domain

import "time"

// ChunkType represents the type of PM artifact extracted from a conversation.
type ChunkType string

const (
	ChunkDecision        ChunkType = "decision"
	ChunkActionItem      ChunkType = "action_item"
	ChunkRequirement     ChunkType = "requirement"
	ChunkBlocker         ChunkType = "blocker"
	ChunkScopeChange     ChunkType = "scope_change"
	ChunkContext         ChunkType = "context"
	ChunkResearchFinding ChunkType = "research_finding"
	ChunkStatusUpdate    ChunkType = "status_update"
)

// Status represents the lifecycle state of a chunk or work item.
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusResolved   Status = "resolved"
	StatusArchived   Status = "archived"
)

// Conversation stores the original raw input.
type Conversation struct {
	ID           string    `json:"id"`
	Source       string    `json:"source"`
	RawText      string    `json:"raw_text"`
	Summary      string    `json:"summary"`
	Participants []string  `json:"participants"`
	Date         time.Time `json:"date"`
	CreatedAt    time.Time `json:"created_at"`
}

// Chunk is a PM-relevant piece of knowledge extracted from a conversation.
type Chunk struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Content        string    `json:"content"`
	Type           ChunkType `json:"type"`
	Tags           []string  `json:"tags"`
	Systems        []string  `json:"systems"`
	People         []string  `json:"people"`
	Status         Status    `json:"status"`
	ActionItem     string    `json:"action_item,omitempty"`
	Embedding      []float32 `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// WorkItem is a trackable unit of work derived from chunks.
type WorkItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // task, bug, spec, spike
	Status      Status    `json:"status"`
	Priority    int       `json:"priority"` // 1=urgent, 4=low
	AssigneeIDs []string  `json:"assignee_ids"`
	ParentID    string    `json:"parent_id,omitempty"`
	Labels      []string  `json:"labels"`
	ChunkIDs    []string  `json:"chunk_ids"` // linked context
	DueDate     string    `json:"due_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Relation links two entities (chunks or work items).
type Relation struct {
	ID       string `json:"id"`
	FromID   string `json:"from_id"`
	FromType string `json:"from_type"` // chunk, work_item
	ToID     string `json:"to_id"`
	ToType   string `json:"to_type"`
	Type     string `json:"type"` // blocks, relates_to, parent_of, derived_from
}

// SearchQuery represents a semantic or structured search request.
type SearchQuery struct {
	Text   string    `json:"text,omitempty"`
	Type   ChunkType `json:"type,omitempty"`
	Status Status    `json:"status,omitempty"`
	People []string  `json:"people,omitempty"`
	Tags   []string  `json:"tags,omitempty"`
	Limit  int       `json:"limit,omitempty"`
}

// SearchResult wraps a chunk with its similarity score.
type SearchResult struct {
	Chunk Chunk   `json:"chunk"`
	Score float32 `json:"score"`
}

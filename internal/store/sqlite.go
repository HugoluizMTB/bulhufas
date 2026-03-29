package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HugoluizMTB/bulhufas/internal/domain"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLite(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *SQLiteStore) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			source TEXT,
			raw_text TEXT,
			summary TEXT,
			participants TEXT,
			date TEXT,
			created_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS chunks (
			id TEXT PRIMARY KEY,
			conversation_id TEXT,
			content TEXT,
			type TEXT,
			tags TEXT,
			systems TEXT,
			people TEXT,
			status TEXT DEFAULT 'pending',
			action_item TEXT,
			created_at TEXT,
			updated_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS work_items (
			id TEXT PRIMARY KEY,
			title TEXT,
			description TEXT,
			type TEXT,
			status TEXT DEFAULT 'pending',
			priority INTEGER DEFAULT 3,
			assignee_ids TEXT,
			parent_id TEXT,
			labels TEXT,
			chunk_ids TEXT,
			due_date TEXT,
			created_at TEXT,
			updated_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS relations (
			id TEXT PRIMARY KEY,
			from_id TEXT,
			from_type TEXT,
			to_id TEXT,
			to_type TEXT,
			type TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_type ON chunks(type)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_status ON chunks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_conversation ON chunks(conversation_id)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("migration: %w", err)
		}
	}

	return nil
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func fromJSON(s string, v any) {
	json.Unmarshal([]byte(s), v)
}

func (s *SQLiteStore) SaveConversation(_ context.Context, conv *domain.Conversation) error {
	if conv.ID == "" {
		conv.ID = uuid.New().String()
	}
	if conv.CreatedAt.IsZero() {
		conv.CreatedAt = time.Now()
	}

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO conversations (id, source, raw_text, summary, participants, date, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		conv.ID, conv.Source, conv.RawText, conv.Summary,
		toJSON(conv.Participants), conv.Date.Format(time.RFC3339), conv.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (s *SQLiteStore) GetConversation(_ context.Context, id string) (*domain.Conversation, error) {
	row := s.db.QueryRow(`SELECT id, source, raw_text, summary, participants, date, created_at FROM conversations WHERE id = ?`, id)

	var conv domain.Conversation
	var participants, date, createdAt string

	err := row.Scan(&conv.ID, &conv.Source, &conv.RawText, &conv.Summary, &participants, &date, &createdAt)
	if err != nil {
		return nil, err
	}

	fromJSON(participants, &conv.Participants)
	conv.Date, _ = time.Parse(time.RFC3339, date)
	conv.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	return &conv, nil
}

func (s *SQLiteStore) ListConversations(_ context.Context, limit, offset int) ([]domain.Conversation, error) {
	rows, err := s.db.Query(`SELECT id, source, summary, participants, date, created_at FROM conversations ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []domain.Conversation
	for rows.Next() {
		var conv domain.Conversation
		var participants, date, createdAt string
		rows.Scan(&conv.ID, &conv.Source, &conv.Summary, &participants, &date, &createdAt)
		fromJSON(participants, &conv.Participants)
		conv.Date, _ = time.Parse(time.RFC3339, date)
		conv.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		convs = append(convs, conv)
	}

	return convs, nil
}

func (s *SQLiteStore) SaveChunk(_ context.Context, chunk *domain.Chunk) error {
	if chunk.ID == "" {
		chunk.ID = uuid.New().String()
	}
	now := time.Now()
	if chunk.CreatedAt.IsZero() {
		chunk.CreatedAt = now
	}
	chunk.UpdatedAt = now

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO chunks (id, conversation_id, content, type, tags, systems, people, status, action_item, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		chunk.ID, chunk.ConversationID, chunk.Content, string(chunk.Type),
		toJSON(chunk.Tags), toJSON(chunk.Systems), toJSON(chunk.People),
		string(chunk.Status), chunk.ActionItem,
		chunk.CreatedAt.Format(time.RFC3339), chunk.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (s *SQLiteStore) GetChunk(_ context.Context, id string) (*domain.Chunk, error) {
	row := s.db.QueryRow(`SELECT id, conversation_id, content, type, tags, systems, people, status, action_item, created_at, updated_at FROM chunks WHERE id = ?`, id)
	return scanChunk(row)
}

func scanChunk(row *sql.Row) (*domain.Chunk, error) {
	var chunk domain.Chunk
	var tags, systems, people, createdAt, updatedAt, chunkType, status string

	err := row.Scan(&chunk.ID, &chunk.ConversationID, &chunk.Content, &chunkType,
		&tags, &systems, &people, &status, &chunk.ActionItem, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	chunk.Type = domain.ChunkType(chunkType)
	chunk.Status = domain.Status(status)
	fromJSON(tags, &chunk.Tags)
	fromJSON(systems, &chunk.Systems)
	fromJSON(people, &chunk.People)
	chunk.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	chunk.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &chunk, nil
}

func (s *SQLiteStore) UpdateChunk(_ context.Context, chunk *domain.Chunk) error {
	chunk.UpdatedAt = time.Now()

	_, err := s.db.Exec(
		`UPDATE chunks SET content=?, type=?, tags=?, systems=?, people=?, status=?, action_item=?, updated_at=? WHERE id=?`,
		chunk.Content, string(chunk.Type), toJSON(chunk.Tags), toJSON(chunk.Systems),
		toJSON(chunk.People), string(chunk.Status), chunk.ActionItem,
		chunk.UpdatedAt.Format(time.RFC3339), chunk.ID,
	)
	return err
}

func (s *SQLiteStore) DeleteChunk(_ context.Context, id string) error {
	_, err := s.db.Exec(`DELETE FROM chunks WHERE id = ?`, id)
	return err
}

func (s *SQLiteStore) ListChunks(_ context.Context, filters ChunkFilters) ([]domain.Chunk, error) {
	query := `SELECT id, conversation_id, content, type, tags, systems, people, status, action_item, created_at, updated_at FROM chunks WHERE 1=1`
	var args []any

	if filters.Type != "" {
		query += ` AND type = ?`
		args = append(args, string(filters.Type))
	}
	if filters.Status != "" {
		query += ` AND status = ?`
		args = append(args, string(filters.Status))
	}
	if filters.ConversationID != "" {
		query += ` AND conversation_id = ?`
		args = append(args, filters.ConversationID)
	}
	if len(filters.People) > 0 {
		for _, p := range filters.People {
			query += ` AND people LIKE ?`
			args = append(args, "%"+p+"%")
		}
	}
	if len(filters.Tags) > 0 {
		for _, t := range filters.Tags {
			query += ` AND tags LIKE ?`
			args = append(args, "%"+t+"%")
		}
	}

	query += ` ORDER BY created_at DESC`

	if filters.Limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, filters.Limit)
	}
	if filters.Offset > 0 {
		query += fmt.Sprintf(` OFFSET %d`, filters.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []domain.Chunk
	for rows.Next() {
		var chunk domain.Chunk
		var tags, systems, people, createdAt, updatedAt, chunkType, status string

		rows.Scan(&chunk.ID, &chunk.ConversationID, &chunk.Content, &chunkType,
			&tags, &systems, &people, &status, &chunk.ActionItem, &createdAt, &updatedAt)

		chunk.Type = domain.ChunkType(chunkType)
		chunk.Status = domain.Status(status)
		fromJSON(tags, &chunk.Tags)
		fromJSON(systems, &chunk.Systems)
		fromJSON(people, &chunk.People)
		chunk.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		chunk.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func (s *SQLiteStore) SaveWorkItem(_ context.Context, item *domain.WorkItem) error {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO work_items (id, title, description, type, status, priority, assignee_ids, parent_id, labels, chunk_ids, due_date, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.Title, item.Description, item.Type, string(item.Status), item.Priority,
		toJSON(item.AssigneeIDs), item.ParentID, toJSON(item.Labels), toJSON(item.ChunkIDs),
		item.DueDate, item.CreatedAt.Format(time.RFC3339), item.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (s *SQLiteStore) GetWorkItem(_ context.Context, id string) (*domain.WorkItem, error) {
	row := s.db.QueryRow(`SELECT id, title, description, type, status, priority, assignee_ids, parent_id, labels, chunk_ids, due_date, created_at, updated_at FROM work_items WHERE id = ?`, id)

	var item domain.WorkItem
	var assignees, labels, chunkIDs, createdAt, updatedAt, status string

	err := row.Scan(&item.ID, &item.Title, &item.Description, &item.Type, &status, &item.Priority,
		&assignees, &item.ParentID, &labels, &chunkIDs, &item.DueDate, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	item.Status = domain.Status(status)
	fromJSON(assignees, &item.AssigneeIDs)
	fromJSON(labels, &item.Labels)
	fromJSON(chunkIDs, &item.ChunkIDs)
	item.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	item.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &item, nil
}

func (s *SQLiteStore) UpdateWorkItem(_ context.Context, item *domain.WorkItem) error {
	item.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		`UPDATE work_items SET title=?, description=?, type=?, status=?, priority=?, assignee_ids=?, parent_id=?, labels=?, chunk_ids=?, due_date=?, updated_at=? WHERE id=?`,
		item.Title, item.Description, item.Type, string(item.Status), item.Priority,
		toJSON(item.AssigneeIDs), item.ParentID, toJSON(item.Labels), toJSON(item.ChunkIDs),
		item.DueDate, item.UpdatedAt.Format(time.RFC3339), item.ID,
	)
	return err
}

func (s *SQLiteStore) DeleteWorkItem(_ context.Context, id string) error {
	_, err := s.db.Exec(`DELETE FROM work_items WHERE id = ?`, id)
	return err
}

func (s *SQLiteStore) ListWorkItems(_ context.Context, filters WorkItemFilters) ([]domain.WorkItem, error) {
	query := `SELECT id, title, description, type, status, priority, assignee_ids, parent_id, labels, chunk_ids, due_date, created_at, updated_at FROM work_items WHERE 1=1`
	var args []any

	if filters.Type != "" {
		query += ` AND type = ?`
		args = append(args, filters.Type)
	}
	if filters.Status != "" {
		query += ` AND status = ?`
		args = append(args, string(filters.Status))
	}
	if filters.ParentID != "" {
		query += ` AND parent_id = ?`
		args = append(args, filters.ParentID)
	}
	if len(filters.Labels) > 0 {
		for _, l := range filters.Labels {
			query += ` AND labels LIKE ?`
			args = append(args, "%"+l+"%")
		}
	}

	query += ` ORDER BY priority ASC, created_at DESC`
	if filters.Limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, filters.Limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.WorkItem
	for rows.Next() {
		var item domain.WorkItem
		var assignees, labels, chunkIDs, createdAt, updatedAt, status string

		rows.Scan(&item.ID, &item.Title, &item.Description, &item.Type, &status, &item.Priority,
			&assignees, &item.ParentID, &labels, &chunkIDs, &item.DueDate, &createdAt, &updatedAt)

		item.Status = domain.Status(status)
		fromJSON(assignees, &item.AssigneeIDs)
		fromJSON(labels, &item.Labels)
		fromJSON(chunkIDs, &item.ChunkIDs)
		item.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		item.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		items = append(items, item)
	}

	return items, nil
}

func (s *SQLiteStore) SaveRelation(_ context.Context, rel *domain.Relation) error {
	if rel.ID == "" {
		rel.ID = uuid.New().String()
	}
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO relations (id, from_id, from_type, to_id, to_type, type) VALUES (?, ?, ?, ?, ?, ?)`,
		rel.ID, rel.FromID, rel.FromType, rel.ToID, rel.ToType, rel.Type,
	)
	return err
}

func (s *SQLiteStore) DeleteRelation(_ context.Context, id string) error {
	_, err := s.db.Exec(`DELETE FROM relations WHERE id = ?`, id)
	return err
}

func (s *SQLiteStore) GetRelations(_ context.Context, entityID string) ([]domain.Relation, error) {
	rows, err := s.db.Query(`SELECT id, from_id, from_type, to_id, to_type, type FROM relations WHERE from_id = ? OR to_id = ?`, entityID, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rels []domain.Relation
	for rows.Next() {
		var rel domain.Relation
		rows.Scan(&rel.ID, &rel.FromID, &rel.FromType, &rel.ToID, &rel.ToType, &rel.Type)
		rels = append(rels, rel)
	}

	return rels, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}


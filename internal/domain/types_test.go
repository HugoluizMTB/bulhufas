package domain

import "testing"

func TestChunkTypeValues(t *testing.T) {
	tests := []struct {
		name string
		ct   ChunkType
		want string
	}{
		{"decision", ChunkDecision, "decision"},
		{"action_item", ChunkActionItem, "action_item"},
		{"requirement", ChunkRequirement, "requirement"},
		{"blocker", ChunkBlocker, "blocker"},
		{"scope_change", ChunkScopeChange, "scope_change"},
		{"context", ChunkContext, "context"},
		{"research_finding", ChunkResearchFinding, "research_finding"},
		{"status_update", ChunkStatusUpdate, "status_update"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.ct) != tt.want {
				t.Errorf("ChunkType %s = %q, want %q", tt.name, tt.ct, tt.want)
			}
		})
	}
}

func TestStatusValues(t *testing.T) {
	tests := []struct {
		name string
		s    Status
		want string
	}{
		{"pending", StatusPending, "pending"},
		{"in_progress", StatusInProgress, "in_progress"},
		{"resolved", StatusResolved, "resolved"},
		{"archived", StatusArchived, "archived"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.s) != tt.want {
				t.Errorf("Status %s = %q, want %q", tt.name, tt.s, tt.want)
			}
		})
	}
}

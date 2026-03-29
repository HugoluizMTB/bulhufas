package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/HugoluizMTB/bulhufas/internal/domain"
	"github.com/HugoluizMTB/bulhufas/internal/store"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func RunStdio(h *Handler, s store.Store) error {
	srv := mcpserver.NewMCPServer("bulhufas", "0.1.0",
		mcpserver.WithToolCapabilities(false),
	)

	srv.AddTool(
		mcplib.NewTool("save_conversation",
			mcplib.WithDescription("Save a conversation with extracted structured chunks"),
			mcplib.WithString("source", mcplib.Required(), mcplib.Description("Origin: whatsapp, slack, meeting, email")),
			mcplib.WithString("raw_text", mcplib.Description("Original raw conversation text")),
			mcplib.WithString("summary", mcplib.Required(), mcplib.Description("Brief summary")),
			mcplib.WithString("participants_json", mcplib.Description("JSON array of participant names")),
			mcplib.WithString("chunks_json", mcplib.Required(), mcplib.Description("JSON array of chunk objects with: content, type, tags, systems, people, status, action_item")),
		),
		func(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
			source, _ := req.RequireString("source")
			summary, _ := req.RequireString("summary")
			rawText, _ := req.RequireString("raw_text")
			participantsJSON, _ := req.RequireString("participants_json")
			chunksJSON, _ := req.RequireString("chunks_json")

			conv := &domain.Conversation{
				Source:  source,
				RawText: rawText,
				Summary: summary,
			}
			json.Unmarshal([]byte(participantsJSON), &conv.Participants)

			var chunks []domain.Chunk
			json.Unmarshal([]byte(chunksJSON), &chunks)

			if err := h.SaveConversation(ctx, conv, chunks); err != nil {
				return mcplib.NewToolResultError(err.Error()), nil
			}

			return mcplib.NewToolResultText(fmt.Sprintf("Saved conversation %s with %d chunks", conv.ID, len(chunks))), nil
		},
	)

	srv.AddTool(
		mcplib.NewTool("search",
			mcplib.WithDescription("Semantic search across all stored chunks by meaning similarity"),
			mcplib.WithString("text", mcplib.Required(), mcplib.Description("Search query")),
			mcplib.WithNumber("limit", mcplib.Description("Max results, default 5")),
		),
		func(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
			text, _ := req.RequireString("text")
			limit := 5
			if l, err := req.RequireFloat("limit"); err == nil {
				limit = int(l)
			}

			results, err := h.Search(ctx, domain.SearchQuery{Text: text, Limit: limit})
			if err != nil {
				return mcplib.NewToolResultError(err.Error()), nil
			}

			b, _ := json.MarshalIndent(results, "", "  ")
			return mcplib.NewToolResultText(string(b)), nil
		},
	)

	srv.AddTool(
		mcplib.NewTool("list_chunks",
			mcplib.WithDescription("List chunks with optional filters"),
			mcplib.WithString("type", mcplib.Description("Filter: decision, action_item, requirement, blocker, scope_change, context, research_finding, status_update")),
			mcplib.WithString("status", mcplib.Description("Filter: pending, in_progress, resolved, archived")),
		),
		func(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
			t, _ := req.RequireString("type")
			st, _ := req.RequireString("status")

			chunks, err := s.ListChunks(ctx, store.ChunkFilters{
				Type:   domain.ChunkType(t),
				Status: domain.Status(st),
				Limit:  50,
			})
			if err != nil {
				return mcplib.NewToolResultError(err.Error()), nil
			}

			b, _ := json.MarshalIndent(chunks, "", "  ")
			return mcplib.NewToolResultText(string(b)), nil
		},
	)

	srv.AddTool(
		mcplib.NewTool("update_status",
			mcplib.WithDescription("Update chunk status"),
			mcplib.WithString("id", mcplib.Required(), mcplib.Description("Chunk ID")),
			mcplib.WithString("status", mcplib.Required(), mcplib.Description("New status: pending, in_progress, resolved, archived")),
		),
		func(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
			id, _ := req.RequireString("id")
			status, _ := req.RequireString("status")

			if err := h.UpdateChunkStatus(ctx, id, domain.Status(status)); err != nil {
				return mcplib.NewToolResultError(err.Error()), nil
			}
			return mcplib.NewToolResultText("Status updated"), nil
		},
	)

	srv.AddTool(
		mcplib.NewTool("delete_chunk",
			mcplib.WithDescription("Delete a chunk by ID"),
			mcplib.WithString("id", mcplib.Required(), mcplib.Description("Chunk ID")),
		),
		func(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
			id, _ := req.RequireString("id")

			if err := h.DeleteChunk(ctx, id); err != nil {
				return mcplib.NewToolResultError(err.Error()), nil
			}
			return mcplib.NewToolResultText("Chunk deleted"), nil
		},
	)

	srv.AddTool(
		mcplib.NewTool("list_actions",
			mcplib.WithDescription("List all pending action items"),
		),
		func(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
			allChunks, err := s.ListChunks(ctx, store.ChunkFilters{
				Status: domain.StatusPending,
				Limit:  100,
			})
			if err != nil {
				return mcplib.NewToolResultError(err.Error()), nil
			}

			var actions []domain.Chunk
			for _, c := range allChunks {
				if c.ActionItem != "" {
					actions = append(actions, c)
				}
			}

			b, _ := json.MarshalIndent(actions, "", "  ")
			return mcplib.NewToolResultText(string(b)), nil
		},
	)

	return mcpserver.ServeStdio(srv)
}

<div align="center">

# bulhufas

**RAG-powered project management that captures what PM tools miss.**

[![Build](https://github.com/HugoluizMTB/bulhufas/actions/workflows/ci.yml/badge.svg)](https://github.com/HugoluizMTB/bulhufas/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/HugoluizMTB/bulhufas.svg)](https://pkg.go.dev/github.com/HugoluizMTB/bulhufas)
[![Go Report Card](https://goreportcard.com/badge/github.com/HugoluizMTB/bulhufas)](https://goreportcard.com/report/github.com/HugoluizMTB/bulhufas)
[![License](https://img.shields.io/github/license/HugoluizMTB/bulhufas)](LICENSE)

[Getting Started](#getting-started) · [How It Works](#how-it-works) · [Self-Host](#self-hosting) · [Contributing](CONTRIBUTING.md)

</div>

---

Teams make decisions in Slack, WhatsApp, and meetings — then none of it reaches the PM tool. **bulhufas** captures raw conversations, extracts structured project artifacts (decisions, action items, blockers, scope changes), and makes them searchable via semantic embeddings.

It works as an [MCP server](https://modelcontextprotocol.io), so your AI coding assistant becomes the interface.

## Features

- **Conversation → Structure** — Paste raw chat, get structured chunks: decisions, action items, blockers, requirements, scope changes
- **Semantic Search** — Find context by meaning, not keywords. "What did we decide about auth?" finds the right chunk even if "auth" isn't in the text
- **CRUD on Knowledge** — Update status, add context, archive outdated chunks. Your knowledge base stays current
- **MCP Native** — Works directly inside Claude Code, Cursor, or any MCP-compatible client
- **Single Binary** — One Go binary, embedded vector store. No external databases required
- **Self-Hostable** — Docker image under 15MB. Deploy anywhere: Coolify, Railway, Hetzner, AWS, GCP

## How It Works

```
You paste a conversation into your AI assistant
         ↓
The LLM extracts structured chunks with metadata
         ↓
bulhufas stores chunks + embeddings (via Ollama)
         ↓
Later: "what's pending from last week?" → semantic search returns relevant chunks
```

### What Gets Captured

| Chunk Type | Example |
|-----------|---------|
| `decision` | "We chose WebSockets over polling for real-time updates" |
| `action_item` | "Hugo will create read-only DB credentials by Friday" |
| `blocker` | "Can't deploy until the SSL cert is renewed" |
| `requirement` | "Client needs CSV export for the finance report" |
| `scope_change` | "Auth module expanded to include SSO" |
| `context` | "The legacy API returns XML, not JSON" |
| `research_finding` | "pgvector outperforms pinecone for our dataset size" |
| `status_update` | "Payment integration is live in staging" |

## Getting Started

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Ollama](https://ollama.ai) (for local embeddings)
- [Docker](https://docs.docker.com/get-docker/) (optional, for containerized deployment)

### Install

```bash
go install github.com/HugoluizMTB/bulhufas/cmd/server@latest
```

### Run with Ollama

```bash
# Pull the embedding model (~274MB, runs on CPU)
ollama pull nomic-embed-text

# Start bulhufas
bulhufas
```

### Use as MCP Server

Add to your Claude Code or Cursor config:

```json
{
  "mcpServers": {
    "bulhufas": {
      "command": "bulhufas",
      "args": ["--mcp"]
    }
  }
}
```

Then in your AI assistant:

```
@bulhufas save this conversation [paste chat]
@bulhufas what's pending from renan?
@bulhufas context about card reconciliation
@bulhufas mark chunk-abc as resolved
```

## Self-Hosting

### Docker

```bash
docker run -d \
  --name bulhufas \
  -p 8420:8420 \
  -v bulhufas-data:/data \
  ghcr.io/hugoluizmtb/bulhufas:latest
```

### Docker Compose

```bash
git clone https://github.com/HugoluizMTB/bulhufas.git
cd bulhufas
docker compose up -d
```

See [docker-compose.yml](docker-compose.yml) for the full configuration including Ollama.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8420` | Server port |
| `DATA_DIR` | `/data` | Persistent storage directory |
| `OLLAMA_URL` | `http://localhost:11434` | Ollama API endpoint |
| `EMBED_MODEL` | `nomic-embed-text` | Embedding model name |
| `LOG_LEVEL` | `info` | Log verbosity: debug, info, warn, error |

## Architecture

```
cmd/server/          → entrypoint
internal/
├── domain/          → core types: Conversation, Chunk, WorkItem, Relation
├── mcp/             → MCP protocol handlers
├── store/           → persistence interface + SQLite implementation
├── vectorstore/     → embedded vector search (chromem-go)
├── embedder/        → Ollama client for generating embeddings
└── chunker/         → text chunking logic
web/                 → React dashboard (future)
```

All external dependencies are behind interfaces. Swap Ollama for OpenAI, SQLite for Postgres, or chromem-go for pgvector — without touching business logic.

## Roadmap

- [x] Core domain types and interfaces
- [x] MCP handler with save/search/update/delete
- [ ] SQLite store implementation
- [ ] Ollama embedder integration
- [ ] chromem-go vector store
- [ ] MCP server protocol (stdio transport)
- [ ] Docker image
- [ ] React dashboard
- [ ] Slack plugin
- [ ] Remote MCP via SSE transport

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for setup instructions, code style, and PR process.

## License

[Apache License 2.0](LICENSE) — use it freely, even commercially. Patent protection included.

---

<div align="center">

Created by [@HugoluizMTB](https://github.com/HugoluizMTB)

</div>

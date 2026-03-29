package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/HugoluizMTB/bulhufas/internal/embedder"
	"github.com/HugoluizMTB/bulhufas/internal/mcp"
	"github.com/HugoluizMTB/bulhufas/internal/store"
	"github.com/HugoluizMTB/bulhufas/internal/vectorstore"
)

func main() {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	os.MkdirAll(dataDir, 0755)

	log.Println("loading embedding model...")
	emb, err := embedder.NewHugot(dataDir + "/models")
	if err != nil {
		log.Fatalf("embedder: %v", err)
	}

	log.Println("initializing vector store...")
	vectors, err := vectorstore.NewChromem(dataDir+"/vectors", emb)
	if err != nil {
		log.Fatalf("vectorstore: %v", err)
	}

	log.Println("initializing database...")
	db, err := store.NewSQLite(dataDir + "/bulhufas.db")
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	handler := mcp.NewHandler(db, vectors, emb)

	if len(os.Args) > 1 && os.Args[1] == "--mcp" {
		log.Println("starting MCP stdio server...")
		if err := mcp.RunStdio(handler, db); err != nil {
			log.Fatalf("mcp: %v", err)
		}
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8420"
	}

	server := mcp.NewServer(handler, db)
	log.Printf("bulhufas ready on :%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), server); err != nil {
		log.Fatalf("server: %v", err)
	}
}

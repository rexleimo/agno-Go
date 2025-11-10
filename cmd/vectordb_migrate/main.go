package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rexleimo/agno-go/internal/vectordb/migrate"
)

func main() {
	var (
		action       = flag.String("action", "up", "Migration action: up|down")
		provider     = flag.String("provider", "chroma", "VectorDB provider: chroma")
		collection   = flag.String("collection", "", "Collection name")
		chromaURL    = flag.String("chroma-url", os.Getenv("CHROMA_URL"), "Chroma base URL (default from CHROMA_URL)")
		chromaTenant = flag.String("chroma-tenant", os.Getenv("CHROMA_TENANT"), "Chroma tenant")
		chromaDB     = flag.String("chroma-db", os.Getenv("CHROMA_DB"), "Chroma database")
		distance     = flag.String("distance", os.Getenv("VECTOR_DISTANCE"), "Distance function: l2|cosine|ip")
		timeout      = flag.Duration("timeout", 30*time.Second, "Operation timeout")
	)
	flag.Parse()

	if *collection == "" {
		fmt.Fprintln(os.Stderr, "--collection is required")
		os.Exit(2)
	}

	opts := migrate.Options{
		Provider:       *provider,
		Collection:     *collection,
		ChromaBaseURL:  *chromaURL,
		ChromaTenant:   *chromaTenant,
		ChromaDatabase: *chromaDB,
		Distance:       *distance,
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	var err error
	switch *action {
	case "up":
		err = migrate.Up(ctx, opts)
	case "down":
		err = migrate.Down(ctx, opts)
	default:
		fmt.Fprintln(os.Stderr, "invalid --action, expected up|down")
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "migration error:", err)
		os.Exit(1)
	}
	fmt.Println("OK")
}

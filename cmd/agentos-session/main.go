package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rexleimo/agno-go/internal/http/router"
	"github.com/rexleimo/agno-go/internal/session/service"
	"github.com/rexleimo/agno-go/internal/session/store"
	postgresstore "github.com/rexleimo/agno-go/internal/session/store/postgres"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storesConfig, defaultID, err := loadStoreConfig()
	if err != nil {
		log.Fatalf("failed to load store configuration: %v", err)
	}

	storeInstances := make(map[string]*postgresstore.Store, len(storesConfig))
	storeInterfaces := make(map[string]store.Store, len(storesConfig))
	for id, cfg := range storesConfig {
		store, err := postgresstore.New(ctx, postgresstore.Config{DSN: cfg})
		if err != nil {
			log.Fatalf("failed to initialise store %s: %v", id, err)
		}
		storeInstances[id] = store
		storeInterfaces[id] = store
	}
	defer func() {
		for _, store := range storeInstances {
			store.Close()
		}
	}()

	serviceInstance, err := service.New(service.Config{Stores: storeInterfaces, DefaultDB: defaultID})
	if err != nil {
		log.Fatalf("failed to initialise session service: %v", err)
	}

	handler := router.New(serviceInstance)

	port := os.Getenv("AGNO_SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("Go session service listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	<-sigCh
	log.Println("shutting down session service")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}

func loadStoreConfig() (map[string]string, string, error) {
	if raw := strings.TrimSpace(os.Getenv("AGNO_SESSION_DSN_MAP")); raw != "" {
		var mapping map[string]string
		if err := json.Unmarshal([]byte(raw), &mapping); err != nil {
			return nil, "", err
		}
		if len(mapping) == 0 {
			return nil, "", service.ErrNoStoresConfigured
		}
		return mapping, pickDefault(mapping), nil
	}

	if dsn := strings.TrimSpace(os.Getenv("AGNO_PG_DSN")); dsn != "" {
		return map[string]string{"default": dsn}, "default", nil
	}

	if dsn := strings.TrimSpace(os.Getenv("DATABASE_URL")); dsn != "" {
		return map[string]string{"default": dsn}, "default", nil
	}

	return nil, "", service.ErrNoStoresConfigured
}

func pickDefault(mapping map[string]string) string {
	if defaultID := strings.TrimSpace(os.Getenv("AGNO_DEFAULT_DB_ID")); defaultID != "" {
		if _, ok := mapping[defaultID]; ok {
			return defaultID
		}
	}
	for id := range mapping {
		return id
	}
	return ""
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/varadharajaan/tracedeck-agent/backend/internal/api"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/logging"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/store"
)

func main() {
	addr := flag.String("addr", constants.DefaultBackendAddr, "localhost bind address")
	logDir := flag.String("log-dir", constants.DefaultLogDir, "backend log directory")
	logLevel := flag.String("log-level", constants.DefaultLogLevel, "log level: trace, debug, info, warn, error")
	dataPath := flag.String("data-path", constants.DefaultDataPath, "backend state JSON path")
	memoryStore := flag.Bool("memory-store", false, "use in-memory backend state")
	apiKey := flag.String("api-key", "", "optional local API key for API routes")
	apiKeyTenantID := flag.String("api-key-tenant-id", "", "optional tenant scope for the local API key")
	apiKeyActorID := flag.String("api-key-actor-id", constants.AuditActorLocalAPI, "actor id attached to local API key requests")
	apiKeyRoleID := flag.String("api-key-role-id", constants.RoleBusinessManager, "role id attached to local API key requests")
	flag.Parse()

	logger, closeLogger, err := logging.New(*logDir, *logLevel)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func() {
		if err := closeLogger(); err != nil {
			slog.Default().Warn("close backend logger", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var repo store.Repository
	if *memoryStore {
		repo = store.NewMemory()
	} else {
		persistent, err := store.NewPersistent(*dataPath)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		repo = persistent
	}

	server := api.NewServerWithAuth(repo, logger, api.AuthConfig{
		APIKey:   *apiKey,
		TenantID: *apiKeyTenantID,
		ActorID:  *apiKeyActorID,
		RoleID:   *apiKeyRoleID,
	})
	if err := api.Serve(ctx, *addr, server.Handler(), logger); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

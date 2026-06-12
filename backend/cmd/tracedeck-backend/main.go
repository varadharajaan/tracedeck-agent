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

	server := api.NewServer(store.NewMemory(), logger)
	if err := api.Serve(ctx, *addr, server.Handler(), logger); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

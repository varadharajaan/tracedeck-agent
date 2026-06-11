package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(logDir string, level string) (*slog.Logger, func() error, error) {
	if logDir == "" {
		logDir = constants.DefaultLogDir
	}
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		return nil, nil, fmt.Errorf("create log dir: %w", err)
	}

	sink := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, constants.DefaultLogFileName),
		MaxSize:    constants.LogRotationMaxSizeMB,
		MaxBackups: constants.LogRotationMaxFiles,
		MaxAge:     constants.LogRotationMaxAgeDay,
		Compress:   true,
	}

	handler := slog.NewJSONHandler(sink, &slog.HandlerOptions{
		Level: parseLevel(level),
	})

	return slog.New(handler), sink.Close, nil
}

func parseLevel(level string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case constants.LogLevelTrace, constants.LogLevelDebug:
		return slog.LevelDebug
	case constants.LogLevelWarn:
		return slog.LevelWarn
	case constants.LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

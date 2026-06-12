package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/varadharajaan/tracedeck-agent/backend/internal/constants"
)

func New(logDir string, level string) (*slog.Logger, func() error, error) {
	if logDir == "" {
		logDir = constants.DefaultLogDir
	}
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		return nil, nil, fmt.Errorf("create backend log dir: %w", err)
	}

	writer := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, constants.BackendName+".log"),
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: parseLevel(level)})
	return slog.New(handler), writer.Close, nil
}

func parseLevel(level string) slog.Level {
	switch level {
	case "trace", "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

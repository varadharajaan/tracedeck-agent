package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type capture struct {
	ReceivedAt    time.Time `json:"received_at"`
	Method        string    `json:"method"`
	Path          string    `json:"path"`
	ContentType   string    `json:"content_type"`
	ResourceLogs  int       `json:"resource_logs"`
	LogRecords    int       `json:"log_records"`
	PrivacySafe   bool      `json:"privacy_safe"`
	ForbiddenHits []string  `json:"forbidden_hits"`
}

func main() {
	addr := flag.String("addr", "127.0.0.1:4318", "listen address")
	outputPath := flag.String("output", "data/local/output/fake-otlp-last.json", "capture output path")
	readyPath := flag.String("ready", "data/local/output/fake-otlp-ready.json", "ready output path")
	flag.Parse()

	mux := http.NewServeMux()
	server := &http.Server{Addr: *addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/v1/logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 2*1024*1024))
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		defer func() {
			_ = r.Body.Close()
		}()
		record := inspect(r, body)
		if err := writeJSON(*outputPath, record); err != nil {
			http.Error(w, "write capture", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{}`))
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = server.Shutdown(ctx)
		}()
	})

	if err := writeJSON(*readyPath, map[string]any{
		"status": "ready",
		"addr":   *addr,
	}); err != nil {
		log.Fatalf("write ready file: %v", err)
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("fake otlp receiver failed: %v", err)
	}
}

func inspect(r *http.Request, body []byte) capture {
	var payload map[string]any
	_ = json.Unmarshal(body, &payload)
	resourceLogs, logRecords := countLogs(payload)
	forbiddenHits := forbiddenMatches(string(body))
	return capture{
		ReceivedAt:    time.Now().UTC(),
		Method:        r.Method,
		Path:          r.URL.Path,
		ContentType:   r.Header.Get("Content-Type"),
		ResourceLogs:  resourceLogs,
		LogRecords:    logRecords,
		PrivacySafe:   len(forbiddenHits) == 0,
		ForbiddenHits: forbiddenHits,
	}
}

func countLogs(payload map[string]any) (int, int) {
	resourceLogs, _ := payload["resourceLogs"].([]any)
	totalRecords := 0
	for _, resourceLog := range resourceLogs {
		resourceMap, _ := resourceLog.(map[string]any)
		scopeLogs, _ := resourceMap["scopeLogs"].([]any)
		for _, scopeLog := range scopeLogs {
			scopeMap, _ := scopeLog.(map[string]any)
			records, _ := scopeMap["logRecords"].([]any)
			totalRecords += len(records)
		}
	}
	return len(resourceLogs), totalRecords
}

func forbiddenMatches(body string) []string {
	candidates := []string{
		"password",
		"screenshot",
		"raw_url",
		"page_title",
		"cookie",
		"token",
		"private title",
		"https://",
		"http://private",
	}
	hits := make([]string, 0)
	for _, candidate := range candidates {
		if strings.Contains(body, candidate) {
			hits = append(hits, candidate)
		}
	}
	return hits
}

func writeJSON(path string, value any) error {
	cleanPath, err := cleanLocalOutputPath(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cleanPath), 0o750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	return os.WriteFile(cleanPath, append(data, '\n'), 0o600) // #nosec G304,G703 -- cleanLocalOutputPath rejects absolute and parent-traversal paths.
}

func cleanLocalOutputPath(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	if filepath.IsAbs(cleanPath) || cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("output path must be relative to the repo workspace: %s", path)
	}
	return cleanPath, nil
}

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const appName = "tracedeck-agent"

func main() {
	out := flag.String("out", "", "browser history fixture output path")
	flag.Parse()

	if *out == "" {
		fatalf("missing --out")
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o750); err != nil {
		fatalf("create fixture dir: %v", err)
	}
	if err := os.RemoveAll(*out); err != nil {
		fatalf("remove existing fixture: %v", err)
	}

	db, err := sql.Open("sqlite", *out)
	if err != nil {
		fatalf("open fixture sqlite: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			fatalf("close fixture sqlite: %v", err)
		}
	}()

	if _, err := db.Exec(`
CREATE TABLE urls (
  id INTEGER PRIMARY KEY,
  url TEXT NOT NULL,
  title TEXT,
  visit_count INTEGER NOT NULL,
  last_visit_time INTEGER NOT NULL
)`); err != nil {
		fatalf("create urls table: %v", err)
	}

	lastVisit := chromeMicroseconds(time.Now().UTC())
	entries := []struct {
		rawURL     string
		title      string
		visitCount int
	}{
		{rawURL: "https://www.youtube.com/watch?v=traceDeckSmoke123", title: "private title must not persist", visitCount: 2},
		{rawURL: "https://learn.microsoft.com/training", title: "private title must not persist", visitCount: 3},
		{rawURL: "https://www.instagram.com/", title: "private title must not persist", visitCount: 1},
	}

	for _, entry := range entries {
		if _, err := db.Exec(`INSERT INTO urls (url, title, visit_count, last_visit_time) VALUES (?, ?, ?, ?)`, entry.rawURL, entry.title, entry.visitCount, lastVisit); err != nil {
			fatalf("insert fixture row: %v", err)
		}
	}
}

func chromeMicroseconds(value time.Time) int64 {
	return value.UTC().Sub(time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)).Microseconds()
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, appName+": "+format+"\n", args...)
	os.Exit(1)
}

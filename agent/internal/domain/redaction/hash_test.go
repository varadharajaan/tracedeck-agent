package redaction

import "testing"

func TestHashPath(t *testing.T) {
	t.Parallel()

	got := HashPath(`C:\Users\Student\movie.mp4`)
	if got == "" {
		t.Fatal("expected hash")
	}
	if got == `C:\Users\Student\movie.mp4` {
		t.Fatal("hash must not expose raw path")
	}
	if got != HashPath(`c:\users\student\movie.mp4`) {
		t.Fatal("hash should normalize path casing")
	}
}

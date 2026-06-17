package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func main() {
	outDir := flag.String("out-dir", "data/local/webpush", "directory for generated Web Push VAPID key files")
	flag.Parse()

	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		fatalf("generate VAPID keys: %v", err)
	}
	if err := os.MkdirAll(*outDir, 0o750); err != nil {
		fatalf("create output directory: %v", err)
	}
	publicPath := filepath.Join(*outDir, "vapid-public.key")
	privatePath := filepath.Join(*outDir, "vapid-private.key")
	if err := os.WriteFile(publicPath, []byte(publicKey+"\n"), 0o600); err != nil {
		fatalf("write public key: %v", err)
	}
	if err := os.WriteFile(privatePath, []byte(privateKey+"\n"), 0o600); err != nil {
		fatalf("write private key: %v", err)
	}
	fmt.Printf("web_push_vapid_public_key=%s\n", publicPath)
	fmt.Printf("web_push_vapid_private_key=%s\n", privatePath)
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

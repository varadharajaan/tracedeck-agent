package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:2525", "listen address")
	outDir := flag.String("out-dir", "data/local/fake-smtp", "captured message output directory")
	readyFile := flag.String("ready-file", "", "ready file path")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o750); err != nil {
		log.Fatalf("create output dir: %v", err)
	}
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	if *readyFile != "" {
		if err := os.MkdirAll(filepath.Dir(*readyFile), 0o750); err != nil {
			log.Fatalf("create ready dir: %v", err)
		}
		if err := os.WriteFile(*readyFile, []byte(listener.Addr().String()), 0o600); err != nil {
			log.Fatalf("write ready file: %v", err)
		}
	}
	log.Printf("fake smtp listening on %s", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handleConn(conn, *outDir)
	}
}

func handleConn(conn net.Conn, outDir string) {
	defer func() {
		_ = conn.Close()
	}()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	writeLine(writer, "220 tracedeck fake smtp")

	var message []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		command := strings.TrimRight(line, "\r\n")
		upper := strings.ToUpper(command)
		switch {
		case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
			writeLine(writer, "250-tracedeck")
			writeLine(writer, "250 OK")
		case strings.HasPrefix(upper, "MAIL FROM:"):
			writeLine(writer, "250 OK")
		case strings.HasPrefix(upper, "RCPT TO:"):
			writeLine(writer, "250 OK")
		case strings.HasPrefix(upper, "DATA"):
			writeLine(writer, "354 End data with <CR><LF>.<CR><LF>")
			message = message[:0]
			for {
				dataLine, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				trimmed := strings.TrimRight(dataLine, "\r\n")
				if trimmed == "." {
					if err := writeMessage(outDir, message); err != nil {
						writeLine(writer, "451 "+err.Error())
						break
					}
					writeLine(writer, "250 queued")
					break
				}
				message = append(message, dataLine)
			}
		case strings.HasPrefix(upper, "RSET"):
			message = message[:0]
			writeLine(writer, "250 OK")
		case strings.HasPrefix(upper, "QUIT"):
			writeLine(writer, "221 bye")
			return
		default:
			writeLine(writer, "250 OK")
		}
	}
}

func writeLine(writer *bufio.Writer, line string) {
	_, _ = fmt.Fprintf(writer, "%s\r\n", line)
	_ = writer.Flush()
}

func writeMessage(outDir string, lines []string) error {
	name := time.Now().UTC().Format("20060102T150405.000000000Z") + ".eml"
	path := filepath.Join(outDir, name)
	return os.WriteFile(path, []byte(strings.Join(lines, "")), 0o600)
}

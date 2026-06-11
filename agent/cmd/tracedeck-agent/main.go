package main

import (
	"fmt"
	"os"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/Sol1de/go-grimhold/internal/cli"
	"github.com/Sol1de/go-grimhold/internal/client"
	"github.com/Sol1de/go-grimhold/internal/world"
)

func main() {
	w, err := world.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cli.Run(w, client.New(), os.Stdin, os.Stdout)
}

package main

import (
	"os"

	"github.com/Sol1de/go-grimhold/internal/cli"
)

func main() {
	cli.Run(os.Stdin, os.Stdout)
}

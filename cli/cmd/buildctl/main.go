package main

import (
	"os"

	"github.com/build-assistant/cli/cmd/buildctl/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}


package main

import (
	"github.com/spring-financial-group/jx-semanticcheck/cmd/app"
	"os"
)

// Entrypoint for the command
func main() {
	if err := app.Run(nil); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

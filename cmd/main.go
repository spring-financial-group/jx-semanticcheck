package main

import (
	"os"
	"spring-financial-group/jx-semanticcheck/cmd/app"
)

// Entrypoint for the command
func main() {
	if err := app.Run(nil); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

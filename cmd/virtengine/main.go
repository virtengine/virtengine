package main

import (
	"os"

	_ "github.com/virtengine/virtengine/sdk/go/sdkutil"

	"github.com/virtengine/virtengine/cmd/virtengine/cmd"
)

// In main we call the rootCmd
func main() {
	rootCmd, _ := cmd.NewRootCmd()

	if err := cmd.Execute(rootCmd, "VE"); err != nil {
		os.Exit(1)
	}
}

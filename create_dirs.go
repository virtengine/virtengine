//go:build ignore

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	// Create HPC directories
	dirs := []string{
		"sdk/proto/node/virtengine/hpc/v1",
		"sdk/go/node/hpc/v1",
	}

	for _, dir := range dirs {
		cleanPath := filepath.FromSlash(dir)
		err := os.MkdirAll(cleanPath, 0755)
		if err != nil {
			fmt.Println("Error creating", cleanPath, ":", err)
			os.Exit(1)
		}
		fmt.Println("Created directory:", cleanPath)
	}

	// Copy proto files to target location
	protoFiles := map[string]string{
		"hpc_types.proto.txt":   "sdk/proto/node/virtengine/hpc/v1/types.proto",
		"hpc_tx.proto.txt":      "sdk/proto/node/virtengine/hpc/v1/tx.proto",
		"hpc_query.proto.txt":   "sdk/proto/node/virtengine/hpc/v1/query.proto",
		"hpc_genesis.proto.txt": "sdk/proto/node/virtengine/hpc/v1/genesis.proto",
	}

	for src, dst := range protoFiles {
		dstPath := filepath.FromSlash(dst)
		if err := copyFile(src, dstPath); err != nil {
			fmt.Printf("Error copying %s to %s: %v\n", src, dstPath, err)
			os.Exit(1)
		}
		fmt.Printf("Copied %s -> %s\n", src, dstPath)
	}

	fmt.Println("\nHPC proto files created successfully!")
	fmt.Println("Next steps:")
	fmt.Println("  1. cd sdk")
	fmt.Println("  2. buf generate (or ./script/protocgen.sh go github.com/virtengine/virtengine/sdk/go/node go)")
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

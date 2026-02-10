//go:build ignore
// +build ignore

// Small utility to convert bech32 addresses from one prefix to another
package main

import (
	"fmt"
	"os"

	"github.com/cosmos/btcutil/bech32"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run convert_address.go <address> <new_prefix>")
		os.Exit(1)
	}

	oldAddr := os.Args[1]
	newPrefix := os.Args[2]

	// Decode the address (limit 1023 for Cosmos bech32)
	_, data, err := bech32.Decode(oldAddr, 1023)
	if err != nil {
		fmt.Printf("Error decoding address: %v\n", err)
		os.Exit(1)
	}

	// Re-encode with new prefix
	newAddr, err := bech32.Encode(newPrefix, data)
	if err != nil {
		fmt.Printf("Error encoding address: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(newAddr)
}

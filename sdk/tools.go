//go:build tools
// +build tools

package api

import (
	_ "github.com/99designs/keyring"
	_ "github.com/grpc-ecosystem/grpc-gateway/runtime"
	_ "google.golang.org/grpc"

	_ "github.com/pseudomuto/protoc-gen-doc"

	// TODO https://github.com/akash-network/support/issues/77
	_ "github.com/regen-network/cosmos-proto/protoc-gen-gocosmos"
)

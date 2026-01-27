//go:build tools
// +build tools

package _go

import (
	_ "github.com/99designs/keyring"
	_ "github.com/grpc-ecosystem/grpc-gateway/runtime"
	_ "google.golang.org/grpc"

	_ "github.com/pseudomuto/protoc-gen-doc"

	_ "github.com/cosmos/ibc-go/v10/modules/core/02-client"
	_ "github.com/cosmos/gogoproto/gogoproto"
	_ "github.com/cosmos/gogoproto/protoc-gen-combo"
	_ "github.com/cosmos/gogoproto/protoc-gen-gocosmos"
	_ "github.com/cosmos/gogoproto/protoc-gen-gofast"
	_ "github.com/cosmos/gogoproto/protoc-gen-gogo"
	_ "github.com/cosmos/gogoproto/protoc-gen-gogofast"
	_ "github.com/cosmos/gogoproto/protoc-gen-gogofaster"
	_ "github.com/cosmos/gogoproto/protoc-gen-gogoslick"
	_ "github.com/cosmos/gogoproto/protoc-gen-gogotypes"
	_ "github.com/cosmos/gogoproto/protoc-gen-gostring"
	_ "github.com/cosmos/gogoproto/protoc-min-version"

	_ "k8s.io/api"
	_ "k8s.io/apimachinery"
)

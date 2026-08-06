//go:build tools

// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/cosmos/gogoproto/protoc-gen-gocosmos"
	_ "github.com/goware/modvendor"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger"
)

// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package scenarios

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/virtengine/virtengine/tests/load/framework"
)

// VEIDSubmitScenario executes deterministic VEID submission probes.
type VEIDSubmitScenario struct {
	transport scenarioTransport
	accounts  []string
	scopeSize int
	backend   scenarioBackend
	sequence  atomic.Uint64
}

// NewVEIDSubmitScenario creates a new VEID submit scenario.
func NewVEIDSubmitScenario(grpcEndpoint string, accounts []string) *VEIDSubmitScenario {
	return &VEIDSubmitScenario{
		transport: scenarioTransport{endpoint: grpcEndpoint},
		accounts:  append([]string(nil), accounts...),
		scopeSize: 32 * 1024,
		backend:   newDeterministicBackend(),
	}
}

// Name returns the scenario name.
func (s *VEIDSubmitScenario) Name() string {
	return "veid_submit"
}

// Setup initializes the scenario transport.
func (s *VEIDSubmitScenario) Setup(ctx context.Context) error {
	if len(s.accounts) == 0 {
		return fmt.Errorf("veid_submit requires at least one account")
	}
	return s.transport.setup(ctx)
}

// Execute performs a deterministic VEID submission.
func (s *VEIDSubmitScenario) Execute(ctx context.Context) (*framework.ExecutionResult, error) {
	start := time.Now()
	sequence := s.sequence.Add(1)
	account := s.accounts[(sequence-1)%uint64(len(s.accounts))]
	payload := buildScopePayload(sequence, s.scopeSize)

	receipt, err := s.backend.SubmitVEID(ctx, account, payload)
	if err != nil {
		return &framework.ExecutionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    err,
		}, nil
	}

	metadata := cloneMetadata(receipt.Metadata)
	metadata["submission_id"] = receipt.ID

	return &framework.ExecutionResult{
		Success:  true,
		Duration: time.Since(start),
		Metadata: metadata,
	}, nil
}

// Teardown closes the transport.
func (s *VEIDSubmitScenario) Teardown(ctx context.Context) error {
	_ = ctx
	return s.transport.teardown()
}

func buildScopePayload(sequence uint64, size int) []byte {
	if size <= 0 {
		size = 1
	}

	payload := make([]byte, size)
	var offset int

	var seed [8]byte
	binary.BigEndian.PutUint64(seed[:], sequence)
	chunk := sha256.Sum256(seed[:])

	for offset < len(payload) {
		written := copy(payload[offset:], chunk[:])
		offset += written
		chunk = sha256.Sum256(chunk[:])
	}

	return payload
}

func cloneMetadata(metadata map[string]interface{}) map[string]interface{} {
	cloned := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}

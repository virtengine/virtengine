// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	providerv1beta4 "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

func TestFileSubmitterLeaseFencesStaleOwnerAndPersistsTokenAcrossRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "submitter-lease.json")
	base := time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)
	ownerA, err := NewFileSubmitterLease(path, "provider-a")
	require.NoError(t, err)
	ownerB, err := NewFileSubmitterLease(path, "provider-b")
	require.NoError(t, err)
	ownerA.now = func() time.Time { return base }
	ownerB.now = func() time.Time { return base }

	tokenA, err := ownerA.Acquire(context.Background(), "provider-mutation:account", time.Minute)
	require.NoError(t, err)
	require.Equal(t, uint64(1), tokenA)
	_, err = ownerB.Acquire(context.Background(), "provider-mutation:account", time.Minute)
	require.Error(t, err)

	ownerA.now = func() time.Time { return base.Add(2 * time.Minute) }
	ownerB.now = func() time.Time { return base.Add(2 * time.Minute) }
	tokenB, err := ownerB.Acquire(context.Background(), "provider-mutation:account", time.Minute)
	require.NoError(t, err)
	require.Greater(t, tokenB, tokenA)
	require.False(t, ownerA.Held(context.Background(), "provider-mutation:account", tokenA))
	require.ErrorIs(t, ownerA.Renew(context.Background(), "provider-mutation:account", tokenA, time.Minute), ErrSubmitterLeaseNotHeld)
	require.NoError(t, ownerA.Release(context.Background(), "provider-mutation:account", tokenA))
	require.True(t, ownerB.Held(context.Background(), "provider-mutation:account", tokenB))

	restarted, err := NewFileSubmitterLease(path, "provider-c")
	require.NoError(t, err)
	restarted.now = func() time.Time { return base.Add(4 * time.Minute) }
	tokenC, err := restarted.Acquire(context.Background(), "provider-mutation:account", time.Minute)
	require.NoError(t, err)
	require.Greater(t, tokenC, tokenB)
}

func TestFileSubmitterLeaseSplitBrainHasExactlyOneOwner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "submitter-lease.json")
	base := time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)
	owners := make([]*FileSubmitterLease, 8)
	results := make(chan uint64, len(owners))
	for i := range owners {
		var err error
		owners[i], err = NewFileSubmitterLease(path, string(rune('a'+i)))
		require.NoError(t, err)
		owners[i].now = func() time.Time { return base }
		go func(lease *FileSubmitterLease) {
			token, acquireErr := lease.Acquire(context.Background(), "shared", time.Minute)
			if acquireErr == nil {
				results <- token
				return
			}
			results <- 0
		}(owners[i])
	}
	var winners []uint64
	for range owners {
		if token := <-results; token != 0 {
			winners = append(winners, token)
		}
	}
	require.Equal(t, []uint64{1}, winners)
}

func TestProductionMutationSubmitterRejectsLocalLease(t *testing.T) {
	address, keyManager := newMutationKeyManagerForLeaseTest(t)
	cfg := DefaultProviderMutationSubmitterConfig()
	cfg.ChainID = "chain"
	cfg.ProviderAddress = address
	cfg.QueueStatePath = filepath.Join(t.TempDir(), "queue.json")
	cfg.Chain = newMutationChainFake()
	cfg.Lease = NewLocalSubmitterLease()
	_, err := NewProviderMutationSubmitter(cfg, keyManager)
	require.ErrorContains(t, err, "process-local")

	cfg.Lease = nil
	_, err = NewProviderMutationSubmitter(cfg, keyManager)
	require.ErrorContains(t, err, "explicit durable")
}

func TestProviderMutationSubmitterStandbyTakesOverWithHigherFence(t *testing.T) {
	address, keyManager := newMutationKeyManagerForLeaseTest(t)
	dir := t.TempDir()
	chain := newMutationChainFake()
	firstLease, err := NewFileSubmitterLease(filepath.Join(dir, "lease.json"), "pod-a")
	require.NoError(t, err)
	secondLease, err := NewFileSubmitterLease(filepath.Join(dir, "lease.json"), "pod-b")
	require.NoError(t, err)
	newSubmitter := func(lease SubmitterLease) *ProviderMutationSubmitter {
		cfg := DefaultProviderMutationSubmitterConfig()
		cfg.ChainID = "chain"
		cfg.ProviderAddress = address
		cfg.QueueStatePath = filepath.Join(dir, "queue.json")
		cfg.Chain = chain
		cfg.Lease = lease
		cfg.LeaseTTL = 90 * time.Millisecond
		cfg.PollInterval = 10 * time.Millisecond
		cfg.FinalityBlocks = 0
		submitter, createErr := NewProviderMutationSubmitter(cfg, keyManager)
		require.NoError(t, createErr)
		return submitter
	}
	first := newSubmitter(firstLease)
	second := newSubmitter(secondLease)
	require.NoError(t, first.Start(context.Background()))
	require.NoError(t, second.Start(context.Background()))
	t.Cleanup(func() {
		_ = first.Stop(context.Background())
		_ = second.Stop(context.Background())
	})
	require.True(t, first.Readiness(context.Background()).Ready)
	require.False(t, second.Readiness(context.Background()).Ready)
	firstToken := first.leaseToken

	result, err := first.Submit(context.Background(), MutationProviderDelete, &providerv1beta4.MsgDeleteProvider{Owner: address})
	require.NoError(t, err)
	require.True(t, result.Final)
	chain.mu.Lock()
	require.Len(t, chain.broadcasts, 1)
	chain.mu.Unlock()

	require.NoError(t, first.Stop(context.Background()))
	require.Eventually(t, func() bool { return second.Readiness(context.Background()).Ready }, time.Second, 10*time.Millisecond)
	require.Greater(t, second.leaseToken, firstToken)
	replayed, err := second.Submit(context.Background(), MutationProviderDelete, &providerv1beta4.MsgDeleteProvider{Owner: address})
	require.NoError(t, err)
	require.True(t, replayed.Existed)
	chain.mu.Lock()
	require.Len(t, chain.broadcasts, 1)
	chain.mu.Unlock()
}

func newMutationKeyManagerForLeaseTest(t *testing.T) (string, *KeyManager) {
	t.Helper()
	kmConfig := DefaultKeyManagerConfig()
	kmConfig.StorageType = KeyStorageTypeMemory
	km, err := NewKeyManager(kmConfig)
	require.NoError(t, err)
	require.NoError(t, km.Unlock(""))
	key, err := km.GenerateKey("provider")
	require.NoError(t, err)
	address, err := ManagedKeyAccountAddress(key)
	require.NoError(t, err)
	key.ProviderAddress = address
	return address, km
}

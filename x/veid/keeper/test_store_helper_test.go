// Package keeper_test provides test utilities for VEID keeper tests.
//
// This file provides shared helper functions to properly manage IAVL store lifecycle
// and prevent goroutine leaks from IAVL nodeDB pruning background goroutines.
package keeper_test

import (
	"io"

	"cosmossdk.io/store"
)

// CloseStoreIfNeeded closes the CommitMultiStore if it implements io.Closer.
// This is critical for stopping IAVL nodeDB pruning goroutines that would
// otherwise leak after tests complete.
func CloseStoreIfNeeded(stateStore store.CommitMultiStore) {
	if stateStore == nil {
		return
	}
	if closer, ok := stateStore.(io.Closer); ok {
		_ = closer.Close()
	}
}

package keeper

import (
	"bytes"
	"sort"
	"sync"

	"github.com/virtengine/virtengine/x/veid/types"
)

const (
	inferenceReceiptBufferMaxHeights       = int64(8)
	inferenceReceiptBufferRetentionHeights = int64(6)
	inferenceReceiptBufferMaxReplayEntries = 256
)

type inferenceReceiptBuffer struct {
	mu              sync.Mutex
	resultsByHeight map[int64]map[string]types.VerificationResult
	replayByContext map[string]inferenceReceiptReplayEntry
	replayByNonce   map[string]string
}

type inferenceReceiptReplayEntry struct {
	height        int64
	contextDigest string
	receiptDigest string
	nonceDigest   string
	requestID     string
}

type inferenceReceiptBufferInsertResult struct {
	ExactReplay bool
}

func newInferenceReceiptBuffer() *inferenceReceiptBuffer {
	return &inferenceReceiptBuffer{
		resultsByHeight: make(map[int64]map[string]types.VerificationResult),
		replayByContext: make(map[string]inferenceReceiptReplayEntry),
		replayByNonce:   make(map[string]string),
	}
}

func (b *inferenceReceiptBuffer) insert(height int64, result types.VerificationResult, replay inferenceReceiptReplayCheck) (inferenceReceiptBufferInsertResult, error) {
	if b == nil {
		return inferenceReceiptBufferInsertResult{}, types.ErrInvalidVerificationResult.Wrap("inference receipt buffer is not configured")
	}
	if height <= 0 || result.BlockHeight != height {
		return inferenceReceiptBufferInsertResult{}, types.ErrInvalidVerificationResult.Wrap("result height does not match carrier height")
	}
	if err := result.Validate(); err != nil {
		return inferenceReceiptBufferInsertResult{}, err
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.pruneLocked(height)

	if existing, ok := b.replayByContext[replay.ContextDigest]; ok {
		if existing.receiptDigest != replay.ReceiptDigest {
			return inferenceReceiptBufferInsertResult{}, types.ErrInvalidVerificationResult.Wrap("inference receipt context replay changed digest")
		}
		if existing.height != height || existing.requestID != result.RequestID {
			return inferenceReceiptBufferInsertResult{}, types.ErrInvalidVerificationResult.Wrap("inference receipt exact replay changed carrier binding")
		}
		if nonceContext := b.replayByNonce[replay.NonceDigest]; nonceContext != "" && nonceContext != replay.ContextDigest {
			return inferenceReceiptBufferInsertResult{}, types.ErrInvalidVerificationResult.Wrap("inference receipt nonce replay changed context")
		}
		if err := b.insertResultLocked(height, result, true); err != nil {
			return inferenceReceiptBufferInsertResult{}, err
		}
		return inferenceReceiptBufferInsertResult{ExactReplay: true}, nil
	}
	if existingContext := b.replayByNonce[replay.NonceDigest]; existingContext != "" {
		return inferenceReceiptBufferInsertResult{}, types.ErrInvalidVerificationResult.Wrap("inference receipt nonce replay changed context")
	}
	if err := b.insertResultLocked(height, result, false); err != nil {
		return inferenceReceiptBufferInsertResult{}, err
	}
	b.replayByContext[replay.ContextDigest] = inferenceReceiptReplayEntry{
		height:        height,
		contextDigest: replay.ContextDigest,
		receiptDigest: replay.ReceiptDigest,
		nonceDigest:   replay.NonceDigest,
		requestID:     result.RequestID,
	}
	b.replayByNonce[replay.NonceDigest] = replay.ContextDigest
	b.pruneReplayCountLocked()
	return inferenceReceiptBufferInsertResult{}, nil
}

func (b *inferenceReceiptBuffer) stageResult(height int64, result types.VerificationResult) error {
	if b == nil {
		return types.ErrInvalidVerificationResult.Wrap("inference receipt buffer is not configured")
	}
	if height <= 0 || result.BlockHeight != height {
		return types.ErrInvalidVerificationResult.Wrap("result height does not match carrier height")
	}
	if err := result.Validate(); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pruneLocked(height)
	return b.insertResultLocked(height, result, false)
}

func (b *inferenceReceiptBuffer) insertResultLocked(height int64, result types.VerificationResult, exactReplay bool) error {
	results := b.resultsByHeight[height]
	if results == nil {
		results = make(map[string]types.VerificationResult)
		b.resultsByHeight[height] = results
	}
	if existing, ok := results[result.RequestID]; ok {
		if !inferenceReplayResultsMatch(existing, result) {
			return types.ErrInvalidVerificationResult.Wrap("inference receipt exact replay staged result mismatch")
		}
		return nil
	}
	if exactReplay {
		return types.ErrInvalidVerificationResult.Wrap("inference receipt exact replay missing staged result")
	}
	if len(results) >= MaxVoteExtensionResults {
		return types.ErrInvalidVerificationResult.Wrap("pre-consensus result limit exceeded")
	}
	results[result.RequestID] = cloneVerificationResult(result)
	return nil
}

func (b *inferenceReceiptBuffer) snapshot(height int64, currentHeight int64) []types.VerificationResult {
	if b == nil || height <= 0 {
		return []types.VerificationResult{}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pruneLocked(currentHeight)

	results := b.resultsByHeight[height]
	if len(results) == 0 {
		return []types.VerificationResult{}
	}
	out := make([]types.VerificationResult, 0, len(results))
	for _, result := range results {
		out = append(out, cloneVerificationResult(result))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].RequestID < out[j].RequestID
	})
	return out
}

func (b *inferenceReceiptBuffer) clearHeight(height int64) {
	if b == nil || height <= 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.resultsByHeight, height)
	for contextDigest, replay := range b.replayByContext {
		if replay.height == height {
			delete(b.replayByContext, contextDigest)
			delete(b.replayByNonce, replay.nonceDigest)
		}
	}
}

func (b *inferenceReceiptBuffer) pruneLocked(currentHeight int64) {
	if currentHeight <= 0 {
		return
	}
	minHeight := currentHeight - inferenceReceiptBufferRetentionHeights
	for height := range b.resultsByHeight {
		if height < minHeight {
			delete(b.resultsByHeight, height)
		}
	}
	for contextDigest, replay := range b.replayByContext {
		if replay.height < minHeight {
			delete(b.replayByContext, contextDigest)
			delete(b.replayByNonce, replay.nonceDigest)
		}
	}
	for int64(len(b.resultsByHeight)) > inferenceReceiptBufferMaxHeights {
		oldest := int64(0)
		for height := range b.resultsByHeight {
			if oldest == 0 || height < oldest {
				oldest = height
			}
		}
		delete(b.resultsByHeight, oldest)
		for contextDigest, replay := range b.replayByContext {
			if replay.height == oldest {
				delete(b.replayByContext, contextDigest)
				delete(b.replayByNonce, replay.nonceDigest)
			}
		}
	}
}

func (b *inferenceReceiptBuffer) pruneReplayCountLocked() {
	for len(b.replayByContext) > inferenceReceiptBufferMaxReplayEntries {
		var oldest inferenceReceiptReplayEntry
		oldestKey := ""
		for contextDigest, replay := range b.replayByContext {
			if oldestKey == "" || replay.height < oldest.height || (replay.height == oldest.height && contextDigest < oldestKey) {
				oldestKey = contextDigest
				oldest = replay
			}
		}
		delete(b.replayByContext, oldestKey)
		delete(b.replayByNonce, oldest.nonceDigest)
	}
}

func cloneVerificationResult(result types.VerificationResult) types.VerificationResult {
	clone := result
	clone.InputHash = bytes.Clone(result.InputHash)
	clone.ReasonCodes = append([]types.ReasonCode(nil), result.ReasonCodes...)
	clone.ScopeResults = cloneScopeVerificationResults(result.ScopeResults)
	if result.Metadata != nil {
		clone.Metadata = make(map[string]string, len(result.Metadata))
		for key, value := range result.Metadata {
			clone.Metadata[key] = value
		}
	}
	return clone
}

func cloneScopeVerificationResults(results []types.ScopeVerificationResult) []types.ScopeVerificationResult {
	if len(results) == 0 {
		return nil
	}
	out := make([]types.ScopeVerificationResult, len(results))
	copy(out, results)
	for i := range out {
		out[i].ReasonCodes = append([]types.ReasonCode(nil), results[i].ReasonCodes...)
	}
	return out
}

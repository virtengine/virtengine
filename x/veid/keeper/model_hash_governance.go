// Package keeper provides keeper functions for the VEID module.
//
// VE-21A: Model Hash Governance Integration
// This file bridges the ML scoring pipeline with on-chain model governance.
// It ensures validators use the consensus-approved model by verifying the
// model hash against on-chain state before scoring.
package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Model Hash Governance Bridge
// ============================================================================

// ValidateModelForScoring verifies the local model matches the on-chain
// approved model before allowing it for consensus scoring. Returns nil
// if valid, error if the model is not approved.
func (k Keeper) ValidateModelForScoring(ctx sdk.Context, modelType string, localModelPath string) error {
	if modelType == "" {
		return types.ErrInvalidModelType.Wrap("model type cannot be empty")
	}

	if localModelPath == "" {
		return types.ErrInvalidModelInfo.Wrap("model path cannot be empty")
	}

	// Get active model from on-chain state
	activeModel, err := k.GetActiveModel(ctx, modelType)
	if err != nil {
		// No active model on-chain: scoring with any model is not valid
		return fmt.Errorf("no active model on-chain for type %s: %w", modelType, err)
	}

	// Compute hash of local model files
	localHash, err := ComputeLocalModelHash(localModelPath)
	if err != nil {
		return fmt.Errorf("failed to compute local model hash: %w", err)
	}

	// Validate against on-chain hash
	if err := k.ValidateModelHash(ctx, modelType, localHash); err != nil {
		k.Logger(ctx).Error("model hash mismatch for scoring",
			"model_type", modelType,
			"expected_hash", activeModel.SHA256Hash,
			"local_hash", localHash,
			"model_path", localModelPath,
		)
		return err
	}

	k.Logger(ctx).Info("model hash validated for scoring",
		"model_type", modelType,
		"model_id", activeModel.ModelID,
		"hash", localHash,
	)

	return nil
}

// ComputeLocalModelHash computes SHA-256 hash of model files at a given path.
// Files are hashed in sorted order for determinism. Metadata files like
// MODEL_HASH.txt and manifest.json are excluded from the hash.
func ComputeLocalModelHash(modelPath string) (string, error) {
	info, err := os.Stat(modelPath)
	if err != nil {
		return "", fmt.Errorf("model path does not exist: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("model path is not a directory: %s", modelPath)
	}

	hasher := sha256.New()

	// Collect all files
	var files []string
	err = filepath.Walk(modelPath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		// Skip metadata files that are generated after model export
		switch fi.Name() {
		case "export_metadata.json", "MODEL_HASH.txt", "manifest.json":
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to walk model directory: %w", err)
	}

	// Sort for determinism
	sort.Strings(files)

	for _, fpath := range files {
		// Hash relative path
		rel, err := filepath.Rel(modelPath, fpath)
		if err != nil {
			return "", fmt.Errorf("failed to compute relative path: %w", err)
		}
		// Normalize path separators for cross-platform consistency
		rel = filepath.ToSlash(rel)
		hasher.Write([]byte(rel))

		// Hash file contents
		f, err := os.Open(fpath)
		if err != nil {
			return "", fmt.Errorf("failed to open %s: %w", fpath, err)
		}
		if _, err := io.Copy(hasher, f); err != nil {
			f.Close()
			return "", fmt.Errorf("failed to hash %s: %w", fpath, err)
		}
		f.Close()
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ComputeLocalModelHash is a Keeper method wrapper for the standalone ComputeLocalModelHash function.
// It computes SHA-256 hash of model files at a given path with the context available.
func (k Keeper) ComputeLocalModelHash(ctx sdk.Context, modelPath string) (string, error) {
	return ComputeLocalModelHash(modelPath)
}

// EnsureModelGovernanceCompliance checks that the validator's local model
// setup is compliant with on-chain governance requirements. This should
// be called during node startup or BeginBlock to detect stale models.
func (k Keeper) EnsureModelGovernanceCompliance(
	ctx sdk.Context,
	validatorAddr string,
	localModelVersions map[string]string, // modelType -> local hash
) (*types.ValidatorModelReport, error) {
	if validatorAddr == "" {
		return nil, types.ErrInvalidAddress.Wrap("validator address cannot be empty")
	}

	if len(localModelVersions) == 0 {
		return nil, types.ErrInvalidModelInfo.Wrap("no local model versions provided")
	}

	// Report to on-chain state
	if err := k.ReportValidatorModelVersions(ctx, validatorAddr, localModelVersions); err != nil {
		return nil, fmt.Errorf("failed to report model versions: %w", err)
	}

	// Retrieve the report
	report, found := k.GetValidatorModelReport(ctx, validatorAddr)
	if !found {
		return nil, fmt.Errorf("failed to retrieve model report for %s", validatorAddr)
	}

	if !report.IsSynced {
		k.Logger(ctx).Warn("validator model version mismatch",
			"validator", validatorAddr,
			"mismatched_models", report.MismatchedModels,
		)
	}

	return report, nil
}

// GetActiveModelHash returns the SHA-256 hash of the active model for a type.
// Returns empty string and false if no active model exists.
func (k Keeper) GetActiveModelHash(ctx sdk.Context, modelType string) (string, bool) {
	model, err := k.GetActiveModel(ctx, modelType)
	if err != nil {
		return "", false
	}
	return model.SHA256Hash, true
}

// IsModelHashApproved checks whether a given hash matches the currently
// active on-chain model for the specified type.
func (k Keeper) IsModelHashApproved(ctx sdk.Context, modelType string, hash string) bool {
	return k.ValidateModelHash(ctx, modelType, hash) == nil
}

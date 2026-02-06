# Task 29B: Model Hash Computation + Governance

**ID:** 29B  
**Title:** feat(veid): Model hash computation + governance registration  
**Priority:** P0 (Critical Blocker)  
**Wave:** 2 (Sequential after 29A)  
**Estimated LOC:** ~1000  
**Dependencies:** 29A (Trained Models)  
**Blocking:** Production VEID Deployment  

---

## Problem Statement

The VEID module has a model registry with fields for `MODEL_HASH` but these are **empty strings**. Without approved model hashes:

1. Validators cannot verify they're using governance-approved ML models
2. Malicious validators could use different models
3. VEID scores are not consensus-safe (different validators = different results)
4. Chain state becomes inconsistent

### Current State Analysis

```go
// x/veid/types/params.go - Current state
type ModelConfig struct {
    ModelHash    string  // ❌ EMPTY STRING
    ModelVersion string  // Set but not verified
    MinScore     float64
}

func DefaultParams() Params {
    return Params{
        FacialModel: ModelConfig{
            ModelHash:    "",  // ❌ Not set
            ModelVersion: "1.0.0",
            MinScore:     0.85,
        },
        // ... same for other models
    }
}
```

---

## Acceptance Criteria

### AC-1: Hash Computation Script
- [ ] Create `scripts/compute_model_hash.sh` script
- [ ] Compute SHA256 of frozen model graph (.pb files)
- [ ] Include model version in hash metadata
- [ ] Output JSON with hash, version, timestamp
- [ ] Deterministic across platforms (same bytes = same hash)

### AC-2: Default Params Update
- [ ] Update `x/veid/types/params.go` with computed hashes
- [ ] Include hash for facial_model
- [ ] Include hash for liveness_model
- [ ] Include hash for ocr_model
- [ ] Add model version alongside hash

### AC-3: Hash Verification in Inference
- [ ] Modify `pkg/inference/scorer.go` to verify model hash on load
- [ ] Compare loaded model hash against chain params
- [ ] Reject model if hash doesn't match approved list
- [ ] Log warnings for hash mismatches

### AC-4: Governance Proposal for Model Updates
- [ ] Create `MsgProposeModelHash` message type
- [ ] Implement proposal handler in x/veid keeper
- [ ] Integrate with x/gov for voting
- [ ] Support multiple model hashes (for version upgrades)
- [ ] Add CLI command `virtengine tx veid propose-model-hash`

### AC-5: CLI Commands
- [ ] `virtengine query veid model-hashes` - List approved hashes
- [ ] `virtengine tx veid propose-model-hash` - Submit proposal
- [ ] `virtengine query veid verify-model` - Verify local model against chain

### AC-6: Integration Tests
- [ ] Test hash computation matches expected values
- [ ] Test hash verification rejects tampered models
- [ ] Test governance proposal flow
- [ ] E2E test for model upgrade via governance

---

## Technical Requirements

### Hash Computation Algorithm

```bash
#!/bin/bash
# scripts/compute_model_hash.sh

compute_hash() {
    local model_path=$1
    local model_name=$2
    
    # Use frozen graph for deterministic hashing
    local frozen_pb="${model_path}/${model_name}_frozen.pb"
    
    if [ ! -f "$frozen_pb" ]; then
        echo "Error: Frozen graph not found: $frozen_pb" >&2
        exit 1
    fi
    
    # SHA256 hash of the frozen graph
    local hash=$(sha256sum "$frozen_pb" | cut -d' ' -f1)
    
    # Get model version from metrics.json
    local version=$(jq -r '.version' "${model_path}/metrics.json")
    
    # Output JSON
    cat << EOF
{
    "model_name": "$model_name",
    "model_hash": "$hash",
    "model_version": "$version",
    "algorithm": "sha256",
    "computed_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "source_file": "$frozen_pb"
}
EOF
}

# Compute all model hashes
compute_hash "ml/facial_verification/weights" "facial_model"
compute_hash "ml/liveness_detection/weights" "liveness_model"
compute_hash "ml/ocr_extraction/weights" "ocr_model"
```

### Updated Params Structure

```go
// x/veid/types/params.go

type ModelConfig struct {
    // ModelHash is the SHA256 hash of the frozen model graph
    // Used to verify validators are running approved models
    ModelHash string `json:"model_hash"`
    
    // ModelVersion is the semantic version of the model
    ModelVersion string `json:"model_version"`
    
    // MinScore is the minimum acceptable score for verification
    MinScore float64 `json:"min_score"`
    
    // ActivationHeight is the block height when this model becomes active
    // Allows for coordinated model upgrades across all validators
    ActivationHeight int64 `json:"activation_height,omitempty"`
}

func DefaultParams() Params {
    return Params{
        FacialModel: ModelConfig{
            ModelHash:    "a1b2c3d4e5f6...", // Computed from 29A
            ModelVersion: "1.0.0",
            MinScore:     0.85,
        },
        LivenessModel: ModelConfig{
            ModelHash:    "f6e5d4c3b2a1...", // Computed from 29A
            ModelVersion: "1.0.0",
            MinScore:     0.90,
        },
        OcrModel: ModelConfig{
            ModelHash:    "1a2b3c4d5e6f...", // Computed from 29A
            ModelVersion: "1.0.0",
            MinScore:     0.95,
        },
    }
}
```

### Hash Verification in Inference

```go
// pkg/inference/scorer.go

type Scorer struct {
    modelPath    string
    expectedHash string
    verified     bool
}

func NewScorer(modelPath string, expectedHash string) (*Scorer, error) {
    s := &Scorer{
        modelPath:    modelPath,
        expectedHash: expectedHash,
    }
    
    // Verify model hash before allowing inference
    if err := s.verifyModelHash(); err != nil {
        return nil, fmt.Errorf("model hash verification failed: %w", err)
    }
    
    return s, nil
}

func (s *Scorer) verifyModelHash() error {
    frozenPath := filepath.Join(s.modelPath, "model_frozen.pb")
    
    // Read frozen graph
    data, err := os.ReadFile(frozenPath)
    if err != nil {
        return fmt.Errorf("failed to read frozen graph: %w", err)
    }
    
    // Compute SHA256
    hash := sha256.Sum256(data)
    computedHash := hex.EncodeToString(hash[:])
    
    // Compare with expected
    if computedHash != s.expectedHash {
        return fmt.Errorf(
            "model hash mismatch: expected %s, got %s",
            s.expectedHash,
            computedHash,
        )
    }
    
    s.verified = true
    return nil
}

func (s *Scorer) Score(input []byte) (float64, error) {
    if !s.verified {
        return 0, errors.New("model not verified - hash check failed")
    }
    // ... actual inference
}
```

### Governance Proposal Message

```go
// x/veid/types/msgs.go

// MsgProposeModelHash proposes a new model hash for governance approval
type MsgProposeModelHash struct {
    Authority    string `json:"authority"`     // gov module account
    ModelName    string `json:"model_name"`    // facial, liveness, ocr
    ModelHash    string `json:"model_hash"`    // SHA256 hash
    ModelVersion string `json:"model_version"` // Semantic version
    ActivationHeight int64 `json:"activation_height"` // When to activate
    Justification string `json:"justification"` // Why this model
}

func (msg MsgProposeModelHash) ValidateBasic() error {
    if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
        return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority: %s", err)
    }
    
    validModels := map[string]bool{
        "facial": true, "liveness": true, "ocr": true,
    }
    if !validModels[msg.ModelName] {
        return fmt.Errorf("invalid model name: %s", msg.ModelName)
    }
    
    if len(msg.ModelHash) != 64 { // SHA256 hex = 64 chars
        return fmt.Errorf("invalid hash length: expected 64, got %d", len(msg.ModelHash))
    }
    
    return nil
}
```

### CLI Implementation

```go
// cmd/virtengine/cmd/tx/veid/propose_model_hash.go

func ProposeModelHashCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "propose-model-hash [model-name] [hash] [version]",
        Short: "Propose a new model hash for governance approval",
        Args:  cobra.ExactArgs(3),
        RunE: func(cmd *cobra.Command, args []string) error {
            clientCtx, err := client.GetClientTxContext(cmd)
            if err != nil {
                return err
            }
            
            modelName := args[0]
            modelHash := args[1]
            modelVersion := args[2]
            
            activationHeight, _ := cmd.Flags().GetInt64("activation-height")
            justification, _ := cmd.Flags().GetString("justification")
            
            msg := &types.MsgProposeModelHash{
                Authority:        authtypes.NewModuleAddress(govtypes.ModuleName).String(),
                ModelName:        modelName,
                ModelHash:        modelHash,
                ModelVersion:     modelVersion,
                ActivationHeight: activationHeight,
                Justification:    justification,
            }
            
            return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
        },
    }
    
    cmd.Flags().Int64("activation-height", 0, "Block height to activate new model")
    cmd.Flags().String("justification", "", "Justification for model change")
    
    return cmd
}
```

---

## Files to Create/Modify

### New Files
| Path | Description |
|------|-------------|
| `scripts/compute_model_hash.sh` | Hash computation script |
| `scripts/compute_model_hash.go` | Go version for CI |
| `x/veid/keeper/model_registry.go` | Model hash registry keeper |
| `x/veid/types/msg_propose_model_hash.go` | Proposal message |
| `cmd/virtengine/cmd/tx/veid/propose_model_hash.go` | CLI command |
| `cmd/virtengine/cmd/query/veid/model_hashes.go` | Query CLI |
| `tests/e2e/veid_model_hash_test.go` | E2E tests |

### Files to Modify
| Path | Changes |
|------|---------|
| `x/veid/types/params.go` | Add computed hashes to defaults |
| `x/veid/keeper/keeper.go` | Add model verification |
| `pkg/inference/scorer.go` | Add hash verification before inference |
| `x/veid/module.go` | Register new message handlers |

---

## Implementation Steps

### Step 1: Run Hash Computation
```bash
# After 29A completes
./scripts/compute_model_hash.sh > model_hashes.json
```

### Step 2: Update Default Params
Copy computed hashes into `x/veid/types/params.go`

### Step 3: Implement Hash Verification
Add `verifyModelHash()` to `pkg/inference/scorer.go`

### Step 4: Add Governance Proposal
Create `MsgProposeModelHash` and handler

### Step 5: Add CLI Commands
Implement query and tx commands

### Step 6: Write Tests
- Unit tests for hash computation
- Unit tests for hash verification
- E2E tests for governance flow

---

## Validation Checklist

- [ ] Hash computation script works on all platforms
- [ ] Default params updated with real hashes
- [ ] Inference package verifies hash before scoring
- [ ] Governance proposal can update model hashes
- [ ] CLI commands work correctly
- [ ] Tests pass

---

## Vibe-Kanban Task ID

`ce12c02a-6b82-4f06-a78a-a07f38151c51`

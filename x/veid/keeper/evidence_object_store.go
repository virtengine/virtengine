package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/data_vault/contracts"
	"github.com/virtengine/virtengine/x/veid/types"
)

const evidenceObjectReferenceKeyKind = "reference/"

func (k Keeper) SetEvidenceObjectRef(ctx sdk.Context, ref contracts.EvidenceObjectRef) error {
	if err := ref.Validate(); err != nil {
		return fmt.Errorf("validate evidence object reference: %w", err)
	}
	store := ctx.KVStore(k.skey)
	key := evidenceObjectRefKey(ref.ObjectCommitment)
	if existingBytes := store.Get(key); existingBytes != nil {
		var existing contracts.EvidenceObjectRef
		if err := json.Unmarshal(existingBytes, &existing); err != nil || existing.Validate() != nil || existing.ObjectCommitment != ref.ObjectCommitment {
			return errors.New("existing evidence object reference is invalid")
		}
		if existing == ref {
			return nil
		}
		if !sameEvidenceObjectIdentity(existing, ref) {
			return errors.New("evidence object commitment-bound fields are immutable")
		}
		if existing.State == contracts.RetentionDeleted {
			return errors.New("deleted evidence object state is terminal")
		}
		if existing.State == contracts.RetentionLegalHold && ref.State != contracts.RetentionLegalHold {
			return errors.New("evidence object legal hold cannot be bypassed")
		}
		if !contracts.CanTransitionRetention(existing.State, ref.State) {
			return fmt.Errorf("invalid retention transition %s -> %s", existing.State, ref.State)
		}
		if ref.UpdatedHeight < existing.UpdatedHeight || ref.UpdatedUnix < existing.UpdatedUnix {
			return errors.New("evidence object update coordinates must be monotonic")
		}
	}
	bz, err := json.Marshal(ref)
	if err != nil {
		return fmt.Errorf("marshal evidence object reference: %w", err)
	}
	if err := validatePayloadFreeEvidenceJSON(bz); err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

func sameEvidenceObjectIdentity(first, second contracts.EvidenceObjectRef) bool {
	return first.Version == second.Version &&
		first.CommitmentDomain == second.CommitmentDomain &&
		first.ObjectCommitment == second.ObjectCommitment &&
		first.Fields() == second.Fields()
}

func (k Keeper) GetEvidenceObjectRef(ctx sdk.Context, objectCommitment string) (contracts.EvidenceObjectRef, bool) {
	bz := ctx.KVStore(k.skey).Get(evidenceObjectRefKey(objectCommitment))
	if bz == nil {
		return contracts.EvidenceObjectRef{}, false
	}
	var ref contracts.EvidenceObjectRef
	if json.Unmarshal(bz, &ref) != nil || ref.Validate() != nil || ref.ObjectCommitment != objectCommitment {
		return contracts.EvidenceObjectRef{}, false
	}
	return ref, true
}

func evidenceObjectRefKey(objectCommitment string) []byte {
	key := make([]byte, 0, len(types.PrefixEvidenceObjectRef)+len(evidenceObjectReferenceKeyKind)+len(objectCommitment))
	key = append(key, types.PrefixEvidenceObjectRef...)
	key = append(key, evidenceObjectReferenceKeyKind...)
	return append(key, objectCommitment...)
}

func validatePayloadFreeEvidenceJSON(bz []byte) error {
	var value map[string]json.RawMessage
	if err := json.Unmarshal(bz, &value); err != nil {
		return fmt.Errorf("decode evidence reference JSON: %w", err)
	}
	for _, forbidden := range []string{
		"backend_uri", "backend_ref", "ciphertext", "wrapped_key", "nonce", "opening",
		"raw_evidence", "biometric_template", "issuer_subject", "subject_id", "plaintext_identifier",
	} {
		for key := range value {
			if strings.Contains(strings.ToLower(key), forbidden) {
				return errors.New("evidence reference contains forbidden field " + key)
			}
		}
	}
	return nil
}

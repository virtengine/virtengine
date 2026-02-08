package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// =====================================================================
// Email Verification Storage
// =====================================================================

func (k Keeper) SetEmailVerificationRecord(ctx sdk.Context, record *types.EmailVerificationRecord) error {
	if record == nil {
		return types.ErrInvalidEmail.Wrap("email verification record cannot be nil")
	}
	if err := record.Validate(); err != nil {
		return err
	}

	bz, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal email verification record: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(k.emailVerificationKey(record.VerificationID), bz)

	if record.EmailHash != "" {
		store.Set(k.emailVerificationByAccountKey(record.AccountAddress, record.EmailHash), []byte(record.VerificationID))
	}

	return nil
}

func (k Keeper) GetEmailVerificationRecord(ctx sdk.Context, verificationID string) (*types.EmailVerificationRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(k.emailVerificationKey(verificationID))
	if bz == nil {
		return nil, false
	}

	var record types.EmailVerificationRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		k.Logger(ctx).Error("failed to unmarshal email verification record", "error", err, "verification_id", verificationID)
		return nil, false
	}

	return &record, true
}

func (k Keeper) GetEmailVerificationByAccountAndHash(ctx sdk.Context, account string, emailHash string) (string, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(k.emailVerificationByAccountKey(account, emailHash))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

// =====================================================================
// Email Nonce Tracking
// =====================================================================

func (k Keeper) IsEmailNonceUsed(ctx sdk.Context, nonceHash string) bool {
	store := ctx.KVStore(k.skey)
	return store.Has(k.emailNonceKey(nonceHash))
}

func (k Keeper) SetEmailUsedNonce(ctx sdk.Context, record *types.UsedNonceRecord) error {
	if record == nil {
		return types.ErrInvalidEmail.Wrap("used nonce record cannot be nil")
	}

	bz, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal used nonce record: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(k.emailNonceKey(record.NonceHash), bz)
	return nil
}

// =====================================================================
// Key Helpers
// =====================================================================

func (k Keeper) emailVerificationKey(verificationID string) []byte {
	key := make([]byte, 0, len(types.PrefixEmailVerification)+len(verificationID))
	key = append(key, types.PrefixEmailVerification...)
	key = append(key, []byte(verificationID)...)
	return key
}

func (k Keeper) emailVerificationByAccountKey(account string, emailHash string) []byte {
	key := make([]byte, 0, len(types.PrefixEmailByAccount)+len(account)+1+len(emailHash))
	key = append(key, types.PrefixEmailByAccount...)
	key = append(key, []byte(account)...)
	key = append(key, byte('/'))
	key = append(key, []byte(emailHash)...)
	return key
}

func (k Keeper) emailNonceKey(nonceHash string) []byte {
	key := make([]byte, 0, len(types.PrefixUsedNonce)+len(nonceHash))
	key = append(key, types.PrefixUsedNonce...)
	key = append(key, []byte(nonceHash)...)
	return key
}

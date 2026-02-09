package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// =====================================================================
// SMS Verification Storage
// =====================================================================

func (k Keeper) SetSMSVerificationRecord(ctx sdk.Context, record *types.SMSVerificationRecord) error {
	if record == nil {
		return types.ErrInvalidPhone.Wrap("sms verification record cannot be nil")
	}
	if err := record.Validate(); err != nil {
		return err
	}

	bz, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal sms verification record: %w", err)
	}

	store := ctx.KVStore(k.skey)
	store.Set(k.smsVerificationKey(record.VerificationID), bz)

	if record.PhoneHash.Hash != "" {
		store.Set(k.smsVerificationByAccountKey(record.AccountAddress, record.PhoneHash.Hash), []byte(record.VerificationID))
	}

	return nil
}

func (k Keeper) GetSMSVerificationRecord(ctx sdk.Context, verificationID string) (*types.SMSVerificationRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(k.smsVerificationKey(verificationID))
	if bz == nil {
		return nil, false
	}

	var record types.SMSVerificationRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		k.Logger(ctx).Error("failed to unmarshal sms verification record", "error", err, "verification_id", verificationID)
		return nil, false
	}

	return &record, true
}

func (k Keeper) GetSMSVerificationByAccountAndHash(ctx sdk.Context, account string, phoneHash string) (string, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(k.smsVerificationByAccountKey(account, phoneHash))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

// =====================================================================
// Key Helpers
// =====================================================================

func (k Keeper) smsVerificationKey(verificationID string) []byte {
	key := make([]byte, 0, len(types.PrefixSMSVerification)+len(verificationID))
	key = append(key, types.PrefixSMSVerification...)
	key = append(key, []byte(verificationID)...)
	return key
}

func (k Keeper) smsVerificationByAccountKey(account string, phoneHash string) []byte {
	key := make([]byte, 0, len(types.PrefixSMSByAccount)+len(account)+1+len(phoneHash))
	key = append(key, types.PrefixSMSByAccount...)
	key = append(key, []byte(account)...)
	key = append(key, byte('/'))
	key = append(key, []byte(phoneHash)...)
	return key
}

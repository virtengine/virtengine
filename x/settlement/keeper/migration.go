package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/types"
)

// SetMigrationAuditEntry stores a migration audit entry.
func (k Keeper) SetMigrationAuditEntry(ctx sdk.Context, entry types.MigrationAuditEntry) error {
	if err := entry.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&entry)
	if err != nil {
		return err
	}
	store.Set(types.MigrationAuditKey(entry.RecordType, entry.RecordID), bz)
	return nil
}

// MigrateEncryptedPayloads backfills and clears legacy plaintext fields.
func (k Keeper) MigrateEncryptedPayloads(ctx sdk.Context) error {
	now := ctx.BlockTime().Unix()

	var migrateErr error

	k.WithFiatPayoutPreferences(ctx, func(pref types.FiatPayoutPreference) bool {
		cleared := make([]string, 0, 2)
		if pref.DestinationRef != "" && (pref.EncryptedPayload == nil || pref.DestinationRef != pref.EncryptedPayload.EnvelopeRef) {
			if pref.DestinationHash == "" {
				pref.DestinationHash = types.HashDestination(pref.DestinationRef)
			}
			pref.DestinationRef = ""
			cleared = append(cleared, "destination_ref")
		}
		if pref.EncryptedPayload == nil && pref.Enabled {
			pref.Enabled = false
			cleared = append(cleared, "enabled")
		}

		if len(cleared) == 0 {
			return false
		}

		if err := k.SetFiatPayoutPreference(ctx, pref); err != nil {
			migrateErr = err
			return true
		}

		_ = k.SetMigrationAuditEntry(ctx, types.MigrationAuditEntry{
			RecordType:    "fiat_payout_preference",
			RecordID:      pref.Provider,
			ClearedFields: cleared,
			Timestamp:     now,
			Note:          "cleared plaintext payout preference fields during encryption migration",
		})

		return false
	})

	if migrateErr != nil {
		return migrateErr
	}

	k.WithFiatConversions(ctx, func(conversion types.FiatConversionRecord) bool {
		cleared := make([]string, 0, 1)
		if conversion.DestinationRef != "" && (conversion.EncryptedPayload == nil || conversion.DestinationRef != conversion.EncryptedPayload.EnvelopeRef) {
			if conversion.DestinationHash == "" {
				conversion.DestinationHash = types.HashDestination(conversion.DestinationRef)
			}
			conversion.DestinationRef = ""
			cleared = append(cleared, "destination_ref")
		}

		if len(cleared) == 0 {
			return false
		}

		if err := k.SetFiatConversion(ctx, conversion); err != nil {
			migrateErr = err
			return true
		}

		_ = k.SetMigrationAuditEntry(ctx, types.MigrationAuditEntry{
			RecordType:    "fiat_conversion",
			RecordID:      conversion.ConversionID,
			ClearedFields: cleared,
			Timestamp:     now,
			Note:          "cleared plaintext conversion fields during encryption migration",
		})

		return false
	})

	return migrateErr
}

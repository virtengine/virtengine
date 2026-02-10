package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/encryption/types"
)

// EnvelopeRecord stores an encrypted envelope for rotation tracking.
type EnvelopeRecord struct {
	Envelope      types.EncryptedPayloadEnvelope `json:"envelope"`
	CreatedAt     int64                          `json:"created_at"`
	UpdatedAt     int64                          `json:"updated_at"`
	RotationCount uint32                         `json:"rotation_count"`
}

// ReencryptionJobStatus tracks the status of a reencryption job.
type ReencryptionJobStatus string

const (
	ReencryptionJobPending   ReencryptionJobStatus = "pending"
	ReencryptionJobCompleted ReencryptionJobStatus = "completed"
	ReencryptionJobFailed    ReencryptionJobStatus = "failed"
)

// ReencryptionJob tracks an envelope reencryption request.
type ReencryptionJob struct {
	JobID          string                `json:"job_id"`
	EnvelopeHash   string                `json:"envelope_hash"`
	OldFingerprint string                `json:"old_fingerprint"`
	NewFingerprint string                `json:"new_fingerprint"`
	Status         ReencryptionJobStatus `json:"status"`
	Attempts       uint32                `json:"attempts"`
	LastError      string                `json:"last_error,omitempty"`
	CreatedAt      int64                 `json:"created_at"`
	UpdatedAt      int64                 `json:"updated_at"`
}

// rotationState stores rotation progress alongside a record.
type rotationState struct {
	Record types.KeyRotationRecord `json:"record"`
	Cursor []byte                  `json:"cursor,omitempty"`
}

// ReencryptionWorker performs envelope re-encryption.
type ReencryptionWorker interface {
	ReencryptEnvelope(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope, oldKey, newKey types.RecipientKeyRecord) (*types.EncryptedPayloadEnvelope, error)
}

// RotateRecipientKey rotates a key and queues reencryption jobs for affected envelopes.
func (k Keeper) RotateRecipientKey(ctx sdk.Context, address sdk.AccAddress, oldFingerprint string, newPublicKey []byte, newAlgorithmID, newLabel, reason string, newKeyTTLSeconds uint64) (string, error) {
	params := k.GetParams(ctx)

	oldRecord, found := k.GetRecipientKeyByFingerprint(ctx, oldFingerprint)
	if !found {
		return "", types.ErrKeyNotFound.Wrapf("key fingerprint %s not found", oldFingerprint)
	}

	if oldRecord.RevokedAt != 0 {
		return "", types.ErrKeyRevoked.Wrapf("key %s is revoked", oldFingerprint)
	}

	if newAlgorithmID == "" {
		newAlgorithmID = oldRecord.AlgorithmID
	}

	if !types.IsAlgorithmAllowed(&params, newAlgorithmID) {
		return "", types.ErrUnsupportedAlgorithm.Wrapf("algorithm %s is not allowed", newAlgorithmID)
	}

	newAlgInfo, err := types.GetAlgorithmInfo(newAlgorithmID)
	if err != nil {
		return "", err
	}
	if len(newPublicKey) != newAlgInfo.KeySize {
		return "", types.ErrInvalidPublicKey.Wrapf("expected %d bytes, got %d", newAlgInfo.KeySize, len(newPublicKey))
	}

	newFingerprint, err := k.RegisterRecipientKey(ctx, address, newPublicKey, newAlgorithmID, newLabel)
	if err != nil {
		return "", err
	}

	if newKeyTTLSeconds > 0 {
		if err := k.applyKeyTTL(ctx, newFingerprint, newKeyTTLSeconds); err != nil {
			return "", err
		}
	}

	if err := k.deprecateKey(ctx, oldFingerprint); err != nil {
		return "", err
	}

	rotationID := fmt.Sprintf("%s-%d", oldFingerprint, ctx.BlockHeight())
	oldAlgInfo, err := types.GetAlgorithmInfo(oldRecord.AlgorithmID)
	if err != nil {
		return "", err
	}

	rotationReason := types.KeyRotationReason(reason)
	if rotationReason == "" {
		rotationReason = types.KeyRotationScheduled
	}

	record := types.NewKeyRotationRecord(
		rotationID,
		address.String(),
		rotationReason,
		oldRecord.AlgorithmID,
		oldAlgInfo.Version,
		newAlgorithmID,
		newAlgInfo.Version,
		oldFingerprint,
		newFingerprint,
		ctx.BlockTime(),
		7,
	)
	record.Status = types.KeyRotationStatusInTransition

	state := rotationState{
		Record: *record,
	}

	if err := k.setRotationState(ctx, rotationID, state); err != nil {
		return "", err
	}

	queued, done, cursor, err := k.queueReencryptionJobs(ctx, oldFingerprint, newFingerprint, params.RotationBatchSize, nil)
	if err != nil {
		return "", err
	}

	state.Record.EnvelopesPending = uint64(queued)
	if !done {
		state.Cursor = cursor
	}
	if err := k.setRotationState(ctx, rotationID, state); err != nil {
		return "", err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventKeyRotatedPB{
		Address:        address.String(),
		OldFingerprint: oldFingerprint,
		NewFingerprint: newFingerprint,
		RotatedAt:      ctx.BlockTime().Unix(),
	}); err != nil {
		return "", err
	}

	if k.hooks != nil {
		_ = k.hooks.AfterKeyRotated(ctx, address, oldFingerprint, newFingerprint)
	}

	return newFingerprint, nil
}

func (k Keeper) deprecateKey(ctx sdk.Context, fingerprint string) error {
	record, found := k.GetRecipientKeyByFingerprint(ctx, fingerprint)
	if !found {
		return types.ErrKeyNotFound.Wrapf("key fingerprint %s not found", fingerprint)
	}

	if record.DeprecatedAt != 0 {
		return nil
	}

	addr, err := sdk.AccAddressFromBech32(record.Address)
	if err != nil {
		return err
	}
	storeRecord, ok := k.getRecipientKeyStore(ctx, addr, fingerprint)
	if !ok {
		return types.ErrKeyNotFound.Wrapf("key fingerprint %s not found", fingerprint)
	}

	storeRecord.DeprecatedAt = ctx.BlockTime().Unix()
	return k.setRecipientKeyStore(ctx, storeRecord)
}

func (k Keeper) applyKeyTTL(ctx sdk.Context, fingerprint string, ttlSeconds uint64) error {
	record, found := k.GetRecipientKeyByFingerprint(ctx, fingerprint)
	if !found {
		return types.ErrKeyNotFound.Wrapf("key fingerprint %s not found", fingerprint)
	}

	addr, err := sdk.AccAddressFromBech32(record.Address)
	if err != nil {
		return err
	}

	storeRecord, ok := k.getRecipientKeyStore(ctx, addr, fingerprint)
	if !ok {
		return types.ErrKeyNotFound.Wrapf("key fingerprint %s not found", fingerprint)
	}

	ttl := safeInt64FromUint64(ttlSeconds)
	storeRecord.ExpiresAt = ctx.BlockTime().Add(time.Duration(ttl) * time.Second).Unix()
	return k.setRecipientKeyStore(ctx, storeRecord)
}

// StoreEnvelope stores an envelope record.
func (k Keeper) StoreEnvelope(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope) ([]byte, error) {
	if envelope == nil {
		return nil, types.ErrInvalidEnvelope.Wrap("envelope cannot be nil")
	}

	hash := envelope.Hash()
	record := EnvelopeRecord{
		Envelope:      *envelope,
		CreatedAt:     ctx.BlockTime().Unix(),
		UpdatedAt:     ctx.BlockTime().Unix(),
		RotationCount: 0,
	}

	bz, err := json.Marshal(&record)
	if err != nil {
		return nil, err
	}

	ctx.KVStore(k.skey).Set(types.EnvelopeRecordKey(hash), bz)
	return hash, nil
}

// GetEnvelope retrieves an envelope record.
func (k Keeper) GetEnvelope(ctx sdk.Context, hash []byte) (EnvelopeRecord, bool) {
	bz := ctx.KVStore(k.skey).Get(types.EnvelopeRecordKey(hash))
	if bz == nil {
		return EnvelopeRecord{}, false
	}

	var record EnvelopeRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return EnvelopeRecord{}, false
	}
	return record, true
}

// ProcessReencryptionJobs processes queued reencryption jobs.
func (k Keeper) ProcessReencryptionJobs(ctx sdk.Context, limit uint32, worker ReencryptionWorker) (uint32, error) {
	if worker == nil {
		return 0, types.ErrReencryptionJobFailed.Wrap("reencryption worker is nil")
	}

	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixReencryptionJob)
	defer func() {
		_ = iter.Close()
	}()

	var processed uint32
	for ; iter.Valid(); iter.Next() {
		if processed >= limit {
			break
		}

		var job ReencryptionJob
		if err := json.Unmarshal(iter.Value(), &job); err != nil {
			continue
		}

		if job.Status != ReencryptionJobPending {
			continue
		}

		envelopeHash, err := hex.DecodeString(job.EnvelopeHash)
		if err != nil {
			job.Status = ReencryptionJobFailed
			job.LastError = "invalid envelope hash"
			job.UpdatedAt = ctx.BlockTime().Unix()
			_ = k.setReencryptionJob(ctx, job)
			continue
		}
		record, ok := k.GetEnvelope(ctx, envelopeHash)
		if !ok {
			job.Status = ReencryptionJobFailed
			job.LastError = "envelope not found"
			job.UpdatedAt = ctx.BlockTime().Unix()
			_ = k.setReencryptionJob(ctx, job)
			continue
		}

		oldKey, ok := k.GetRecipientKeyByFingerprint(ctx, job.OldFingerprint)
		if !ok {
			job.Status = ReencryptionJobFailed
			job.LastError = "old key not found"
			job.UpdatedAt = ctx.BlockTime().Unix()
			_ = k.setReencryptionJob(ctx, job)
			continue
		}

		newKey, ok := k.GetRecipientKeyByFingerprint(ctx, job.NewFingerprint)
		if !ok {
			job.Status = ReencryptionJobFailed
			job.LastError = "new key not found"
			job.UpdatedAt = ctx.BlockTime().Unix()
			_ = k.setReencryptionJob(ctx, job)
			continue
		}

		updatedEnvelope, err := worker.ReencryptEnvelope(ctx, &record.Envelope, oldKey, newKey)
		if err != nil {
			job.Attempts++
			job.LastError = err.Error()
			job.UpdatedAt = ctx.BlockTime().Unix()
			if job.Attempts >= 3 {
				job.Status = ReencryptionJobFailed
			}
			_ = k.setReencryptionJob(ctx, job)
			continue
		}

		record.Envelope = *updatedEnvelope
		record.RotationCount++
		record.UpdatedAt = ctx.BlockTime().Unix()
		if err := k.setEnvelopeRecord(ctx, envelopeHash, record); err != nil {
			job.Status = ReencryptionJobFailed
			job.LastError = err.Error()
			job.UpdatedAt = ctx.BlockTime().Unix()
			_ = k.setReencryptionJob(ctx, job)
			continue
		}

		job.Status = ReencryptionJobCompleted
		job.UpdatedAt = ctx.BlockTime().Unix()
		_ = k.setReencryptionJob(ctx, job)
		processed++
	}

	return processed, nil
}

func (k Keeper) queueReencryptionJobs(ctx sdk.Context, oldFingerprint, newFingerprint string, limit uint32, startAfter []byte) (uint32, bool, []byte, error) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixEnvelopeRecord)
	defer func() {
		_ = iter.Close()
	}()

	var queued uint32
	var lastKey []byte
	skipping := len(startAfter) > 0

	for ; iter.Valid(); iter.Next() {
		if skipping {
			if bytes.Compare(iter.Key(), startAfter) <= 0 {
				continue
			}
			skipping = false
		}

		if queued >= limit {
			return queued, false, iter.Key(), nil
		}

		var record EnvelopeRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}

		if !envelopeReferencesKey(&record.Envelope, oldFingerprint) {
			continue
		}

		jobID := buildJobID(iter.Key(), oldFingerprint, newFingerprint)
		envelopeHash := iter.Key()[len(types.PrefixEnvelopeRecord):]
		job := ReencryptionJob{
			JobID:          jobID,
			EnvelopeHash:   hex.EncodeToString(envelopeHash),
			OldFingerprint: oldFingerprint,
			NewFingerprint: newFingerprint,
			Status:         ReencryptionJobPending,
			Attempts:       0,
			CreatedAt:      ctx.BlockTime().Unix(),
			UpdatedAt:      ctx.BlockTime().Unix(),
		}

		if err := k.setReencryptionJob(ctx, job); err != nil {
			return queued, false, lastKey, err
		}

		queued++
		lastKey = iter.Key()
	}

	return queued, true, lastKey, nil
}

func envelopeReferencesKey(envelope *types.EncryptedPayloadEnvelope, fingerprint string) bool {
	for _, keyID := range envelope.RecipientKeyIDs {
		if types.NormalizeRecipientKeyID(keyID) == fingerprint {
			return true
		}
	}
	return false
}

func buildJobID(envelopeKey []byte, oldFingerprint, newFingerprint string) string {
	h := sha256.New()
	h.Write(envelopeKey)
	h.Write([]byte(oldFingerprint))
	h.Write([]byte(newFingerprint))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (k Keeper) setEnvelopeRecord(ctx sdk.Context, hash []byte, record EnvelopeRecord) error {
	bz, err := json.Marshal(&record)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.EnvelopeRecordKey(hash), bz)
	return nil
}

func (k Keeper) setReencryptionJob(ctx sdk.Context, job ReencryptionJob) error {
	bz, err := json.Marshal(&job)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.ReencryptionJobKey([]byte(job.JobID)), bz)
	return nil
}

func (k Keeper) setRotationState(ctx sdk.Context, rotationID string, state rotationState) error {
	bz, err := json.Marshal(&state)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.KeyRotationRecordKey([]byte(rotationID)), bz)
	return nil
}

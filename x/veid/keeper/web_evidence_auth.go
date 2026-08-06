package keeper

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

const (
	maxWebEvidenceClockSkew = 5 * time.Minute
	maxWebEvidenceAge       = 24 * time.Hour
)

type webEvidenceValidation struct {
	SignerKey         *types.SignerKeyInfo
	FullContextDigest string
	GlobalNonceDigest string
	MetadataDigest    string
	ExactReplay       bool
}

func (k Keeper) validateWebEvidenceSubmission(
	ctx sdk.Context,
	att *types.VerificationAttestation,
	evidence types.WebEvidenceContext,
	accountSignature []byte,
) (*webEvidenceValidation, error) {
	if att == nil {
		return nil, types.ErrInvalidAttestation.Wrap("attestation cannot be nil")
	}
	if err := att.Validate(); err != nil {
		return nil, err
	}
	if err := evidence.ValidateAttestationMetadata(att); err != nil {
		return nil, err
	}
	if evidence.ChainID != ctx.ChainID() {
		return nil, types.ErrInvalidAttestation.Wrap("web evidence chain_id mismatch")
	}
	if evidence.AccountAddress != att.Subject.AccountAddress {
		return nil, types.ErrInvalidAttestation.Wrap("web evidence account address mismatch")
	}
	if att.Subject.ID != "did:virtengine:"+att.Subject.AccountAddress {
		return nil, types.ErrInvalidAttestation.Wrap("web evidence subject ID does not match account address")
	}
	if evidence.EvidenceType != att.Type {
		return nil, types.ErrInvalidAttestation.Wrap("web evidence type mismatch")
	}
	if evidence.IssuerID != att.Issuer.ID ||
		evidence.IssuerKeyID != att.Issuer.KeyID ||
		evidence.IssuerFingerprint != att.Issuer.KeyFingerprint {
		return nil, types.ErrInvalidAttestation.Wrap("web evidence issuer mismatch")
	}
	if evidence.IssuerAlgorithm != att.Proof.Type {
		return nil, types.ErrInvalidAttestation.Wrap("web evidence proof algorithm mismatch")
	}
	if evidence.Nonce != att.Nonce || evidence.Nonce != att.Proof.Nonce {
		return nil, types.ErrInvalidAttestation.Wrap("web evidence nonce mismatch")
	}
	if !evidence.IssuedAt.Equal(att.IssuedAt.UTC()) || !evidence.ExpiresAt.Equal(att.ExpiresAt.UTC()) {
		return nil, types.ErrInvalidTimestamp.Wrap("web evidence attestation timestamp mismatch")
	}
	if !att.Proof.Created.UTC().Equal(evidence.IssuedAt) {
		return nil, types.ErrInvalidTimestamp.Wrap("web evidence proof created timestamp mismatch")
	}
	if att.Proof.VerificationMethod != evidence.IssuerID+"#"+evidence.IssuerKeyID {
		return nil, types.ErrInvalidAttestation.Wrap("web evidence proof verification method mismatch")
	}
	if att.Proof.ProofPurpose != "assertionMethod" {
		return nil, types.ErrInvalidAttestation.Wrap("web evidence proof purpose mismatch")
	}

	key, err := k.resolveSignerKey(ctx, evidence.IssuerKeyID, evidence.IssuerFingerprint)
	if err != nil {
		return nil, err
	}
	if err := k.validateWebEvidenceSignerKey(ctx, key, evidence); err != nil {
		return nil, err
	}
	if err := k.verifyWebEvidenceIssuerSignature(att, evidence, key); err != nil {
		return nil, err
	}
	if err := k.verifyWebEvidenceAccountAuthorization(ctx, evidence, accountSignature); err != nil {
		return nil, err
	}
	if err := validateWebEvidenceFreshness(ctx, evidence); err != nil {
		return nil, err
	}
	return k.checkWebEvidenceReplay(ctx, evidence, att.Metadata, key)
}

func (k Keeper) validateWebEvidenceSignerKey(ctx sdk.Context, key *types.SignerKeyInfo, evidence types.WebEvidenceContext) error {
	if key == nil {
		return types.ErrSignerKeyNotFound.Wrap("signer key not found")
	}
	if err := key.Validate(); err != nil {
		return err
	}
	if key.KeyID != evidence.IssuerKeyID || key.Fingerprint != evidence.IssuerFingerprint {
		return types.ErrInvalidSignerKey.Wrap("signer key does not match evidence issuer")
	}
	if key.SignerID != evidence.IssuerID {
		return types.ErrInvalidSignerKey.Wrap("signer key signer_id does not match evidence issuer")
	}
	if key.Algorithm != evidence.IssuerAlgorithm || key.Algorithm != types.ProofTypeEd25519 {
		return types.ErrInvalidSignerKey.Wrapf("unsupported web evidence signer algorithm: %s", key.Algorithm)
	}
	if !key.State.CanVerify() {
		if key.State == types.SignerKeyStateRevoked {
			return types.ErrSignerKeyRevoked.Wrap("signer key is revoked")
		}
		if key.State == types.SignerKeyStateExpired {
			return types.ErrSignerKeyExpired.Wrap("signer key is expired")
		}
		return types.ErrInvalidSignerKey.Wrapf("signer key cannot verify in state: %s", key.State)
	}

	blockTime := ctx.BlockTime().UTC()
	if key.ActivatedAt == nil || blockTime.Before(key.ActivatedAt.UTC()) {
		return types.ErrInvalidSignerKey.Wrap("signer key is not active yet")
	}
	if evidence.IssuedAt.Before(key.ActivatedAt.UTC()) {
		return types.ErrInvalidSignerKey.Wrap("web evidence was issued before signer key activation")
	}
	if key.ExpiresAt != nil && !blockTime.Before(key.ExpiresAt.UTC()) {
		return types.ErrSignerKeyExpired.Wrap("signer key is expired")
	}
	if key.RevokedAt != nil && !blockTime.Before(key.RevokedAt.UTC()) {
		return types.ErrSignerKeyRevoked.Wrap("signer key is revoked")
	}
	if err := validateWebEvidenceSignerKeyHeight(ctx, key); err != nil {
		return err
	}
	if err := validateWebEvidenceSignerKeyMetadata(key, evidence); err != nil {
		return err
	}
	return nil
}

func validateWebEvidenceSignerKeyHeight(ctx sdk.Context, key *types.SignerKeyInfo) error {
	if key.Metadata == nil {
		return nil
	}
	if raw := key.Metadata[types.SignerKeyMetadataActivationHeight]; raw != "" {
		height, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || height < 0 {
			return types.ErrInvalidSignerKey.Wrap("invalid signer activation height")
		}
		if ctx.BlockHeight() < height {
			return types.ErrInvalidSignerKey.Wrap("signer key is not active at block height")
		}
	}
	if raw := key.Metadata[types.SignerKeyMetadataExpiryHeight]; raw != "" {
		height, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || height < 0 {
			return types.ErrInvalidSignerKey.Wrap("invalid signer expiry height")
		}
		if ctx.BlockHeight() >= height {
			return types.ErrSignerKeyExpired.Wrap("signer key expired at block height")
		}
	}
	if raw := key.Metadata[types.SignerKeyMetadataRevokedHeight]; raw != "" {
		height, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || height < 0 {
			return types.ErrInvalidSignerKey.Wrap("invalid signer revoked height")
		}
		if ctx.BlockHeight() >= height {
			return types.ErrSignerKeyRevoked.Wrap("signer key revoked at block height")
		}
	}
	return nil
}

func validateWebEvidenceSignerKeyMetadata(key *types.SignerKeyInfo, evidence types.WebEvidenceContext) error {
	if key.Metadata == nil {
		return types.ErrInvalidSignerKey.Wrap("signer key committed evidence type policy is required")
	}
	raw := key.Metadata[types.SignerKeyMetadataEvidenceTypes]
	if raw == "" {
		return types.ErrInvalidSignerKey.Wrap("signer key committed evidence type policy is required")
	}
	allowed := false
	for _, item := range strings.Split(raw, ",") {
		if strings.TrimSpace(item) == string(evidence.EvidenceType) {
			allowed = true
			break
		}
	}
	if !allowed {
		return types.ErrInvalidSignerKey.Wrap("signer key is not authorized for evidence type")
	}
	if expected := key.Metadata[types.SignerKeyMetadataServiceMetadataHash]; expected != "" {
		if evidence.ServiceMetadataHash == "" || evidence.ServiceMetadataHash != expected {
			return types.ErrInvalidSignerKey.Wrap("signer key service metadata mismatch")
		}
	}
	return nil
}

func (k Keeper) verifyWebEvidenceIssuerSignature(
	att *types.VerificationAttestation,
	evidence types.WebEvidenceContext,
	key *types.SignerKeyInfo,
) error {
	if att.Proof.Type != types.ProofTypeEd25519 {
		return types.ErrInvalidSignerKey.Wrapf("unsupported attestation proof type: %s", att.Proof.Type)
	}
	if len(key.PublicKey) != ed25519.PublicKeySize {
		return types.ErrInvalidSignerKey.Wrap("invalid issuer public key size")
	}
	signature, err := att.GetProofBytes()
	if err != nil {
		return types.ErrInvalidAttestation.Wrapf("invalid issuer signature encoding: %v", err)
	}
	if len(signature) != ed25519.SignatureSize {
		return types.ErrAttestationSignatureInvalid.Wrap("invalid issuer signature size")
	}
	signBytes, err := evidence.IssuerSignBytes()
	if err != nil {
		return err
	}
	if !ed25519.Verify(key.PublicKey, signBytes, signature) {
		return types.ErrAttestationSignatureInvalid.Wrap("issuer signature invalid")
	}
	return nil
}

func (k Keeper) verifyWebEvidenceAccountAuthorization(
	ctx sdk.Context,
	evidence types.WebEvidenceContext,
	accountSignature []byte,
) error {
	if len(accountSignature) == 0 {
		return types.ErrInvalidUserSignature.Wrap("account authorization signature is required")
	}
	account, err := sdk.AccAddressFromBech32(evidence.AccountAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrap("invalid web evidence account address")
	}
	wallet, found := k.GetWallet(ctx, account)
	if !found {
		return types.ErrWalletNotFound.Wrap("wallet not found for web evidence account")
	}
	if !wallet.IsActive() {
		return types.ErrWalletInactive.Wrap("wallet is not active")
	}
	if wallet.AccountAddress != evidence.AccountAddress {
		return types.ErrInvalidWallet.Wrap("wallet account address mismatch")
	}
	signBytes, err := evidence.AccountAuthorizationBytes()
	if err != nil {
		return err
	}
	return k.verifySignature(wallet.BindingPubKey, signBytes, accountSignature, "web evidence account authorization")
}

func validateWebEvidenceFreshness(ctx sdk.Context, evidence types.WebEvidenceContext) error {
	blockTime := ctx.BlockTime().UTC()
	if evidence.IssuedAt.After(blockTime.Add(maxWebEvidenceClockSkew)) {
		return types.ErrInvalidTimestamp.Wrap("web evidence issued_at is after block time")
	}
	if !blockTime.Before(evidence.ExpiresAt) {
		return types.ErrAttestationExpired.Wrap("web evidence is expired")
	}
	if blockTime.Sub(evidence.IssuedAt) > maxWebEvidenceAge {
		return types.ErrAttestationExpired.Wrap("web evidence is stale")
	}
	return nil
}

func (k Keeper) checkWebEvidenceReplay(ctx sdk.Context, evidence types.WebEvidenceContext, metadata map[string]string, key *types.SignerKeyInfo) (*webEvidenceValidation, error) {
	fullContextDigest, err := webEvidenceFullContextDigest(evidence)
	if err != nil {
		return nil, err
	}
	globalNonceDigest, err := webEvidenceGlobalNonceDigest(evidence, key)
	if err != nil {
		return nil, err
	}
	metadataDigest, err := webEvidenceMetadataDigest(metadata)
	if err != nil {
		return nil, err
	}
	store := ctx.KVStore(k.skey)
	if existingMetadataDigest := string(store.Get(webEvidenceReplayStoreKey(fullContextDigest))); existingMetadataDigest != "" {
		existingFullContextDigest := string(store.Get(webEvidenceReplayNonceStoreKey(globalNonceDigest)))
		if existingFullContextDigest != fullContextDigest || existingMetadataDigest != metadataDigest {
			return nil, types.ErrNonceAlreadyUsed.Wrap("web evidence exact replay digest mismatch")
		}
		return &webEvidenceValidation{
			SignerKey:         key,
			FullContextDigest: fullContextDigest,
			GlobalNonceDigest: globalNonceDigest,
			MetadataDigest:    metadataDigest,
			ExactReplay:       true,
		}, nil
	}
	if existingFullContextDigest := string(store.Get(webEvidenceReplayNonceStoreKey(globalNonceDigest))); existingFullContextDigest != "" {
		return nil, types.ErrNonceAlreadyUsed.Wrap("web evidence nonce replay detected")
	}
	return &webEvidenceValidation{
		SignerKey:         key,
		FullContextDigest: fullContextDigest,
		GlobalNonceDigest: globalNonceDigest,
		MetadataDigest:    metadataDigest,
	}, nil
}

func (k Keeper) recordWebEvidenceReplay(ctx sdk.Context, validation *webEvidenceValidation) error {
	if validation == nil {
		return types.ErrInvalidAttestation.Wrap("web evidence validation result is required")
	}
	if validation.ExactReplay {
		return nil
	}
	store := ctx.KVStore(k.skey)
	if store.Has(webEvidenceReplayStoreKey(validation.FullContextDigest)) || store.Has(webEvidenceReplayNonceStoreKey(validation.GlobalNonceDigest)) {
		return types.ErrNonceAlreadyUsed.Wrap("web evidence replay detected")
	}
	store.Set(webEvidenceReplayStoreKey(validation.FullContextDigest), []byte(validation.MetadataDigest))
	store.Set(webEvidenceReplayNonceStoreKey(validation.GlobalNonceDigest), []byte(validation.FullContextDigest))
	return nil
}

func webEvidenceFullContextDigest(evidence types.WebEvidenceContext) (string, error) {
	bz, err := evidence.IssuerSignBytes()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(append([]byte("VEID_WEB_EVIDENCE_FULL_CONTEXT_V1:"), bz...))
	return hex.EncodeToString(sum[:]), nil
}

func webEvidenceGlobalNonceDigest(evidence types.WebEvidenceContext, key *types.SignerKeyInfo) (string, error) {
	if key == nil {
		return "", types.ErrInvalidSignerKey.Wrap("signer key is required for web evidence replay")
	}
	payload := struct {
		Domain            string `json:"domain"`
		Version           string `json:"version"`
		IssuerID          string `json:"issuer_id"`
		IssuerKeyID       string `json:"issuer_key_id"`
		IssuerFingerprint string `json:"issuer_fingerprint"`
		KeySequence       uint64 `json:"key_sequence"`
		Nonce             string `json:"nonce"`
	}{
		Domain:            "VEID_WEB_EVIDENCE_GLOBAL_NONCE_V1",
		Version:           types.WebEvidenceVersion,
		IssuerID:          evidence.IssuerID,
		IssuerKeyID:       evidence.IssuerKeyID,
		IssuerFingerprint: evidence.IssuerFingerprint,
		KeySequence:       key.SequenceNumber,
		Nonce:             evidence.Nonce,
	}
	bz, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(bz)
	return hex.EncodeToString(sum[:]), nil
}

func webEvidenceMetadataDigest(metadata map[string]string) (string, error) {
	fields := make([]types.WebEvidenceField, 0, len(metadata))
	for key, value := range metadata {
		if key == "" {
			return "", types.ErrInvalidAttestation.Wrap("web evidence metadata key is required")
		}
		fields = append(fields, types.WebEvidenceField{Name: key, Value: value})
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})
	env := struct {
		Domain  string                   `json:"domain"`
		Version string                   `json:"version"`
		Fields  []types.WebEvidenceField `json:"fields"`
	}{
		Domain:  "VEID_WEB_EVIDENCE_ATTESTATION_METADATA_V1",
		Version: types.WebEvidenceVersion,
		Fields:  fields,
	}
	bz, err := json.Marshal(env)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(bz)
	return hex.EncodeToString(sum[:]), nil
}

func webEvidenceReplayStoreKey(digest string) []byte {
	key := make([]byte, 0, len(types.PrefixWebEvidenceReplay)+len(digest))
	key = append(key, types.PrefixWebEvidenceReplay...)
	key = append(key, []byte(digest)...)
	return key
}

func webEvidenceReplayNonceStoreKey(digest string) []byte {
	key := make([]byte, 0, len(types.PrefixWebEvidenceReplayNonce)+len(digest))
	key = append(key, types.PrefixWebEvidenceReplayNonce...)
	key = append(key, []byte(digest)...)
	return key
}

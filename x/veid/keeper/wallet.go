package keeper

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Identity Wallet Keeper Methods
// ============================================================================

// walletStore is the stored format of an identity wallet
type walletStore struct {
	WalletID            string                           `json:"wallet_id"`
	AccountAddress      string                           `json:"account_address"`
	CreatedAt           int64                            `json:"created_at"`
	UpdatedAt           int64                            `json:"updated_at"`
	Status              types.WalletStatus               `json:"status"`
	ScopeRefs           []types.ScopeReference           `json:"scope_refs"`
	DerivedFeatures     derivedFeaturesStore             `json:"derived_features"`
	CurrentScore        uint32                           `json:"current_score"`
	ScoreStatus         types.AccountStatus              `json:"score_status"`
	VerificationHistory []verificationHistoryEntryStore  `json:"verification_history"`
	ConsentSettings     consentSettingsStore             `json:"consent_settings"`
	BindingSignature    []byte                           `json:"binding_signature"`
	BindingPubKey       []byte                           `json:"binding_pub_key"`
	LastBindingAt       int64                            `json:"last_binding_at"`
	Tier                types.IdentityTier               `json:"tier"`
	Metadata            map[string]string                `json:"metadata,omitempty"`
}

// derivedFeaturesStore is the stored format of derived features
type derivedFeaturesStore struct {
	FaceEmbeddingHash []byte            `json:"face_embedding_hash,omitempty"`
	DocFieldHashes    map[string][]byte `json:"doc_field_hashes,omitempty"`
	BiometricHash     []byte            `json:"biometric_hash,omitempty"`
	LivenessProofHash []byte            `json:"liveness_proof_hash,omitempty"`
	LastComputedAt    int64             `json:"last_computed_at"`
	ModelVersion      string            `json:"model_version"`
	ComputedBy        string            `json:"computed_by,omitempty"`
	BlockHeight       int64             `json:"block_height"`
	FeatureVersion    uint32            `json:"feature_version"`
}

// verificationHistoryEntryStore is the stored format of a verification history entry
type verificationHistoryEntryStore struct {
	EntryID          string              `json:"entry_id"`
	Timestamp        int64               `json:"timestamp"`
	BlockHeight      int64               `json:"block_height"`
	PreviousScore    uint32              `json:"previous_score"`
	NewScore         uint32              `json:"new_score"`
	PreviousStatus   types.AccountStatus `json:"previous_status"`
	NewStatus        types.AccountStatus `json:"new_status"`
	ScopesEvaluated  []string            `json:"scopes_evaluated,omitempty"`
	ModelVersion     string              `json:"model_version"`
	ValidatorAddress string              `json:"validator_address,omitempty"`
	Reason           string              `json:"reason,omitempty"`
}

// consentSettingsStore is the stored format of consent settings
type consentSettingsStore struct {
	ScopeConsents              map[string]scopeConsentStore `json:"scope_consents"`
	ShareWithProviders         bool                         `json:"share_with_providers"`
	ShareForVerification       bool                         `json:"share_for_verification"`
	AllowReVerification        bool                         `json:"allow_re_verification"`
	AllowDerivedFeatureSharing bool                         `json:"allow_derived_feature_sharing"`
	GlobalExpiresAt            *int64                       `json:"global_expires_at,omitempty"`
	LastUpdatedAt              int64                        `json:"last_updated_at"`
	ConsentVersion             uint32                       `json:"consent_version"`
}

// scopeConsentStore is the stored format of a scope consent
type scopeConsentStore struct {
	ScopeID            string   `json:"scope_id"`
	Granted            bool     `json:"granted"`
	GrantedAt          *int64   `json:"granted_at,omitempty"`
	RevokedAt          *int64   `json:"revoked_at,omitempty"`
	ExpiresAt          *int64   `json:"expires_at,omitempty"`
	Purpose            string   `json:"purpose,omitempty"`
	GrantedToProviders []string `json:"granted_to_providers,omitempty"`
	Restrictions       []string `json:"restrictions,omitempty"`
}

// walletToStore converts a wallet to its stored format
func walletToStore(w *types.IdentityWallet) *walletStore {
	ws := &walletStore{
		WalletID:         w.WalletID,
		AccountAddress:   w.AccountAddress,
		CreatedAt:        w.CreatedAt.Unix(),
		UpdatedAt:        w.UpdatedAt.Unix(),
		Status:           w.Status,
		ScopeRefs:        w.ScopeRefs,
		CurrentScore:     w.CurrentScore,
		ScoreStatus:      w.ScoreStatus,
		BindingSignature: w.BindingSignature,
		BindingPubKey:    w.BindingPubKey,
		LastBindingAt:    w.LastBindingAt.Unix(),
		Tier:             w.Tier,
		Metadata:         w.Metadata,
	}

	// Convert derived features
	ws.DerivedFeatures = derivedFeaturesStore{
		FaceEmbeddingHash: w.DerivedFeatures.FaceEmbeddingHash,
		DocFieldHashes:    w.DerivedFeatures.DocFieldHashes,
		BiometricHash:     w.DerivedFeatures.BiometricHash,
		LivenessProofHash: w.DerivedFeatures.LivenessProofHash,
		LastComputedAt:    w.DerivedFeatures.LastComputedAt.Unix(),
		ModelVersion:      w.DerivedFeatures.ModelVersion,
		ComputedBy:        w.DerivedFeatures.ComputedBy,
		BlockHeight:       w.DerivedFeatures.BlockHeight,
		FeatureVersion:    w.DerivedFeatures.FeatureVersion,
	}

	// Convert verification history
	ws.VerificationHistory = make([]verificationHistoryEntryStore, len(w.VerificationHistory))
	for i, entry := range w.VerificationHistory {
		ws.VerificationHistory[i] = verificationHistoryEntryStore{
			EntryID:          entry.EntryID,
			Timestamp:        entry.Timestamp.Unix(),
			BlockHeight:      entry.BlockHeight,
			PreviousScore:    entry.PreviousScore,
			NewScore:         entry.NewScore,
			PreviousStatus:   entry.PreviousStatus,
			NewStatus:        entry.NewStatus,
			ScopesEvaluated:  entry.ScopesEvaluated,
			ModelVersion:     entry.ModelVersion,
			ValidatorAddress: entry.ValidatorAddress,
			Reason:           entry.Reason,
		}
	}

	// Convert consent settings
	ws.ConsentSettings = consentSettingsStore{
		ScopeConsents:              make(map[string]scopeConsentStore),
		ShareWithProviders:         w.ConsentSettings.ShareWithProviders,
		ShareForVerification:       w.ConsentSettings.ShareForVerification,
		AllowReVerification:        w.ConsentSettings.AllowReVerification,
		AllowDerivedFeatureSharing: w.ConsentSettings.AllowDerivedFeatureSharing,
		LastUpdatedAt:              w.ConsentSettings.LastUpdatedAt.Unix(),
		ConsentVersion:             w.ConsentSettings.ConsentVersion,
	}
	if w.ConsentSettings.GlobalExpiresAt != nil {
		ts := w.ConsentSettings.GlobalExpiresAt.Unix()
		ws.ConsentSettings.GlobalExpiresAt = &ts
	}
	for scopeID, consent := range w.ConsentSettings.ScopeConsents {
		sc := scopeConsentStore{
			ScopeID:            consent.ScopeID,
			Granted:            consent.Granted,
			Purpose:            consent.Purpose,
			GrantedToProviders: consent.GrantedToProviders,
			Restrictions:       consent.Restrictions,
		}
		if consent.GrantedAt != nil {
			ts := consent.GrantedAt.Unix()
			sc.GrantedAt = &ts
		}
		if consent.RevokedAt != nil {
			ts := consent.RevokedAt.Unix()
			sc.RevokedAt = &ts
		}
		if consent.ExpiresAt != nil {
			ts := consent.ExpiresAt.Unix()
			sc.ExpiresAt = &ts
		}
		ws.ConsentSettings.ScopeConsents[scopeID] = sc
	}

	return ws
}

// storeToWallet converts stored format to a wallet
func storeToWallet(ws *walletStore) *types.IdentityWallet {
	w := &types.IdentityWallet{
		WalletID:         ws.WalletID,
		AccountAddress:   ws.AccountAddress,
		CreatedAt:        time.Unix(ws.CreatedAt, 0),
		UpdatedAt:        time.Unix(ws.UpdatedAt, 0),
		Status:           ws.Status,
		ScopeRefs:        ws.ScopeRefs,
		CurrentScore:     ws.CurrentScore,
		ScoreStatus:      ws.ScoreStatus,
		BindingSignature: ws.BindingSignature,
		BindingPubKey:    ws.BindingPubKey,
		LastBindingAt:    time.Unix(ws.LastBindingAt, 0),
		Tier:             ws.Tier,
		Metadata:         ws.Metadata,
	}

	// Convert derived features
	w.DerivedFeatures = types.DerivedFeatures{
		FaceEmbeddingHash: ws.DerivedFeatures.FaceEmbeddingHash,
		DocFieldHashes:    ws.DerivedFeatures.DocFieldHashes,
		BiometricHash:     ws.DerivedFeatures.BiometricHash,
		LivenessProofHash: ws.DerivedFeatures.LivenessProofHash,
		LastComputedAt:    time.Unix(ws.DerivedFeatures.LastComputedAt, 0),
		ModelVersion:      ws.DerivedFeatures.ModelVersion,
		ComputedBy:        ws.DerivedFeatures.ComputedBy,
		BlockHeight:       ws.DerivedFeatures.BlockHeight,
		FeatureVersion:    ws.DerivedFeatures.FeatureVersion,
	}
	if w.DerivedFeatures.DocFieldHashes == nil {
		w.DerivedFeatures.DocFieldHashes = make(map[string][]byte)
	}

	// Convert verification history
	w.VerificationHistory = make([]types.VerificationHistoryEntry, len(ws.VerificationHistory))
	for i, entry := range ws.VerificationHistory {
		w.VerificationHistory[i] = types.VerificationHistoryEntry{
			EntryID:          entry.EntryID,
			Timestamp:        time.Unix(entry.Timestamp, 0),
			BlockHeight:      entry.BlockHeight,
			PreviousScore:    entry.PreviousScore,
			NewScore:         entry.NewScore,
			PreviousStatus:   entry.PreviousStatus,
			NewStatus:        entry.NewStatus,
			ScopesEvaluated:  entry.ScopesEvaluated,
			ModelVersion:     entry.ModelVersion,
			ValidatorAddress: entry.ValidatorAddress,
			Reason:           entry.Reason,
		}
	}

	// Convert consent settings
	w.ConsentSettings = types.ConsentSettings{
		ScopeConsents:              make(map[string]types.ScopeConsent),
		ShareWithProviders:         ws.ConsentSettings.ShareWithProviders,
		ShareForVerification:       ws.ConsentSettings.ShareForVerification,
		AllowReVerification:        ws.ConsentSettings.AllowReVerification,
		AllowDerivedFeatureSharing: ws.ConsentSettings.AllowDerivedFeatureSharing,
		LastUpdatedAt:              time.Unix(ws.ConsentSettings.LastUpdatedAt, 0),
		ConsentVersion:             ws.ConsentSettings.ConsentVersion,
	}
	if ws.ConsentSettings.GlobalExpiresAt != nil {
		t := time.Unix(*ws.ConsentSettings.GlobalExpiresAt, 0)
		w.ConsentSettings.GlobalExpiresAt = &t
	}
	for scopeID, sc := range ws.ConsentSettings.ScopeConsents {
		consent := types.ScopeConsent{
			ScopeID:            sc.ScopeID,
			Granted:            sc.Granted,
			Purpose:            sc.Purpose,
			GrantedToProviders: sc.GrantedToProviders,
			Restrictions:       sc.Restrictions,
		}
		if sc.GrantedAt != nil {
			t := time.Unix(*sc.GrantedAt, 0)
			consent.GrantedAt = &t
		}
		if sc.RevokedAt != nil {
			t := time.Unix(*sc.RevokedAt, 0)
			consent.RevokedAt = &t
		}
		if sc.ExpiresAt != nil {
			t := time.Unix(*sc.ExpiresAt, 0)
			consent.ExpiresAt = &t
		}
		w.ConsentSettings.ScopeConsents[scopeID] = consent
	}

	return w
}

// GenerateWalletID generates a unique wallet ID from an account address
func GenerateWalletID(accountAddress string) string {
	hash := sha256.Sum256([]byte("VEID_WALLET:" + accountAddress))
	return fmt.Sprintf("wallet_%x", hash[:16])
}

// CreateWallet creates a new identity wallet for an account
func (k Keeper) CreateWallet(ctx sdk.Context, accountAddr sdk.AccAddress, bindingSignature, bindingPubKey []byte) (*types.IdentityWallet, error) {
	// Check if wallet already exists
	if _, found := k.GetWallet(ctx, accountAddr); found {
		return nil, types.ErrWalletAlreadyExists.Wrap("wallet already exists for this account")
	}

	// Generate wallet ID
	walletID := GenerateWalletID(accountAddr.String())

	// Verify the binding signature
	if err := k.VerifyWalletBindingSignature(walletID, accountAddr.String(), bindingPubKey, bindingSignature); err != nil {
		return nil, err
	}

	// Create the wallet
	wallet := types.NewIdentityWallet(
		walletID,
		accountAddr.String(),
		ctx.BlockTime(),
		bindingSignature,
		bindingPubKey,
	)

	// Store the wallet
	if err := k.SetWallet(ctx, wallet); err != nil {
		return nil, err
	}

	// Store wallet ID -> address mapping
	k.setWalletIDMapping(ctx, walletID, accountAddr)

	k.Logger(ctx).Info("Created identity wallet",
		"wallet_id", walletID,
		"account", accountAddr.String(),
	)

	return wallet, nil
}

// GetWallet returns an identity wallet by account address
func (k Keeper) GetWallet(ctx sdk.Context, address sdk.AccAddress) (*types.IdentityWallet, bool) {
	store := ctx.KVStore(k.skey)
	key := types.IdentityWalletKey(address.Bytes())
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var ws walletStore
	if err := json.Unmarshal(bz, &ws); err != nil {
		return nil, false
	}

	return storeToWallet(&ws), true
}

// GetWalletByID returns an identity wallet by wallet ID
func (k Keeper) GetWalletByID(ctx sdk.Context, walletID string) (*types.IdentityWallet, bool) {
	address := k.getWalletIDMapping(ctx, walletID)
	if address == nil {
		return nil, false
	}
	return k.GetWallet(ctx, address)
}

// SetWallet stores an identity wallet
func (k Keeper) SetWallet(ctx sdk.Context, wallet *types.IdentityWallet) error {
	if err := wallet.Validate(); err != nil {
		return err
	}

	address, err := sdk.AccAddressFromBech32(wallet.AccountAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	ws := walletToStore(wallet)
	bz, err := json.Marshal(ws)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.IdentityWalletKey(address.Bytes()), bz)
	return nil
}

// setWalletIDMapping stores the wallet ID -> address mapping
func (k Keeper) setWalletIDMapping(ctx sdk.Context, walletID string, address sdk.AccAddress) {
	store := ctx.KVStore(k.skey)
	store.Set(types.WalletByIDKey(walletID), address.Bytes())
}

// getWalletIDMapping returns the address for a wallet ID
func (k Keeper) getWalletIDMapping(ctx sdk.Context, walletID string) sdk.AccAddress {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.WalletByIDKey(walletID))
	if bz == nil {
		return nil
	}
	return sdk.AccAddress(bz)
}

// AddScopeToWallet adds a scope reference to a wallet
func (k Keeper) AddScopeToWallet(
	ctx sdk.Context,
	accountAddr sdk.AccAddress,
	scopeRef types.ScopeReference,
	userSignature []byte,
) error {
	// Get the wallet
	wallet, found := k.GetWallet(ctx, accountAddr)
	if !found {
		return types.ErrWalletNotFound.Wrap("wallet not found for account")
	}

	// Check wallet is active
	if !wallet.IsActive() {
		return types.ErrWalletInactive.Wrap("wallet is not active")
	}

	// Verify user signature
	if err := k.VerifyAddScopeSignature(accountAddr.String(), scopeRef.ScopeID, wallet.BindingPubKey, userSignature); err != nil {
		return err
	}

	// Check if scope already exists
	if _, exists := wallet.GetScopeReference(scopeRef.ScopeID); exists {
		return types.ErrScopeAlreadyInWallet.Wrapf("scope %s already in wallet", scopeRef.ScopeID)
	}

	// Add the scope reference
	wallet.AddScopeReference(scopeRef)
	wallet.UpdatedAt = ctx.BlockTime()

	// Save the wallet
	if err := k.SetWallet(ctx, wallet); err != nil {
		return err
	}

	k.Logger(ctx).Info("Added scope to wallet",
		"wallet_id", wallet.WalletID,
		"scope_id", scopeRef.ScopeID,
		"scope_type", scopeRef.ScopeType,
	)

	return nil
}

// RevokeScopeFromWallet revokes a scope from a wallet
func (k Keeper) RevokeScopeFromWallet(
	ctx sdk.Context,
	accountAddr sdk.AccAddress,
	scopeID string,
	reason string,
	userSignature []byte,
) error {
	// Get the wallet
	wallet, found := k.GetWallet(ctx, accountAddr)
	if !found {
		return types.ErrWalletNotFound.Wrap("wallet not found for account")
	}

	// Verify user signature
	if err := k.VerifyRevokeScopeSignature(accountAddr.String(), scopeID, wallet.BindingPubKey, userSignature); err != nil {
		return err
	}

	// Check if scope exists
	if _, exists := wallet.GetScopeReference(scopeID); !exists {
		return types.ErrScopeNotInWallet.Wrapf("scope %s not in wallet", scopeID)
	}

	// Revoke the scope
	if !wallet.RevokeScopeReference(scopeID, reason, ctx.BlockTime()) {
		return types.ErrScopeNotInWallet.Wrapf("failed to revoke scope %s", scopeID)
	}

	// Also revoke consent for this scope
	wallet.ConsentSettings.RevokeScopeConsent(scopeID)

	// Save the wallet
	if err := k.SetWallet(ctx, wallet); err != nil {
		return err
	}

	k.Logger(ctx).Info("Revoked scope from wallet",
		"wallet_id", wallet.WalletID,
		"scope_id", scopeID,
		"reason", reason,
	)

	return nil
}

// UpdateConsent updates consent settings for a wallet
func (k Keeper) UpdateConsent(
	ctx sdk.Context,
	accountAddr sdk.AccAddress,
	update types.ConsentUpdateRequest,
	userSignature []byte,
) error {
	// Get the wallet
	wallet, found := k.GetWallet(ctx, accountAddr)
	if !found {
		return types.ErrWalletNotFound.Wrap("wallet not found for account")
	}

	// Verify user signature for consent update
	if err := k.VerifyConsentUpdateSignature(accountAddr.String(), update.ScopeID, update.GrantConsent, wallet.BindingPubKey, userSignature); err != nil {
		return err
	}

	// Apply the consent update
	wallet.ConsentSettings.ApplyConsentUpdate(update)
	wallet.UpdatedAt = ctx.BlockTime()

	// Save the wallet
	if err := k.SetWallet(ctx, wallet); err != nil {
		return err
	}

	k.Logger(ctx).Info("Updated consent settings",
		"wallet_id", wallet.WalletID,
		"scope_id", update.ScopeID,
		"grant", update.GrantConsent,
		"version", wallet.ConsentSettings.ConsentVersion,
	)

	return nil
}

// UpdateDerivedFeatures updates derived features for a wallet (validator only)
func (k Keeper) UpdateDerivedFeatures(
	ctx sdk.Context,
	accountAddr sdk.AccAddress,
	update *types.DerivedFeaturesUpdate,
) error {
	// Validate the update
	if err := update.Validate(); err != nil {
		return err
	}

	// Get the wallet
	wallet, found := k.GetWallet(ctx, accountAddr)
	if !found {
		return types.ErrWalletNotFound.Wrap("wallet not found for account")
	}

	// Apply the update
	update.Apply(&wallet.DerivedFeatures, ctx.BlockHeight(), ctx.BlockTime())
	wallet.UpdatedAt = ctx.BlockTime()

	// Save the wallet
	if err := k.SetWallet(ctx, wallet); err != nil {
		return err
	}

	k.Logger(ctx).Info("Updated derived features",
		"wallet_id", wallet.WalletID,
		"model_version", update.ModelVersion,
		"validator", update.ValidatorAddress,
	)

	return nil
}

// RebindWallet rebinds a wallet with a new key (for key rotation)
func (k Keeper) RebindWallet(
	ctx sdk.Context,
	accountAddr sdk.AccAddress,
	newBindingSignature []byte,
	newBindingPubKey []byte,
	oldSignature []byte,
) error {
	// Get the wallet
	wallet, found := k.GetWallet(ctx, accountAddr)
	if !found {
		return types.ErrWalletNotFound.Wrap("wallet not found for account")
	}

	// Verify old signature (signs the new public key)
	if err := k.VerifyRebindSignature(wallet.BindingPubKey, newBindingPubKey, oldSignature); err != nil {
		return types.ErrInvalidBindingSignature.Wrap("old signature verification failed")
	}

	// Verify new binding signature
	if err := k.VerifyWalletBindingSignature(wallet.WalletID, accountAddr.String(), newBindingPubKey, newBindingSignature); err != nil {
		return err
	}

	// Rebind the wallet
	wallet.Rebind(newBindingSignature, newBindingPubKey, ctx.BlockTime())

	// Save the wallet
	if err := k.SetWallet(ctx, wallet); err != nil {
		return err
	}

	k.Logger(ctx).Info("Rebound wallet",
		"wallet_id", wallet.WalletID,
		"account", accountAddr.String(),
	)

	return nil
}

// GetWalletPublicMetadata returns non-sensitive public metadata about a wallet
func (k Keeper) GetWalletPublicMetadata(ctx sdk.Context, accountAddr sdk.AccAddress) (types.PublicWalletInfo, bool) {
	wallet, found := k.GetWallet(ctx, accountAddr)
	if !found {
		return types.PublicWalletInfo{}, false
	}
	return wallet.ToPublicInfo(), true
}

// UpdateWalletScore updates the wallet's score and adds a verification history entry
func (k Keeper) UpdateWalletScore(
	ctx sdk.Context,
	accountAddr sdk.AccAddress,
	newScore uint32,
	newStatus types.AccountStatus,
	modelVersion string,
	validatorAddress string,
	scopesEvaluated []string,
	reason string,
) error {
	// Get the wallet
	wallet, found := k.GetWallet(ctx, accountAddr)
	if !found {
		return types.ErrWalletNotFound.Wrap("wallet not found for account")
	}

	// Create verification history entry
	entryID := fmt.Sprintf("vhe_%d_%d", ctx.BlockHeight(), ctx.BlockTime().Unix())
	entry := types.VerificationHistoryEntry{
		EntryID:          entryID,
		Timestamp:        ctx.BlockTime(),
		BlockHeight:      ctx.BlockHeight(),
		PreviousScore:    wallet.CurrentScore,
		NewScore:         newScore,
		PreviousStatus:   wallet.ScoreStatus,
		NewStatus:        newStatus,
		ScopesEvaluated:  scopesEvaluated,
		ModelVersion:     modelVersion,
		ValidatorAddress: validatorAddress,
		Reason:           reason,
	}

	// Update score and add history
	wallet.UpdateScore(newScore, newStatus)
	wallet.AddVerificationHistoryEntry(entry)

	// Save the wallet
	if err := k.SetWallet(ctx, wallet); err != nil {
		return err
	}

	k.Logger(ctx).Info("Updated wallet score",
		"wallet_id", wallet.WalletID,
		"previous_score", entry.PreviousScore,
		"new_score", newScore,
		"status", newStatus,
	)

	return nil
}

// ============================================================================
// Signature Verification
// ============================================================================

// VerifyWalletBindingSignature verifies a wallet binding signature
func (k Keeper) VerifyWalletBindingSignature(walletID, accountAddress string, pubKey, signature []byte) error {
	msg := types.GetWalletBindingMessage(walletID, accountAddress)
	return k.verifySignature(pubKey, msg, signature, "wallet binding")
}

// VerifyAddScopeSignature verifies a signature for adding a scope
func (k Keeper) VerifyAddScopeSignature(sender, scopeID string, pubKey, signature []byte) error {
	msg := types.GetAddScopeSigningMessage(sender, scopeID)
	return k.verifySignature(pubKey, msg, signature, "add scope")
}

// VerifyRevokeScopeSignature verifies a signature for revoking a scope
func (k Keeper) VerifyRevokeScopeSignature(sender, scopeID string, pubKey, signature []byte) error {
	msg := types.GetRevokeScopeSigningMessage(sender, scopeID)
	return k.verifySignature(pubKey, msg, signature, "revoke scope")
}

// VerifyConsentUpdateSignature verifies a signature for consent update
func (k Keeper) VerifyConsentUpdateSignature(sender, scopeID string, grant bool, pubKey, signature []byte) error {
	grantStr := "revoke"
	if grant {
		grantStr = "grant"
	}
	msg := []byte("VEID_CONSENT_UPDATE:" + sender + ":" + scopeID + ":" + grantStr)
	return k.verifySignature(pubKey, msg, signature, "consent update")
}

// VerifyRebindSignature verifies the old signature during rebind
func (k Keeper) VerifyRebindSignature(oldPubKey, newPubKey, signature []byte) error {
	// The old key signs the new public key to authorize the rebind
	return k.verifySignature(oldPubKey, newPubKey, signature, "rebind authorization")
}

// verifySignature is a helper to verify ed25519 signatures
func (k Keeper) verifySignature(pubKey, message, signature []byte, context string) error {
	if len(pubKey) != ed25519.PublicKeySize {
		return types.ErrInvalidUserSignature.Wrapf("invalid public key size for %s", context)
	}

	if len(signature) != ed25519.SignatureSize {
		return types.ErrInvalidUserSignature.Wrapf("invalid signature size for %s", context)
	}

	// Hash the message if it's not already 32 bytes
	var msgToVerify []byte
	if len(message) == 32 {
		msgToVerify = message
	} else {
		hash := sha256.Sum256(message)
		msgToVerify = hash[:]
	}

	if !ed25519.Verify(pubKey, msgToVerify, signature) {
		return types.ErrInvalidUserSignature.Wrapf("%s signature verification failed", context)
	}

	return nil
}

// ============================================================================
// Wallet Iterators
// ============================================================================

// WithWallets iterates over all identity wallets
func (k Keeper) WithWallets(ctx sdk.Context, fn func(wallet *types.IdentityWallet) bool) {
	store := ctx.KVStore(k.skey)
	iter := store.Iterator(types.PrefixIdentityWallet, storetypes.PrefixEndBytes(types.PrefixIdentityWallet))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var ws walletStore
		if err := json.Unmarshal(iter.Value(), &ws); err != nil {
			continue
		}

		wallet := storeToWallet(&ws)
		if fn(wallet) {
			break
		}
	}
}

// CountWallets returns the total number of wallets
func (k Keeper) CountWallets(ctx sdk.Context) int {
	count := 0
	k.WithWallets(ctx, func(_ *types.IdentityWallet) bool {
		count++
		return false
	})
	return count
}

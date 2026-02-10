package keeper

import (
	"encoding/json"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/config/types"
)

// IKeeper defines the interface for the config keeper
type IKeeper interface {
	// Client management
	RegisterClient(ctx sdk.Context, client types.ApprovedClient) error
	UpdateClient(ctx sdk.Context, clientID string, name, description, minVersion, maxVersion string, allowedScopes []string, updatedBy string) error
	SuspendClient(ctx sdk.Context, clientID string, reason string, suspendedBy string) error
	RevokeClient(ctx sdk.Context, clientID string, reason string, revokedBy string) error
	ReactivateClient(ctx sdk.Context, clientID string, reason string, reactivatedBy string) error
	GetClient(ctx sdk.Context, clientID string) (types.ApprovedClient, bool)
	ListClients(ctx sdk.Context) []types.ApprovedClient
	ListClientsByStatus(ctx sdk.Context, status types.ClientStatus) []types.ApprovedClient

	// Validation
	IsClientApproved(ctx sdk.Context, clientID string) bool
	ValidateClientSignature(ctx sdk.Context, clientID string, signature []byte, payloadHash []byte) error
	ValidateClientVersion(ctx sdk.Context, clientID string, version string) error
	ValidateScopePermission(ctx sdk.Context, clientID string, scopeType string) error

	// Signature verification helper
	VerifyUploadSignatures(
		ctx sdk.Context,
		clientID string,
		clientVersion string,
		clientSignature []byte,
		userSignature []byte,
		payloadHash []byte,
		salt []byte,
		userAddress sdk.AccAddress,
	) error

	// Authorization
	IsAdmin(ctx sdk.Context, address sdk.AccAddress) bool

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
}

// Keeper of the config store
type Keeper struct {
	skey storetypes.StoreKey
	cdc  codec.BinaryCodec

	// The address capable of executing a MsgUpdateParams message.
	// This should be the x/gov module account.
	authority string
}

// NewKeeper creates and returns an instance for config keeper
func NewKeeper(cdc codec.BinaryCodec, skey storetypes.StoreKey, authority string) Keeper {
	return Keeper{
		cdc:       cdc,
		skey:      skey,
		authority: authority,
	}
}

// Codec returns keeper codec
func (k Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}

// StoreKey returns store key
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.skey
}

// GetAuthority returns the module's authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// ============================================================================
// Parameters
// ============================================================================

// paramsStore is the stored format of params
type paramsStore struct {
	RequireClientSignature  bool     `json:"require_client_signature"`
	RequireUserSignature    bool     `json:"require_user_signature"`
	RequireSaltBinding      bool     `json:"require_salt_binding"`
	MaxClientsPerRegistrar  uint32   `json:"max_clients_per_registrar"`
	AllowGovernanceOverride bool     `json:"allow_governance_override"`
	DefaultMinVersion       string   `json:"default_min_version"`
	AdminAddresses          []string `json:"admin_addresses"`
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&paramsStore{
		RequireClientSignature:  params.RequireClientSignature,
		RequireUserSignature:    params.RequireUserSignature,
		RequireSaltBinding:      params.RequireSaltBinding,
		MaxClientsPerRegistrar:  params.MaxClientsPerRegistrar,
		AllowGovernanceOverride: params.AllowGovernanceOverride,
		DefaultMinVersion:       params.DefaultMinVersion,
		AdminAddresses:          params.AdminAddresses,
	})
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey(), bz)
	return nil
}

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ParamsKey())
	if bz == nil {
		return types.DefaultParams()
	}

	var ps paramsStore
	if err := json.Unmarshal(bz, &ps); err != nil {
		return types.DefaultParams()
	}

	return types.Params{
		RequireClientSignature:  ps.RequireClientSignature,
		RequireUserSignature:    ps.RequireUserSignature,
		RequireSaltBinding:      ps.RequireSaltBinding,
		MaxClientsPerRegistrar:  ps.MaxClientsPerRegistrar,
		AllowGovernanceOverride: ps.AllowGovernanceOverride,
		DefaultMinVersion:       ps.DefaultMinVersion,
		AdminAddresses:          ps.AdminAddresses,
	}
}

// ============================================================================
// Authorization
// ============================================================================

// IsAdmin checks if an address is an admin
func (k Keeper) IsAdmin(ctx sdk.Context, address sdk.AccAddress) bool {
	// Authority (governance) is always admin
	if address.String() == k.authority {
		return true
	}

	params := k.GetParams(ctx)
	addrStr := address.String()
	for _, admin := range params.AdminAddresses {
		if admin == addrStr {
			return true
		}
	}

	return false
}

// ============================================================================
// Approved Clients
// ============================================================================

// approvedClientStore is the stored format of an approved client
type approvedClientStore struct {
	ClientID      string            `json:"client_id"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	PublicKey     []byte            `json:"public_key"`
	KeyType       string            `json:"key_type"`
	MinVersion    string            `json:"min_version"`
	MaxVersion    string            `json:"max_version,omitempty"`
	AllowedScopes []string          `json:"allowed_scopes"`
	Status        string            `json:"status"`
	StatusReason  string            `json:"status_reason,omitempty"`
	RegisteredBy  string            `json:"registered_by"`
	RegisteredAt  int64             `json:"registered_at"`
	LastUpdatedAt int64             `json:"last_updated_at"`
	LastUpdatedBy string            `json:"last_updated_by,omitempty"`
	SuspendedAt   *int64            `json:"suspended_at,omitempty"`
	RevokedAt     *int64            `json:"revoked_at,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// RegisterClient registers a new approved client
func (k Keeper) RegisterClient(ctx sdk.Context, client types.ApprovedClient) error {
	if err := client.Validate(); err != nil {
		return err
	}

	// Check if client already exists
	if _, found := k.GetClient(ctx, client.ClientID); found {
		return types.ErrClientAlreadyExists.Wrapf("client %s already exists", client.ClientID)
	}

	// Store the client
	return k.setClient(ctx, &client)
}

// UpdateClient updates an existing approved client
func (k Keeper) UpdateClient(
	ctx sdk.Context,
	clientID string,
	name string,
	description string,
	minVersion string,
	maxVersion string,
	allowedScopes []string,
	updatedBy string,
) error {
	client, found := k.GetClient(ctx, clientID)
	if !found {
		return types.ErrClientNotFound.Wrapf("client %s not found", clientID)
	}

	// Update mutable fields
	if err := client.Update(name, description, minVersion, maxVersion, allowedScopes, updatedBy, ctx.BlockTime()); err != nil {
		return err
	}

	// Store audit entry
	k.storeAuditEntry(ctx, types.NewAuditEntry(
		clientID,
		"update",
		updatedBy,
		ctx.BlockTime(),
		"client updated",
	))

	return k.setClient(ctx, &client)
}

// SuspendClient suspends an approved client
func (k Keeper) SuspendClient(ctx sdk.Context, clientID string, reason string, suspendedBy string) error {
	client, found := k.GetClient(ctx, clientID)
	if !found {
		return types.ErrClientNotFound.Wrapf("client %s not found", clientID)
	}

	if err := client.Suspend(reason, suspendedBy, ctx.BlockTime()); err != nil {
		return err
	}

	// Update status index
	k.removeClientFromStatusIndex(ctx, clientID, types.ClientStatusActive)
	k.addClientToStatusIndex(ctx, clientID, types.ClientStatusSuspended)

	// Store audit entry
	k.storeAuditEntry(ctx, types.NewAuditEntry(
		clientID,
		"suspend",
		suspendedBy,
		ctx.BlockTime(),
		reason,
	))

	return k.setClient(ctx, &client)
}

// RevokeClient revokes an approved client
func (k Keeper) RevokeClient(ctx sdk.Context, clientID string, reason string, revokedBy string) error {
	client, found := k.GetClient(ctx, clientID)
	if !found {
		return types.ErrClientNotFound.Wrapf("client %s not found", clientID)
	}

	previousStatus := client.Status

	if err := client.Revoke(reason, revokedBy, ctx.BlockTime()); err != nil {
		return err
	}

	// Update status index
	k.removeClientFromStatusIndex(ctx, clientID, previousStatus)
	k.addClientToStatusIndex(ctx, clientID, types.ClientStatusRevoked)

	// Store audit entry
	k.storeAuditEntry(ctx, types.NewAuditEntry(
		clientID,
		"revoke",
		revokedBy,
		ctx.BlockTime(),
		reason,
	))

	return k.setClient(ctx, &client)
}

// ReactivateClient reactivates a suspended client
func (k Keeper) ReactivateClient(ctx sdk.Context, clientID string, reason string, reactivatedBy string) error {
	client, found := k.GetClient(ctx, clientID)
	if !found {
		return types.ErrClientNotFound.Wrapf("client %s not found", clientID)
	}

	if err := client.Reactivate(reason, reactivatedBy, ctx.BlockTime()); err != nil {
		return err
	}

	// Update status index
	k.removeClientFromStatusIndex(ctx, clientID, types.ClientStatusSuspended)
	k.addClientToStatusIndex(ctx, clientID, types.ClientStatusActive)

	// Store audit entry
	k.storeAuditEntry(ctx, types.NewAuditEntry(
		clientID,
		"reactivate",
		reactivatedBy,
		ctx.BlockTime(),
		reason,
	))

	return k.setClient(ctx, &client)
}

// GetClient returns an approved client by ID
func (k Keeper) GetClient(ctx sdk.Context, clientID string) (types.ApprovedClient, bool) {
	store := ctx.KVStore(k.skey)
	key := types.ApprovedClientKey(clientID)
	bz := store.Get(key)
	if bz == nil {
		return types.ApprovedClient{}, false
	}

	var cs approvedClientStore
	if err := json.Unmarshal(bz, &cs); err != nil {
		return types.ApprovedClient{}, false
	}

	return k.clientStoreToClient(cs), true
}

// ListClients returns all approved clients
func (k Keeper) ListClients(ctx sdk.Context) []types.ApprovedClient {
	var clients []types.ApprovedClient

	k.iterateClients(ctx, func(client types.ApprovedClient) bool {
		clients = append(clients, client)
		return false
	})

	return clients
}

// ListClientsByStatus returns all clients with a given status
func (k Keeper) ListClientsByStatus(ctx sdk.Context, status types.ClientStatus) []types.ApprovedClient {
	var clients []types.ApprovedClient

	k.iterateClients(ctx, func(client types.ApprovedClient) bool {
		if client.Status == status {
			clients = append(clients, client)
		}
		return false
	})

	return clients
}

// IsClientApproved checks if a client is approved and active
func (k Keeper) IsClientApproved(ctx sdk.Context, clientID string) bool {
	client, found := k.GetClient(ctx, clientID)
	if !found {
		return false
	}
	return client.IsActive()
}

// ValidateScopePermission checks if a client can submit a specific scope type
func (k Keeper) ValidateScopePermission(ctx sdk.Context, clientID string, scopeType string) error {
	client, found := k.GetClient(ctx, clientID)
	if !found {
		return types.ErrClientNotFound.Wrapf("client %s not found", clientID)
	}

	if !client.IsActive() {
		if client.IsSuspended() {
			return types.ErrClientSuspended.Wrapf("client %s is suspended", clientID)
		}
		if client.IsRevoked() {
			return types.ErrClientRevoked.Wrapf("client %s is revoked", clientID)
		}
		return types.ErrClientNotApproved.Wrapf("client %s is not active", clientID)
	}

	if !client.CanSubmitScope(scopeType) {
		return types.ErrScopeNotAllowed.Wrapf("client %s cannot submit scope type %s", clientID, scopeType)
	}

	return nil
}

// setClient stores an approved client
func (k Keeper) setClient(ctx sdk.Context, client *types.ApprovedClient) error {
	cs := approvedClientStore{
		ClientID:      client.ClientID,
		Name:          client.Name,
		Description:   client.Description,
		PublicKey:     client.PublicKey,
		KeyType:       string(client.KeyType),
		MinVersion:    client.MinVersion,
		MaxVersion:    client.MaxVersion,
		AllowedScopes: client.AllowedScopes,
		Status:        string(client.Status),
		StatusReason:  client.StatusReason,
		RegisteredBy:  client.RegisteredBy,
		RegisteredAt:  client.RegisteredAt.Unix(),
		LastUpdatedAt: client.LastUpdatedAt.Unix(),
		LastUpdatedBy: client.LastUpdatedBy,
		Metadata:      client.Metadata,
	}

	if client.SuspendedAt != nil {
		ts := client.SuspendedAt.Unix()
		cs.SuspendedAt = &ts
	}
	if client.RevokedAt != nil {
		ts := client.RevokedAt.Unix()
		cs.RevokedAt = &ts
	}

	bz, err := json.Marshal(&cs)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ApprovedClientKey(client.ClientID), bz)
	return nil
}

// clientStoreToClient converts a clientStore to ApprovedClient
func (k Keeper) clientStoreToClient(cs approvedClientStore) types.ApprovedClient {
	client := types.ApprovedClient{
		ClientID:      cs.ClientID,
		Name:          cs.Name,
		Description:   cs.Description,
		PublicKey:     cs.PublicKey,
		KeyType:       types.KeyType(cs.KeyType),
		MinVersion:    cs.MinVersion,
		MaxVersion:    cs.MaxVersion,
		AllowedScopes: cs.AllowedScopes,
		Status:        types.ClientStatus(cs.Status),
		StatusReason:  cs.StatusReason,
		RegisteredBy:  cs.RegisteredBy,
		RegisteredAt:  time.Unix(cs.RegisteredAt, 0),
		LastUpdatedAt: time.Unix(cs.LastUpdatedAt, 0),
		LastUpdatedBy: cs.LastUpdatedBy,
		Metadata:      cs.Metadata,
	}

	if cs.SuspendedAt != nil {
		t := time.Unix(*cs.SuspendedAt, 0)
		client.SuspendedAt = &t
	}
	if cs.RevokedAt != nil {
		t := time.Unix(*cs.RevokedAt, 0)
		client.RevokedAt = &t
	}

	return client
}

// iterateClients iterates over all approved clients
func (k Keeper) iterateClients(ctx sdk.Context, fn func(client types.ApprovedClient) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.ApprovedClientPrefixKey())
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var cs approvedClientStore
		if err := json.Unmarshal(iter.Value(), &cs); err != nil {
			continue
		}

		if fn(k.clientStoreToClient(cs)) {
			break
		}
	}
}

// addClientToStatusIndex adds a client to the status index
func (k Keeper) addClientToStatusIndex(ctx sdk.Context, clientID string, status types.ClientStatus) {
	store := ctx.KVStore(k.skey)
	store.Set(types.ApprovedClientByStatusKey(status, clientID), []byte{1})
}

// removeClientFromStatusIndex removes a client from the status index
func (k Keeper) removeClientFromStatusIndex(ctx sdk.Context, clientID string, status types.ClientStatus) {
	store := ctx.KVStore(k.skey)
	store.Delete(types.ApprovedClientByStatusKey(status, clientID))
}

// storeAuditEntry stores an audit log entry
func (k Keeper) storeAuditEntry(ctx sdk.Context, entry *types.AuditEntry) {
	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(entry)
	if err != nil {
		return
	}

	store.Set(types.ClientAuditLogKey(entry.ClientID, entry.Timestamp.UnixNano()), bz)
}

// GetAuditHistory returns the audit history for a client
func (k Keeper) GetAuditHistory(ctx sdk.Context, clientID string) []types.AuditEntry {
	var entries []types.AuditEntry

	store := ctx.KVStore(k.skey)
	prefix := types.ClientAuditLogPrefixKey(clientID)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var entry types.AuditEntry
		if err := json.Unmarshal(iter.Value(), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries
}

// InitGenesis initializes the module's state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, gs *types.GenesisState) error {
	// Set params
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}

	// Register initial clients
	for _, client := range gs.ApprovedClients {
		if err := k.RegisterClient(ctx, client); err != nil {
			return err
		}
		// Add to status index
		k.addClientToStatusIndex(ctx, client.ClientID, client.Status)
	}

	return nil
}

// ExportGenesis exports the module's state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		ApprovedClients: k.ListClients(ctx),
		Params:          k.GetParams(ctx),
	}
}

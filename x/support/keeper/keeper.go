package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	rolestypes "github.com/virtengine/virtengine/x/roles/types"
	types "github.com/virtengine/virtengine/x/support/types" //nolint:staticcheck // Deprecated types retained for compatibility.
)

// IKeeper defines the interface for the support keeper
type IKeeper interface {
	// External reference management
	RegisterExternalRef(ctx sdk.Context, ref *types.ExternalTicketRef) error
	GetExternalRef(ctx sdk.Context, resourceType types.ResourceType, resourceID string) (types.ExternalTicketRef, bool)
	UpdateExternalRef(ctx sdk.Context, ref *types.ExternalTicketRef) error
	RemoveExternalRef(ctx sdk.Context, resourceType types.ResourceType, resourceID string) error

	// Query methods
	GetExternalRefsByOwner(ctx sdk.Context, ownerAddr sdk.AccAddress) []types.ExternalTicketRef
	WithExternalRefs(ctx sdk.Context, fn func(ref types.ExternalTicketRef) bool)

	// Support requests
	CreateSupportRequest(ctx sdk.Context, request *types.SupportRequest) error
	GetSupportRequest(ctx sdk.Context, id types.SupportRequestID) (types.SupportRequest, bool)
	UpdateSupportRequest(ctx sdk.Context, request *types.SupportRequest) error
	GetSupportRequestsBySubmitter(ctx sdk.Context, submitter sdk.AccAddress) []types.SupportRequest
	GetSupportRequestsByStatus(ctx sdk.Context, status types.SupportStatus) []types.SupportRequest
	WithSupportRequests(ctx sdk.Context, fn func(req types.SupportRequest) bool)
	ArchiveSupportRequest(ctx sdk.Context, id types.SupportRequestID, reason string, archivedBy string) error
	PurgeSupportRequestPayload(ctx sdk.Context, id types.SupportRequestID, reason string, purgedBy string) error

	// Support responses
	AddSupportResponse(ctx sdk.Context, response *types.SupportResponse) error
	GetSupportResponses(ctx sdk.Context, requestID types.SupportRequestID) []types.SupportResponse
	GetSupportResponse(ctx sdk.Context, responseID types.SupportResponseID) (types.SupportResponse, bool)

	// Support events
	EmitSupportEvent(ctx sdk.Context, event types.SupportEvent) error
	GetEventSequence(ctx sdk.Context) uint64
	IncrementEventSequence(ctx sdk.Context) uint64
	SetEventSequence(ctx sdk.Context, seq uint64)
	GetEventCheckpoint(ctx sdk.Context, subscriberID string) (*types.SupportEventCheckpoint, bool)
	SetEventCheckpoint(ctx sdk.Context, checkpoint *types.SupportEventCheckpoint) error

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
}

// Keeper of the support store
type Keeper struct {
	skey storetypes.StoreKey
	cdc  codec.BinaryCodec

	// The address capable of executing a MsgUpdateParams message.
	// This should be the x/gov module account.
	authority string

	encryptionKeeper EncryptionKeeper
	rolesKeeper      RolesKeeper
}

// EncryptionKeeper defines the interface for encryption module interactions.
type EncryptionKeeper interface {
	ValidateEnvelope(ctx sdk.Context, envelope *encryptiontypes.EncryptedPayloadEnvelope) error
	ValidateEnvelopeRecipients(ctx sdk.Context, envelope *encryptiontypes.EncryptedPayloadEnvelope) ([]string, error)
	GetActiveRecipientKey(ctx sdk.Context, address sdk.AccAddress) (encryptiontypes.RecipientKeyRecord, bool)
}

// RolesKeeper defines the interface for roles checks.
type RolesKeeper interface {
	HasRole(ctx sdk.Context, address sdk.AccAddress, role rolestypes.Role) bool
	IsAdmin(ctx sdk.Context, addr sdk.AccAddress) bool
}

// NewKeeper creates and returns an instance for support keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	skey storetypes.StoreKey,
	authority string,
	encryptionKeeper EncryptionKeeper,
	rolesKeeper RolesKeeper,
) Keeper {
	return Keeper{
		cdc:              cdc,
		skey:             skey,
		authority:        authority,
		encryptionKeeper: encryptionKeeper,
		rolesKeeper:      rolesKeeper,
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

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// ============================================================================
// Parameters
// ============================================================================

// paramsStore is the stored format of params
type paramsStore struct {
	AllowedExternalSystems   []string              `json:"allowed_external_systems"`
	AllowedExternalDomains   []string              `json:"allowed_external_domains"`
	SupportRecipientKeyIDs   []string              `json:"support_recipient_key_ids"`
	RequireSupportRecipients bool                  `json:"require_support_recipients"`
	MaxResponsesPerRequest   uint32                `json:"max_responses_per_request"`
	DefaultRetentionPolicy   types.RetentionPolicy `json:"default_retention_policy"`
	RetentionQueueBatchSize  uint32                `json:"retention_queue_batch_size"`
	RetentionQueueMaxRetries uint32                `json:"retention_queue_max_retries"`
	RetentionQueueRetryBase  int64                 `json:"retention_queue_retry_base_seconds"`
	RetentionQueueRetryMax   int64                 `json:"retention_queue_retry_max_seconds"`
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&paramsStore{
		AllowedExternalSystems:   params.AllowedExternalSystems,
		AllowedExternalDomains:   params.AllowedExternalDomains,
		SupportRecipientKeyIDs:   params.SupportRecipientKeyIDs,
		RequireSupportRecipients: params.RequireSupportRecipients,
		MaxResponsesPerRequest:   params.MaxResponsesPerRequest,
		DefaultRetentionPolicy:   params.DefaultRetentionPolicy,
		RetentionQueueBatchSize:  params.RetentionQueueBatchSize,
		RetentionQueueMaxRetries: params.RetentionQueueMaxRetries,
		RetentionQueueRetryBase:  params.RetentionQueueRetryBaseSeconds,
		RetentionQueueRetryMax:   params.RetentionQueueRetryMaxSeconds,
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
		AllowedExternalSystems:         ps.AllowedExternalSystems,
		AllowedExternalDomains:         ps.AllowedExternalDomains,
		SupportRecipientKeyIDs:         ps.SupportRecipientKeyIDs,
		RequireSupportRecipients:       ps.RequireSupportRecipients,
		MaxResponsesPerRequest:         ps.MaxResponsesPerRequest,
		DefaultRetentionPolicy:         ps.DefaultRetentionPolicy,
		RetentionQueueBatchSize:        ps.RetentionQueueBatchSize,
		RetentionQueueMaxRetries:       ps.RetentionQueueMaxRetries,
		RetentionQueueRetryBaseSeconds: ps.RetentionQueueRetryBase,
		RetentionQueueRetryMaxSeconds:  ps.RetentionQueueRetryMax,
	}
}

// ============================================================================
// External Reference Management
// ============================================================================

// externalRefStore is the stored format of an external ticket reference
type externalRefStore struct {
	ResourceID       string `json:"resource_id"`
	ResourceType     string `json:"resource_type"`
	ExternalSystem   string `json:"external_system"`
	ExternalTicketID string `json:"external_ticket_id"`
	ExternalURL      string `json:"external_url,omitempty"`
	CreatedAt        int64  `json:"created_at"`
	CreatedBy        string `json:"created_by"`
	UpdatedAt        int64  `json:"updated_at"`
}

// RegisterExternalRef registers a new external ticket reference
func (k Keeper) RegisterExternalRef(ctx sdk.Context, ref *types.ExternalTicketRef) error {
	if err := ref.Validate(); err != nil {
		return err
	}

	// Check if ref already exists
	if _, found := k.GetExternalRef(ctx, ref.ResourceType, ref.ResourceID); found {
		return types.ErrRefAlreadyExists.Wrapf("ref for %s/%s already exists", ref.ResourceType, ref.ResourceID)
	}

	// Validate external system is allowed
	params := k.GetParams(ctx)
	if !params.IsSystemAllowed(ref.ExternalSystem) {
		return types.ErrInvalidExternalSystem.Wrapf("system %s is not allowed", ref.ExternalSystem)
	}

	// Set timestamps
	now := ctx.BlockTime()
	ref.CreatedAt = now
	ref.UpdatedAt = now

	// Store the reference
	return k.setExternalRef(ctx, ref)
}

// GetExternalRef returns an external ticket reference
func (k Keeper) GetExternalRef(ctx sdk.Context, resourceType types.ResourceType, resourceID string) (types.ExternalTicketRef, bool) {
	store := ctx.KVStore(k.skey)
	key := types.ExternalRefKey(resourceType, resourceID)
	bz := store.Get(key)
	if bz == nil {
		return types.ExternalTicketRef{}, false
	}

	var rs externalRefStore
	if err := json.Unmarshal(bz, &rs); err != nil {
		return types.ExternalTicketRef{}, false
	}

	return k.refStoreToRef(rs), true
}

// UpdateExternalRef updates an existing external ticket reference
func (k Keeper) UpdateExternalRef(ctx sdk.Context, ref *types.ExternalTicketRef) error {
	// Check if ref exists
	existing, found := k.GetExternalRef(ctx, ref.ResourceType, ref.ResourceID)
	if !found {
		return types.ErrRefNotFound.Wrapf("ref for %s/%s not found", ref.ResourceType, ref.ResourceID)
	}

	// Preserve original creation info
	ref.CreatedAt = existing.CreatedAt
	ref.CreatedBy = existing.CreatedBy
	ref.UpdatedAt = ctx.BlockTime()

	return k.setExternalRef(ctx, ref)
}

// RemoveExternalRef removes an external ticket reference
func (k Keeper) RemoveExternalRef(ctx sdk.Context, resourceType types.ResourceType, resourceID string) error {
	ref, found := k.GetExternalRef(ctx, resourceType, resourceID)
	if !found {
		return types.ErrRefNotFound.Wrapf("ref for %s/%s not found", resourceType, resourceID)
	}

	store := ctx.KVStore(k.skey)

	// Remove owner index
	ownerAddr, _ := sdk.AccAddressFromBech32(ref.CreatedBy)
	store.Delete(types.ExternalRefByOwnerKey(ownerAddr.Bytes(), resourceType, resourceID))

	// Remove the reference
	store.Delete(types.ExternalRefKey(resourceType, resourceID))

	return nil
}

// setExternalRef stores an external ticket reference
func (k Keeper) setExternalRef(ctx sdk.Context, ref *types.ExternalTicketRef) error {
	rs := externalRefStore{
		ResourceID:       ref.ResourceID,
		ResourceType:     string(ref.ResourceType),
		ExternalSystem:   string(ref.ExternalSystem),
		ExternalTicketID: ref.ExternalTicketID,
		ExternalURL:      ref.ExternalURL,
		CreatedAt:        ref.CreatedAt.Unix(),
		CreatedBy:        ref.CreatedBy,
		UpdatedAt:        ref.UpdatedAt.Unix(),
	}

	bz, err := json.Marshal(&rs)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ExternalRefKey(ref.ResourceType, ref.ResourceID), bz)

	// Add owner index
	ownerAddr, _ := sdk.AccAddressFromBech32(ref.CreatedBy)
	store.Set(types.ExternalRefByOwnerKey(ownerAddr.Bytes(), ref.ResourceType, ref.ResourceID), []byte{1})

	return nil
}

// refStoreToRef converts a stored format to ExternalTicketRef
func (k Keeper) refStoreToRef(rs externalRefStore) types.ExternalTicketRef {
	return types.ExternalTicketRef{
		ResourceID:       rs.ResourceID,
		ResourceType:     types.ResourceType(rs.ResourceType),
		ExternalSystem:   types.ExternalSystem(rs.ExternalSystem),
		ExternalTicketID: rs.ExternalTicketID,
		ExternalURL:      rs.ExternalURL,
		CreatedAt:        time.Unix(rs.CreatedAt, 0),
		CreatedBy:        rs.CreatedBy,
		UpdatedAt:        time.Unix(rs.UpdatedAt, 0),
	}
}

// ============================================================================
// Query Methods
// ============================================================================

// GetExternalRefsByOwner returns all external refs created by an owner
func (k Keeper) GetExternalRefsByOwner(ctx sdk.Context, ownerAddr sdk.AccAddress) []types.ExternalTicketRef {
	var refs []types.ExternalTicketRef

	store := ctx.KVStore(k.skey)
	prefix := types.ExternalRefByOwnerPrefixKey(ownerAddr.Bytes())
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// Parse key to get resource type and ID
		key := iter.Key()
		remaining := key[len(prefix):]

		// Find separator between resource type and ID
		for i := range remaining {
			if remaining[i] == '/' {
				resourceType := types.ResourceType(remaining[:i])
				resourceID := string(remaining[i+1:])
				if ref, found := k.GetExternalRef(ctx, resourceType, resourceID); found {
					refs = append(refs, ref)
				}
				break
			}
		}
	}

	return refs
}

// WithExternalRefs iterates over all external refs
func (k Keeper) WithExternalRefs(ctx sdk.Context, fn func(ref types.ExternalTicketRef) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixExternalRef)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var rs externalRefStore
		if err := json.Unmarshal(iter.Value(), &rs); err != nil {
			continue
		}

		if fn(k.refStoreToRef(rs)) {
			break
		}
	}
}

// ========================================================================
// Support Requests
// ========================================================================

// CreateSupportRequest stores a new support request.
func (k Keeper) CreateSupportRequest(ctx sdk.Context, request *types.SupportRequest) error {
	if err := request.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.SupportRequestKey(request.ID.String())
	if store.Has(key) {
		return types.ErrInvalidSupportRequest.Wrap("support request already exists")
	}

	request.Payload.EnsureEnvelopeHash()

	bz, err := json.Marshal(request)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	submitterAddr, _ := sdk.AccAddressFromBech32(request.SubmitterAddress)
	store.Set(types.SupportRequestBySubmitterKey(submitterAddr.Bytes(), request.ID.String()), []byte{1})
	store.Set(types.SupportRequestByStatusKey(request.Status, request.ID.String()), []byte{1})

	k.enqueueRetention(ctx, request)

	return nil
}

// GetSupportRequest retrieves a support request by ID.
func (k Keeper) GetSupportRequest(ctx sdk.Context, id types.SupportRequestID) (types.SupportRequest, bool) {
	store := ctx.KVStore(k.skey)
	key := types.SupportRequestKey(id.String())
	bz := store.Get(key)
	if bz == nil {
		return types.SupportRequest{}, false
	}

	var req types.SupportRequest
	if err := json.Unmarshal(bz, &req); err != nil {
		return types.SupportRequest{}, false
	}
	return req, true
}

// UpdateSupportRequest updates an existing support request.
func (k Keeper) UpdateSupportRequest(ctx sdk.Context, request *types.SupportRequest) error {
	if err := request.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.SupportRequestKey(request.ID.String())
	existingBz := store.Get(key)
	if existingBz == nil {
		return types.ErrSupportRequestNotFound
	}

	var existing types.SupportRequest
	if err := json.Unmarshal(existingBz, &existing); err != nil {
		return err
	}

	request.Payload.EnsureEnvelopeHash()

	bz, err := json.Marshal(request)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	// Ensure indexes exist (idempotent)
	submitterAddr, _ := sdk.AccAddressFromBech32(request.SubmitterAddress)
	store.Set(types.SupportRequestBySubmitterKey(submitterAddr.Bytes(), request.ID.String()), []byte{1})
	if existing.Status != request.Status {
		store.Delete(types.SupportRequestByStatusKey(existing.Status, request.ID.String()))
	}
	store.Set(types.SupportRequestByStatusKey(request.Status, request.ID.String()), []byte{1})

	k.enqueueRetention(ctx, request)

	return nil
}

// GetSupportRequestsBySubmitter returns requests for a submitter.
func (k Keeper) GetSupportRequestsBySubmitter(ctx sdk.Context, submitter sdk.AccAddress) []types.SupportRequest {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.SupportRequestBySubmitterPrefixKey(submitter.Bytes()))
	defer iter.Close()

	var requests []types.SupportRequest
	for ; iter.Valid(); iter.Next() {
		requestID := string(iter.Key()[len(types.SupportRequestBySubmitterPrefixKey(submitter.Bytes())):])
		id, err := types.ParseSupportRequestID(requestID)
		if err != nil {
			continue
		}
		req, found := k.GetSupportRequest(ctx, id)
		if found {
			requests = append(requests, req)
		}
	}
	return requests
}

// GetSupportRequestsByStatus returns requests for a status.
func (k Keeper) GetSupportRequestsByStatus(ctx sdk.Context, status types.SupportStatus) []types.SupportRequest {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.SupportRequestByStatusPrefixKey(status))
	defer iter.Close()

	var requests []types.SupportRequest
	for ; iter.Valid(); iter.Next() {
		requestID := string(iter.Key()[len(types.SupportRequestByStatusPrefixKey(status)):])
		id, err := types.ParseSupportRequestID(requestID)
		if err != nil {
			continue
		}
		req, found := k.GetSupportRequest(ctx, id)
		if found {
			requests = append(requests, req)
		}
	}
	return requests
}

// WithSupportRequests iterates over all support requests.
func (k Keeper) WithSupportRequests(ctx sdk.Context, fn func(req types.SupportRequest) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixSupportRequest)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var req types.SupportRequest
		if err := json.Unmarshal(iter.Value(), &req); err != nil {
			continue
		}
		if fn(req) {
			break
		}
	}
}

// ArchiveSupportRequest archives a support request.
func (k Keeper) ArchiveSupportRequest(ctx sdk.Context, id types.SupportRequestID, reason string, archivedBy string) error {
	req, found := k.GetSupportRequest(ctx, id)
	if !found {
		return types.ErrSupportRequestNotFound
	}
	if req.Archived {
		return nil
	}

	req.MarkArchived(reason, ctx.BlockTime())
	req.AppendAuditEntry(types.NewSupportAuditEntry(
		types.SupportAuditActionRequestArchived,
		archivedBy,
		reason,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	))
	if err := k.UpdateSupportRequest(ctx, &req); err != nil {
		return err
	}

	seq := k.IncrementEventSequence(ctx)
	event := types.SupportRequestArchivedEvent{
		EventType:   string(types.SupportEventTypeRequestArchived),
		EventID:     fmt.Sprintf("%s/%d", req.ID.String(), seq),
		BlockHeight: ctx.BlockHeight(),
		Sequence:    seq,
		TicketID:    req.ID.String(),
		ArchivedBy:  archivedBy,
		Reason:      reason,
		Timestamp:   ctx.BlockTime().Unix(),
	}
	return k.EmitSupportEvent(ctx, event)
}

// PurgeSupportRequestPayload purges the encrypted payload.
func (k Keeper) PurgeSupportRequestPayload(ctx sdk.Context, id types.SupportRequestID, reason string, purgedBy string) error {
	req, found := k.GetSupportRequest(ctx, id)
	if !found {
		return types.ErrSupportRequestNotFound
	}
	if req.Purged {
		return nil
	}

	req.Payload = req.Payload.CloneWithoutEnvelope()
	req.MarkPurged(reason, ctx.BlockTime())
	req.AppendAuditEntry(types.NewSupportAuditEntry(
		types.SupportAuditActionRequestPurged,
		purgedBy,
		reason,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	))
	if err := k.UpdateSupportRequest(ctx, &req); err != nil {
		return err
	}

	seq := k.IncrementEventSequence(ctx)
	event := types.SupportRequestPurgedEvent{
		EventType:   string(types.SupportEventTypeRequestPurged),
		EventID:     fmt.Sprintf("%s/%d", req.ID.String(), seq),
		BlockHeight: ctx.BlockHeight(),
		Sequence:    seq,
		TicketID:    req.ID.String(),
		PurgedBy:    purgedBy,
		Reason:      reason,
		Timestamp:   ctx.BlockTime().Unix(),
	}
	return k.EmitSupportEvent(ctx, event)
}

// AddSupportResponse stores a response to a request.
func (k Keeper) AddSupportResponse(ctx sdk.Context, response *types.SupportResponse) error {
	if err := response.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.SupportResponseKey(response.RequestID.String(), response.ID.Sequence)
	if store.Has(key) {
		return types.ErrInvalidSupportResponse.Wrap("response already exists")
	}

	response.Payload.EnsureEnvelopeHash()

	bz, err := json.Marshal(response)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	store.Set(types.SupportResponseByRequestKey(response.RequestID.String(), response.ID.Sequence), []byte{1})

	return nil
}

// GetSupportResponses returns responses for a request.
func (k Keeper) GetSupportResponses(ctx sdk.Context, requestID types.SupportRequestID) []types.SupportResponse {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.SupportResponseByRequestPrefixKey(requestID.String()))
	defer iter.Close()

	var responses []types.SupportResponse
	for ; iter.Valid(); iter.Next() {
		seqStr := string(iter.Key()[len(types.SupportResponseByRequestPrefixKey(requestID.String())):])
		seq, err := strconv.ParseUint(seqStr, 10, 64)
		if err != nil {
			continue
		}
		resp, found := k.GetSupportResponse(ctx, types.SupportResponseID{
			RequestID: requestID,
			Sequence:  seq,
		})
		if found {
			responses = append(responses, resp)
		}
	}
	return responses
}

// GetSupportResponse returns a response by ID.
func (k Keeper) GetSupportResponse(ctx sdk.Context, responseID types.SupportResponseID) (types.SupportResponse, bool) {
	store := ctx.KVStore(k.skey)
	key := types.SupportResponseKey(responseID.RequestID.String(), responseID.Sequence)
	bz := store.Get(key)
	if bz == nil {
		return types.SupportResponse{}, false
	}
	var resp types.SupportResponse
	if err := json.Unmarshal(bz, &resp); err != nil {
		return types.SupportResponse{}, false
	}
	return resp, true
}

// SupportRequestSequence returns the next request sequence for a submitter.
func (k Keeper) SupportRequestSequence(ctx sdk.Context, submitter sdk.AccAddress) uint64 {
	store := ctx.KVStore(k.skey)
	key := types.SupportRequestSequenceKey(submitter.Bytes())
	bz := store.Get(key)
	if bz == nil {
		return 0
	}
	var seq uint64
	if err := json.Unmarshal(bz, &seq); err != nil {
		return 0
	}
	return seq
}

// IncrementSupportRequestSequence increments and stores submitter sequence.
func (k Keeper) IncrementSupportRequestSequence(ctx sdk.Context, submitter sdk.AccAddress) uint64 {
	store := ctx.KVStore(k.skey)
	seq := k.SupportRequestSequence(ctx, submitter) + 1
	bz, _ := json.Marshal(seq) //nolint:errchkjson
	store.Set(types.SupportRequestSequenceKey(submitter.Bytes()), bz)
	return seq
}

// SetSupportRequestSequence sets the submitter sequence.
func (k Keeper) SetSupportRequestSequence(ctx sdk.Context, submitter sdk.AccAddress, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz, _ := json.Marshal(seq) //nolint:errchkjson
	store.Set(types.SupportRequestSequenceKey(submitter.Bytes()), bz)
}

// TicketNumberSequence returns the global ticket number sequence.
func (k Keeper) TicketNumberSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SupportTicketNumberKey())
	if bz == nil {
		return 0
	}
	var seq uint64
	if err := json.Unmarshal(bz, &seq); err != nil {
		return 0
	}
	return seq
}

// IncrementTicketNumberSequence increments and stores ticket number sequence.
func (k Keeper) IncrementTicketNumberSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	seq := k.TicketNumberSequence(ctx) + 1
	bz, _ := json.Marshal(seq) //nolint:errchkjson
	store.Set(types.SupportTicketNumberKey(), bz)
	return seq
}

// SetTicketNumberSequence sets the global ticket number sequence.
func (k Keeper) SetTicketNumberSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz, _ := json.Marshal(seq) //nolint:errchkjson
	store.Set(types.SupportTicketNumberKey(), bz)
}

// SupportResponseSequence returns the next response sequence for a request.
func (k Keeper) SupportResponseSequence(ctx sdk.Context, requestID types.SupportRequestID) uint64 {
	store := ctx.KVStore(k.skey)
	key := types.SupportResponseSequenceKey(requestID.String())
	bz := store.Get(key)
	if bz == nil {
		return 0
	}
	var seq uint64
	if err := json.Unmarshal(bz, &seq); err != nil {
		return 0
	}
	return seq
}

// IncrementSupportResponseSequence increments and stores response sequence.
func (k Keeper) IncrementSupportResponseSequence(ctx sdk.Context, requestID types.SupportRequestID) uint64 {
	store := ctx.KVStore(k.skey)
	seq := k.SupportResponseSequence(ctx, requestID) + 1
	bz, _ := json.Marshal(seq) //nolint:errchkjson
	store.Set(types.SupportResponseSequenceKey(requestID.String()), bz)
	return seq
}

// SetSupportResponseSequence sets the response sequence for a request.
func (k Keeper) SetSupportResponseSequence(ctx sdk.Context, requestID string, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz, _ := json.Marshal(seq) //nolint:errchkjson
	store.Set(types.SupportResponseSequenceKey(requestID), bz)
}

// ProcessRetentionPolicies checks and applies retention policies.
func (k Keeper) ProcessRetentionPolicies(ctx sdk.Context) (int, int) {
	params := k.GetParams(ctx)
	now := ctx.BlockTime()
	batch := int(params.RetentionQueueBatchSize)
	archived := k.processRetentionQueue(ctx, types.RetentionActionArchive, now, batch)
	purged := k.processRetentionQueue(ctx, types.RetentionActionPurge, now, batch)
	return archived, purged
}

func (k Keeper) enqueueRetention(ctx sdk.Context, request *types.SupportRequest) {
	if request == nil {
		return
	}
	if request.RetentionPolicy == nil {
		k.clearRetentionQueueEntry(ctx, request.ID.String(), types.RetentionActionArchive)
		k.clearRetentionQueueEntry(ctx, request.ID.String(), types.RetentionActionPurge)
		return
	}

	if request.Archived {
		k.clearRetentionQueueEntry(ctx, request.ID.String(), types.RetentionActionArchive)
	} else if archiveAt, ok := request.RetentionPolicy.ArchiveAt(); ok {
		k.ensureRetentionQueueEntry(ctx, request.ID.String(), types.RetentionActionArchive, archiveAt)
	}

	if request.Purged {
		k.clearRetentionQueueEntry(ctx, request.ID.String(), types.RetentionActionPurge)
	} else if purgeAt, ok := request.RetentionPolicy.PurgeAt(); ok {
		k.ensureRetentionQueueEntry(ctx, request.ID.String(), types.RetentionActionPurge, purgeAt)
	}
}

// BuildRetentionStatus builds retention status for a support request.
func (k Keeper) BuildRetentionStatus(ctx sdk.Context, request types.SupportRequest) types.RetentionStatus {
	status := types.RetentionStatus{
		TicketID: request.ID.String(),
		Archived: request.Archived,
		Purged:   request.Purged,
	}
	if request.RetentionPolicy != nil {
		if archiveAt, ok := request.RetentionPolicy.ArchiveAt(); ok {
			archiveAt = archiveAt.UTC()
			status.ArchiveAt = &archiveAt
		}
		if purgeAt, ok := request.RetentionPolicy.PurgeAt(); ok {
			purgeAt = purgeAt.UTC()
			status.PurgeAt = &purgeAt
		}
	}
	if entry, found := k.GetRetentionQueueEntry(ctx, request.ID.String(), types.RetentionActionArchive); found {
		status.ArchiveQueue = &entry
	}
	if entry, found := k.GetRetentionQueueEntry(ctx, request.ID.String(), types.RetentionActionPurge); found {
		status.PurgeQueue = &entry
	}
	return status
}

func (k Keeper) processRetentionQueue(ctx sdk.Context, action types.RetentionAction, now time.Time, limit int) int {
	if limit <= 0 {
		return 0
	}
	store := ctx.KVStore(k.skey)
	params := k.GetParams(ctx)

	var prefix []byte
	switch action {
	case types.RetentionActionArchive:
		prefix = types.PrefixSupportArchiveQueue
	case types.RetentionActionPurge:
		prefix = types.PrefixSupportPurgeQueue
	default:
		return 0
	}

	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	type queued struct {
		key       []byte
		scheduled int64
		requestID string
	}
	queuedItems := make([]queued, 0, limit)
	nowUnix := now.Unix()

	for ; iter.Valid(); iter.Next() {
		scheduled, requestID, ok := types.ParseRetentionQueueKey(prefix, iter.Key())
		if !ok {
			continue
		}
		if scheduled > nowUnix {
			// Keys are ordered by timestamp for new format; stop once we reach future entries.
			break
		}
		queuedItems = append(queuedItems, queued{
			key:       append([]byte(nil), iter.Key()...),
			scheduled: scheduled,
			requestID: requestID,
		})
		if len(queuedItems) >= limit {
			break
		}
	}

	processed := 0
	for _, item := range queuedItems {
		entry := types.RetentionQueueEntry{}
		bz := store.Get(item.key)
		if bz != nil {
			if err := json.Unmarshal(bz, &entry); err != nil {
				store.Delete(item.key)
				continue
			}
		}
		if entry.RequestID == "" {
			entry.RequestID = item.requestID
		}
		entry.Action = action
		entry.ScheduledAt = time.Unix(item.scheduled, 0).UTC()

		if err := k.applyRetentionAction(ctx, entry); err != nil {
			if errors.Is(err, types.ErrSupportRequestNotFound) {
				k.deleteRetentionQueueEntry(ctx, entry, item.scheduled)
				processed++
				continue
			}
			entry.Attempts++
			entry.LastError = err.Error()
			nowUTC := now.UTC()
			entry.LastAttemptAt = &nowUTC

			if entry.Attempts >= params.RetentionQueueMaxRetries {
				k.deleteRetentionQueueEntry(ctx, entry, item.scheduled)
				processed++
				continue
			}

			delay := retentionBackoffDelay(entry.Attempts, params.RetentionQueueRetryBaseSeconds, params.RetentionQueueRetryMaxSeconds)
			entry.ScheduledAt = nowUTC.Add(delay)
			k.rescheduleRetentionQueueEntry(ctx, entry, item.scheduled)
			processed++
			continue
		}

		k.deleteRetentionQueueEntry(ctx, entry, item.scheduled)
		processed++
	}

	return processed
}

func (k Keeper) applyRetentionAction(ctx sdk.Context, entry types.RetentionQueueEntry) error {
	reqID, err := types.ParseSupportRequestID(entry.RequestID)
	if err != nil {
		return err
	}

	switch entry.Action {
	case types.RetentionActionArchive:
		return k.ArchiveSupportRequest(ctx, reqID, "retention policy", "system")
	case types.RetentionActionPurge:
		return k.PurgeSupportRequestPayload(ctx, reqID, "retention policy", "system")
	default:
		return errors.New("unknown retention action")
	}
}

func retentionBackoffDelay(attempts uint32, baseSeconds int64, maxSeconds int64) time.Duration {
	if attempts == 0 {
		return time.Duration(baseSeconds) * time.Second
	}
	base := time.Duration(baseSeconds) * time.Second
	delay := base << (attempts - 1)
	max := time.Duration(maxSeconds) * time.Second
	if delay > max {
		return max
	}
	return delay
}

func (k Keeper) ensureRetentionQueueEntry(ctx sdk.Context, requestID string, action types.RetentionAction, scheduledAt time.Time) {
	if _, found := k.GetRetentionQueueEntry(ctx, requestID, action); found {
		return
	}
	entry := types.RetentionQueueEntry{
		RequestID:   requestID,
		Action:      action,
		ScheduledAt: scheduledAt.UTC(),
	}
	_ = k.setRetentionQueueEntry(ctx, entry)
}

func (k Keeper) GetRetentionQueueEntry(ctx sdk.Context, requestID string, action types.RetentionAction) (types.RetentionQueueEntry, bool) {
	store := ctx.KVStore(k.skey)
	key := types.SupportRetentionIndexKey(requestID, action)
	bz := store.Get(key)
	if bz == nil {
		return types.RetentionQueueEntry{}, false
	}
	var entry types.RetentionQueueEntry
	if err := json.Unmarshal(bz, &entry); err != nil {
		return types.RetentionQueueEntry{}, false
	}
	return entry, true
}

func (k Keeper) clearRetentionQueueEntry(ctx sdk.Context, requestID string, action types.RetentionAction) {
	entry, found := k.GetRetentionQueueEntry(ctx, requestID, action)
	if !found {
		return
	}
	k.deleteRetentionQueueEntry(ctx, entry, entry.ScheduledAt.Unix())
}

func (k Keeper) setRetentionQueueEntry(ctx sdk.Context, entry types.RetentionQueueEntry) error {
	if entry.RequestID == "" || !entry.Action.IsValid() || entry.ScheduledAt.IsZero() {
		return types.ErrInvalidRetentionPolicy.Wrap("invalid retention queue entry")
	}
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&entry)
	if err != nil {
		return err
	}

	var queueKey []byte
	switch entry.Action {
	case types.RetentionActionArchive:
		queueKey = types.SupportArchiveQueueKey(entry.ScheduledAt.Unix(), entry.RequestID)
	case types.RetentionActionPurge:
		queueKey = types.SupportPurgeQueueKey(entry.ScheduledAt.Unix(), entry.RequestID)
	default:
		return types.ErrInvalidRetentionPolicy.Wrap("unsupported retention action")
	}

	store.Set(queueKey, bz)
	store.Set(types.SupportRetentionIndexKey(entry.RequestID, entry.Action), bz)
	return nil
}

func (k Keeper) deleteRetentionQueueEntry(ctx sdk.Context, entry types.RetentionQueueEntry, scheduledAt int64) {
	store := ctx.KVStore(k.skey)

	var queueKey []byte
	switch entry.Action {
	case types.RetentionActionArchive:
		queueKey = types.SupportArchiveQueueKey(scheduledAt, entry.RequestID)
	case types.RetentionActionPurge:
		queueKey = types.SupportPurgeQueueKey(scheduledAt, entry.RequestID)
	default:
		return
	}

	store.Delete(queueKey)
	store.Delete(types.SupportRetentionIndexKey(entry.RequestID, entry.Action))
}

func (k Keeper) rescheduleRetentionQueueEntry(ctx sdk.Context, entry types.RetentionQueueEntry, previousScheduledAt int64) {
	k.deleteRetentionQueueEntry(ctx, entry, previousScheduledAt)
	_ = k.setRetentionQueueEntry(ctx, entry)
}

// ========================================================================
// Support Events
// ========================================================================

// EmitSupportEvent emits a support event.
func (k Keeper) EmitSupportEvent(ctx sdk.Context, event types.SupportEvent) error {
	payloadJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	attributes := []sdk.Attribute{
		sdk.NewAttribute("event_type", string(event.GetEventType())),
		sdk.NewAttribute("event_id", event.GetEventID()),
		sdk.NewAttribute("block_height", fmt.Sprintf("%d", event.GetBlockHeight())),
		sdk.NewAttribute("sequence", fmt.Sprintf("%d", event.GetSequence())),
		sdk.NewAttribute("payload_json", string(payloadJSON)),
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent("support_event", attributes...),
	)

	return ctx.EventManager().EmitTypedEvent(&types.SupportEventWrapper{
		EventType:   string(event.GetEventType()),
		EventID:     event.GetEventID(),
		BlockHeight: event.GetBlockHeight(),
		Sequence:    event.GetSequence(),
	})
}

// GetEventSequence returns the current event sequence.
func (k Keeper) GetEventSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SupportEventSequenceKey())
	if bz == nil {
		return 0
	}

	var seq uint64
	if err := json.Unmarshal(bz, &seq); err != nil {
		return 0
	}
	return seq
}

// IncrementEventSequence increments and stores the event sequence.
func (k Keeper) IncrementEventSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	seq := k.GetEventSequence(ctx) + 1
	bz, _ := json.Marshal(seq) //nolint:errchkjson
	store.Set(types.SupportEventSequenceKey(), bz)
	return seq
}

// SetEventSequence sets the event sequence to a specific value.
func (k Keeper) SetEventSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz, _ := json.Marshal(seq) //nolint:errchkjson
	store.Set(types.SupportEventSequenceKey(), bz)
}

// GetEventCheckpoint returns a checkpoint for a subscriber.
func (k Keeper) GetEventCheckpoint(ctx sdk.Context, subscriberID string) (*types.SupportEventCheckpoint, bool) {
	store := ctx.KVStore(k.skey)
	key := types.SupportEventCheckpointKey(subscriberID)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var checkpoint types.SupportEventCheckpoint
	if err := json.Unmarshal(bz, &checkpoint); err != nil {
		return nil, false
	}
	return &checkpoint, true
}

// SetEventCheckpoint sets a checkpoint for a subscriber.
func (k Keeper) SetEventCheckpoint(ctx sdk.Context, checkpoint *types.SupportEventCheckpoint) error {
	if checkpoint == nil {
		return nil
	}
	store := ctx.KVStore(k.skey)
	checkpoint.UpdatedAt = ctx.BlockTime().UTC()
	bz, err := json.Marshal(checkpoint)
	if err != nil {
		return err
	}
	store.Set(types.SupportEventCheckpointKey(checkpoint.SubscriberID), bz)
	return nil
}

// NewGRPCQuerier returns a new GRPCQuerier
func (k Keeper) NewGRPCQuerier() GRPCQuerier {
	return GRPCQuerier{Keeper: k}
}

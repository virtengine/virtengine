package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	rolestypes "github.com/virtengine/virtengine/x/roles/types"
	"github.com/virtengine/virtengine/x/support/types"
)

// IKeeper defines the interface for the support keeper
type IKeeper interface {
	// Ticket management
	CreateTicket(ctx sdk.Context, ticket *types.SupportTicket) error
	GetTicket(ctx sdk.Context, ticketID string) (types.SupportTicket, bool)
	SetTicket(ctx sdk.Context, ticket *types.SupportTicket) error
	DeleteTicket(ctx sdk.Context, ticketID string) error

	// Ticket status transitions
	AssignTicket(ctx sdk.Context, ticketID string, agentAddr sdk.AccAddress, assignedBy sdk.AccAddress) error
	ResolveTicket(ctx sdk.Context, ticketID string, resolvedBy sdk.AccAddress, resolutionRef string) error
	CloseTicket(ctx sdk.Context, ticketID string, closedBy sdk.AccAddress, reason string) error
	ReopenTicket(ctx sdk.Context, ticketID string, reopenedBy sdk.AccAddress, reason string) error

	// Response management
	AddResponse(ctx sdk.Context, ticketID string, response *types.TicketResponse) error
	GetResponse(ctx sdk.Context, ticketID string, responseIndex uint32) (types.TicketResponse, bool)
	GetResponses(ctx sdk.Context, ticketID string) []types.TicketResponse

	// Ticket queries
	GetTicketsByCustomer(ctx sdk.Context, customerAddr sdk.AccAddress) []types.SupportTicket
	GetTicketsByProvider(ctx sdk.Context, providerAddr sdk.AccAddress) []types.SupportTicket
	GetTicketsByAgent(ctx sdk.Context, agentAddr sdk.AccAddress) []types.SupportTicket
	GetTicketsByStatus(ctx sdk.Context, status types.TicketStatus) []types.SupportTicket

	// Authorization checks
	CanViewTicket(ctx sdk.Context, viewer sdk.AccAddress, ticket *types.SupportTicket) bool
	CanRespondToTicket(ctx sdk.Context, responder sdk.AccAddress, ticket *types.SupportTicket) bool
	CanAssignTicket(ctx sdk.Context, assigner sdk.AccAddress) bool
	CanCloseTicket(ctx sdk.Context, closer sdk.AccAddress, ticket *types.SupportTicket) bool
	IsSupportAgent(ctx sdk.Context, addr sdk.AccAddress) bool
	IsSupportAdmin(ctx sdk.Context, addr sdk.AccAddress) bool

	// Rate limiting
	CheckRateLimit(ctx sdk.Context, addr sdk.AccAddress) error
	IncrementRateLimit(ctx sdk.Context, addr sdk.AccAddress)

	// Sequence management
	GetNextTicketID(ctx sdk.Context) string
	GetTicketSequence(ctx sdk.Context) uint64
	SetTicketSequence(ctx sdk.Context, seq uint64)

	// Iterators
	WithTickets(ctx sdk.Context, fn func(ticket types.SupportTicket) bool)

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
}

// RolesKeeper defines the interface for the roles keeper that support needs
type RolesKeeper interface {
	HasRole(ctx sdk.Context, address sdk.AccAddress, role rolestypes.Role) bool
	IsAdmin(ctx sdk.Context, addr sdk.AccAddress) bool
}

// Keeper of the support store
type Keeper struct {
	skey storetypes.StoreKey
	cdc  codec.BinaryCodec

	// The address capable of executing a MsgUpdateParams message.
	// This should be the x/gov module account.
	authority string

	// rolesKeeper is the roles keeper reference for authorization checks
	rolesKeeper RolesKeeper
}

// NewKeeper creates and returns an instance for support keeper
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

// SetRolesKeeper sets the roles keeper reference
func (k *Keeper) SetRolesKeeper(rolesKeeper RolesKeeper) {
	k.rolesKeeper = rolesKeeper
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
	MaxTicketsPerCustomerPerDay uint32   `json:"max_tickets_per_customer_per_day"`
	MaxResponsesPerTicket       uint32   `json:"max_responses_per_ticket"`
	TicketCooldownSeconds       uint32   `json:"ticket_cooldown_seconds"`
	AutoCloseAfterDays          uint32   `json:"auto_close_after_days"`
	MaxOpenTicketsPerCustomer   uint32   `json:"max_open_tickets_per_customer"`
	ReopenWindowDays            uint32   `json:"reopen_window_days"`
	AllowedCategories           []string `json:"allowed_categories"`
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&paramsStore{
		MaxTicketsPerCustomerPerDay: params.MaxTicketsPerCustomerPerDay,
		MaxResponsesPerTicket:       params.MaxResponsesPerTicket,
		TicketCooldownSeconds:       params.TicketCooldownSeconds,
		AutoCloseAfterDays:          params.AutoCloseAfterDays,
		MaxOpenTicketsPerCustomer:   params.MaxOpenTicketsPerCustomer,
		ReopenWindowDays:            params.ReopenWindowDays,
		AllowedCategories:           params.AllowedCategories,
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
		MaxTicketsPerCustomerPerDay: ps.MaxTicketsPerCustomerPerDay,
		MaxResponsesPerTicket:       ps.MaxResponsesPerTicket,
		TicketCooldownSeconds:       ps.TicketCooldownSeconds,
		AutoCloseAfterDays:          ps.AutoCloseAfterDays,
		MaxOpenTicketsPerCustomer:   ps.MaxOpenTicketsPerCustomer,
		ReopenWindowDays:            ps.ReopenWindowDays,
		AllowedCategories:           ps.AllowedCategories,
	}
}

// ============================================================================
// Ticket Management
// ============================================================================

// ticketStore is the stored format of a support ticket
type ticketStore struct {
	TicketID         string   `json:"ticket_id"`
	CustomerAddress  string   `json:"customer_address"`
	ProviderAddress  string   `json:"provider_address,omitempty"`
	ResourceRefType  string   `json:"resource_ref_type,omitempty"`
	ResourceRefID    string   `json:"resource_ref_id,omitempty"`
	ResourceRefOwner string   `json:"resource_ref_owner,omitempty"`
	Status           uint8    `json:"status"`
	Priority         uint8    `json:"priority"`
	Category         string   `json:"category"`
	EncryptedPayload []byte   `json:"encrypted_payload"`
	AssignedTo       string   `json:"assigned_to,omitempty"`
	ResponseCount    uint32   `json:"response_count"`
	CreatedAt        int64    `json:"created_at"`
	UpdatedAt        int64    `json:"updated_at"`
	ResolvedAt       *int64   `json:"resolved_at,omitempty"`
	ClosedAt         *int64   `json:"closed_at,omitempty"`
	ResolutionRef    string   `json:"resolution_ref,omitempty"`
}

// CreateTicket creates a new support ticket
func (k Keeper) CreateTicket(ctx sdk.Context, ticket *types.SupportTicket) error {
	if err := ticket.Validate(); err != nil {
		return err
	}

	// Check if ticket already exists
	if _, found := k.GetTicket(ctx, ticket.TicketID); found {
		return types.ErrTicketAlreadyExists.Wrapf("ticket %s already exists", ticket.TicketID)
	}

	// Validate category is allowed
	params := k.GetParams(ctx)
	if !params.IsCategoryAllowed(ticket.Category) {
		return types.ErrInvalidCategory.Wrapf("category %s is not allowed", ticket.Category)
	}

	// Store the ticket
	if err := k.SetTicket(ctx, ticket); err != nil {
		return err
	}

	// Update indexes
	k.indexTicketByCustomer(ctx, ticket)
	if ticket.ProviderAddress != "" {
		k.indexTicketByProvider(ctx, ticket)
	}
	k.indexTicketByStatus(ctx, ticket)

	return nil
}

// GetTicket returns a ticket by ID
func (k Keeper) GetTicket(ctx sdk.Context, ticketID string) (types.SupportTicket, bool) {
	store := ctx.KVStore(k.skey)
	key := types.TicketKey(ticketID)
	bz := store.Get(key)
	if bz == nil {
		return types.SupportTicket{}, false
	}

	var ts ticketStore
	if err := json.Unmarshal(bz, &ts); err != nil {
		return types.SupportTicket{}, false
	}

	return k.ticketStoreToTicket(ts), true
}

// SetTicket stores a ticket
func (k Keeper) SetTicket(ctx sdk.Context, ticket *types.SupportTicket) error {
	// Serialize encrypted payload
	payloadBz, err := json.Marshal(ticket.EncryptedPayload)
	if err != nil {
		return err
	}

	ts := ticketStore{
		TicketID:         ticket.TicketID,
		CustomerAddress:  ticket.CustomerAddress,
		ProviderAddress:  ticket.ProviderAddress,
		ResourceRefType:  ticket.ResourceRef.Type,
		ResourceRefID:    ticket.ResourceRef.ID,
		ResourceRefOwner: ticket.ResourceRef.Owner,
		Status:           uint8(ticket.Status),
		Priority:         uint8(ticket.Priority),
		Category:         ticket.Category,
		EncryptedPayload: payloadBz,
		AssignedTo:       ticket.AssignedTo,
		ResponseCount:    ticket.ResponseCount,
		CreatedAt:        ticket.CreatedAt.Unix(),
		UpdatedAt:        ticket.UpdatedAt.Unix(),
		ResolutionRef:    ticket.ResolutionRef,
	}

	if ticket.ResolvedAt != nil {
		ts.ResolvedAt = new(int64)
		*ts.ResolvedAt = ticket.ResolvedAt.Unix()
	}
	if ticket.ClosedAt != nil {
		ts.ClosedAt = new(int64)
		*ts.ClosedAt = ticket.ClosedAt.Unix()
	}

	bz, err := json.Marshal(&ts)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.TicketKey(ticket.TicketID), bz)
	return nil
}

// DeleteTicket deletes a ticket and its indexes
func (k Keeper) DeleteTicket(ctx sdk.Context, ticketID string) error {
	ticket, found := k.GetTicket(ctx, ticketID)
	if !found {
		return types.ErrTicketNotFound.Wrapf("ticket %s not found", ticketID)
	}

	store := ctx.KVStore(k.skey)

	// Remove indexes
	customerAddr, _ := sdk.AccAddressFromBech32(ticket.CustomerAddress)
	store.Delete(types.TicketsByCustomerKey(customerAddr.Bytes(), ticketID))

	if ticket.ProviderAddress != "" {
		providerAddr, _ := sdk.AccAddressFromBech32(ticket.ProviderAddress)
		store.Delete(types.TicketsByProviderKey(providerAddr.Bytes(), ticketID))
	}

	if ticket.AssignedTo != "" {
		agentAddr, _ := sdk.AccAddressFromBech32(ticket.AssignedTo)
		store.Delete(types.TicketsByAgentKey(agentAddr.Bytes(), ticketID))
	}

	store.Delete(types.TicketsByStatusKey(ticket.Status, ticketID))

	// Delete the ticket itself
	store.Delete(types.TicketKey(ticketID))

	return nil
}

// ticketStoreToTicket converts a ticketStore to SupportTicket
func (k Keeper) ticketStoreToTicket(ts ticketStore) types.SupportTicket {
	ticket := types.SupportTicket{
		TicketID:        ts.TicketID,
		CustomerAddress: ts.CustomerAddress,
		ProviderAddress: ts.ProviderAddress,
		ResourceRef: types.ResourceReference{
			Type:  ts.ResourceRefType,
			ID:    ts.ResourceRefID,
			Owner: ts.ResourceRefOwner,
		},
		Status:        types.TicketStatus(ts.Status),
		Priority:      types.TicketPriority(ts.Priority),
		Category:      ts.Category,
		AssignedTo:    ts.AssignedTo,
		ResponseCount: ts.ResponseCount,
		CreatedAt:     time.Unix(ts.CreatedAt, 0),
		UpdatedAt:     time.Unix(ts.UpdatedAt, 0),
		ResolutionRef: ts.ResolutionRef,
	}

	// Unmarshal encrypted payload
	if len(ts.EncryptedPayload) > 0 {
		_ = json.Unmarshal(ts.EncryptedPayload, &ticket.EncryptedPayload)
	}

	if ts.ResolvedAt != nil {
		t := time.Unix(*ts.ResolvedAt, 0)
		ticket.ResolvedAt = &t
	}
	if ts.ClosedAt != nil {
		t := time.Unix(*ts.ClosedAt, 0)
		ticket.ClosedAt = &t
	}

	return ticket
}

// ============================================================================
// Ticket Status Transitions
// ============================================================================

// AssignTicket assigns a ticket to a support agent
func (k Keeper) AssignTicket(ctx sdk.Context, ticketID string, agentAddr sdk.AccAddress, assignedBy sdk.AccAddress) error {
	ticket, found := k.GetTicket(ctx, ticketID)
	if !found {
		return types.ErrTicketNotFound.Wrapf("ticket %s not found", ticketID)
	}

	if !ticket.Status.CanTransitionTo(types.TicketStatusAssigned) {
		return types.ErrInvalidTicketStatus.Wrapf("cannot assign ticket in status %s", ticket.Status)
	}

	// Remove old agent index if any
	store := ctx.KVStore(k.skey)
	if ticket.AssignedTo != "" {
		oldAgentAddr, _ := sdk.AccAddressFromBech32(ticket.AssignedTo)
		store.Delete(types.TicketsByAgentKey(oldAgentAddr.Bytes(), ticketID))
	}

	// Remove old status index
	store.Delete(types.TicketsByStatusKey(ticket.Status, ticketID))

	// Update ticket
	ticket.AssignedTo = agentAddr.String()
	ticket.Status = types.TicketStatusAssigned
	ticket.UpdatedAt = ctx.BlockTime()

	if err := k.SetTicket(ctx, &ticket); err != nil {
		return err
	}

	// Add new indexes
	k.indexTicketByAgent(ctx, &ticket)
	k.indexTicketByStatus(ctx, &ticket)

	return nil
}

// ResolveTicket marks a ticket as resolved
func (k Keeper) ResolveTicket(ctx sdk.Context, ticketID string, resolvedBy sdk.AccAddress, resolutionRef string) error {
	ticket, found := k.GetTicket(ctx, ticketID)
	if !found {
		return types.ErrTicketNotFound.Wrapf("ticket %s not found", ticketID)
	}

	if !ticket.Status.CanTransitionTo(types.TicketStatusResolved) {
		return types.ErrInvalidTicketStatus.Wrapf("cannot resolve ticket in status %s", ticket.Status)
	}

	// Remove old status index
	store := ctx.KVStore(k.skey)
	store.Delete(types.TicketsByStatusKey(ticket.Status, ticketID))

	// Update ticket
	now := ctx.BlockTime()
	ticket.Status = types.TicketStatusResolved
	ticket.UpdatedAt = now
	ticket.ResolvedAt = &now
	ticket.ResolutionRef = resolutionRef

	if err := k.SetTicket(ctx, &ticket); err != nil {
		return err
	}

	// Add new status index
	k.indexTicketByStatus(ctx, &ticket)

	return nil
}

// CloseTicket closes a ticket
func (k Keeper) CloseTicket(ctx sdk.Context, ticketID string, closedBy sdk.AccAddress, reason string) error {
	ticket, found := k.GetTicket(ctx, ticketID)
	if !found {
		return types.ErrTicketNotFound.Wrapf("ticket %s not found", ticketID)
	}

	if !ticket.Status.CanTransitionTo(types.TicketStatusClosed) {
		return types.ErrInvalidTicketStatus.Wrapf("cannot close ticket in status %s", ticket.Status)
	}

	// Remove old status index
	store := ctx.KVStore(k.skey)
	store.Delete(types.TicketsByStatusKey(ticket.Status, ticketID))

	// Update ticket
	now := ctx.BlockTime()
	ticket.Status = types.TicketStatusClosed
	ticket.UpdatedAt = now
	ticket.ClosedAt = &now

	if err := k.SetTicket(ctx, &ticket); err != nil {
		return err
	}

	// Add new status index
	k.indexTicketByStatus(ctx, &ticket)

	return nil
}

// ReopenTicket reopens a closed or resolved ticket
func (k Keeper) ReopenTicket(ctx sdk.Context, ticketID string, reopenedBy sdk.AccAddress, reason string) error {
	ticket, found := k.GetTicket(ctx, ticketID)
	if !found {
		return types.ErrTicketNotFound.Wrapf("ticket %s not found", ticketID)
	}

	if !ticket.Status.CanTransitionTo(types.TicketStatusOpen) {
		return types.ErrCannotReopen.Wrapf("cannot reopen ticket in status %s", ticket.Status)
	}

	// Check reopen window
	params := k.GetParams(ctx)
	if ticket.ClosedAt != nil {
		reopenDeadline := ticket.ClosedAt.AddDate(0, 0, int(params.ReopenWindowDays))
		if ctx.BlockTime().After(reopenDeadline) {
			return types.ErrCannotReopen.Wrapf("reopen window of %d days has passed", params.ReopenWindowDays)
		}
	}

	// Remove old status index
	store := ctx.KVStore(k.skey)
	store.Delete(types.TicketsByStatusKey(ticket.Status, ticketID))

	// Update ticket
	ticket.Status = types.TicketStatusOpen
	ticket.UpdatedAt = ctx.BlockTime()
	ticket.ClosedAt = nil
	ticket.ResolvedAt = nil
	ticket.ResolutionRef = ""

	if err := k.SetTicket(ctx, &ticket); err != nil {
		return err
	}

	// Add new status index
	k.indexTicketByStatus(ctx, &ticket)

	return nil
}

// ============================================================================
// Response Management
// ============================================================================

// responseStore is the stored format of a ticket response
type responseStore struct {
	TicketID         string `json:"ticket_id"`
	ResponseIndex    uint32 `json:"response_index"`
	ResponderAddress string `json:"responder_address"`
	IsAgent          bool   `json:"is_agent"`
	EncryptedPayload []byte `json:"encrypted_payload"`
	CreatedAt        int64  `json:"created_at"`
}

// AddResponse adds a response to a ticket
func (k Keeper) AddResponse(ctx sdk.Context, ticketID string, response *types.TicketResponse) error {
	if err := response.Validate(); err != nil {
		return err
	}

	ticket, found := k.GetTicket(ctx, ticketID)
	if !found {
		return types.ErrTicketNotFound.Wrapf("ticket %s not found", ticketID)
	}

	if ticket.Status == types.TicketStatusClosed {
		return types.ErrTicketClosed.Wrapf("cannot respond to closed ticket %s", ticketID)
	}

	// Check max responses
	params := k.GetParams(ctx)
	if ticket.ResponseCount >= params.MaxResponsesPerTicket {
		return types.ErrMaxResponsesExceeded.Wrapf("ticket %s has reached max responses", ticketID)
	}

	// Set response index
	response.ResponseIndex = ticket.ResponseCount

	// Serialize encrypted payload
	payloadBz, err := json.Marshal(response.EncryptedPayload)
	if err != nil {
		return err
	}

	rs := responseStore{
		TicketID:         response.TicketID,
		ResponseIndex:    response.ResponseIndex,
		ResponderAddress: response.ResponderAddress,
		IsAgent:          response.IsAgent,
		EncryptedPayload: payloadBz,
		CreatedAt:        response.CreatedAt.Unix(),
	}

	bz, err := json.Marshal(&rs)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.TicketResponseKey(ticketID, response.ResponseIndex), bz)

	// Update ticket response count and status
	ticket.ResponseCount++
	ticket.UpdatedAt = ctx.BlockTime()

	// If agent responds, mark as in progress
	if response.IsAgent && ticket.Status == types.TicketStatusAssigned {
		store.Delete(types.TicketsByStatusKey(ticket.Status, ticketID))
		ticket.Status = types.TicketStatusInProgress
		k.indexTicketByStatus(ctx, &ticket)
	}

	// If customer responds while pending, mark as in progress
	if !response.IsAgent && ticket.Status == types.TicketStatusPendingCustomer {
		store.Delete(types.TicketsByStatusKey(ticket.Status, ticketID))
		ticket.Status = types.TicketStatusInProgress
		k.indexTicketByStatus(ctx, &ticket)
	}

	return k.SetTicket(ctx, &ticket)
}

// GetResponse returns a specific response
func (k Keeper) GetResponse(ctx sdk.Context, ticketID string, responseIndex uint32) (types.TicketResponse, bool) {
	store := ctx.KVStore(k.skey)
	key := types.TicketResponseKey(ticketID, responseIndex)
	bz := store.Get(key)
	if bz == nil {
		return types.TicketResponse{}, false
	}

	var rs responseStore
	if err := json.Unmarshal(bz, &rs); err != nil {
		return types.TicketResponse{}, false
	}

	return k.responseStoreToResponse(rs), true
}

// GetResponses returns all responses for a ticket
func (k Keeper) GetResponses(ctx sdk.Context, ticketID string) []types.TicketResponse {
	var responses []types.TicketResponse

	store := ctx.KVStore(k.skey)
	prefix := types.TicketResponsePrefixKey(ticketID)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var rs responseStore
		if err := json.Unmarshal(iter.Value(), &rs); err != nil {
			continue
		}
		responses = append(responses, k.responseStoreToResponse(rs))
	}

	return responses
}

// responseStoreToResponse converts a responseStore to TicketResponse
func (k Keeper) responseStoreToResponse(rs responseStore) types.TicketResponse {
	response := types.TicketResponse{
		TicketID:         rs.TicketID,
		ResponseIndex:    rs.ResponseIndex,
		ResponderAddress: rs.ResponderAddress,
		IsAgent:          rs.IsAgent,
		CreatedAt:        time.Unix(rs.CreatedAt, 0),
	}

	// Unmarshal encrypted payload
	if len(rs.EncryptedPayload) > 0 {
		_ = json.Unmarshal(rs.EncryptedPayload, &response.EncryptedPayload)
	}

	return response
}

// ============================================================================
// Indexes
// ============================================================================

func (k Keeper) indexTicketByCustomer(ctx sdk.Context, ticket *types.SupportTicket) {
	customerAddr, _ := sdk.AccAddressFromBech32(ticket.CustomerAddress)
	store := ctx.KVStore(k.skey)
	store.Set(types.TicketsByCustomerKey(customerAddr.Bytes(), ticket.TicketID), []byte{1})
}

func (k Keeper) indexTicketByProvider(ctx sdk.Context, ticket *types.SupportTicket) {
	providerAddr, _ := sdk.AccAddressFromBech32(ticket.ProviderAddress)
	store := ctx.KVStore(k.skey)
	store.Set(types.TicketsByProviderKey(providerAddr.Bytes(), ticket.TicketID), []byte{1})
}

func (k Keeper) indexTicketByAgent(ctx sdk.Context, ticket *types.SupportTicket) {
	agentAddr, _ := sdk.AccAddressFromBech32(ticket.AssignedTo)
	store := ctx.KVStore(k.skey)
	store.Set(types.TicketsByAgentKey(agentAddr.Bytes(), ticket.TicketID), []byte{1})
}

func (k Keeper) indexTicketByStatus(ctx sdk.Context, ticket *types.SupportTicket) {
	store := ctx.KVStore(k.skey)
	store.Set(types.TicketsByStatusKey(ticket.Status, ticket.TicketID), []byte{1})
}

// ============================================================================
// Ticket Queries
// ============================================================================

// GetTicketsByCustomer returns all tickets for a customer
func (k Keeper) GetTicketsByCustomer(ctx sdk.Context, customerAddr sdk.AccAddress) []types.SupportTicket {
	var tickets []types.SupportTicket

	store := ctx.KVStore(k.skey)
	prefix := types.TicketsByCustomerPrefixKey(customerAddr.Bytes())
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		ticketID := string(key[len(prefix):])
		if ticket, found := k.GetTicket(ctx, ticketID); found {
			tickets = append(tickets, ticket)
		}
	}

	return tickets
}

// GetTicketsByProvider returns all tickets related to a provider
func (k Keeper) GetTicketsByProvider(ctx sdk.Context, providerAddr sdk.AccAddress) []types.SupportTicket {
	var tickets []types.SupportTicket

	store := ctx.KVStore(k.skey)
	prefix := types.TicketsByProviderPrefixKey(providerAddr.Bytes())
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		ticketID := string(key[len(prefix):])
		if ticket, found := k.GetTicket(ctx, ticketID); found {
			tickets = append(tickets, ticket)
		}
	}

	return tickets
}

// GetTicketsByAgent returns all tickets assigned to an agent
func (k Keeper) GetTicketsByAgent(ctx sdk.Context, agentAddr sdk.AccAddress) []types.SupportTicket {
	var tickets []types.SupportTicket

	store := ctx.KVStore(k.skey)
	prefix := types.TicketsByAgentPrefixKey(agentAddr.Bytes())
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		ticketID := string(key[len(prefix):])
		if ticket, found := k.GetTicket(ctx, ticketID); found {
			tickets = append(tickets, ticket)
		}
	}

	return tickets
}

// GetTicketsByStatus returns all tickets with a specific status
func (k Keeper) GetTicketsByStatus(ctx sdk.Context, status types.TicketStatus) []types.SupportTicket {
	var tickets []types.SupportTicket

	store := ctx.KVStore(k.skey)
	prefix := types.TicketsByStatusPrefixKey(status)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		ticketID := string(key[len(prefix):])
		if ticket, found := k.GetTicket(ctx, ticketID); found {
			tickets = append(tickets, ticket)
		}
	}

	return tickets
}

// WithTickets iterates over all tickets
func (k Keeper) WithTickets(ctx sdk.Context, fn func(ticket types.SupportTicket) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixTicket)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var ts ticketStore
		if err := json.Unmarshal(iter.Value(), &ts); err != nil {
			continue
		}

		if fn(k.ticketStoreToTicket(ts)) {
			break
		}
	}
}

// ============================================================================
// Authorization Checks
// ============================================================================

// CanViewTicket checks if an address can view a ticket
func (k Keeper) CanViewTicket(ctx sdk.Context, viewer sdk.AccAddress, ticket *types.SupportTicket) bool {
	viewerStr := viewer.String()

	// Customer can always view their own tickets
	if viewerStr == ticket.CustomerAddress {
		return true
	}

	// Provider can view tickets related to them
	if viewerStr == ticket.ProviderAddress {
		return true
	}

	// Assigned agent can view
	if viewerStr == ticket.AssignedTo {
		return true
	}

	// Support agents and admins can view all tickets
	if k.IsSupportAgent(ctx, viewer) || k.IsSupportAdmin(ctx, viewer) {
		return true
	}

	return false
}

// CanRespondToTicket checks if an address can respond to a ticket
func (k Keeper) CanRespondToTicket(ctx sdk.Context, responder sdk.AccAddress, ticket *types.SupportTicket) bool {
	responderStr := responder.String()

	// Customer can respond to their own tickets
	if responderStr == ticket.CustomerAddress {
		return true
	}

	// Assigned agent can respond
	if responderStr == ticket.AssignedTo {
		return true
	}

	// Support admins can respond to any ticket
	if k.IsSupportAdmin(ctx, responder) {
		return true
	}

	return false
}

// CanAssignTicket checks if an address can assign tickets
func (k Keeper) CanAssignTicket(ctx sdk.Context, assigner sdk.AccAddress) bool {
	// Only support admins can assign tickets
	return k.IsSupportAdmin(ctx, assigner)
}

// CanCloseTicket checks if an address can close a ticket
func (k Keeper) CanCloseTicket(ctx sdk.Context, closer sdk.AccAddress, ticket *types.SupportTicket) bool {
	closerStr := closer.String()

	// Customer can close their own tickets
	if closerStr == ticket.CustomerAddress {
		return true
	}

	// Support admins can close any ticket
	if k.IsSupportAdmin(ctx, closer) {
		return true
	}

	return false
}

// IsSupportAgent checks if an address has the SupportAgent role
func (k Keeper) IsSupportAgent(ctx sdk.Context, addr sdk.AccAddress) bool {
	if k.rolesKeeper == nil {
		return false
	}
	return k.rolesKeeper.HasRole(ctx, addr, rolestypes.RoleSupportAgent)
}

// IsSupportAdmin checks if an address has admin privileges for support
func (k Keeper) IsSupportAdmin(ctx sdk.Context, addr sdk.AccAddress) bool {
	if k.rolesKeeper == nil {
		return false
	}
	return k.rolesKeeper.IsAdmin(ctx, addr)
}

// ============================================================================
// Rate Limiting
// ============================================================================

// CheckRateLimit checks if the address has exceeded rate limits
func (k Keeper) CheckRateLimit(ctx sdk.Context, addr sdk.AccAddress) error {
	params := k.GetParams(ctx)

	// Get day timestamp (truncate to day)
	dayTimestamp := ctx.BlockTime().Truncate(24 * time.Hour).Unix()

	store := ctx.KVStore(k.skey)
	key := types.RateLimitKey(addr.Bytes(), dayTimestamp)
	bz := store.Get(key)

	var count uint32
	if bz != nil {
		count = binary.BigEndian.Uint32(bz)
	}

	if count >= params.MaxTicketsPerCustomerPerDay {
		return types.ErrRateLimitExceeded.Wrapf("maximum %d tickets per day exceeded", params.MaxTicketsPerCustomerPerDay)
	}

	// Check max open tickets
	openTickets := 0
	tickets := k.GetTicketsByCustomer(ctx, addr)
	for _, t := range tickets {
		if t.Status.IsActive() {
			openTickets++
		}
	}

	if uint32(openTickets) >= params.MaxOpenTicketsPerCustomer {
		return types.ErrRateLimitExceeded.Wrapf("maximum %d open tickets exceeded", params.MaxOpenTicketsPerCustomer)
	}

	return nil
}

// IncrementRateLimit increments the rate limit counter for an address
func (k Keeper) IncrementRateLimit(ctx sdk.Context, addr sdk.AccAddress) {
	// Get day timestamp (truncate to day)
	dayTimestamp := ctx.BlockTime().Truncate(24 * time.Hour).Unix()

	store := ctx.KVStore(k.skey)
	key := types.RateLimitKey(addr.Bytes(), dayTimestamp)
	bz := store.Get(key)

	var count uint32
	if bz != nil {
		count = binary.BigEndian.Uint32(bz)
	}

	count++
	countBz := make([]byte, 4)
	binary.BigEndian.PutUint32(countBz, count)
	store.Set(key, countBz)
}

// ============================================================================
// Sequence Management
// ============================================================================

// GetNextTicketID generates and returns the next ticket ID
func (k Keeper) GetNextTicketID(ctx sdk.Context) string {
	seq := k.GetTicketSequence(ctx)
	ticketID := fmt.Sprintf("TKT-%08d", seq)
	k.SetTicketSequence(ctx, seq+1)
	return ticketID
}

// GetTicketSequence returns the current ticket sequence
func (k Keeper) GetTicketSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.TicketSequenceKey())
	if bz == nil {
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

// SetTicketSequence sets the ticket sequence
func (k Keeper) SetTicketSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	store.Set(types.TicketSequenceKey(), bz)
}

// NewGRPCQuerier returns a new GRPCQuerier
func (k Keeper) NewGRPCQuerier() GRPCQuerier {
	return GRPCQuerier{Keeper: k}
}

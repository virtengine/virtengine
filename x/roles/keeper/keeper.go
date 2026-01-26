package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.akt.dev/node/x/roles/types"
)

// IKeeper defines the interface for the roles keeper
type IKeeper interface {
	// Role management
	AssignRole(ctx sdk.Context, address sdk.AccAddress, role types.Role, assignedBy sdk.AccAddress) error
	RevokeRole(ctx sdk.Context, address sdk.AccAddress, role types.Role, revokedBy sdk.AccAddress) error
	HasRole(ctx sdk.Context, address sdk.AccAddress, role types.Role) bool
	GetAccountRoles(ctx sdk.Context, address sdk.AccAddress) []types.RoleAssignment
	GetRoleMembers(ctx sdk.Context, role types.Role) []types.RoleAssignment

	// Account state management
	SetAccountState(ctx sdk.Context, address sdk.AccAddress, state types.AccountState, reason string, modifiedBy sdk.AccAddress) error
	GetAccountState(ctx sdk.Context, address sdk.AccAddress) (types.AccountStateRecord, bool)
	IsAccountOperational(ctx sdk.Context, address sdk.AccAddress) bool

	// Genesis account management
	IsGenesisAccount(ctx sdk.Context, address sdk.AccAddress) bool
	AddGenesisAccount(ctx sdk.Context, address sdk.AccAddress) error
	GetGenesisAccounts(ctx sdk.Context) []sdk.AccAddress

	// Authorization checks
	CanAssignRole(ctx sdk.Context, sender sdk.AccAddress, targetRole types.Role) bool
	CanRevokeRole(ctx sdk.Context, sender sdk.AccAddress, targetRole types.Role) bool
	CanModifyAccountState(ctx sdk.Context, sender sdk.AccAddress) bool

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
}

// Keeper of the roles store
type Keeper struct {
	skey storetypes.StoreKey
	cdc  codec.BinaryCodec

	// The address capable of executing a MsgUpdateParams message.
	// This should be the x/gov module account.
	authority string
}

// NewKeeper creates and returns an instance for roles keeper
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

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := k.cdc.Marshal(&paramsStore{
		MaxRolesPerAccount: params.MaxRolesPerAccount,
		AllowSelfRevoke:    params.AllowSelfRevoke,
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
	k.cdc.MustUnmarshal(bz, &ps)
	return types.Params{
		MaxRolesPerAccount: ps.MaxRolesPerAccount,
		AllowSelfRevoke:    ps.AllowSelfRevoke,
	}
}

// paramsStore is the stored format of params
type paramsStore struct {
	MaxRolesPerAccount uint32 `json:"max_roles_per_account"`
	AllowSelfRevoke    bool   `json:"allow_self_revoke"`
}

// AssignRole assigns a role to an account
func (k Keeper) AssignRole(ctx sdk.Context, address sdk.AccAddress, role types.Role, assignedBy sdk.AccAddress) error {
	if !role.IsValid() {
		return types.ErrInvalidRole
	}

	// Check if role is already assigned
	if k.HasRole(ctx, address, role) {
		return types.ErrRoleAlreadyAssigned
	}

	// Check max roles per account
	params := k.GetParams(ctx)
	currentRoles := k.GetAccountRoles(ctx, address)
	if uint32(len(currentRoles)) >= params.MaxRolesPerAccount {
		return types.ErrInvalidRole.Wrapf("account has reached max roles limit: %d", params.MaxRolesPerAccount)
	}

	store := ctx.KVStore(k.skey)

	assignment := roleAssignmentStore{
		AssignedBy: assignedBy.String(),
		AssignedAt: ctx.BlockTime().Unix(),
	}

	bz, err := k.cdc.Marshal(&assignment)
	if err != nil {
		return err
	}

	// Store the role assignment
	key := types.RoleAssignmentKey(address.Bytes(), role)
	store.Set(key, bz)

	// Add to role members index
	memberKey := types.RoleMembersKey(role, address.Bytes())
	store.Set(memberKey, []byte{1})

	// If assigning GenesisAccount role, also add to genesis accounts
	if role == types.RoleGenesisAccount {
		genesisKey := types.GenesisAccountKey(address.Bytes())
		store.Set(genesisKey, []byte{1})
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventRoleAssigned{
		Address:    address.String(),
		Role:       role.String(),
		AssignedBy: assignedBy.String(),
	})
	if err != nil {
		return err
	}

	return nil
}

// RevokeRole revokes a role from an account
func (k Keeper) RevokeRole(ctx sdk.Context, address sdk.AccAddress, role types.Role, revokedBy sdk.AccAddress) error {
	if !role.IsValid() {
		return types.ErrInvalidRole
	}

	// Check if role exists
	if !k.HasRole(ctx, address, role) {
		return types.ErrRoleNotFound
	}

	store := ctx.KVStore(k.skey)

	// Delete the role assignment
	key := types.RoleAssignmentKey(address.Bytes(), role)
	store.Delete(key)

	// Remove from role members index
	memberKey := types.RoleMembersKey(role, address.Bytes())
	store.Delete(memberKey)

	// If revoking GenesisAccount role, also remove from genesis accounts
	if role == types.RoleGenesisAccount {
		genesisKey := types.GenesisAccountKey(address.Bytes())
		store.Delete(genesisKey)
	}

	// Emit event
	err := ctx.EventManager().EmitTypedEvent(&types.EventRoleRevoked{
		Address:   address.String(),
		Role:      role.String(),
		RevokedBy: revokedBy.String(),
	})
	if err != nil {
		return err
	}

	return nil
}

// HasRole checks if an account has a specific role
func (k Keeper) HasRole(ctx sdk.Context, address sdk.AccAddress, role types.Role) bool {
	store := ctx.KVStore(k.skey)
	key := types.RoleAssignmentKey(address.Bytes(), role)
	return store.Has(key)
}

// GetAccountRoles returns all roles assigned to an account
func (k Keeper) GetAccountRoles(ctx sdk.Context, address sdk.AccAddress) []types.RoleAssignment {
	store := ctx.KVStore(k.skey)
	prefix := types.RoleAssignmentPrefixKey(address.Bytes())

	var roles []types.RoleAssignment
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		// Extract role from key (last byte)
		roleValue := types.Role(key[len(key)-1])

		var assignment roleAssignmentStore
		k.cdc.MustUnmarshal(iter.Value(), &assignment)

		roles = append(roles, types.RoleAssignment{
			Address:    address.String(),
			Role:       roleValue,
			AssignedBy: assignment.AssignedBy,
			AssignedAt: assignment.AssignedAt,
		})
	}

	return roles
}

// GetRoleMembers returns all accounts with a specific role
func (k Keeper) GetRoleMembers(ctx sdk.Context, role types.Role) []types.RoleAssignment {
	store := ctx.KVStore(k.skey)
	prefix := types.RoleMembersPrefixKey(role)

	var members []types.RoleAssignment
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		// Extract address from key (after prefix + role byte)
		addrBytes := key[len(prefix):]
		addr := sdk.AccAddress(addrBytes)

		// Get the full role assignment
		assignmentKey := types.RoleAssignmentKey(addrBytes, role)
		assignmentBz := store.Get(assignmentKey)
		if assignmentBz != nil {
			var assignment roleAssignmentStore
			k.cdc.MustUnmarshal(assignmentBz, &assignment)

			members = append(members, types.RoleAssignment{
				Address:    addr.String(),
				Role:       role,
				AssignedBy: assignment.AssignedBy,
				AssignedAt: assignment.AssignedAt,
			})
		}
	}

	return members
}

// roleAssignmentStore is the stored format of a role assignment
type roleAssignmentStore struct {
	AssignedBy string `json:"assigned_by"`
	AssignedAt int64  `json:"assigned_at"`
}

// SetAccountState sets the state of an account
func (k Keeper) SetAccountState(ctx sdk.Context, address sdk.AccAddress, state types.AccountState, reason string, modifiedBy sdk.AccAddress) error {
	if !state.IsValid() {
		return types.ErrInvalidAccountState
	}

	store := ctx.KVStore(k.skey)
	key := types.AccountStateKey(address.Bytes())

	var previousState types.AccountState
	existingBz := store.Get(key)
	if existingBz != nil {
		var existing accountStateStore
		k.cdc.MustUnmarshal(existingBz, &existing)
		previousState = types.AccountState(existing.State)

		// Check if transition is allowed
		if !previousState.CanTransitionTo(state) {
			return types.ErrInvalidStateTransition.Wrapf(
				"cannot transition from %s to %s",
				previousState.String(),
				state.String(),
			)
		}
	} else {
		previousState = types.AccountStateUnspecified
	}

	record := accountStateStore{
		State:         uint8(state),
		Reason:        reason,
		ModifiedBy:    modifiedBy.String(),
		ModifiedAt:    ctx.BlockTime().Unix(),
		PreviousState: uint8(previousState),
	}

	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventAccountStateChanged{
		Address:       address.String(),
		PreviousState: previousState.String(),
		NewState:      state.String(),
		ModifiedBy:    modifiedBy.String(),
		Reason:        reason,
	})
	if err != nil {
		return err
	}

	return nil
}

// GetAccountState returns the state of an account
func (k Keeper) GetAccountState(ctx sdk.Context, address sdk.AccAddress) (types.AccountStateRecord, bool) {
	store := ctx.KVStore(k.skey)
	key := types.AccountStateKey(address.Bytes())

	bz := store.Get(key)
	if bz == nil {
		return types.AccountStateRecord{}, false
	}

	var record accountStateStore
	k.cdc.MustUnmarshal(bz, &record)

	return types.AccountStateRecord{
		Address:       address.String(),
		State:         types.AccountState(record.State),
		Reason:        record.Reason,
		ModifiedBy:    record.ModifiedBy,
		ModifiedAt:    record.ModifiedAt,
		PreviousState: types.AccountState(record.PreviousState),
	}, true
}

// IsAccountOperational checks if an account can perform normal operations
func (k Keeper) IsAccountOperational(ctx sdk.Context, address sdk.AccAddress) bool {
	record, found := k.GetAccountState(ctx, address)
	if !found {
		// Default to operational for new accounts
		return true
	}
	return record.State.IsOperational()
}

// accountStateStore is the stored format of an account state
type accountStateStore struct {
	State         uint8  `json:"state"`
	Reason        string `json:"reason"`
	ModifiedBy    string `json:"modified_by"`
	ModifiedAt    int64  `json:"modified_at"`
	PreviousState uint8  `json:"previous_state"`
}

// IsGenesisAccount checks if an account is a genesis account
func (k Keeper) IsGenesisAccount(ctx sdk.Context, address sdk.AccAddress) bool {
	store := ctx.KVStore(k.skey)
	key := types.GenesisAccountKey(address.Bytes())
	return store.Has(key)
}

// AddGenesisAccount adds an account as a genesis account
func (k Keeper) AddGenesisAccount(ctx sdk.Context, address sdk.AccAddress) error {
	store := ctx.KVStore(k.skey)
	key := types.GenesisAccountKey(address.Bytes())
	store.Set(key, []byte{1})

	// Also assign the GenesisAccount role
	return k.AssignRole(ctx, address, types.RoleGenesisAccount, address)
}

// GetGenesisAccounts returns all genesis accounts
func (k Keeper) GetGenesisAccounts(ctx sdk.Context) []sdk.AccAddress {
	store := ctx.KVStore(k.skey)
	prefix := types.PrefixGenesisAccount

	var accounts []sdk.AccAddress
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		addrBytes := key[len(prefix):]
		accounts = append(accounts, sdk.AccAddress(addrBytes))
	}

	return accounts
}

// CanAssignRole checks if a sender can assign a specific role
func (k Keeper) CanAssignRole(ctx sdk.Context, sender sdk.AccAddress, targetRole types.Role) bool {
	// GenesisAccount can assign any role
	if k.IsGenesisAccount(ctx, sender) {
		return true
	}

	// Check if sender is an administrator
	if k.HasRole(ctx, sender, types.RoleAdministrator) {
		return types.RoleAdministrator.CanAssignRole(targetRole)
	}

	return false
}

// CanRevokeRole checks if a sender can revoke a specific role
func (k Keeper) CanRevokeRole(ctx sdk.Context, sender sdk.AccAddress, targetRole types.Role) bool {
	// Same logic as assign
	return k.CanAssignRole(ctx, sender, targetRole)
}

// CanModifyAccountState checks if a sender can modify account states
func (k Keeper) CanModifyAccountState(ctx sdk.Context, sender sdk.AccAddress) bool {
	// GenesisAccount can always modify
	if k.IsGenesisAccount(ctx, sender) {
		return true
	}

	// Administrators can modify
	if k.HasRole(ctx, sender, types.RoleAdministrator) {
		return true
	}

	return false
}

// NewQuerier returns a new Querier
func (k Keeper) NewQuerier() Querier {
	return Querier{Keeper: k}
}

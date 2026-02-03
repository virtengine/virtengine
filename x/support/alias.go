package support

import (
	"github.com/virtengine/virtengine/x/support/keeper"
	types "github.com/virtengine/virtengine/x/support/types" //nolint:staticcheck // deprecated types retained for compatibility
)

const (
	// ModuleName is the module name
	ModuleName = types.ModuleName

	// StoreKey is the store key
	StoreKey = types.StoreKey

	// RouterKey is the router key
	RouterKey = types.RouterKey

	// QuerierRoute is the querier route
	QuerierRoute = types.QuerierRoute
)

var (
	// DefaultGenesisState returns the default genesis state
	DefaultGenesisState = types.DefaultGenesisState

	// NewKeeper creates a new keeper
	NewKeeper = keeper.NewKeeper

	// NewMsgServerImpl creates a new message server
	NewMsgServerImpl = keeper.NewMsgServerImpl
)

// Type aliases for types package
type (
	// GenesisState is the genesis state for the support module
	GenesisState = types.GenesisState

	// Params defines the module parameters
	Params = types.Params

	// ExternalTicketRef represents a reference to an external support ticket
	ExternalTicketRef = types.ExternalTicketRef

	// ResourceType represents the type of on-chain resource
	ResourceType = types.ResourceType

	// ExternalSystem represents the external service desk system
	ExternalSystem = types.ExternalSystem

	// MsgRegisterExternalTicket registers an external ticket reference
	MsgRegisterExternalTicket = types.MsgRegisterExternalTicket

	// MsgUpdateExternalTicket updates an external ticket reference
	MsgUpdateExternalTicket = types.MsgUpdateExternalTicket

	// MsgRemoveExternalTicket removes an external ticket reference
	MsgRemoveExternalTicket = types.MsgRemoveExternalTicket

	// MsgUpdateParams updates params
	MsgUpdateParams = types.MsgUpdateParams

	// Keeper is the support keeper
	Keeper = keeper.Keeper
)

// Resource type constants
const (
	ResourceTypeDeployment = types.ResourceTypeDeployment
	ResourceTypeLease      = types.ResourceTypeLease
	ResourceTypeOrder      = types.ResourceTypeOrder
	ResourceTypeProvider   = types.ResourceTypeProvider
)

// External system constants
const (
	ExternalSystemWaldur = types.ExternalSystemWaldur
	ExternalSystemJira   = types.ExternalSystemJira
)

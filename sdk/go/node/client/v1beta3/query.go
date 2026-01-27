package v1beta3

import (
	evdtypes "cosmossdk.io/x/evidence/types"
	feegranttypes "cosmossdk.io/x/feegrant"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	staketypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	atypes "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
	btypes "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
	ctypes "github.com/virtengine/virtengine/sdk/go/node/cert/v1"
	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/v1"
	mtypes "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	otypes "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
	ptypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

var _ QueryClient = (*queryClient)(nil)

type sdkQueryClient struct {
	auth      authtypes.QueryClient
	authz     authz.QueryClient
	bank      banktypes.QueryClient
	distr     disttypes.QueryClient
	evidence  evdtypes.QueryClient
	feegrant  feegranttypes.QueryClient
	govLegacy govtypes.QueryClient
	gov       v1.QueryClient
	mint      minttypes.QueryClient
	slashing  slashtypes.QueryClient
	staking   staketypes.QueryClient
	upgrade   upgradetypes.QueryClient
	params    paramstypes.QueryClient
	wasm      wasmtypes.QueryClient
	oracle    otypes.QueryClient
	bme       btypes.QueryClient
}

type queryClient struct {
	dclient dtypes.QueryClient
	eclient etypes.QueryClient
	mclient mtypes.QueryClient
	pclient ptypes.QueryClient
	aclient atypes.QueryClient
	cclient ctypes.QueryClient
	sdk     sdkQueryClient
	cctx    sdkclient.Context
}

// NewQueryClient creates new query client instance based on a Cosmos SDK client context.
func NewQueryClient(cctx sdkclient.Context) QueryClient {
	return newQueryClient(cctx)
}

func newQueryClient(cctx sdkclient.Context) *queryClient {
	return &queryClient{
		dclient: dtypes.NewQueryClient(cctx),
		eclient: etypes.NewQueryClient(cctx),
		mclient: mtypes.NewQueryClient(cctx),
		pclient: ptypes.NewQueryClient(cctx),
		aclient: atypes.NewQueryClient(cctx),
		cclient: ctypes.NewQueryClient(cctx),
		sdk: sdkQueryClient{
			auth:      authtypes.NewQueryClient(cctx),
			authz:     authz.NewQueryClient(cctx),
			bank:      banktypes.NewQueryClient(cctx),
			distr:     disttypes.NewQueryClient(cctx),
			evidence:  evdtypes.NewQueryClient(cctx),
			feegrant:  feegranttypes.NewQueryClient(cctx),
			govLegacy: govtypes.NewQueryClient(cctx),
			gov:       v1.NewQueryClient(cctx),
			mint:      minttypes.NewQueryClient(cctx),
			slashing:  slashtypes.NewQueryClient(cctx),
			staking:   staketypes.NewQueryClient(cctx),
			upgrade:   upgradetypes.NewQueryClient(cctx),
			params:    paramstypes.NewQueryClient(cctx),
			wasm:      wasmtypes.NewQueryClient(cctx),
			oracle:    otypes.NewQueryClient(cctx),
			bme:       btypes.NewQueryClient(cctx),
		},
		cctx: cctx,
	}
}

// ClientContext returns the client's Cosmos SDK client context.
func (c *queryClient) ClientContext() sdkclient.Context {
	return c.cctx
}

func (c *queryClient) Deployment() dtypes.QueryClient {
	return c.dclient
}

func (c *queryClient) Escrow() etypes.QueryClient {
	return c.eclient
}

func (c *queryClient) Market() mtypes.QueryClient {
	return c.mclient
}

func (c *queryClient) Provider() ptypes.QueryClient {
	return c.pclient
}

func (c *queryClient) Audit() atypes.QueryClient {
	return c.aclient
}

// Certs implements QueryClient by returning the certs Akash SDK query client.
func (c *queryClient) Certs() ctypes.QueryClient {
	return c.cclient
}

// Auth implements QueryClient by returning the auth Cosmos SDK query client.
func (c *queryClient) Auth() authtypes.QueryClient {
	return c.sdk.auth
}

// Authz implements QueryClient by returning the authz Cosmos SDK query client.
func (c *queryClient) Authz() authz.QueryClient {
	return c.sdk.authz
}

// Bank implements QueryClient by returning the bank Cosmos SDK query client.
func (c *queryClient) Bank() banktypes.QueryClient {
	return c.sdk.bank
}

// Distribution implements QueryClient by returning the distribution Cosmos SDK query client.
func (c *queryClient) Distribution() disttypes.QueryClient {
	return c.sdk.distr
}

// Evidence implements QueryClient by returning the evidence Cosmos SDK query client.
func (c *queryClient) Evidence() evdtypes.QueryClient {
	return c.sdk.evidence
}

// Feegrant implements QueryClient by returning the feegrant Cosmos SDK query client.
func (c *queryClient) Feegrant() feegranttypes.QueryClient {
	return c.sdk.feegrant
}

// Gov implements QueryClient by returning the governance Cosmos SDK query client.
func (c *queryClient) Gov() v1.QueryClient {
	return c.sdk.gov
}

// GovLegacy implements QueryClient by returning the governance Cosmos SDK query client.
func (c *queryClient) GovLegacy() govtypes.QueryClient {
	return c.sdk.govLegacy
}

// Mint implements QueryClient by returning the mint Cosmos SDK query client.
func (c *queryClient) Mint() minttypes.QueryClient {
	return c.sdk.mint
}

// Slashing implements QueryClient by returning the slashing Cosmos SDK query client.
func (c *queryClient) Slashing() slashtypes.QueryClient {
	return c.sdk.slashing
}

// Staking implements QueryClient by returning the staking Cosmos SDK query client.
func (c *queryClient) Staking() staketypes.QueryClient {
	return c.sdk.staking
}

// Upgrade implements QueryClient by returning the upgrade Cosmos SDK query client.
func (c *queryClient) Upgrade() upgradetypes.QueryClient {
	return c.sdk.upgrade
}

// Params implements QueryClient by returning the params Cosmos SDK query client.
func (c *queryClient) Params() paramstypes.QueryClient {
	return c.sdk.params
}

// Wasm implements QueryClient by returning the wasm query client.
func (c *queryClient) Wasm() wasmtypes.QueryClient {
	return c.sdk.wasm
}

// Oracle implements QueryClient by returning the oracle query client.
func (c *queryClient) Oracle() otypes.QueryClient {
	return c.sdk.oracle
}

// BME implements QueryClient by returning the bme query client.
func (c *queryClient) BME() btypes.QueryClient {
	return c.sdk.bme
}

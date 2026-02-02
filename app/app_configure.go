package app

import (
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibchost "github.com/cosmos/ibc-go/v10/modules/core/exported"

	audittypes "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
	taketypes "github.com/virtengine/virtengine/sdk/go/node/take/v1"

	bmetypes "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
	"github.com/virtengine/virtengine/x/audit"
	"github.com/virtengine/virtengine/x/benchmark"
	benchmarktypes "github.com/virtengine/virtengine/x/benchmark/types"
	"github.com/virtengine/virtengine/x/bme"
	"github.com/virtengine/virtengine/x/cert"
	"github.com/virtengine/virtengine/x/config"
	configtypes "github.com/virtengine/virtengine/x/config/types"
	"github.com/virtengine/virtengine/x/delegation"
	delegationtypes "github.com/virtengine/virtengine/x/delegation/types"
	"github.com/virtengine/virtengine/x/deployment"
	"github.com/virtengine/virtengine/x/enclave"
	enclavetypes "github.com/virtengine/virtengine/x/enclave/types"
	"github.com/virtengine/virtengine/x/encryption"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/escrow"
	"github.com/virtengine/virtengine/x/fraud"
	fraudtypes "github.com/virtengine/virtengine/x/fraud/types"
	"github.com/virtengine/virtengine/x/hpc"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	"github.com/virtengine/virtengine/x/market"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
	"github.com/virtengine/virtengine/x/marketplace"
	"github.com/virtengine/virtengine/x/mfa"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	"github.com/virtengine/virtengine/x/oracle"
	oracletypes "github.com/virtengine/virtengine/x/oracle/types"
	"github.com/virtengine/virtengine/x/provider"
	"github.com/virtengine/virtengine/x/review"
	reviewtypes "github.com/virtengine/virtengine/x/review/types"
	"github.com/virtengine/virtengine/x/roles"
	rolestypes "github.com/virtengine/virtengine/x/roles/types"
	"github.com/virtengine/virtengine/x/settlement"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
	virtstaking "github.com/virtengine/virtengine/x/staking"
	virtstakingtypes "github.com/virtengine/virtengine/x/staking/types"
	"github.com/virtengine/virtengine/x/take"
	"github.com/virtengine/virtengine/x/veid"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

func virtengineModuleBasics() []module.AppModuleBasic {
	return []module.AppModuleBasic{
		take.AppModuleBasic{},
		escrow.AppModuleBasic{},
		deployment.AppModuleBasic{},
		market.AppModuleBasic{},
		marketplace.AppModuleBasic{},
		provider.AppModuleBasic{},
		audit.AppModuleBasic{},
		cert.AppModuleBasic{},
		// VirtEngine patent modules (AU2024203136A1)
		encryption.AppModuleBasic{},
		roles.AppModuleBasic{},
		veid.AppModuleBasic{},
		mfa.AppModuleBasic{},
		config.AppModuleBasic{},
		hpc.AppModuleBasic{},
		benchmark.AppModuleBasic{},
		enclave.AppModuleBasic{},
		settlement.AppModuleBasic{},
		fraud.AppModuleBasic{},
		review.AppModuleBasic{},
		delegation.AppModuleBasic{},
		virtstaking.AppModuleBasic{},
		bme.AppModuleBasic{},
		oracle.AppModuleBasic{},
	}
}

// OrderInitGenesis returns module names in order for init genesis calls.
// NOTE: The genutils module must occur after staking so that pools are
// properly initialized with tokens from genesis accounts.
// NOTE: Capability module must occur first so that it can initialize any capabilities
// so that other modules that want to create or claim capabilities afterwards in InitChain
// can do so safely.
func OrderInitGenesis(_ []string) []string {
	return []string{
		authtypes.ModuleName,
		authz.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		vestingtypes.ModuleName,
		paramstypes.ModuleName,
		audittypes.ModuleName,
		upgradetypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		ibchost.ModuleName,
		evidencetypes.ModuleName,
		transfertypes.ModuleName,
		consensustypes.ModuleName,
		feegrant.ModuleName,
		cert.ModuleName,
		taketypes.ModuleName,
		escrow.ModuleName,
		deployment.ModuleName,
		provider.ModuleName,
		market.ModuleName,
		marketplacetypes.ModuleName,
		// VirtEngine patent modules (AU2024203136A1)
		encryptiontypes.ModuleName,
		rolestypes.ModuleName,
		veidtypes.ModuleName,
		mfatypes.ModuleName,
		configtypes.ModuleName,
		hpctypes.ModuleName,
		benchmarktypes.ModuleName,
		enclavetypes.ModuleName,
		settlementtypes.ModuleName,
		fraudtypes.ModuleName,
		reviewtypes.ModuleName,
		delegationtypes.ModuleName,
		virtstakingtypes.ModuleName,
		bmetypes.ModuleName,
		oracletypes.ModuleName,
		genutiltypes.ModuleName,
	}
}

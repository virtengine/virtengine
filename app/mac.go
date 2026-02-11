package app

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	emodule "github.com/virtengine/virtengine/sdk/go/node/escrow/module"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
	virtstaking "github.com/virtengine/virtengine/x/staking/types"
)

func ModuleAccountPerms() map[string][]string {
	return map[string][]string{
		authtypes.FeeCollectorName:     nil,
		emodule.ModuleName:             nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		virtstaking.ModuleName:         {authtypes.Minter}, // virt_staking needs Minter for epoch rewards
		settlementtypes.ModuleName:     nil,
	}
}

func ModuleAccountAddrs() map[string]bool {
	perms := ModuleAccountPerms()
	addrs := make(map[string]bool, len(perms))
	for k := range perms {
		addrs[authtypes.NewModuleAddress(k).String()] = true
	}
	return addrs
}

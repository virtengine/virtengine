package provider

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/virtengine/virtengine/provider/bidengine"
	mparams "github.com/virtengine/virtengine/x/market/types"
	mtypes "github.com/virtengine/virtengine/x/market/types"
)

type Config struct {
	ClusterWaitReadyDuration        time.Duration
	ClusterPublicHostname           string
	ClusterExternalPortQuantity     uint
	InventoryResourcePollPeriod     time.Duration
	InventoryResourceDebugFrequency uint
	BidPricingStrategy              bidengine.BidPricingStrategy
	BidDeposit                      sdk.Coin
	CPUCommitLevel                  float64
	MemoryCommitLevel               float64
	StorageCommitLevel              float64
	BlockedHostnames                []string
	DeploymentIngressStaticHosts    bool
	BidTimeout                      time.Duration
	ManifestTimeout                 time.Duration

	BalanceCheckerCfg BalanceCheckerConfig
}

func NewDefaultConfig() Config {
	return Config{
		ClusterWaitReadyDuration: time.Second * 5,
		BidDeposit:               mtypes.DefaultBidMinDeposit,
		BalanceCheckerCfg: BalanceCheckerConfig{
			PollingPeriod:           5 * time.Minute,
			MinimumBalanceThreshold: mparams.DefaultBidMinDeposit.Amount.Mul(sdk.NewIntFromUint64(2)).Uint64(),
			WithdrawalPeriod:        24 * time.Hour,
		},
	}
}

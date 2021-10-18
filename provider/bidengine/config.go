package bidengine

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/virtengine/virtengine/types"
)

type Config struct {
	PricingStrategy BidPricingStrategy
	Deposit         sdk.Coin
	BidTimeout      time.Duration
	Attributes      types.Attributes
}

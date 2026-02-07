package treasury

import (
	"errors"
	"time"

	sdkmath "cosmossdk.io/math"
)

var (
	ErrNoAdapters           = errors.New("no exchange adapters available")
	ErrQuoteUnavailable     = errors.New("no valid quotes available")
	ErrExecutionUnavailable = errors.New("exchange execution unavailable")
	ErrPolicyBlocked        = errors.New("withdrawal blocked by policy")
	ErrApprovalRequired     = errors.New("withdrawal requires approval")
	ErrWithdrawalNotFound   = errors.New("withdrawal request not found")
)

type AdapterType string

const (
	AdapterTypeDEX AdapterType = "dex"
	AdapterTypeCEX AdapterType = "cex"
)

type ExchangeRequest struct {
	FromAsset          string
	ToAsset            string
	Amount             sdkmath.Int
	SlippageBps        int64
	Deadline           time.Time
	ClientRequestID    string
	PreferredTypeOrder []AdapterType
}

type ExchangeQuote struct {
	ID           string
	AdapterName  string
	AdapterType  AdapterType
	FromAsset    string
	ToAsset      string
	InputAmount  sdkmath.Int
	OutputAmount sdkmath.Int
	FeeAmount    sdkmath.Int
	FeeAsset     string
	SlippageBps  int64
	ExpiresAt    time.Time
	QuotedAt     time.Time
}

type ExchangeExecution struct {
	Quote       ExchangeQuote
	TxID        string
	FilledInput sdkmath.Int
	FilledOut   sdkmath.Int
	ExecutedAt  time.Time
}

type BestExecutionPolicy struct {
	MaxSlippageBps   int64
	MaxFeeBps        int64
	PreferTypeOrder  []AdapterType
	RequireHealthy   bool
	AllowExpired     bool
	MinOutputAmount  sdkmath.Int
	MaxExecutionTime time.Duration
}

func DefaultBestExecutionPolicy() BestExecutionPolicy {
	return BestExecutionPolicy{
		MaxSlippageBps:   75,
		MaxFeeBps:        50,
		PreferTypeOrder:  []AdapterType{AdapterTypeDEX, AdapterTypeCEX},
		RequireHealthy:   true,
		AllowExpired:     false,
		MinOutputAmount:  sdkmath.ZeroInt(),
		MaxExecutionTime: 5 * time.Second,
	}
}

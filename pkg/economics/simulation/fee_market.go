package simulation

import (
	"math/big"
	"sort"
	"time"

	"github.com/virtengine/virtengine/pkg/economics"
)

// FeeMarketSimulator simulates fee market dynamics.
type FeeMarketSimulator struct {
	params economics.TokenomicsParams
}

// NewFeeMarketSimulator creates a new fee market simulator.
func NewFeeMarketSimulator(params economics.TokenomicsParams) *FeeMarketSimulator {
	return &FeeMarketSimulator{params: params}
}

// Transaction represents a simulated transaction for fee analysis.
type Transaction struct {
	Hash        string   `json:"hash"`
	GasUsed     int64    `json:"gas_used"`
	GasPrice    int64    `json:"gas_price"`
	FeePaid     *big.Int `json:"fee_paid"`
	TxType      string   `json:"tx_type"`
	IsSpam      bool     `json:"is_spam"`
	Timestamp   time.Time `json:"timestamp"`
}

// FeeMarketSnapshot is a point-in-time snapshot of fee market state.
type FeeMarketSnapshot struct {
	BlockHeight     int64    `json:"block_height"`
	TotalTxs        int64    `json:"total_txs"`
	TotalGasUsed    int64    `json:"total_gas_used"`
	TotalFees       *big.Int `json:"total_fees"`
	AvgGasPrice     int64    `json:"avg_gas_price"`
	MedianGasPrice  int64    `json:"median_gas_price"`
	MaxGasPrice     int64    `json:"max_gas_price"`
	MinGasPrice     int64    `json:"min_gas_price"`
	BlockUtilization int64   `json:"block_utilization_bps"` // basis points
	SpamTxCount     int64    `json:"spam_tx_count"`
}

// SimulateFeeMarket simulates fee market dynamics over a period.
func (s *FeeMarketSimulator) SimulateFeeMarket(
	initialState economics.NetworkState,
	txsPerBlock [][]Transaction,
	takeRateBPS int64,
) FeeMarketSimulationResult {
	result := FeeMarketSimulationResult{
		InitialState: initialState,
		Snapshots:    make([]FeeMarketSnapshot, 0, len(txsPerBlock)),
		TakeRateBPS:  takeRateBPS,
	}

	totalFees := big.NewInt(0)
	totalGasUsed := int64(0)
	totalTxs := int64(0)
	totalSpamTxs := int64(0)
	allGasPrices := make([]int64, 0)

	for blockIdx, txs := range txsPerBlock {
		snapshot := s.processBlock(int64(blockIdx)+initialState.BlockHeight, txs)
		result.Snapshots = append(result.Snapshots, snapshot)

		totalFees.Add(totalFees, snapshot.TotalFees)
		totalGasUsed += snapshot.TotalGasUsed
		totalTxs += snapshot.TotalTxs
		totalSpamTxs += snapshot.SpamTxCount

		for _, tx := range txs {
			allGasPrices = append(allGasPrices, tx.GasPrice)
		}
	}

	// Calculate protocol revenue (take rate)
	protocolRevenue := new(big.Int).Mul(totalFees, big.NewInt(takeRateBPS))
	protocolRevenue.Div(protocolRevenue, big.NewInt(10000))

	// Calculate validator revenue
	validatorRevenue := new(big.Int).Sub(totalFees, protocolRevenue)

	// Calculate average and median gas prices
	var avgGasPrice, medianGasPrice int64
	if len(allGasPrices) > 0 {
		sum := int64(0)
		for _, p := range allGasPrices {
			sum += p
		}
		avgGasPrice = sum / int64(len(allGasPrices))

		sort.Slice(allGasPrices, func(i, j int) bool {
			return allGasPrices[i] < allGasPrices[j]
		})
		medianGasPrice = allGasPrices[len(allGasPrices)/2]
	}

	// Calculate fee volatility (coefficient of variation)
	var feeVolatility float64
	if avgGasPrice > 0 && len(allGasPrices) > 1 {
		variance := int64(0)
		for _, p := range allGasPrices {
			diff := p - avgGasPrice
			variance += diff * diff
		}
		variance /= int64(len(allGasPrices) - 1)
		// Volatility = stddev / mean (simplified)
		feeVolatility = float64(isqrt(variance)) / float64(avgGasPrice)
	}

	// Calculate spam resistance score (0-100)
	spamResistance := s.calculateSpamResistance(totalTxs, totalSpamTxs, avgGasPrice)

	result.TotalFeesCollected = totalFees
	result.ProtocolRevenue = protocolRevenue
	result.ValidatorRevenue = validatorRevenue
	result.TotalTxCount = totalTxs
	result.TotalGasUsed = totalGasUsed
	result.AvgGasPrice = avgGasPrice
	result.MedianGasPrice = medianGasPrice
	result.FeeVolatility = feeVolatility
	result.SpamTxCount = totalSpamTxs
	result.SpamResistanceScore = spamResistance

	// Generate recommendations
	result.Recommendations = s.generateFeeRecommendations(result)

	return result
}

// FeeMarketSimulationResult contains fee market simulation results.
type FeeMarketSimulationResult struct {
	InitialState         economics.NetworkState     `json:"initial_state"`
	TakeRateBPS          int64                      `json:"take_rate_bps"`
	Snapshots            []FeeMarketSnapshot        `json:"snapshots"`
	TotalFeesCollected   *big.Int                   `json:"total_fees_collected"`
	ProtocolRevenue      *big.Int                   `json:"protocol_revenue"`
	ValidatorRevenue     *big.Int                   `json:"validator_revenue"`
	TotalTxCount         int64                      `json:"total_tx_count"`
	TotalGasUsed         int64                      `json:"total_gas_used"`
	AvgGasPrice          int64                      `json:"avg_gas_price"`
	MedianGasPrice       int64                      `json:"median_gas_price"`
	FeeVolatility        float64                    `json:"fee_volatility"`
	SpamTxCount          int64                      `json:"spam_tx_count"`
	SpamResistanceScore  int64                      `json:"spam_resistance_score"`
	Recommendations      []economics.Recommendation `json:"recommendations"`
}

// processBlock processes transactions in a block.
func (s *FeeMarketSimulator) processBlock(blockHeight int64, txs []Transaction) FeeMarketSnapshot {
	snapshot := FeeMarketSnapshot{
		BlockHeight:  blockHeight,
		TotalTxs:     int64(len(txs)),
		TotalFees:    big.NewInt(0),
		MinGasPrice:  int64(^uint64(0) >> 1), // Max int64
	}

	if len(txs) == 0 {
		snapshot.MinGasPrice = 0
		return snapshot
	}

	gasPrices := make([]int64, 0, len(txs))

	for _, tx := range txs {
		snapshot.TotalGasUsed += tx.GasUsed
		snapshot.TotalFees.Add(snapshot.TotalFees, tx.FeePaid)

		gasPrices = append(gasPrices, tx.GasPrice)

		if tx.GasPrice > snapshot.MaxGasPrice {
			snapshot.MaxGasPrice = tx.GasPrice
		}
		if tx.GasPrice < snapshot.MinGasPrice {
			snapshot.MinGasPrice = tx.GasPrice
		}

		if tx.IsSpam {
			snapshot.SpamTxCount++
		}
	}

	// Calculate average and median
	sum := int64(0)
	for _, p := range gasPrices {
		sum += p
	}
	snapshot.AvgGasPrice = sum / int64(len(gasPrices))

	sort.Slice(gasPrices, func(i, j int) bool {
		return gasPrices[i] < gasPrices[j]
	})
	snapshot.MedianGasPrice = gasPrices[len(gasPrices)/2]

	// Block utilization (assuming 10M gas limit per block)
	const blockGasLimit = 10000000
	snapshot.BlockUtilization = (snapshot.TotalGasUsed * 10000) / blockGasLimit

	return snapshot
}

// calculateSpamResistance calculates spam resistance score (0-100).
func (s *FeeMarketSimulator) calculateSpamResistance(totalTxs, spamTxs, avgGasPrice int64) int64 {
	score := int64(100)

	// Penalize high spam ratio
	if totalTxs > 0 {
		spamRatio := (spamTxs * 100) / totalTxs
		score -= spamRatio
	}

	// Reward high gas prices (makes spam expensive)
	// Assuming 100 is minimum, 1000 is good
	if avgGasPrice < s.params.MinGasPrice {
		score -= 20
	} else if avgGasPrice < s.params.MinGasPrice*2 {
		score -= 10
	}

	if score < 0 {
		score = 0
	}

	return score
}

// generateFeeRecommendations generates fee market recommendations.
func (s *FeeMarketSimulator) generateFeeRecommendations(result FeeMarketSimulationResult) []economics.Recommendation {
	var recommendations []economics.Recommendation

	// High fee volatility
	if result.FeeVolatility > 0.5 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "fee_market",
			Priority:    "medium",
			Title:       "High Fee Volatility",
			Description: "Gas price volatility is high, causing unpredictable transaction costs.",
			Impact:      "Poor user experience and difficulty in cost estimation.",
			Action:      "Consider implementing EIP-1559 style base fee with predictable adjustments.",
		})
	}

	// Low spam resistance
	if result.SpamResistanceScore < 50 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "security",
			Priority:    "high",
			Title:       "Low Spam Resistance",
			Description: "The network is vulnerable to spam attacks due to low gas prices.",
			Impact:      "Network congestion and degraded performance during spam attacks.",
			Action:      "Increase minimum gas price or implement priority fee mechanism.",
		})
	}

	// Take rate analysis
	if result.TakeRateBPS > 1000 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "fee_market",
			Priority:    "low",
			Title:       "High Take Rate",
			Description: "Protocol take rate exceeds 10%, which may discourage usage.",
			Impact:      "Reduced transaction volume and validator earnings.",
			Action:      "Consider reducing take rate to increase network competitiveness.",
		})
	} else if result.TakeRateBPS < 100 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "fee_market",
			Priority:    "medium",
			Title:       "Low Take Rate",
			Description: "Protocol take rate is below 1%, limiting protocol revenue.",
			Impact:      "Reduced funding for protocol development and community pool.",
			Action:      "Consider increasing take rate to fund ongoing development.",
		})
	}

	return recommendations
}

// AnalyzeFeeMarket provides detailed fee market analysis.
func (s *FeeMarketSimulator) AnalyzeFeeMarket(
	state economics.NetworkState,
	historicalFees []int64,
) economics.FeeMarketAnalysis {
	if len(historicalFees) == 0 {
		return economics.FeeMarketAnalysis{
			SpamResistance: 0,
		}
	}

	// Calculate average fee
	sum := int64(0)
	for _, f := range historicalFees {
		sum += f
	}
	avgFee := sum / int64(len(historicalFees))

	// Calculate median fee
	sorted := make([]int64, len(historicalFees))
	copy(sorted, historicalFees)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	medianFee := sorted[len(sorted)/2]

	// Calculate volatility
	var variance int64
	for _, f := range historicalFees {
		diff := f - avgFee
		variance += diff * diff
	}
	variance /= int64(len(historicalFees))
	volatility := float64(isqrt(variance)) / float64(avgFee)

	// Estimate yearly revenue (assuming current transaction volume)
	txsPerDay := state.TransactionVolume
	if txsPerDay == nil {
		txsPerDay = big.NewInt(10000) // Default estimate
	}
	yearlyRevenue := new(big.Int).Mul(txsPerDay, big.NewInt(avgFee))
	yearlyRevenue.Mul(yearlyRevenue, big.NewInt(365))

	// Market efficiency (simplified: 1 - volatility, capped at 0-1)
	marketEfficiency := 1.0 - volatility
	if marketEfficiency < 0 {
		marketEfficiency = 0
	}
	if marketEfficiency > 1 {
		marketEfficiency = 1
	}

	// Spam resistance score
	spamResistance := s.calculateSpamResistanceFromFees(avgFee, medianFee)

	return economics.FeeMarketAnalysis{
		AverageFeeBPS:       avgFee,
		MedianFeeBPS:        medianFee,
		FeeVolatility:       volatility,
		RevenueEstimateYear: yearlyRevenue,
		MarketEfficiency:    marketEfficiency,
		SpamResistance:      spamResistance,
	}
}

// calculateSpamResistanceFromFees calculates spam resistance from fee data.
func (s *FeeMarketSimulator) calculateSpamResistanceFromFees(avgFee, medianFee int64) int64 {
	score := int64(50) // Base score

	// Higher fees = better spam resistance
	if avgFee >= s.params.MinGasPrice*5 {
		score += 30
	} else if avgFee >= s.params.MinGasPrice*2 {
		score += 15
	}

	// Median close to average indicates fair market (not manipulated)
	if avgFee > 0 {
		ratio := (medianFee * 100) / avgFee
		if ratio >= 80 && ratio <= 120 {
			score += 20
		} else if ratio >= 60 && ratio <= 140 {
			score += 10
		}
	}

	if score > 100 {
		score = 100
	}

	return score
}

// SimulateTakeRateImpact simulates the impact of different take rates.
func (s *FeeMarketSimulator) SimulateTakeRateImpact(
	totalFees *big.Int,
	takeRates []int64,
) []TakeRateImpact {
	results := make([]TakeRateImpact, len(takeRates))

	for i, rate := range takeRates {
		protocolRevenue := new(big.Int).Mul(totalFees, big.NewInt(rate))
		protocolRevenue.Div(protocolRevenue, big.NewInt(10000))

		validatorRevenue := new(big.Int).Sub(totalFees, protocolRevenue)

		// Estimate validator APR impact (simplified)
		// Assuming validators receive fees on top of block rewards
		validatorAprImpact := int64(0)
		if rate > s.params.DefaultTakeRateBPS {
			validatorAprImpact = -(rate - s.params.DefaultTakeRateBPS) / 10
		} else if rate < s.params.DefaultTakeRateBPS {
			validatorAprImpact = (s.params.DefaultTakeRateBPS - rate) / 10
		}

		results[i] = TakeRateImpact{
			TakeRateBPS:        rate,
			ProtocolRevenue:    protocolRevenue,
			ValidatorRevenue:   validatorRevenue,
			ValidatorAPRImpact: validatorAprImpact,
		}
	}

	return results
}

// TakeRateImpact contains the impact analysis of a take rate.
type TakeRateImpact struct {
	TakeRateBPS        int64    `json:"take_rate_bps"`
	ProtocolRevenue    *big.Int `json:"protocol_revenue"`
	ValidatorRevenue   *big.Int `json:"validator_revenue"`
	ValidatorAPRImpact int64    `json:"validator_apr_impact_bps"`
}

// isqrt computes integer square root.
func isqrt(n int64) int64 {
	if n < 0 {
		return 0
	}
	if n == 0 {
		return 0
	}

	x := n
	y := (x + 1) / 2
	for y < x {
		x = y
		y = (x + n/x) / 2
	}
	return x
}

// GenerateSampleTransactions generates sample transactions for testing.
func GenerateSampleTransactions(count int, avgGasPrice int64, spamRatio float64) []Transaction {
	txs := make([]Transaction, count)

	for i := 0; i < count; i++ {
		// Simulate varying gas prices (simple model: +/- 50%)
		priceVariation := int64(float64(avgGasPrice) * (0.5 + float64(i%100)/100.0))
		
		isSpam := float64(i%100) < spamRatio*100

		gasUsed := int64(21000) // Base transaction
		if i%10 == 0 {
			gasUsed = 100000 // Complex transaction
		}

		txs[i] = Transaction{
			Hash:      "tx" + formatAmount(int64(i)),
			GasUsed:   gasUsed,
			GasPrice:  priceVariation,
			FeePaid:   big.NewInt(gasUsed * priceVariation),
			TxType:    "transfer",
			IsSpam:    isSpam,
			Timestamp: time.Now(),
		}
	}

	return txs
}


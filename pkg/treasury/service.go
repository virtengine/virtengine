package treasury

import (
	"context"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
)

// TreasuryService coordinates exchange routing, custody, FX and ledger accounting.
type TreasuryService struct {
	exchange *ExchangeRouter
	custody  *CustodyManager
	ledger   *Ledger
	fx       *FXService
}

func NewTreasuryService(exchange *ExchangeRouter, custody *CustodyManager, ledger *Ledger, fx *FXService) *TreasuryService {
	return &TreasuryService{
		exchange: exchange,
		custody:  custody,
		ledger:   ledger,
		fx:       fx,
	}
}

func (s *TreasuryService) QuoteConversion(ctx context.Context, req ExchangeRequest) (ExchangeQuote, error) {
	return s.exchange.SelectBestQuote(ctx, req)
}

func (s *TreasuryService) ExecuteConversion(ctx context.Context, req ExchangeRequest, reference string) (ExchangeExecution, error) {
	exec, err := s.exchange.ExecuteBestQuote(ctx, req)
	if err != nil {
		return ExchangeExecution{}, err
	}

	if s.ledger != nil {
		s.ledger.RecordConversion(LedgerEntry{
			FromAsset:    exec.Quote.FromAsset,
			ToAsset:      exec.Quote.ToAsset,
			InputAmount:  exec.FilledInput,
			OutputAmount: exec.FilledOut,
			FeeAmount:    exec.Quote.FeeAmount,
			FeeAsset:     exec.Quote.FeeAsset,
			Reference:    reference,
			Timestamp:    exec.ExecutedAt,
		})
	}

	return exec, nil
}

func (s *TreasuryService) RequestWithdrawal(asset string, amount sdkmath.Int, destination string) (*WithdrawalRequest, error) {
	if s.custody == nil {
		return nil, fmt.Errorf("custody manager not configured")
	}
	return s.custody.RequestWithdrawal(asset, amount, destination)
}

func (s *TreasuryService) ApproveWithdrawal(id string, approver string) (*WithdrawalRequest, error) {
	if s.custody == nil {
		return nil, fmt.Errorf("custody manager not configured")
	}
	return s.custody.ApproveWithdrawal(id, approver)
}

func (s *TreasuryService) ExecuteWithdrawal(id string) (*WithdrawalRequest, error) {
	if s.custody == nil {
		return nil, fmt.Errorf("custody manager not configured")
	}
	return s.custody.ExecuteWithdrawal(id)
}

func (s *TreasuryService) FXSnapshot(ctx context.Context, baseAsset, quoteAsset string) (FXSnapshot, error) {
	if s.fx == nil {
		return FXSnapshot{}, fmt.Errorf("fx service not configured")
	}
	return s.fx.GetRate(ctx, baseAsset, quoteAsset)
}

func (s *TreasuryService) FXHistorical(baseAsset, quoteAsset string, at time.Time) (FXSnapshot, bool) {
	if s.fx == nil {
		return FXSnapshot{}, false
	}
	return s.fx.GetHistoricalRate(baseAsset, quoteAsset, at)
}

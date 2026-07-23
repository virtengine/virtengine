// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
)

var ErrFiatConversionQueryUnavailable = errors.New("fiat conversion query unavailable")

const (
	fiatChainStateSwapPending     = "SWAP_PENDING"
	fiatChainStateSwapSubmitted   = "SWAP_SUBMITTED"
	fiatChainStateSwapSettled     = "SWAP_SETTLED"
	fiatChainStatePayoutPending   = "PAYOUT_PENDING" // Settlement chain state, not a local work state.
	fiatChainStatePayoutSubmitted = "PAYOUT_SUBMITTED"
)

// FiatConversionQuery is the read-only chain boundary used by the worker.
type FiatConversionQuery interface {
	Params(ctx context.Context) (settlementv1.Params, error)
	GetConversion(ctx context.Context, conversionID string) (*settlementv1.FiatConversionRecord, error)
	ListNonTerminalConversions(ctx context.Context, provider string) ([]settlementv1.FiatConversionRecord, error)
	ExecutionAuthorization(ctx context.Context, conversionID string) (FiatExecutionAuthorization, error)
}

// FiatExecutionAuthorization is a fresh, read-only view of every consensus
// condition that can stop an external conversion side effect. RPC
// implementations compose only generated settlement query methods.
type FiatExecutionAuthorization struct {
	Conversion      *settlementv1.FiatConversionRecord
	Params          settlementv1.Params
	ActiveCaseID    string
	ActiveHoldCount uint32
}

// RPCFiatConversionQuery uses the existing generated settlement query API.
// The provider query currently returns the complete provider result set; no
// local cursor is advanced unless that query succeeds.
type RPCFiatConversionQuery struct {
	client settlementv1.QueryClient
}

func NewRPCFiatConversionQuery(client *rpcChainClient) (*RPCFiatConversionQuery, error) {
	if client == nil || client.settlementQuery == nil {
		return nil, ErrFiatConversionQueryUnavailable
	}
	return &RPCFiatConversionQuery{client: client.settlementQuery}, nil
}

func (q *RPCFiatConversionQuery) Params(ctx context.Context) (settlementv1.Params, error) {
	if q == nil || q.client == nil {
		return settlementv1.Params{}, ErrFiatConversionQueryUnavailable
	}
	response, err := q.client.Params(ctx, &settlementv1.QueryParamsRequest{})
	if err != nil {
		return settlementv1.Params{}, fmt.Errorf("query settlement params: %w", err)
	}
	return response.Params, nil
}

func (q *RPCFiatConversionQuery) GetConversion(ctx context.Context, conversionID string) (*settlementv1.FiatConversionRecord, error) {
	if q == nil || q.client == nil || strings.TrimSpace(conversionID) == "" {
		return nil, ErrFiatConversionQueryUnavailable
	}
	response, err := q.client.FiatConversion(ctx, &settlementv1.QueryFiatConversionRequest{ConversionId: conversionID})
	if err != nil {
		return nil, fmt.Errorf("query fiat conversion: %w", err)
	}
	if response == nil || response.Conversion == nil {
		return nil, fmt.Errorf("%w: conversion response is empty", ErrFiatConversionQueryUnavailable)
	}
	copyRecord := *response.Conversion
	return &copyRecord, nil
}

func (q *RPCFiatConversionQuery) ListNonTerminalConversions(ctx context.Context, provider string) ([]settlementv1.FiatConversionRecord, error) {
	if q == nil || q.client == nil || strings.TrimSpace(provider) == "" {
		return nil, ErrFiatConversionQueryUnavailable
	}
	response, err := q.client.FiatConversionsByProvider(ctx, &settlementv1.QueryFiatConversionsByProviderRequest{Provider: provider})
	if err != nil {
		return nil, fmt.Errorf("query provider fiat conversions: %w", err)
	}
	if response == nil {
		return nil, fmt.Errorf("%w: provider conversion response is empty", ErrFiatConversionQueryUnavailable)
	}
	result := make([]settlementv1.FiatConversionRecord, 0, len(response.Conversions))
	seen := make(map[string]struct{}, len(response.Conversions))
	for i := range response.Conversions {
		if chainFiatConversionProcessable(response.Conversions[i]) {
			if _, duplicate := seen[response.Conversions[i].ConversionId]; duplicate {
				return nil, fmt.Errorf("%w: duplicate conversion in provider query", ErrFiatConversionQueryUnavailable)
			}
			seen[response.Conversions[i].ConversionId] = struct{}{}
			result = append(result, response.Conversions[i])
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ConversionId < result[j].ConversionId })
	return result, nil
}

func (q *RPCFiatConversionQuery) ExecutionAuthorization(ctx context.Context, conversionID string) (FiatExecutionAuthorization, error) {
	conversion, err := q.GetConversion(ctx, conversionID)
	if err != nil {
		return FiatExecutionAuthorization{}, executionAuthorizationQueryError("conversion", err)
	}
	params, err := q.Params(ctx)
	if err != nil {
		return FiatExecutionAuthorization{}, executionAuthorizationQueryError("params", err)
	}
	subject, err := financialSubjectForConversion(conversion)
	if err != nil {
		return FiatExecutionAuthorization{}, executionAuthorizationQueryError("lineage", err)
	}
	authorization := FiatExecutionAuthorization{Conversion: conversion, Params: params}
	for _, alias := range financialSubjectAliases(subject) {
		response, queryErr := q.client.FinancialCaseBySubject(ctx, &settlementv1.QueryFinancialCaseBySubjectRequest{Subject: alias})
		if queryErr != nil {
			return FiatExecutionAuthorization{}, executionAuthorizationQueryError("financial hold", queryErr)
		}
		if response == nil || response.FinancialCase == nil {
			continue
		}
		financialCase := response.FinancialCase
		if !activeFinancialCaseStatus(financialCase.Status) || financialCase.CaseId == "" || financialCase.ActiveHoldCount == 0 {
			return FiatExecutionAuthorization{}, executionAuthorizationQueryError("financial hold", errors.New("malformed active financial case response"))
		}
		if authorization.ActiveCaseID != "" && authorization.ActiveCaseID != financialCase.CaseId {
			return FiatExecutionAuthorization{}, executionAuthorizationQueryError("financial hold", errors.New("lineage aliases reference multiple active cases"))
		}
		authorization.ActiveCaseID = financialCase.CaseId
		authorization.ActiveHoldCount = financialCase.ActiveHoldCount
	}
	return authorization, nil
}

func executionAuthorizationQueryError(boundary string, err error) error {
	return fmt.Errorf("%w: query %s: %v", ErrFiatConversionQueryUnavailable, boundary, err)
}

func financialSubjectForConversion(conversion *settlementv1.FiatConversionRecord) (settlementv1.FinancialSubject, error) {
	if conversion == nil {
		return settlementv1.FinancialSubject{}, errors.New("conversion is required")
	}
	subject := settlementv1.FinancialSubject{
		OrderId: conversion.OrderId, InvoiceId: conversion.InvoiceId,
		SettlementId: conversion.SettlementId, EscrowId: conversion.EscrowId, LeaseId: conversion.LeaseId,
	}
	switch {
	case conversion.SettlementId != "":
		subject.Type = settlementv1.FinancialSubjectType_FINANCIAL_SUBJECT_TYPE_SETTLEMENT
		subject.PrimaryId = conversion.SettlementId
	case conversion.InvoiceId != "":
		subject.Type = settlementv1.FinancialSubjectType_FINANCIAL_SUBJECT_TYPE_INVOICE
		subject.PrimaryId = conversion.InvoiceId
	case conversion.OrderId != "":
		subject.Type = settlementv1.FinancialSubjectType_FINANCIAL_SUBJECT_TYPE_ORDER
		subject.PrimaryId = conversion.OrderId
	default:
		return settlementv1.FinancialSubject{}, errors.New("conversion has no financial lineage")
	}
	return subject, nil
}

func financialSubjectAliases(subject settlementv1.FinancialSubject) []settlementv1.FinancialSubject {
	aliases := make([]settlementv1.FinancialSubject, 0, 3)
	if subject.SettlementId != "" {
		alias := subject
		alias.Type, alias.PrimaryId = settlementv1.FinancialSubjectType_FINANCIAL_SUBJECT_TYPE_SETTLEMENT, subject.SettlementId
		aliases = append(aliases, alias)
	}
	if subject.InvoiceId != "" {
		alias := subject
		alias.Type, alias.PrimaryId = settlementv1.FinancialSubjectType_FINANCIAL_SUBJECT_TYPE_INVOICE, subject.InvoiceId
		aliases = append(aliases, alias)
	}
	if subject.OrderId != "" {
		alias := subject
		alias.Type, alias.PrimaryId = settlementv1.FinancialSubjectType_FINANCIAL_SUBJECT_TYPE_ORDER, subject.OrderId
		aliases = append(aliases, alias)
	}
	return aliases
}

func activeFinancialCaseStatus(status settlementv1.FinancialCaseStatus) bool {
	switch status {
	case settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_OPEN,
		settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_EVIDENCE,
		settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_REVIEW,
		settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_ESCALATED,
		settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_RESOLVED_PENDING_APPEAL,
		settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_QUARANTINED:
		return true
	default:
		return false
	}
}

func chainFiatConversionProcessable(record settlementv1.FiatConversionRecord) bool {
	if record.ConversionId == "" || record.LegacyQuarantined || chainFiatConversionTerminal(record.State) {
		return false
	}
	switch strings.ToUpper(strings.TrimSpace(record.State)) {
	case "CREATED", fiatChainStateSwapPending, fiatChainStateSwapSubmitted, fiatChainStateSwapSettled, fiatChainStatePayoutPending, "OFFRAMP_PENDING", fiatChainStatePayoutSubmitted:
		return true
	default:
		return false
	}
}

func chainFiatConversionTerminal(state string) bool {
	switch strings.ToUpper(strings.TrimSpace(state)) {
	case "PAYOUT_COMPLETED", "COMPLETED", "FAILED", "CANCELLED":
		return true
	default:
		return false
	}
}

var _ FiatConversionQuery = (*RPCFiatConversionQuery)(nil)

package keeper

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// BorderlineHandler - Handler for borderline verification cases
// ============================================================================

// BorderlineHandler handles cases where ML scores are near thresholds
type BorderlineHandler struct {
	keeper *Keeper
}

// NewBorderlineHandler creates a new borderline handler
func NewBorderlineHandler(k *Keeper) *BorderlineHandler {
	return &BorderlineHandler{keeper: k}
}

// ============================================================================
// Borderline Case Types
// ============================================================================

// BorderlineAction represents the action to take for a borderline case
type BorderlineAction int

const (
	// ActionManualReview queues the case for human review
	ActionManualReview BorderlineAction = iota

	// ActionRequestAdditionalData asks user for more verification data
	ActionRequestAdditionalData

	// ActionApplyPenalty applies a score penalty for suspicious patterns
	ActionApplyPenalty

	// ActionGrantProvisional grants temporary provisional approval
	ActionGrantProvisional

	// ActionRefer refers the case to a specialized review queue
	ActionRefer
)

// String returns the string representation of the action
func (a BorderlineAction) String() string {
	switch a {
	case ActionManualReview:
		return "manual_review"
	case ActionRequestAdditionalData:
		return "request_additional_data"
	case ActionApplyPenalty:
		return "apply_penalty"
	case ActionGrantProvisional:
		return "grant_provisional"
	case ActionRefer:
		return "refer"
	default:
		return "unknown"
	}
}

// BorderlineCase represents a borderline verification case
type BorderlineCase struct {
	// CaseID is the unique identifier for this case
	CaseID string `json:"case_id"`

	// Address is the account address
	Address string `json:"address"`

	// ScopeType is the type of scope being verified
	ScopeType types.ScopeType `json:"scope_type"`

	// Score is the ML score that triggered the borderline case
	Score uint32 `json:"score"`

	// Threshold is the threshold that was nearly met
	Threshold uint32 `json:"threshold"`

	// Margin is how close the score is to the threshold (percentage points)
	Margin uint32 `json:"margin"`

	// FallbackAction is the recommended action for this case
	FallbackAction BorderlineAction `json:"fallback_action"`

	// Status is the current status of the case
	Status BorderlineCaseStatus `json:"status"`

	// CreatedAt is when the case was created (Unix timestamp)
	CreatedAt int64 `json:"created_at"`

	// ExpiresAt is when the case expires if not resolved (Unix timestamp)
	ExpiresAt int64 `json:"expires_at"`

	// ResolvedAt is when the case was resolved (Unix timestamp)
	ResolvedAt int64 `json:"resolved_at,omitempty"`

	// Resolution is the resolution decision
	Resolution string `json:"resolution,omitempty"`

	// ReviewerAddress is the address of the reviewer who resolved the case
	ReviewerAddress string `json:"reviewer_address,omitempty"`

	// Notes contains additional information about the case
	Notes string `json:"notes,omitempty"`

	// BlockHeight is the block height when case was created
	BlockHeight int64 `json:"block_height"`

	// FallbackID links to the BorderlineFallbackRecord if MFA was triggered
	FallbackID string `json:"fallback_id,omitempty"`

	// ProvisionalExpiresAt is when provisional approval expires (if granted)
	ProvisionalExpiresAt int64 `json:"provisional_expires_at,omitempty"`
}

// BorderlineCaseStatus represents the status of a borderline case
type BorderlineCaseStatus string

const (
	// CaseStatusPending indicates the case is awaiting action
	CaseStatusPending BorderlineCaseStatus = "pending"

	// CaseStatusInReview indicates the case is being manually reviewed
	CaseStatusInReview BorderlineCaseStatus = "in_review"

	// CaseStatusAwaitingData indicates waiting for additional data from user
	CaseStatusAwaitingData BorderlineCaseStatus = "awaiting_data"

	// CaseStatusProvisional indicates provisional approval was granted
	CaseStatusProvisional BorderlineCaseStatus = "provisional"

	// CaseStatusResolved indicates the case has been resolved
	CaseStatusResolved BorderlineCaseStatus = "resolved"

	// CaseStatusExpired indicates the case expired without resolution
	CaseStatusExpired BorderlineCaseStatus = "expired"

	// CaseStatusReferred indicates the case was referred elsewhere
	CaseStatusReferred BorderlineCaseStatus = "referred"
)

// ManualReviewRequest represents a request for manual review
type ManualReviewRequest struct {
	// CaseID is the borderline case ID
	CaseID string `json:"case_id"`

	// Requester is the address requesting review
	Requester string `json:"requester"`

	// Reason is why manual review is needed
	Reason string `json:"reason"`

	// Priority is the review priority (1-5, 1 highest)
	Priority int `json:"priority"`

	// EvidenceHashes are hashes of supporting evidence
	EvidenceHashes []string `json:"evidence_hashes,omitempty"`

	// RequestedAt is when review was requested
	RequestedAt int64 `json:"requested_at"`
}

// ProvisionalApproval represents a temporary approval with conditions
type ProvisionalApproval struct {
	// CaseID is the borderline case ID
	CaseID string `json:"case_id"`

	// Address is the approved account address
	Address string `json:"address"`

	// ApprovedAt is when provisional approval was granted
	ApprovedAt int64 `json:"approved_at"`

	// ExpiresAt is when the provisional approval expires
	ExpiresAt int64 `json:"expires_at"`

	// Conditions are the conditions of the provisional approval
	Conditions []string `json:"conditions"`

	// RequiredActions are actions the user must complete before full approval
	RequiredActions []string `json:"required_actions"`

	// TemporaryScore is the temporary score assigned during provisional period
	TemporaryScore uint32 `json:"temporary_score"`

	// OriginalScore is the original borderline score
	OriginalScore uint32 `json:"original_score"`

	// Status is the current status of the provisional approval
	Status ProvisionalStatus `json:"status"`
}

// ProvisionalStatus represents the status of a provisional approval
type ProvisionalStatus string

const (
	// ProvisionalStatusActive indicates the provisional approval is active
	ProvisionalStatusActive ProvisionalStatus = "active"

	// ProvisionalStatusConverted indicates it was converted to full approval
	ProvisionalStatusConverted ProvisionalStatus = "converted"

	// ProvisionalStatusExpired indicates the provisional approval expired
	ProvisionalStatusExpired ProvisionalStatus = "expired"

	// ProvisionalStatusRevoked indicates the provisional approval was revoked
	ProvisionalStatusRevoked ProvisionalStatus = "revoked"
)

// ============================================================================
// Store Prefixes for Borderline Handler
// ============================================================================

var (
	// PrefixBorderlineCase is the prefix for borderline case storage
	PrefixBorderlineCase = []byte{0x80}

	// PrefixBorderlineCaseByAccount is the prefix for case lookup by account
	PrefixBorderlineCaseByAccount = []byte{0x81}

	// PrefixManualReviewQueue is the prefix for manual review queue
	PrefixManualReviewQueue = []byte{0x82}

	// PrefixProvisionalApproval is the prefix for provisional approvals
	PrefixProvisionalApproval = []byte{0x83}

	// PrefixPendingBorderlineCase is the prefix for pending case expiry tracking
	PrefixPendingBorderlineCase = []byte{0x84}
)

// ============================================================================
// Borderline Detection Methods
// ============================================================================

// DetectBorderlineCase checks if a score is within margin of threshold
// and creates a BorderlineCase if so
func (k Keeper) DetectBorderlineCase(
	ctx sdk.Context,
	address string,
	scopeType types.ScopeType,
	score uint32,
) (*BorderlineCase, bool) {
	params := k.GetBorderlineParams(ctx)

	// Check if score is in borderline band
	if !params.IsScoreInBorderlineBand(score) {
		return nil, false
	}

	// Calculate margin from thresholds
	marginFromUpper := params.UpperThreshold - score
	marginFromLower := score - params.LowerThreshold

	// Determine threshold and margin
	var threshold, margin uint32
	if marginFromUpper < marginFromLower {
		threshold = params.UpperThreshold
		margin = marginFromUpper
	} else {
		threshold = params.LowerThreshold
		margin = marginFromLower
	}

	// Determine recommended action based on margin and other factors
	action := k.determineRecommendedAction(ctx, address, score, margin)

	// Generate case ID
	caseID, err := generateFallbackID()
	if err != nil {
		k.Logger(ctx).Error("failed to generate case ID", "error", err)
		return nil, false
	}

	now := ctx.BlockTime().Unix()
	expiresAt := now + params.ChallengeTimeoutSeconds

	borderlineCase := &BorderlineCase{
		CaseID:         caseID,
		Address:        address,
		ScopeType:      scopeType,
		Score:          score,
		Threshold:      threshold,
		Margin:         margin,
		FallbackAction: action,
		Status:         CaseStatusPending,
		CreatedAt:      now,
		ExpiresAt:      expiresAt,
		BlockHeight:    ctx.BlockHeight(),
	}

	return borderlineCase, true
}

// determineRecommendedAction determines the recommended action for a borderline case
func (k Keeper) determineRecommendedAction(
	ctx sdk.Context,
	address string,
	score uint32,
	margin uint32,
) BorderlineAction {
	params := k.GetBorderlineParams(ctx)

	// If very close to upper threshold (within 2 points), suggest MFA first
	if margin <= 2 {
		// Check if account has MFA factors enrolled
		if k.mfaKeeper != nil {
			addr, err := sdk.AccAddressFromBech32(address)
			if err == nil && len(k.mfaKeeper.GetFactorEnrollments(ctx, addr)) > 0 {
				// Has MFA, can use standard fallback
				return ActionManualReview
			}
		}
		// No MFA, request additional data
		return ActionRequestAdditionalData
	}

	// If closer to lower threshold, be more cautious
	if score < params.LowerThreshold+3 {
		return ActionRefer
	}

	// Default to manual review for middle cases
	return ActionManualReview
}

// ============================================================================
// Borderline Case Handling Methods
// ============================================================================

// HandleBorderlineCase applies the appropriate fallback action for a borderline case
func (k Keeper) HandleBorderlineCase(
	ctx sdk.Context,
	borderlineCase *BorderlineCase,
) error {
	switch borderlineCase.FallbackAction {
	case ActionManualReview:
		return k.SubmitForManualReview(ctx, borderlineCase, "Borderline score detected")

	case ActionRequestAdditionalData:
		return k.RequestAdditionalData(ctx, borderlineCase)

	case ActionApplyPenalty:
		return k.applyBorderlinePenalty(ctx, borderlineCase)

	case ActionGrantProvisional:
		return k.GrantProvisionalApproval(ctx, borderlineCase, 24*time.Hour)

	case ActionRefer:
		return k.referBorderlineCase(ctx, borderlineCase)

	default:
		return types.ErrInvalidBorderlineFallback.Wrap("unknown fallback action")
	}
}

// SubmitForManualReview queues a borderline case for human review
func (k Keeper) SubmitForManualReview(
	ctx sdk.Context,
	borderlineCase *BorderlineCase,
	reason string,
) error {
	// Update case status
	borderlineCase.Status = CaseStatusInReview
	borderlineCase.Notes = reason

	// Store the case
	if err := k.setBorderlineCase(ctx, borderlineCase); err != nil {
		return err
	}

	// Create manual review request
	reviewRequest := &ManualReviewRequest{
		CaseID:      borderlineCase.CaseID,
		Requester:   borderlineCase.Address,
		Reason:      reason,
		Priority:    k.calculateReviewPriority(borderlineCase),
		RequestedAt: ctx.BlockTime().Unix(),
	}

	// Add to review queue
	if err := k.addToManualReviewQueue(ctx, reviewRequest); err != nil {
		return err
	}

	// Emit event
	k.emitManualReviewRequestedEvent(ctx, borderlineCase, reason)

	k.Logger(ctx).Info("borderline case submitted for manual review",
		"case_id", borderlineCase.CaseID,
		"address", borderlineCase.Address,
		"score", borderlineCase.Score,
		"priority", reviewRequest.Priority,
	)

	return nil
}

// calculateReviewPriority calculates review priority based on case attributes
func (k Keeper) calculateReviewPriority(borderlineCase *BorderlineCase) int {
	// Priority 1 (highest) to 5 (lowest)
	// Closer to threshold = higher priority
	if borderlineCase.Margin <= 1 {
		return 1
	}
	if borderlineCase.Margin <= 3 {
		return 2
	}
	if borderlineCase.Margin <= 5 {
		return 3
	}
	return 4
}

// RequestAdditionalData marks a case as awaiting additional verification data
func (k Keeper) RequestAdditionalData(
	ctx sdk.Context,
	borderlineCase *BorderlineCase,
) error {
	borderlineCase.Status = CaseStatusAwaitingData
	borderlineCase.Notes = "Additional verification data requested"

	if err := k.setBorderlineCase(ctx, borderlineCase); err != nil {
		return err
	}

	// Emit event
	k.emitAdditionalDataRequestedEvent(ctx, borderlineCase)

	k.Logger(ctx).Info("additional data requested for borderline case",
		"case_id", borderlineCase.CaseID,
		"address", borderlineCase.Address,
	)

	return nil
}

// GrantProvisionalApproval grants temporary approval with conditions
func (k Keeper) GrantProvisionalApproval(
	ctx sdk.Context,
	borderlineCase *BorderlineCase,
	duration time.Duration,
) error {
	now := ctx.BlockTime().Unix()
	expiresAt := now + int64(duration.Seconds())

	// Update case status
	borderlineCase.Status = CaseStatusProvisional
	borderlineCase.ProvisionalExpiresAt = expiresAt

	if err := k.setBorderlineCase(ctx, borderlineCase); err != nil {
		return err
	}

	// Create provisional approval record
	provisionalApproval := &ProvisionalApproval{
		CaseID:         borderlineCase.CaseID,
		Address:        borderlineCase.Address,
		ApprovedAt:     now,
		ExpiresAt:      expiresAt,
		Conditions:     []string{"Must complete MFA verification within provisional period"},
		RequiredActions: []string{"Enroll MFA factor", "Complete identity re-verification"},
		TemporaryScore: borderlineCase.Score,
		OriginalScore:  borderlineCase.Score,
		Status:         ProvisionalStatusActive,
	}

	if err := k.setProvisionalApproval(ctx, provisionalApproval); err != nil {
		return err
	}

	// Update the account's verification status to pending (provisional state)
	if err := k.SetScoreWithDetails(ctx, borderlineCase.Address, borderlineCase.Score, ScoreDetails{
		Status:       types.AccountStatusPending,
		ModelVersion: "borderline-provisional",
		Reason:       fmt.Sprintf("provisional approval for borderline case %s, expires %s", borderlineCase.CaseID, time.Unix(expiresAt, 0).Format(time.RFC3339)),
	}); err != nil {
		return err
	}

	// Emit event
	k.emitProvisionalApprovalGrantedEvent(ctx, borderlineCase, provisionalApproval)

	k.Logger(ctx).Info("provisional approval granted for borderline case",
		"case_id", borderlineCase.CaseID,
		"address", borderlineCase.Address,
		"expires_at", expiresAt,
	)

	return nil
}

// applyBorderlinePenalty applies a score penalty for suspicious patterns
func (k Keeper) applyBorderlinePenalty(
	ctx sdk.Context,
	borderlineCase *BorderlineCase,
) error {
	// Calculate penalty (reduce score by margin to push below threshold)
	penaltyAmount := borderlineCase.Margin + 1
	newScore := borderlineCase.Score - penaltyAmount
	if newScore < 0 {
		newScore = 0
	}

	borderlineCase.Status = CaseStatusResolved
	borderlineCase.Resolution = fmt.Sprintf("penalty applied: -%d points", penaltyAmount)
	borderlineCase.ResolvedAt = ctx.BlockTime().Unix()

	if err := k.setBorderlineCase(ctx, borderlineCase); err != nil {
		return err
	}

	// Update score with penalty
	if err := k.SetScoreWithDetails(ctx, borderlineCase.Address, uint32(newScore), ScoreDetails{
		Status:       types.AccountStatusRejected,
		ModelVersion: "borderline-penalty",
		Reason:       fmt.Sprintf("borderline penalty applied for case %s", borderlineCase.CaseID),
	}); err != nil {
		return err
	}

	k.Logger(ctx).Info("penalty applied for borderline case",
		"case_id", borderlineCase.CaseID,
		"address", borderlineCase.Address,
		"original_score", borderlineCase.Score,
		"new_score", newScore,
	)

	return nil
}

// referBorderlineCase refers a case to a specialized review queue
func (k Keeper) referBorderlineCase(
	ctx sdk.Context,
	borderlineCase *BorderlineCase,
) error {
	borderlineCase.Status = CaseStatusReferred
	borderlineCase.Notes = "Referred for specialized review"

	if err := k.setBorderlineCase(ctx, borderlineCase); err != nil {
		return err
	}

	k.Logger(ctx).Info("borderline case referred for specialized review",
		"case_id", borderlineCase.CaseID,
		"address", borderlineCase.Address,
	)

	return nil
}

// ============================================================================
// Case Query Methods
// ============================================================================

// GetBorderlineCases returns all pending borderline cases
func (k Keeper) GetBorderlineCases(ctx sdk.Context) []BorderlineCase {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, PrefixBorderlineCase)
	defer iterator.Close()

	var cases []BorderlineCase
	for ; iterator.Valid(); iterator.Next() {
		var bc BorderlineCase
		if err := json.Unmarshal(iterator.Value(), &bc); err != nil {
			continue
		}
		cases = append(cases, bc)
	}

	return cases
}

// GetPendingBorderlineCases returns only pending borderline cases
func (k Keeper) GetPendingBorderlineCases(ctx sdk.Context) []BorderlineCase {
	allCases := k.GetBorderlineCases(ctx)
	var pending []BorderlineCase
	for _, c := range allCases {
		if c.Status == CaseStatusPending || c.Status == CaseStatusInReview || c.Status == CaseStatusAwaitingData {
			pending = append(pending, c)
		}
	}
	return pending
}

// GetBorderlineCase retrieves a specific borderline case by ID
func (k Keeper) GetBorderlineCase(ctx sdk.Context, caseID string) (*BorderlineCase, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(borderlineCaseKey(caseID))
	if bz == nil {
		return nil, false
	}

	var bc BorderlineCase
	if err := json.Unmarshal(bz, &bc); err != nil {
		return nil, false
	}

	return &bc, true
}

// GetBorderlineCasesForAccount returns all borderline cases for an account
func (k Keeper) GetBorderlineCasesForAccount(ctx sdk.Context, address string) []BorderlineCase {
	allCases := k.GetBorderlineCases(ctx)
	var accountCases []BorderlineCase
	for _, c := range allCases {
		if c.Address == address {
			accountCases = append(accountCases, c)
		}
	}
	return accountCases
}

// ============================================================================
// Case Resolution Methods
// ============================================================================

// ResolveBorderlineCase resolves a borderline case with a final decision
func (k Keeper) ResolveBorderlineCase(
	ctx sdk.Context,
	caseID string,
	resolution string,
	reviewerAddress string,
	approved bool,
	newScore *uint32,
) error {
	borderlineCase, found := k.GetBorderlineCase(ctx, caseID)
	if !found {
		return types.ErrBorderlineFallbackNotFound.Wrapf("case %s not found", caseID)
	}

	// Check if case can be resolved
	if borderlineCase.Status == CaseStatusResolved || borderlineCase.Status == CaseStatusExpired {
		return types.ErrBorderlineFallbackAlreadyCompleted.Wrapf("case status is %s", borderlineCase.Status)
	}

	now := ctx.BlockTime().Unix()
	borderlineCase.Status = CaseStatusResolved
	borderlineCase.Resolution = resolution
	borderlineCase.ReviewerAddress = reviewerAddress
	borderlineCase.ResolvedAt = now

	if err := k.setBorderlineCase(ctx, borderlineCase); err != nil {
		return err
	}

	// Update account score/status based on resolution
	if approved {
		scoreToSet := borderlineCase.Score
		if newScore != nil {
			scoreToSet = *newScore
		}

		if err := k.SetScoreWithDetails(ctx, borderlineCase.Address, scoreToSet, ScoreDetails{
			Status:       types.AccountStatusVerified,
			ModelVersion: "borderline-resolved",
			Reason:       fmt.Sprintf("borderline case %s resolved: %s", caseID, resolution),
		}); err != nil {
			return err
		}
	} else {
		// Rejected
		if err := k.SetScoreWithDetails(ctx, borderlineCase.Address, borderlineCase.Score, ScoreDetails{
			Status:       types.AccountStatusRejected,
			ModelVersion: "borderline-resolved",
			Reason:       fmt.Sprintf("borderline case %s rejected: %s", caseID, resolution),
		}); err != nil {
			return err
		}
	}

	// Emit resolution event
	k.emitBorderlineCaseResolvedEvent(ctx, borderlineCase, approved)

	k.Logger(ctx).Info("borderline case resolved",
		"case_id", caseID,
		"reviewer", reviewerAddress,
		"approved", approved,
		"resolution", resolution,
	)

	return nil
}

// ProcessExpiredProvisionalApprovals processes expired provisional approvals
func (k Keeper) ProcessExpiredProvisionalApprovals(ctx sdk.Context) int {
	now := ctx.BlockTime().Unix()
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, PrefixProvisionalApproval)
	defer iterator.Close()

	expiredCount := 0

	for ; iterator.Valid(); iterator.Next() {
		var pa ProvisionalApproval
		if err := json.Unmarshal(iterator.Value(), &pa); err != nil {
			continue
		}

		if pa.Status == ProvisionalStatusActive && pa.ExpiresAt <= now {
			// Mark as expired
			pa.Status = ProvisionalStatusExpired
			bz, _ := json.Marshal(pa)
			store.Set(iterator.Key(), bz)

			// Update the account's status
			_ = k.SetScoreWithDetails(ctx, pa.Address, pa.OriginalScore, ScoreDetails{
				Status:       types.AccountStatusExpired,
				ModelVersion: "provisional-expired",
				Reason:       fmt.Sprintf("provisional approval for case %s expired", pa.CaseID),
			})

			expiredCount++

			k.Logger(ctx).Info("provisional approval expired",
				"case_id", pa.CaseID,
				"address", pa.Address,
			)
		}
	}

	if expiredCount > 0 {
		k.Logger(ctx).Info("processed expired provisional approvals", "count", expiredCount)
	}

	return expiredCount
}

// ============================================================================
// Storage Methods
// ============================================================================

func borderlineCaseKey(caseID string) []byte {
	return append(PrefixBorderlineCase, []byte(caseID)...)
}

func provisionalApprovalKey(caseID string) []byte {
	return append(PrefixProvisionalApproval, []byte(caseID)...)
}

func manualReviewQueueKey(priority int, caseID string) []byte {
	key := append(PrefixManualReviewQueue, byte(priority))
	return append(key, []byte(caseID)...)
}

func (k Keeper) setBorderlineCase(ctx sdk.Context, borderlineCase *BorderlineCase) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(borderlineCase)
	if err != nil {
		return err
	}
	store.Set(borderlineCaseKey(borderlineCase.CaseID), bz)
	return nil
}

func (k Keeper) setProvisionalApproval(ctx sdk.Context, approval *ProvisionalApproval) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(approval)
	if err != nil {
		return err
	}
	store.Set(provisionalApprovalKey(approval.CaseID), bz)
	return nil
}

func (k Keeper) addToManualReviewQueue(ctx sdk.Context, request *ManualReviewRequest) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(request)
	if err != nil {
		return err
	}
	store.Set(manualReviewQueueKey(request.Priority, request.CaseID), bz)
	return nil
}

// GetManualReviewQueue returns the manual review queue ordered by priority
func (k Keeper) GetManualReviewQueue(ctx sdk.Context) []ManualReviewRequest {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, PrefixManualReviewQueue)
	defer iterator.Close()

	var requests []ManualReviewRequest
	for ; iterator.Valid(); iterator.Next() {
		var req ManualReviewRequest
		if err := json.Unmarshal(iterator.Value(), &req); err != nil {
			continue
		}
		requests = append(requests, req)
	}

	return requests
}

// GetProvisionalApproval retrieves a provisional approval by case ID
func (k Keeper) GetProvisionalApproval(ctx sdk.Context, caseID string) (*ProvisionalApproval, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(provisionalApprovalKey(caseID))
	if bz == nil {
		return nil, false
	}

	var pa ProvisionalApproval
	if err := json.Unmarshal(bz, &pa); err != nil {
		return nil, false
	}

	return &pa, true
}

// ============================================================================
// Event Emission
// ============================================================================

func (k Keeper) emitManualReviewRequestedEvent(ctx sdk.Context, borderlineCase *BorderlineCase, reason string) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"borderline_manual_review_requested",
			sdk.NewAttribute("case_id", borderlineCase.CaseID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, borderlineCase.Address),
			sdk.NewAttribute(types.AttributeKeyBorderlineScore, fmt.Sprintf("%d", borderlineCase.Score)),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
}

func (k Keeper) emitAdditionalDataRequestedEvent(ctx sdk.Context, borderlineCase *BorderlineCase) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"borderline_additional_data_requested",
			sdk.NewAttribute("case_id", borderlineCase.CaseID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, borderlineCase.Address),
			sdk.NewAttribute(types.AttributeKeyBorderlineScore, fmt.Sprintf("%d", borderlineCase.Score)),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
}

func (k Keeper) emitProvisionalApprovalGrantedEvent(ctx sdk.Context, borderlineCase *BorderlineCase, approval *ProvisionalApproval) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"borderline_provisional_approval_granted",
			sdk.NewAttribute("case_id", borderlineCase.CaseID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, borderlineCase.Address),
			sdk.NewAttribute(types.AttributeKeyBorderlineScore, fmt.Sprintf("%d", borderlineCase.Score)),
			sdk.NewAttribute(types.AttributeKeyExpiresAt, fmt.Sprintf("%d", approval.ExpiresAt)),
			sdk.NewAttribute("conditions", strings.Join(approval.Conditions, ",")),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
}

func (k Keeper) emitBorderlineCaseResolvedEvent(ctx sdk.Context, borderlineCase *BorderlineCase, approved bool) {
	status := "rejected"
	if approved {
		status = "approved"
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"borderline_case_resolved",
			sdk.NewAttribute("case_id", borderlineCase.CaseID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, borderlineCase.Address),
			sdk.NewAttribute(types.AttributeKeyBorderlineScore, fmt.Sprintf("%d", borderlineCase.Score)),
			sdk.NewAttribute("resolution", borderlineCase.Resolution),
			sdk.NewAttribute("reviewer", borderlineCase.ReviewerAddress),
			sdk.NewAttribute(types.AttributeKeyFinalStatus, status),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
}

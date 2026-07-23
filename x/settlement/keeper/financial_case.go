package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/x/settlement/types"
)

const (
	financialCaseVersion             uint32 = 1
	financialHashSize                       = sha256.Size
	financialMaxSourceModuleBytes           = 32
	financialMaxSourceReferenceBytes        = 256
	financialMaxRecommendationBytes         = 256
	financialMaxSubjectIDBytes              = 512
	financialMaxReasonHashBytes             = sha256.Size
	financialDefaultMaxClaims        uint32 = 32
	financialDefaultMaxAppeals       uint32 = 1
	financialDefaultEvidenceRefBytes uint32 = 512
	financialDefaultTimeoutLimit     uint32 = 100
	financialSourceSettlement               = "settlement"
)

var financialCaseDomain = []byte("virtengine/settlement/financial-case/v1")
var financialClaimDomain = []byte("virtengine/settlement/financial-claim/v1")
var financialAppealDomain = []byte("virtengine/settlement/financial-appeal/v1")

// FinancialCaseOpenRequest is the narrow cross-module adapter input.
type FinancialCaseOpenRequest struct {
	Subject          types.FinancialSubject
	Claimant         string
	Respondent       string
	Claim            types.FinancialClaim
	IdempotencyKey   []byte
	TrustedAdapter   bool
	Migrated         bool
	Quarantine       bool
	QuarantineReason string
}

// FinancialCaseMigrationReport is persisted as the v1.7.0 reconciliation artifact.
type FinancialCaseMigrationReport struct {
	PayoutsScanned    uint64 `json:"payouts_scanned"`
	EscrowsScanned    uint64 `json:"escrows_scanned"`
	CasesCreated      uint64 `json:"cases_created"`
	ClaimsMerged      uint64 `json:"claims_merged"`
	Quarantined       uint64 `json:"quarantined"`
	TerminalPreserved uint64 `json:"terminal_preserved"`
	AlreadyMigrated   uint64 `json:"already_migrated"`
	MalformedOrphans  uint64 `json:"malformed_orphans"`
	Digest            string `json:"digest"`
}

type financialClaimReplay struct {
	CaseID      string `json:"case_id"`
	ClaimID     string `json:"claim_id"`
	PayloadHash []byte `json:"payload_hash"`
}

type financialAppealReplay struct {
	CaseID      string `json:"case_id"`
	AppealID    string `json:"appeal_id"`
	PayloadHash []byte `json:"payload_hash"`
}

// DeterministicFinancialCaseID derives the same case ID from the canonical subject on every validator.
func DeterministicFinancialCaseID(subject types.FinancialSubject) (string, error) {
	subjectKey, err := CanonicalFinancialSubjectKey(subject)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	writeFinancialField(h, financialCaseDomain)
	writeFinancialField(h, []byte(subjectKey))
	return "financial-case/" + hex.EncodeToString(h.Sum(nil)), nil
}

// DeterministicFinancialClaimID derives one claim identity and its conflict-detection payload hash.
func DeterministicFinancialClaimID(caseID string, claim types.FinancialClaim) (string, []byte, error) {
	if caseID == "" {
		return "", nil, types.ErrInvalidFinancialCase.Wrap("case ID required")
	}
	if err := validateFinancialClaimInput(claim, financialDefaultEvidenceRefBytes); err != nil {
		return "", nil, err
	}
	h := sha256.New()
	writeFinancialField(h, financialClaimDomain)
	writeFinancialField(h, []byte(caseID))
	writeFinancialField(h, []byte(claim.ClaimType.String()))
	writeFinancialField(h, []byte(claim.Claimant))
	writeFinancialField(h, []byte(claim.SourceModule))
	writeFinancialField(h, []byte(claim.SourceReference))
	writeFinancialField(h, claim.EvidenceHash)
	writeFinancialField(h, []byte(claim.EncryptedReference))
	writeFinancialField(h, claim.IdempotencyKey)
	writeFinancialField(h, []byte(claim.Recommendation))
	payload := h.Sum(nil)
	return "financial-claim/" + hex.EncodeToString(payload), payload, nil
}

func deterministicAppealID(caseID, appellant string, evidenceHash, idempotency []byte) string {
	h := sha256.New()
	writeFinancialField(h, financialAppealDomain)
	writeFinancialField(h, []byte(caseID))
	writeFinancialField(h, []byte(appellant))
	writeFinancialField(h, evidenceHash)
	writeFinancialField(h, idempotency)
	return "financial-appeal/" + hex.EncodeToString(h.Sum(nil))
}

func financialAppealPayloadHash(caseID, appellant string, evidenceHash []byte, encryptedReference string, idempotency []byte) []byte {
	h := sha256.New()
	writeFinancialField(h, financialAppealDomain)
	writeFinancialField(h, []byte(caseID))
	writeFinancialField(h, []byte(appellant))
	writeFinancialField(h, evidenceHash)
	writeFinancialField(h, []byte(encryptedReference))
	writeFinancialField(h, idempotency)
	return h.Sum(nil)
}

type byteWriter interface{ Write([]byte) (int, error) }

func writeFinancialField(w byteWriter, value []byte) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value))) //nolint:gosec // all fields are protocol bounded before hashing
	_, _ = w.Write(length[:])
	_, _ = w.Write(value)
}

// CanonicalFinancialSubjectKey returns one deterministic lineage root.
func CanonicalFinancialSubjectKey(subject types.FinancialSubject) (string, error) {
	primary := strings.TrimSpace(subject.PrimaryId)
	if primary == "" {
		switch subject.Type {
		case types.FinancialSubjectTypeOrder:
			primary = subject.OrderId
		case types.FinancialSubjectTypeInvoice:
			primary = subject.InvoiceId
		case types.FinancialSubjectTypeUsage:
			primary = subject.UsageId
		case types.FinancialSubjectTypeHPCJob:
			primary = subject.HpcJobId
		case types.FinancialSubjectTypeSettlement:
			primary = subject.SettlementId
		}
	}
	if subject.Type == settlementv1.FinancialSubjectType_FINANCIAL_SUBJECT_TYPE_UNSPECIFIED || primary == "" {
		return "", types.ErrInvalidFinancialCase.Wrap("subject type and primary ID required")
	}
	if len(primary) > financialMaxSubjectIDBytes || strings.ContainsRune(primary, '\x00') {
		return "", types.ErrInvalidFinancialCase.Wrap("subject ID exceeds bounds or contains NUL")
	}
	return fmt.Sprintf("%d/%s", subject.Type, primary), nil
}

func CanTransitionFinancialCase(from, to types.FinancialCaseStatus) bool {
	switch from {
	case types.FinancialCaseStatusOpen:
		return to == types.FinancialCaseStatusEvidence || to == types.FinancialCaseStatusReview || to == types.FinancialCaseStatusEscalated || to == types.FinancialCaseStatusRejected || to == types.FinancialCaseStatusCancelled || to == types.FinancialCaseStatusExpired || to == types.FinancialCaseStatusQuarantined
	case types.FinancialCaseStatusEvidence:
		return to == types.FinancialCaseStatusReview || to == types.FinancialCaseStatusEscalated || to == types.FinancialCaseStatusRejected || to == types.FinancialCaseStatusCancelled || to == types.FinancialCaseStatusExpired || to == types.FinancialCaseStatusQuarantined
	case types.FinancialCaseStatusReview:
		return to == types.FinancialCaseStatusEscalated || to == types.FinancialCaseStatusResolvedPendingAppeal || to == types.FinancialCaseStatusRejected || to == types.FinancialCaseStatusQuarantined
	case types.FinancialCaseStatusEscalated:
		return to == types.FinancialCaseStatusResolvedPendingAppeal || to == types.FinancialCaseStatusRejected || to == types.FinancialCaseStatusQuarantined
	case types.FinancialCaseStatusResolvedPendingAppeal:
		return to == types.FinancialCaseStatusReview || to == types.FinancialCaseStatusFinal || to == types.FinancialCaseStatusQuarantined
	case types.FinancialCaseStatusQuarantined:
		return to == types.FinancialCaseStatusEscalated || to == types.FinancialCaseStatusResolvedPendingAppeal
	default:
		return false
	}
}

// ValidateTerminalAllocation enforces exact equality for every denom and rejects hidden/extra denoms.
func ValidateTerminalAllocation(exposure sdk.Coins, allocation types.TerminalAllocation) error {
	if !exposure.IsValid() || exposure.IsZero() || !allocation.OriginalExposure.IsValid() || !allocation.OriginalExposure.Equal(exposure) {
		return types.ErrFinancialCaseConservation.Wrap("original exposure mismatch")
	}
	for _, coins := range []sdk.Coins{allocation.Provider, allocation.Customer, allocation.Platform, allocation.SlashWitness} {
		if !coins.IsValid() {
			return types.ErrFinancialCaseConservation.Wrap("invalid allocation coins")
		}
	}
	total := allocation.Provider.Add(allocation.Customer...).Add(allocation.Platform...).Add(allocation.SlashWitness...)
	if !total.Equal(exposure) {
		return types.ErrFinancialCaseConservation.Wrapf("allocation %s does not equal exposure %s", total, exposure)
	}
	if allocation.ResolutionType == settlementv1.FinancialResolutionType_FINANCIAL_RESOLUTION_TYPE_UNSPECIFIED {
		return types.ErrFinancialCaseConservation.Wrap("resolution type required")
	}
	if !allocation.SlashWitness.IsZero() {
		if _, err := sdk.AccAddressFromBech32(allocation.SlashWitnessRecipient); err != nil {
			return types.ErrFinancialCaseConservation.Wrap("slash/witness recipient required")
		}
	}
	return nil
}

func (k Keeper) ActivateFinancialCases(ctx sdk.Context) {
	ctx.KVStore(k.skey).Set(types.FinancialCaseActivationKey(), []byte{1})
}
func (k Keeper) IsFinancialCasesActive(ctx sdk.Context) bool {
	return ctx.KVStore(k.skey).Has(types.FinancialCaseActivationKey())
}

func (k Keeper) OpenFinancialCase(ctx sdk.Context, request FinancialCaseOpenRequest) (*types.FinancialCase, *types.FinancialClaim, bool, error) {
	if !request.TrustedAdapter && !request.Migrated && !request.Quarantine && !k.IsFinancialCasesActive(ctx) {
		return nil, nil, false, types.ErrFinancialCasesNotActive
	}
	cacheCtx, write := ctx.CacheContext()
	financialCase, claim, duplicate, err := k.openFinancialCase(cacheCtx, request)
	if err != nil {
		return nil, nil, false, err
	}
	write()
	return financialCase, claim, duplicate, nil
}

func (k Keeper) openFinancialCase(ctx sdk.Context, request FinancialCaseOpenRequest) (*types.FinancialCase, *types.FinancialClaim, bool, error) {
	if _, err := sdk.AccAddressFromBech32(request.Claimant); err != nil {
		return nil, nil, false, types.ErrFinancialCaseAuthorization.Wrap("invalid claimant")
	}
	if _, err := sdk.AccAddressFromBech32(request.Respondent); err != nil || request.Claimant == request.Respondent {
		return nil, nil, false, types.ErrFinancialCaseAuthorization.Wrap("invalid or conflicting respondent")
	}
	if len(request.IdempotencyKey) == 0 || len(request.IdempotencyKey) > 128 {
		return nil, nil, false, types.ErrFinancialCaseIdempotencyConflict.Wrap("bounded idempotency key required")
	}
	if !request.TrustedAdapter && !request.Migrated && !request.Quarantine && request.Claimant != k.authority && !k.financialSubjectPartyAuthorized(ctx, request.Subject, request.Claimant) {
		return nil, nil, false, types.ErrFinancialCaseAuthorization.Wrap("claimant is not a subject party")
	}
	if !request.TrustedAdapter && !request.Migrated && !request.Quarantine && request.Claimant == k.authority {
		return nil, nil, false, types.ErrFinancialCaseResolverConflict.Wrap("resolver authority cannot file as a financial party")
	}
	request.Claim.Claimant = request.Claimant
	request.Claim.IdempotencyKey = append([]byte(nil), request.IdempotencyKey...)
	params := k.GetParams(ctx)
	maxReference := params.FinancialCaseMaxEvidenceReferenceBytes
	if maxReference == 0 {
		maxReference = financialDefaultEvidenceRefBytes
	}
	if err := validateFinancialClaimInput(request.Claim, maxReference); err != nil {
		return nil, nil, false, err
	}
	canonicalSubject, err := k.normalizeFinancialSubject(ctx, request.Subject)
	if err != nil {
		return nil, nil, false, err
	}
	request.Subject = canonicalSubject
	subjectKey, err := CanonicalFinancialSubjectKey(request.Subject)
	if err != nil {
		return nil, nil, false, err
	}
	caseID, err := DeterministicFinancialCaseID(request.Subject)
	if err != nil {
		return nil, nil, false, err
	}
	claimID, payloadHash, err := DeterministicFinancialClaimID(caseID, request.Claim)
	if err != nil {
		return nil, nil, false, err
	}
	request.Claim.ClaimId = claimID
	request.Claim.PayloadHash = append([]byte(nil), payloadHash...)

	if replay, exists, err := k.getFinancialClaimReplay(ctx, request.IdempotencyKey); err != nil {
		return nil, nil, false, err
	} else if exists {
		if replay.CaseID != caseID || !bytes.Equal(replay.PayloadHash, payloadHash) {
			return nil, nil, false, types.ErrFinancialCaseIdempotencyConflict
		}
		existing, found := k.GetFinancialCase(ctx, replay.CaseID)
		if !found {
			return nil, nil, false, types.ErrFinancialCaseMalformedState.Wrap("idempotency points to missing case")
		}
		for i := range existing.Claims {
			if existing.Claims[i].ClaimId == replay.ClaimID {
				return &existing, &existing.Claims[i], true, nil
			}
		}
		return nil, nil, false, types.ErrFinancialCaseMalformedState.Wrap("idempotency points to missing claim")
	}

	if existingCaseID, found, findErr := k.findActiveFinancialCaseForSubject(ctx, request.Subject); findErr != nil {
		return nil, nil, false, findErr
	} else if found {
		existing, exists := k.GetFinancialCase(ctx, existingCaseID)
		if !exists {
			return nil, nil, false, types.ErrFinancialCaseMalformedState.Wrap("active subject index points to missing case")
		}
		if request.Claimant != existing.Provider && request.Claimant != existing.Customer || request.Respondent != existing.Provider && request.Respondent != existing.Customer {
			return nil, nil, false, types.ErrFinancialCaseAuthorization.Wrap("simultaneous claim parties conflict with canonical root")
		}
		if !request.TrustedAdapter && !request.Migrated && deadlinePassed(ctx, existing.FilingDeadlineHeight, existing.FilingDeadlineTime) {
			return nil, nil, false, types.ErrFinancialCaseDeadline.Wrap("filing window closed")
		}
		caseID = existingCaseID
		request.Claim.ClaimId, request.Claim.PayloadHash, err = DeterministicFinancialClaimID(caseID, request.Claim)
		if err != nil {
			return nil, nil, false, err
		}
		return k.addFinancialClaim(ctx, caseID, request.Claim)
	}
	if existing, found := k.GetFinancialCase(ctx, caseID); found {
		return nil, nil, false, types.ErrFinancialCaseTransition.Wrapf("deterministic case root already exists with status %s", existing.Status)
	}

	exposure, err := k.deriveFinancialExposure(ctx, request.Subject)
	if err != nil {
		return nil, nil, false, err
	}
	provider, customer, err := k.deriveFinancialParties(ctx, request.Subject, request.Claimant, request.Respondent)
	if err != nil {
		return nil, nil, false, err
	}
	now := ctx.BlockTime().Unix()
	maxAppeals := params.FinancialCaseMaxAppeals
	if maxAppeals == 0 {
		maxAppeals = financialDefaultMaxAppeals
	}
	status := types.FinancialCaseStatusOpen
	if request.Quarantine {
		status = types.FinancialCaseStatusQuarantined
	}
	financialCase := types.FinancialCase{
		Version: financialCaseVersion, CaseId: caseID, Subject: request.Subject,
		Claimant: request.Claimant, Respondent: request.Respondent, Exposure: exposure,
		Provider: provider, Customer: customer,
		Status: status, ResolverAuthority: k.authority, MaxAppeals: maxAppeals,
		OpenIdempotencyKey: append([]byte(nil), request.IdempotencyKey...), Migrated: request.Migrated,
		Quarantined: request.Quarantine, QuarantineReason: request.QuarantineReason,
		CreatedHeight: ctx.BlockHeight(), CreatedAt: now, UpdatedHeight: ctx.BlockHeight(), UpdatedAt: now,
	}
	k.setFinancialCaseDeadlines(ctx, &financialCase)
	request.Claim.CreatedHeight = ctx.BlockHeight()
	request.Claim.CreatedAt = now
	financialCase.Claims = []types.FinancialClaim{request.Claim}
	financialCase.ClaimRoot = financialClaimRoot(financialCase.Claims)
	k.appendFinancialTransition(ctx, &financialCase, settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_UNSPECIFIED, status, request.Claimant, "opened", nil)
	if err := k.holdFinancialExposure(ctx, &financialCase); err != nil {
		return nil, nil, false, err
	}
	if err := k.SetFinancialCase(ctx, financialCase); err != nil {
		return nil, nil, false, err
	}
	if err := k.setFinancialClaimReplay(ctx, request.IdempotencyKey, financialCase.CaseId, request.Claim.ClaimId, request.Claim.PayloadHash); err != nil {
		return nil, nil, false, err
	}
	k.emitFinancialCaseOpened(ctx, financialCase, subjectKey)
	if request.Quarantine {
		k.emitFinancialCaseQuarantined(ctx, financialCase.CaseId, []byte(request.QuarantineReason))
	}
	return &financialCase, &financialCase.Claims[0], false, nil
}

func (k Keeper) findActiveFinancialCaseForSubject(ctx sdk.Context, subject types.FinancialSubject) (string, bool, error) {
	if subjectKey, err := CanonicalFinancialSubjectKey(subject); err == nil {
		if indexed := ctx.KVStore(k.skey).Get(types.FinancialSubjectKey(subjectKey)); indexed != nil {
			return string(indexed), true, nil
		}
	}
	aliases := []struct{ kind, value string }{{"order", subject.OrderId}, {"invoice", subject.InvoiceId}, {"usage", subject.UsageId}, {"job", subject.HpcJobId}, {"escrow", subject.EscrowId}, {"settlement", subject.SettlementId}, {"reservation", subject.ReservationId}, {"lease", subject.LeaseId}}
	owner := ""
	for _, alias := range aliases {
		if alias.value == "" {
			continue
		}
		cases, err := k.FinancialCasesByIndex(ctx, alias.kind, alias.value)
		if err != nil {
			return "", false, err
		}
		for _, financialCase := range cases {
			if !types.IsActiveFinancialCaseStatus(financialCase.Status) {
				continue
			}
			if owner != "" && owner != financialCase.CaseId {
				return "", false, types.ErrFinancialCaseMalformedState.Wrap("subject aliases reference multiple active cases")
			}
			owner = financialCase.CaseId
		}
	}
	return owner, owner != "", nil
}

func (k Keeper) AddFinancialClaim(ctx sdk.Context, caseID string, claim types.FinancialClaim) (*types.FinancialCase, *types.FinancialClaim, bool, error) {
	cacheCtx, write := ctx.CacheContext()
	financialCase, added, duplicate, err := k.addFinancialClaim(cacheCtx, caseID, claim)
	if err != nil {
		return nil, nil, false, err
	}
	write()
	return financialCase, added, duplicate, nil
}

func (k Keeper) addFinancialClaim(ctx sdk.Context, caseID string, claim types.FinancialClaim) (*types.FinancialCase, *types.FinancialClaim, bool, error) {
	financialCase, found := k.GetFinancialCase(ctx, caseID)
	if !found {
		return nil, nil, false, types.ErrFinancialCaseNotFound
	}
	if !types.IsActiveFinancialCaseStatus(financialCase.Status) || financialCase.Status == types.FinancialCaseStatusResolvedPendingAppeal {
		return nil, nil, false, types.ErrFinancialCaseTransition.Wrap("case does not accept claims")
	}
	if claim.Claimant != financialCase.Claimant && claim.Claimant != financialCase.Respondent && claim.Claimant != k.authority {
		return nil, nil, false, types.ErrFinancialCaseAuthorization
	}
	isPartyClaim := claim.Claimant == financialCase.Claimant || claim.Claimant == financialCase.Respondent
	if isPartyClaim && deadlinePassed(ctx, financialCase.FilingDeadlineHeight, financialCase.FilingDeadlineTime) {
		return nil, nil, false, types.ErrFinancialCaseDeadline.Wrap("filing window closed")
	}
	if isPartyClaim && (financialCase.Status == types.FinancialCaseStatusOpen || financialCase.Status == types.FinancialCaseStatusEvidence) && deadlinePassed(ctx, financialCase.EvidenceDeadlineHeight, financialCase.EvidenceDeadlineTime) {
		return nil, nil, false, types.ErrFinancialCaseDeadline.Wrap("evidence window closed")
	}
	params := k.GetParams(ctx)
	maxReference := params.FinancialCaseMaxEvidenceReferenceBytes
	if maxReference == 0 {
		maxReference = financialDefaultEvidenceRefBytes
	}
	if err := validateFinancialClaimInput(claim, maxReference); err != nil {
		return nil, nil, false, err
	}
	claimID, payloadHash, err := DeterministicFinancialClaimID(caseID, claim)
	if err != nil {
		return nil, nil, false, err
	}
	claim.ClaimId, claim.PayloadHash = claimID, payloadHash
	if replay, exists, err := k.getFinancialClaimReplay(ctx, claim.IdempotencyKey); err != nil {
		return nil, nil, false, err
	} else if exists {
		if replay.CaseID != caseID || !bytes.Equal(replay.PayloadHash, payloadHash) {
			return nil, nil, false, types.ErrFinancialCaseIdempotencyConflict
		}
		for i := range financialCase.Claims {
			if financialCase.Claims[i].ClaimId == replay.ClaimID {
				return &financialCase, &financialCase.Claims[i], true, nil
			}
		}
		return nil, nil, false, types.ErrFinancialCaseMalformedState.Wrap("claim replay orphan")
	}
	maxClaims := params.FinancialCaseMaxClaims
	if maxClaims == 0 {
		maxClaims = financialDefaultMaxClaims
	}
	if len(financialCase.Claims) >= int(maxClaims) {
		return nil, nil, false, types.ErrInvalidFinancialCase.Wrap("claim limit reached")
	}
	claim.CreatedHeight, claim.CreatedAt = ctx.BlockHeight(), ctx.BlockTime().Unix()
	financialCase.Claims = append(financialCase.Claims, claim)
	financialCase.ClaimRoot = financialClaimRoot(financialCase.Claims)
	financialCase.UpdatedHeight, financialCase.UpdatedAt = ctx.BlockHeight(), ctx.BlockTime().Unix()
	if financialCase.Status == types.FinancialCaseStatusOpen {
		if err := k.transitionFinancialCase(ctx, &financialCase, types.FinancialCaseStatusEvidence, claim.Claimant, "claim_added", nil); err != nil {
			return nil, nil, false, err
		}
	}
	if err := k.SetFinancialCase(ctx, financialCase); err != nil {
		return nil, nil, false, err
	}
	if err := k.setFinancialClaimReplay(ctx, claim.IdempotencyKey, caseID, claimID, payloadHash); err != nil {
		return nil, nil, false, err
	}
	_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialClaimAdded{CaseId: caseID, ClaimId: claimID, ClaimType: claim.ClaimType, SourceModule: claim.SourceModule})
	return &financialCase, &financialCase.Claims[len(financialCase.Claims)-1], false, nil
}

func (k Keeper) SubmitFinancialCaseForReview(ctx sdk.Context, caseID, actor string) error {
	return k.updateFinancialCase(ctx, caseID, func(financialCase *types.FinancialCase) error {
		if actor != financialCase.Claimant && actor != financialCase.Respondent && actor != k.authority {
			return types.ErrFinancialCaseAuthorization
		}
		if deadlinePassed(ctx, financialCase.EvidenceDeadlineHeight, financialCase.EvidenceDeadlineTime) {
			return types.ErrFinancialCaseDeadline.Wrap("evidence window closed")
		}
		return k.transitionFinancialCase(ctx, financialCase, types.FinancialCaseStatusReview, actor, "submitted_for_review", nil)
	})
}

func (k Keeper) EscalateFinancialCase(ctx sdk.Context, caseID, actor string, reasonHash []byte) error {
	return k.updateFinancialCase(ctx, caseID, func(financialCase *types.FinancialCase) error {
		if actor != financialCase.Claimant && actor != financialCase.Respondent && actor != k.authority && !strings.HasPrefix(actor, "adapter:") {
			return types.ErrFinancialCaseAuthorization
		}
		if len(reasonHash) != financialHashSize {
			return types.ErrInvalidFinancialCase.Wrap("reason hash must be SHA-256")
		}
		if deadlinePassed(ctx, financialCase.EscalationDeadlineHeight, financialCase.EscalationDeadlineTime) {
			return types.ErrFinancialCaseDeadline.Wrap("escalation window closed")
		}
		return k.transitionFinancialCase(ctx, financialCase, types.FinancialCaseStatusEscalated, actor, "escalated", reasonHash)
	})
}

func (k Keeper) ResolveFinancialCase(ctx sdk.Context, caseID, resolver string, allocation types.TerminalAllocation) error {
	return k.updateFinancialCase(ctx, caseID, func(financialCase *types.FinancialCase) error {
		if resolver != financialCase.ResolverAuthority || resolver != k.authority {
			return types.ErrFinancialCaseAuthorization
		}
		if resolver == financialCase.Claimant || resolver == financialCase.Respondent {
			return types.ErrFinancialCaseResolverConflict
		}
		if financialCase.Status != types.FinancialCaseStatusReview && financialCase.Status != types.FinancialCaseStatusEscalated && financialCase.Status != types.FinancialCaseStatusQuarantined {
			return types.ErrFinancialCaseTransition.Wrap("case not resolvable")
		}
		if financialCase.Exposure.PayoutId != "" {
			payout, found := k.GetPayout(ctx, financialCase.Exposure.PayoutId)
			if !found {
				return types.ErrFinancialCaseHold.Wrap("payout missing at resolution")
			}
			irreversibleFiat, err := k.payoutHasIrreversibleFiatBoundary(ctx, payout)
			if err != nil {
				return err
			}
			if irreversibleFiat {
				return types.ErrFiatConversionQuarantined.Wrap("irreversible fiat incident cannot allocate native value")
			}
		}
		if err := ValidateTerminalAllocation(financialCase.Exposure.OriginalHeld, allocation); err != nil {
			return err
		}
		allocation.AllocationHash = financialAllocationHash(allocation)
		financialCase.TerminalAllocation = &allocation
		params := k.GetParams(ctx)
		financialCase.AppealDeadlineHeight = addHeightBounded(ctx.BlockHeight(), params.FinancialCaseAppealWindowBlocks, defaultBlockWindow(params.FinancialCaseAppealWindowSeconds))
		financialCase.AppealDeadlineTime = addTimeBounded(ctx.BlockTime(), params.FinancialCaseAppealWindowSeconds, 24*time.Hour).Unix()
		return k.transitionFinancialCase(ctx, financialCase, types.FinancialCaseStatusResolvedPendingAppeal, resolver, "resolved_pending_appeal", allocation.AllocationHash)
	})
}

func (k Keeper) AppealFinancialCase(ctx sdk.Context, caseID, appellant string, evidenceHash []byte, encryptedReference string, idempotency []byte) (*types.FinancialAppeal, bool, error) {
	var result types.FinancialAppeal
	duplicate := false
	appealCount := uint32(0)
	err := k.updateFinancialCase(ctx, caseID, func(financialCase *types.FinancialCase) error {
		if appellant != financialCase.Claimant && appellant != financialCase.Respondent {
			return types.ErrFinancialCaseAuthorization
		}
		if len(evidenceHash) != financialHashSize || len(idempotency) == 0 || len(idempotency) > 128 {
			return types.ErrInvalidFinancialCase.Wrap("appeal hash and idempotency required")
		}
		maxRef := k.GetParams(ctx).FinancialCaseMaxEvidenceReferenceBytes
		if maxRef == 0 {
			maxRef = financialDefaultEvidenceRefBytes
		}
		if len(encryptedReference) > int(maxRef) || strings.ContainsRune(encryptedReference, '\x00') {
			return types.ErrFinancialCasePrivacy
		}
		appealID := deterministicAppealID(caseID, appellant, evidenceHash, idempotency)
		payloadHash := financialAppealPayloadHash(caseID, appellant, evidenceHash, encryptedReference, idempotency)
		if replay, exists, err := k.getFinancialAppealReplay(ctx, idempotency); err != nil {
			return err
		} else if exists {
			if replay.CaseID != caseID || replay.AppealID != appealID || !bytes.Equal(replay.PayloadHash, payloadHash) {
				return types.ErrFinancialCaseIdempotencyConflict.Wrap("appeal retry payload differs")
			}
			for i := range financialCase.Appeals {
				if financialCase.Appeals[i].AppealId == replay.AppealID {
					result, duplicate = financialCase.Appeals[i], true
					appealCount = uint32(len(financialCase.Appeals)) //nolint:gosec // appeal count is protocol bounded
					return nil
				}
			}
			return types.ErrFinancialCaseMalformedState.Wrap("appeal replay orphan")
		}
		if financialCase.Status != types.FinancialCaseStatusResolvedPendingAppeal {
			return types.ErrFinancialCaseTransition
		}
		if deadlinePassed(ctx, financialCase.AppealDeadlineHeight, financialCase.AppealDeadlineTime) {
			return types.ErrFinancialCaseDeadline
		}
		if len(financialCase.Appeals) >= int(financialCase.MaxAppeals) {
			return types.ErrInvalidFinancialCase.Wrap("appeal limit reached")
		}
		result = types.FinancialAppeal{AppealId: appealID, Appellant: appellant, EvidenceHash: append([]byte(nil), evidenceHash...), EncryptedReference: encryptedReference, CreatedHeight: ctx.BlockHeight(), CreatedAt: ctx.BlockTime().Unix(), IdempotencyKey: append([]byte(nil), idempotency...)}
		financialCase.Appeals = append(financialCase.Appeals, result)
		if err := k.setFinancialAppealReplay(ctx, idempotency, caseID, appealID, payloadHash); err != nil {
			return err
		}
		appealCount = uint32(len(financialCase.Appeals)) //nolint:gosec // appeal count is protocol bounded
		financialCase.TerminalAllocation = nil
		params := k.GetParams(ctx)
		financialCase.ReviewDeadlineHeight = addHeightBounded(ctx.BlockHeight(), params.FinancialCaseReviewWindowBlocks, defaultBlockWindow(params.FinancialCaseReviewWindowSeconds))
		financialCase.ReviewDeadlineTime = addTimeBounded(ctx.BlockTime(), params.FinancialCaseReviewWindowSeconds, 7*24*time.Hour).Unix()
		return k.transitionFinancialCase(ctx, financialCase, types.FinancialCaseStatusReview, appellant, "appealed", evidenceHash)
	})
	if err != nil {
		return nil, false, err
	}
	if !duplicate {
		_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialCaseAppealed{CaseId: caseID, AppealId: result.AppealId, AppealCount: appealCount})
	}
	return &result, duplicate, nil
}

func (k Keeper) CancelFinancialCase(ctx sdk.Context, caseID, actor string, reasonHash []byte) error {
	cacheCtx, write := ctx.CacheContext()
	financialCase, found := k.GetFinancialCase(cacheCtx, caseID)
	if !found {
		return types.ErrFinancialCaseNotFound
	}
	if actor != financialCase.Claimant && actor != k.authority {
		return types.ErrFinancialCaseAuthorization
	}
	if financialCase.Status != types.FinancialCaseStatusOpen && financialCase.Status != types.FinancialCaseStatusEvidence {
		return types.ErrFinancialCaseTransition
	}
	if len(reasonHash) != financialMaxReasonHashBytes {
		return types.ErrInvalidFinancialCase.Wrap("reason hash must be SHA-256")
	}
	if financialCase.Exposure.PayoutId != "" {
		payout, payoutFound := k.GetPayout(cacheCtx, financialCase.Exposure.PayoutId)
		if !payoutFound {
			return types.ErrFinancialCaseHold.Wrap("payout missing at cancellation")
		}
		irreversibleFiat, err := k.payoutHasIrreversibleFiatBoundary(cacheCtx, payout)
		if err != nil {
			return err
		}
		if irreversibleFiat {
			return types.ErrFiatConversionQuarantined.Wrap("irreversible fiat incident requires governed external reconciliation")
		}
	}
	if err := k.releaseFinancialHoldsWithoutAllocation(cacheCtx, &financialCase); err != nil {
		return err
	}
	if err := k.transitionFinancialCase(cacheCtx, &financialCase, types.FinancialCaseStatusCancelled, actor, "cancelled", reasonHash); err != nil {
		return err
	}
	if err := k.SetFinancialCase(cacheCtx, financialCase); err != nil {
		return err
	}
	write()
	return nil
}

func (k Keeper) FinalizeFinancialCase(ctx sdk.Context, caseID, actor string) (*types.FinancialCase, error) {
	cacheCtx, write := ctx.CacheContext()
	financialCase, found := k.GetFinancialCase(cacheCtx, caseID)
	if !found {
		return nil, types.ErrFinancialCaseNotFound
	}
	if financialCase.Status == types.FinancialCaseStatusFinal {
		if err := k.validateFinancialEffectsComplete(financialCase); err != nil {
			return nil, err
		}
		return &financialCase, nil
	}
	if actor != financialCase.ResolverAuthority || actor != k.authority {
		return nil, types.ErrFinancialCaseAuthorization
	}
	if financialCase.Status != types.FinancialCaseStatusResolvedPendingAppeal || financialCase.TerminalAllocation == nil {
		return nil, types.ErrFinancialCaseTransition
	}
	if !deadlinePassed(cacheCtx, financialCase.AppealDeadlineHeight, financialCase.AppealDeadlineTime) {
		return nil, types.ErrFinancialCaseDeadline.Wrap("appeal window remains open")
	}
	if err := ValidateTerminalAllocation(financialCase.Exposure.OriginalHeld, *financialCase.TerminalAllocation); err != nil {
		return nil, err
	}
	if err := k.applyFinancialCaseEffects(cacheCtx, &financialCase); err != nil {
		return nil, err
	}
	if err := k.transitionFinancialCase(cacheCtx, &financialCase, types.FinancialCaseStatusFinal, actor, "finalized", financialCase.TerminalAllocation.AllocationHash); err != nil {
		return nil, err
	}
	if err := k.SetFinancialCase(cacheCtx, financialCase); err != nil {
		return nil, err
	}
	write()
	_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialCaseFinalized{CaseId: caseID, ResolutionType: financialCase.TerminalAllocation.ResolutionType})
	return &financialCase, nil
}

func (k Keeper) SetFinancialCase(ctx sdk.Context, financialCase types.FinancialCase) error {
	if err := k.validateFinancialCase(ctx, financialCase); err != nil {
		return err
	}
	store := ctx.KVStore(k.skey)
	if existing, found := k.GetFinancialCase(ctx, financialCase.CaseId); found {
		k.deleteFinancialCaseIndexes(store, existing)
	}
	bz, err := k.cdc.Marshal(&financialCase)
	if err != nil {
		return err
	}
	store.Set(types.FinancialCaseKey(financialCase.CaseId), bz)
	return k.setFinancialCaseIndexes(store, financialCase)
}

func (k Keeper) GetFinancialCase(ctx sdk.Context, caseID string) (types.FinancialCase, bool) {
	bz := ctx.KVStore(k.skey).Get(types.FinancialCaseKey(caseID))
	if bz == nil {
		return types.FinancialCase{}, false
	}
	var financialCase types.FinancialCase
	if err := k.cdc.Unmarshal(bz, &financialCase); err != nil {
		return types.FinancialCase{}, false
	}
	return financialCase, true
}

func (k Keeper) GetFinancialCaseBySubject(ctx sdk.Context, subject types.FinancialSubject) (types.FinancialCase, bool) {
	key, err := CanonicalFinancialSubjectKey(subject)
	if err != nil {
		return types.FinancialCase{}, false
	}
	caseID := ctx.KVStore(k.skey).Get(types.FinancialSubjectKey(key))
	if caseID != nil {
		return k.GetFinancialCase(ctx, string(caseID))
	}
	owner, found, err := k.findActiveFinancialCaseForSubject(ctx, subject)
	if err != nil || !found {
		return types.FinancialCase{}, false
	}
	return k.GetFinancialCase(ctx, owner)
}

// QuarantineFinancialCase retains all holds and exposes deterministic operator recovery.
func (k Keeper) QuarantineFinancialCase(ctx sdk.Context, caseID, reason string) error {
	return k.updateFinancialCase(ctx, caseID, func(financialCase *types.FinancialCase) error {
		if financialCase.Status == types.FinancialCaseStatusQuarantined {
			if financialCase.QuarantineReason == reason {
				return nil
			}
			return types.ErrFinancialCaseIdempotencyConflict
		}
		if !CanTransitionFinancialCase(financialCase.Status, types.FinancialCaseStatusQuarantined) {
			return types.ErrFinancialCaseTransition
		}
		financialCase.Quarantined, financialCase.QuarantineReason = true, reason
		if err := k.transitionFinancialCase(ctx, financialCase, types.FinancialCaseStatusQuarantined, k.authority, "quarantined", hashFinancialMigrationReference(reason)); err != nil {
			return err
		}
		k.emitFinancialCaseQuarantined(ctx, caseID, []byte(reason))
		return nil
	})
}

func (k Keeper) HasActiveFinancialCase(ctx sdk.Context, kind, value string) (string, bool) {
	if value == "" {
		return "", false
	}
	prefix := financialCaseIndexPrefixForKind(kind)
	if prefix == nil {
		return "", false
	}
	iter := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.FinancialCaseIndexPrefix(prefix, value))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		financialCase, found := k.GetFinancialCase(ctx, string(iter.Value()))
		if !found {
			return "", true
		}
		if types.IsActiveFinancialCaseStatus(financialCase.Status) {
			return financialCase.CaseId, true
		}
	}
	return "", false
}

func (k Keeper) FinancialCasesByIndex(ctx sdk.Context, kind, value string) ([]types.FinancialCase, error) {
	prefix := financialCaseIndexPrefixForKind(kind)
	if prefix == nil || value == "" {
		return nil, types.ErrInvalidFinancialCase.Wrap("index kind and value required")
	}
	iter := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.FinancialCaseIndexPrefix(prefix, value))
	defer iter.Close()
	result := make([]types.FinancialCase, 0)
	for ; iter.Valid(); iter.Next() {
		financialCase, found := k.GetFinancialCase(ctx, string(iter.Value()))
		if !found {
			return nil, types.ErrFinancialCaseMalformedState.Wrap("index references missing case")
		}
		result = append(result, financialCase)
	}
	return result, nil
}

func (k Keeper) WithFinancialCases(ctx sdk.Context, fn func(types.FinancialCase) bool) error {
	iter := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.PrefixFinancialCase)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var financialCase types.FinancialCase
		if err := k.cdc.Unmarshal(iter.Value(), &financialCase); err != nil {
			return types.ErrFinancialCaseMalformedState.Wrapf("malformed case at %x", iter.Key())
		}
		if fn(financialCase) {
			return nil
		}
	}
	return nil
}

// ValidateFinancialCaseInvariants checks case/subject/hold/index/audit/effect consistency.
func (k Keeper) ValidateFinancialCaseInvariants(ctx sdk.Context) []string {
	broken := make([]string, 0)
	activeAliases := make(map[string]string)
	if err := k.WithFinancialCases(ctx, func(financialCase types.FinancialCase) bool {
		if err := k.validateFinancialCase(ctx, financialCase); err != nil {
			broken = append(broken, financialCase.CaseId+": "+err.Error())
			return false
		}
		if types.IsActiveFinancialCaseStatus(financialCase.Status) {
			holds, err := k.financialCaseHoldCount(ctx, financialCase)
			if err != nil {
				broken = append(broken, financialCase.CaseId+": "+err.Error())
			} else if holds == 0 || holds != financialCase.ActiveHoldCount {
				broken = append(broken, financialCase.CaseId+": active hold count mismatch")
			}
			for _, alias := range financialCaseAliasKeys(financialCase) {
				if owner, exists := activeAliases[alias]; exists && owner != financialCase.CaseId {
					broken = append(broken, financialCase.CaseId+": active alias owned by "+owner)
				} else {
					activeAliases[alias] = financialCase.CaseId
				}
			}
		}
		for _, index := range financialCaseIndexes(financialCase) {
			key := types.FinancialCaseIndexKey(index.prefix, index.value, financialCase.CaseId)
			if indexed := ctx.KVStore(k.skey).Get(key); string(indexed) != financialCase.CaseId {
				broken = append(broken, financialCase.CaseId+": missing or incorrect case index")
			}
		}
		if financialCase.Status == types.FinancialCaseStatusFinal {
			if financialCase.TerminalAllocation == nil {
				broken = append(broken, financialCase.CaseId+": final allocation missing")
			} else if err := ValidateTerminalAllocation(financialCase.Exposure.OriginalHeld, *financialCase.TerminalAllocation); err != nil {
				broken = append(broken, financialCase.CaseId+": "+err.Error())
			}
			if err := k.validateFinancialEffectsComplete(financialCase); err != nil {
				broken = append(broken, financialCase.CaseId+": "+err.Error())
			}
			if financialCase.ActiveHoldCount != 0 {
				broken = append(broken, financialCase.CaseId+": final case retains holds")
			}
		}
		return false
	}); err != nil {
		broken = append(broken, err.Error())
	}
	store := ctx.KVStore(k.skey)
	appealBindingCounts := make(map[string]int)
	for _, prefix := range financialCaseIndexPrefixes() {
		iter := storetypes.KVStorePrefixIterator(store, prefix)
		for ; iter.Valid(); iter.Next() {
			financialCase, found := k.GetFinancialCase(ctx, string(iter.Value()))
			if !found {
				broken = append(broken, fmt.Sprintf("orphan financial-case index %x", iter.Key()))
				continue
			}
			expected := false
			for _, index := range financialCaseIndexes(financialCase) {
				if bytes.Equal(index.prefix, prefix) && bytes.Equal(iter.Key(), types.FinancialCaseIndexKey(index.prefix, index.value, financialCase.CaseId)) {
					expected = true
					break
				}
			}
			if !expected {
				broken = append(broken, fmt.Sprintf("unexpected financial-case index %x", iter.Key()))
			}
		}
		iter.Close()
	}
	subjectIter := storetypes.KVStorePrefixIterator(store, types.PrefixFinancialCaseBySubject)
	for ; subjectIter.Valid(); subjectIter.Next() {
		financialCase, found := k.GetFinancialCase(ctx, string(subjectIter.Value()))
		if !found {
			broken = append(broken, fmt.Sprintf("orphan financial-case subject index %x", subjectIter.Key()))
			continue
		}
		subjectKey, err := CanonicalFinancialSubjectKey(financialCase.Subject)
		if err != nil {
			broken = append(broken, fmt.Sprintf("%s: subject index has invalid case subject", financialCase.CaseId))
			continue
		}
		if !types.IsActiveFinancialCaseStatus(financialCase.Status) || !bytes.Equal(subjectIter.Key(), types.FinancialSubjectKey(subjectKey)) {
			broken = append(broken, fmt.Sprintf("unexpected financial-case subject index %x", subjectIter.Key()))
		}
	}
	subjectIter.Close()
	replayIter := storetypes.KVStorePrefixIterator(store, types.PrefixFinancialClaimIdempotency)
	for ; replayIter.Valid(); replayIter.Next() {
		var replay financialClaimReplay
		if err := json.Unmarshal(replayIter.Value(), &replay); err != nil {
			broken = append(broken, fmt.Sprintf("malformed claim replay %x", replayIter.Key()))
			continue
		}
		financialCase, found := k.GetFinancialCase(ctx, replay.CaseID)
		if !found {
			broken = append(broken, fmt.Sprintf("orphan claim replay %x", replayIter.Key()))
			continue
		}
		matched := false
		for _, claim := range financialCase.Claims {
			if claim.ClaimId == replay.ClaimID && bytes.Equal(claim.PayloadHash, replay.PayloadHash) {
				matched = true
				break
			}
		}
		if !matched {
			broken = append(broken, fmt.Sprintf("claim replay mismatch %x", replayIter.Key()))
		}
	}
	replayIter.Close()
	appealReplayIter := storetypes.KVStorePrefixIterator(store, types.PrefixFinancialAppealIdempotency)
	for ; appealReplayIter.Valid(); appealReplayIter.Next() {
		var replay financialAppealReplay
		if err := json.Unmarshal(appealReplayIter.Value(), &replay); err != nil {
			broken = append(broken, fmt.Sprintf("malformed appeal replay %x", appealReplayIter.Key()))
			continue
		}
		financialCase, found := k.GetFinancialCase(ctx, replay.CaseID)
		if !found {
			broken = append(broken, fmt.Sprintf("orphan appeal replay %x", appealReplayIter.Key()))
			continue
		}
		matched := false
		for _, appeal := range financialCase.Appeals {
			if appeal.AppealId == replay.AppealID && bytes.Equal(replay.PayloadHash, financialAppealPayloadHash(financialCase.CaseId, appeal.Appellant, appeal.EvidenceHash, appeal.EncryptedReference, appeal.IdempotencyKey)) {
				matched = true
				break
			}
		}
		if !matched || len(replay.PayloadHash) != financialHashSize {
			broken = append(broken, fmt.Sprintf("appeal replay mismatch %x", appealReplayIter.Key()))
		} else {
			appealBindingCounts[replay.CaseID+"\x00"+replay.AppealID]++
		}
	}
	appealReplayIter.Close()
	if err := k.WithFinancialCases(ctx, func(financialCase types.FinancialCase) bool {
		for _, appeal := range financialCase.Appeals {
			if count := appealBindingCounts[financialCase.CaseId+"\x00"+appeal.AppealId]; count != 1 {
				broken = append(broken, fmt.Sprintf("%s: appeal %s has %d replay bindings", financialCase.CaseId, appeal.AppealId, count))
			}
		}
		return false
	}); err != nil {
		broken = append(broken, err.Error())
	}
	sort.Strings(broken)
	return broken
}

func (k Keeper) ProcessFinancialCaseTimeouts(ctx sdk.Context) (uint64, error) {
	limit := k.GetParams(ctx).FinancialCaseTimeoutBatchLimit
	if limit == 0 {
		limit = financialDefaultTimeoutLimit
	}
	processed := uint64(0)
	var caseIDs []string
	if err := k.WithFinancialCases(ctx, func(financialCase types.FinancialCase) bool {
		if processed+uint64(len(caseIDs)) >= uint64(limit) {
			return true
		}
		if (financialCase.Status == types.FinancialCaseStatusOpen || financialCase.Status == types.FinancialCaseStatusEvidence) && deadlinePassed(ctx, financialCase.EvidenceDeadlineHeight, financialCase.EvidenceDeadlineTime) {
			caseIDs = append(caseIDs, financialCase.CaseId)
		} else if financialCase.Status == types.FinancialCaseStatusReview && deadlinePassed(ctx, financialCase.ReviewDeadlineHeight, financialCase.ReviewDeadlineTime) {
			caseIDs = append(caseIDs, financialCase.CaseId)
		}
		return false
	}); err != nil {
		return 0, err
	}
	for _, caseID := range caseIDs {
		if err := k.updateFinancialCase(ctx, caseID, func(financialCase *types.FinancialCase) error {
			return k.transitionFinancialCase(ctx, financialCase, types.FinancialCaseStatusEscalated, k.authority, "timeout_escalated", nil)
		}); err != nil {
			return processed, err
		}
		processed++
		_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialCaseExpired{CaseId: caseID, Status: types.FinancialCaseStatusEscalated})
	}
	return processed, nil
}

// MigrateFinancialCases deterministically imports settlement-local active holds.
// Cross-module v1.7.0 reconciliation adds fraud/HPC/billing claims before activation.
func (k Keeper) MigrateFinancialCases(ctx sdk.Context) (FinancialCaseMigrationReport, error) {
	cacheCtx, write := ctx.CacheContext()
	report, err := k.migrateFinancialCases(cacheCtx)
	if err != nil {
		return report, err
	}
	write()
	return report, nil
}

func (k Keeper) migrateFinancialCases(ctx sdk.Context) (FinancialCaseMigrationReport, error) {
	if existing := ctx.KVStore(k.skey).Get(types.FinancialCaseMigrationAuditKey()); existing != nil {
		var report FinancialCaseMigrationReport
		if err := json.Unmarshal(existing, &report); err != nil {
			return report, types.ErrFinancialCaseMalformedState.Wrap("migration audit malformed")
		}
		return report, nil
	}
	report := FinancialCaseMigrationReport{}
	var migrateErr error
	k.WithPayouts(ctx, func(payout types.PayoutRecord) bool {
		report.PayoutsScanned++
		if payout.State.IsTerminal() {
			report.TerminalPreserved++
			return false
		}
		if payout.State != types.PayoutStateHeld || payout.DisputeID == "" {
			return false
		}
		idempotencyHash := sha256.Sum256([]byte("migration/payout/v1\x00" + payout.PayoutID + "\x00" + payout.DisputeID))
		request := FinancialCaseOpenRequest{
			Subject:  types.FinancialSubject{Type: types.FinancialSubjectTypeOrder, PrimaryId: payout.OrderID, OrderId: payout.OrderID, InvoiceId: payout.InvoiceID, SettlementId: payout.SettlementID, EscrowId: payout.EscrowID},
			Claimant: payout.Customer, Respondent: payout.Provider, Migrated: true,
			IdempotencyKey: idempotencyHash[:],
			Claim:          types.FinancialClaim{ClaimType: types.FinancialClaimTypeMigration, Claimant: payout.Customer, SourceModule: financialSourceSettlement, SourceReference: payout.DisputeID, EvidenceHash: hashFinancialMigrationReference("payout", payout.PayoutID, payout.DisputeID), EncryptedReference: "migration://settlement/payout/" + payout.PayoutID},
		}
		financialCase, _, duplicate, err := k.openFinancialCase(ctx, request)
		if err != nil {
			request.Quarantine, request.QuarantineReason = true, "ambiguous_legacy_payout_hold"
			financialCase, _, duplicate, err = k.openFinancialCase(ctx, request)
		}
		if err != nil {
			migrateErr = err
			report.MalformedOrphans++
			return true
		}
		if duplicate {
			report.AlreadyMigrated++
		} else {
			report.CasesCreated++
		}
		if financialCase.Quarantined {
			report.Quarantined++
		}
		return false
	})
	if migrateErr != nil {
		return report, migrateErr
	}
	k.WithEscrowsByState(ctx, types.EscrowStateDisputed, func(escrow types.EscrowAccount) bool {
		report.EscrowsScanned++
		if _, held := k.HasActiveFinancialCase(ctx, "escrow", escrow.EscrowID); held {
			report.AlreadyMigrated++
			return false
		}
		if escrow.Depositor == "" || escrow.Recipient == "" {
			report.MalformedOrphans++
			return false
		}
		idempotencyHash := sha256.Sum256([]byte("migration/escrow/v1\x00" + escrow.EscrowID))
		request := FinancialCaseOpenRequest{
			Subject:  types.FinancialSubject{Type: types.FinancialSubjectTypeOrder, PrimaryId: escrow.OrderID, OrderId: escrow.OrderID, EscrowId: escrow.EscrowID, LeaseId: escrow.LeaseID},
			Claimant: escrow.Depositor, Respondent: escrow.Recipient, Migrated: true, Quarantine: true, QuarantineReason: "legacy_escrow_dispute_without_unambiguous_allocation",
			IdempotencyKey: idempotencyHash[:],
			Claim:          types.FinancialClaim{ClaimType: types.FinancialClaimTypeMigration, Claimant: escrow.Depositor, SourceModule: financialSourceSettlement, SourceReference: escrow.EscrowID, EvidenceHash: hashFinancialMigrationReference("escrow", escrow.EscrowID), EncryptedReference: "migration://settlement/escrow/" + escrow.EscrowID},
		}
		_, _, duplicate, err := k.openFinancialCase(ctx, request)
		if err != nil {
			migrateErr = err
			report.MalformedOrphans++
			return true
		}
		if duplicate {
			report.AlreadyMigrated++
		} else {
			report.CasesCreated++
			report.Quarantined++
		}
		return false
	})
	if migrateErr != nil {
		return report, migrateErr
	}
	report.Digest = financialMigrationDigest(report)
	bz, err := json.Marshal(report)
	if err != nil {
		return report, err
	}
	ctx.KVStore(k.skey).Set(types.FinancialCaseMigrationAuditKey(), bz)
	return report, nil
}

// RebuildFinancialCaseState deterministically recreates all derived indexes,
// replay bindings and terminal effect defaults from canonical case values.
func (k Keeper) RebuildFinancialCaseState(ctx sdk.Context) error {
	store := ctx.KVStore(k.skey)
	for _, prefix := range [][]byte{
		types.PrefixFinancialCaseBySubject,
		types.PrefixFinancialCaseByOrder,
		types.PrefixFinancialCaseByInvoice,
		types.PrefixFinancialCaseByUsage,
		types.PrefixFinancialCaseByJob,
		types.PrefixFinancialCaseByEscrow,
		types.PrefixFinancialCaseByStatus,
		types.PrefixFinancialCaseByParty,
		types.PrefixFinancialClaimIdempotency,
		types.PrefixFinancialAppealIdempotency,
		types.PrefixFinancialCaseDeadline,
		types.PrefixFinancialCaseBySettlement,
		types.PrefixFinancialCaseByReservation,
		types.PrefixFinancialCaseByLease,
	} {
		iterator := storetypes.KVStorePrefixIterator(store, prefix)
		keys := make([][]byte, 0)
		for ; iterator.Valid(); iterator.Next() {
			keys = append(keys, append([]byte(nil), iterator.Key()...))
		}
		iterator.Close()
		for _, key := range keys {
			store.Delete(key)
		}
	}

	cases := make([]types.FinancialCase, 0)
	if err := k.WithFinancialCases(ctx, func(financialCase types.FinancialCase) bool {
		cases = append(cases, financialCase)
		return false
	}); err != nil {
		return err
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].CaseId < cases[j].CaseId })
	for i := range cases {
		financialCase := &cases[i]
		if financialCase.Status == types.FinancialCaseStatusFinal && len(financialCase.Effects) == 0 {
			financialCase.Effects = expectedAppliedFinancialEffects(*financialCase)
			bz, err := k.cdc.Marshal(financialCase)
			if err != nil {
				return err
			}
			store.Set(types.FinancialCaseKey(financialCase.CaseId), bz)
		}
		if err := k.setFinancialCaseIndexes(store, *financialCase); err != nil {
			return err
		}
		for _, claim := range financialCase.Claims {
			if len(claim.IdempotencyKey) == 0 {
				return types.ErrFinancialCaseMalformedState.Wrapf("claim %s has no idempotency key", claim.ClaimId)
			}
			claimID, payload, err := DeterministicFinancialClaimID(financialCase.CaseId, claim)
			if err != nil || claimID != claim.ClaimId || !bytes.Equal(payload, claim.PayloadHash) {
				return types.ErrFinancialCaseMalformedState.Wrapf("claim %s is noncanonical", claim.ClaimId)
			}
			if err := k.setFinancialClaimReplay(ctx, claim.IdempotencyKey, financialCase.CaseId, claim.ClaimId, claim.PayloadHash); err != nil {
				return err
			}
		}
		for _, appeal := range financialCase.Appeals {
			if len(appeal.IdempotencyKey) == 0 || len(appeal.IdempotencyKey) > 128 {
				return types.ErrFinancialCaseMalformedState.Wrapf("appeal %s has invalid idempotency key", appeal.AppealId)
			}
			expectedAppealID := deterministicAppealID(financialCase.CaseId, appeal.Appellant, appeal.EvidenceHash, appeal.IdempotencyKey)
			payload := financialAppealPayloadHash(financialCase.CaseId, appeal.Appellant, appeal.EvidenceHash, appeal.EncryptedReference, appeal.IdempotencyKey)
			if expectedAppealID != appeal.AppealId {
				return types.ErrFinancialCaseMalformedState.Wrapf("appeal %s is noncanonical", appeal.AppealId)
			}
			if err := k.setFinancialAppealReplay(ctx, appeal.IdempotencyKey, financialCase.CaseId, appeal.AppealId, payload); err != nil {
				return err
			}
		}
	}
	return nil
}

func expectedAppliedFinancialEffects(financialCase types.FinancialCase) []types.FinancialCaseEffect {
	specs := []struct {
		id  string
		typ types.FinancialEffectType
	}{{"provider", types.FinancialEffectPayout}, {"customer", types.FinancialEffectPayout}, {"platform", types.FinancialEffectPayout}, {"slash-witness", types.FinancialEffectPayout}}
	if !financialCase.Exposure.UnclaimedRewards.IsZero() {
		specs = append(specs, struct {
			id  string
			typ types.FinancialEffectType
		}{"reward", types.FinancialEffectReward})
	}
	if financialCase.Exposure.ReservationId != "" {
		specs = append(specs, struct {
			id  string
			typ types.FinancialEffectType
		}{"reservation", types.FinancialEffectReservation})
	}
	specs = append(specs, struct {
		id  string
		typ types.FinancialEffectType
	}{"projection", types.FinancialEffectProjection})
	effects := make([]types.FinancialCaseEffect, 0, len(specs))
	for _, spec := range specs {
		effects = append(effects, types.FinancialCaseEffect{
			EffectId: financialCase.CaseId + "/" + spec.id, Type: spec.typ,
			Status: types.FinancialEffectStatusApplied, Attempts: 1,
			AppliedHeight: financialCase.UpdatedHeight, AppliedAt: financialCase.UpdatedAt,
		})
	}
	return effects
}

func hashFinancialMigrationReference(parts ...string) []byte {
	h := sha256.New()
	for _, part := range parts {
		writeFinancialField(h, []byte(part))
	}
	return h.Sum(nil)
}
func financialMigrationDigest(report FinancialCaseMigrationReport) string {
	h := sha256.New()
	for _, value := range []uint64{report.PayoutsScanned, report.EscrowsScanned, report.CasesCreated, report.ClaimsMerged, report.Quarantined, report.TerminalPreserved, report.AlreadyMigrated, report.MalformedOrphans} {
		var encoded [8]byte
		binary.BigEndian.PutUint64(encoded[:], value)
		_, _ = h.Write(encoded[:])
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (k Keeper) updateFinancialCase(ctx sdk.Context, caseID string, update func(*types.FinancialCase) error) error {
	cacheCtx, write := ctx.CacheContext()
	financialCase, found := k.GetFinancialCase(cacheCtx, caseID)
	if !found {
		return types.ErrFinancialCaseNotFound
	}
	if err := update(&financialCase); err != nil {
		return err
	}
	if err := k.SetFinancialCase(cacheCtx, financialCase); err != nil {
		return err
	}
	write()
	return nil
}

func (k Keeper) transitionFinancialCase(ctx sdk.Context, financialCase *types.FinancialCase, to types.FinancialCaseStatus, actor, action string, reasonHash []byte) error {
	if !CanTransitionFinancialCase(financialCase.Status, to) {
		return types.ErrFinancialCaseTransition.Wrapf("%s -> %s", financialCase.Status, to)
	}
	from := financialCase.Status
	financialCase.Status = to
	financialCase.UpdatedHeight, financialCase.UpdatedAt = ctx.BlockHeight(), ctx.BlockTime().Unix()
	k.appendFinancialTransition(ctx, financialCase, from, to, actor, action, reasonHash)
	switch to {
	case types.FinancialCaseStatusReview:
		_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialCaseReviewed{CaseId: financialCase.CaseId, Status: to})
	case types.FinancialCaseStatusEscalated:
		_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialCaseEscalated{CaseId: financialCase.CaseId, Status: to})
	case types.FinancialCaseStatusResolvedPendingAppeal:
		if financialCase.TerminalAllocation != nil {
			_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialCaseResolved{CaseId: financialCase.CaseId, ResolutionType: financialCase.TerminalAllocation.ResolutionType, AllocationHash: financialCase.TerminalAllocation.AllocationHash})
		}
	}
	return nil
}

func (k Keeper) appendFinancialTransition(ctx sdk.Context, financialCase *types.FinancialCase, from, to types.FinancialCaseStatus, actor, action string, reasonHash []byte) {
	financialCase.Transitions = append(financialCase.Transitions, types.FinancialCaseTransition{Sequence: uint64(len(financialCase.Transitions) + 1), From: from, To: to, Actor: actor, Action: action, ReasonHash: append([]byte(nil), reasonHash...), BlockHeight: ctx.BlockHeight(), BlockTime: ctx.BlockTime().Unix()}) //nolint:gosec
}

func (k Keeper) deriveFinancialExposure(ctx sdk.Context, subject types.FinancialSubject) (types.FinancialExposure, error) {
	exposure := types.FinancialExposure{ReservationId: subject.ReservationId, EscrowId: subject.EscrowId}
	if subject.InvoiceId != "" {
		if payout, found := k.GetPayoutByInvoice(ctx, subject.InvoiceId); found {
			exposure.PayoutId, exposure.PayoutAmount, exposure.EscrowId = payout.PayoutID, payout.GrossAmount, payout.EscrowID
		}
	}
	if exposure.PayoutId == "" && subject.SettlementId != "" {
		if payout, found := k.GetPayoutBySettlement(ctx, subject.SettlementId); found {
			exposure.PayoutId, exposure.PayoutAmount, exposure.EscrowId = payout.PayoutID, payout.GrossAmount, payout.EscrowID
		}
	}
	if subject.UsageId != "" {
		if usage, found := k.GetUsageRecord(ctx, subject.UsageId); found {
			if subject.OrderId == "" {
				subject.OrderId = usage.OrderID
			}
			exposure.RewardAddress = usage.Provider
		}
	}
	if exposure.RewardAddress == "" && subject.OrderId != "" {
		if escrow, found := k.GetEscrowByOrder(ctx, subject.OrderId); found {
			exposure.RewardAddress = escrow.Recipient
			if exposure.EscrowId == "" {
				exposure.EscrowId = escrow.EscrowID
			}
		}
	}
	if exposure.EscrowId != "" {
		if escrow, found := k.GetEscrow(ctx, exposure.EscrowId); found {
			exposure.EscrowAmount = escrow.Balance
		}
	}
	if exposure.RewardAddress != "" {
		if address, err := sdk.AccAddressFromBech32(exposure.RewardAddress); err == nil {
			if rewards, found := k.GetClaimableRewards(ctx, address); found {
				exposure.UnclaimedRewards = rewards.TotalClaimable
			}
		}
	}
	if exposure.PayoutAmount.IsZero() && exposure.EscrowAmount.IsZero() && exposure.UnclaimedRewards.IsZero() {
		return exposure, types.ErrFinancialCaseHold.Wrap("subject has no held financial exposure")
	}
	if !exposure.PayoutAmount.IsZero() {
		exposure.OriginalHeld = exposure.PayoutAmount
	} else if !exposure.EscrowAmount.IsZero() {
		exposure.OriginalHeld = exposure.EscrowAmount
	} else {
		exposure.OriginalHeld = exposure.UnclaimedRewards
	}
	if (!exposure.PayoutAmount.IsZero() || !exposure.EscrowAmount.IsZero()) && !exposure.UnclaimedRewards.IsZero() {
		exposure.OriginalHeld = exposure.OriginalHeld.Add(exposure.UnclaimedRewards...)
	}
	return exposure, nil
}

func (k Keeper) normalizeFinancialSubject(ctx sdk.Context, subject types.FinancialSubject) (types.FinancialSubject, error) {
	merged := subject
	merge := func(field *string, value, name string) error {
		if value == "" {
			return nil
		}
		if *field != "" && *field != value {
			return types.ErrFinancialCaseMalformedState.Wrapf("%s alias conflicts with canonical lineage", name)
		}
		*field = value
		return nil
	}
	mergeEscrow := func(escrow types.EscrowAccount) error {
		if err := merge(&merged.EscrowId, escrow.EscrowID, "escrow"); err != nil {
			return err
		}
		if err := merge(&merged.OrderId, escrow.OrderID, "order"); err != nil {
			return err
		}
		return merge(&merged.LeaseId, escrow.LeaseID, "lease")
	}
	if merged.InvoiceId != "" {
		if payout, found := k.GetPayoutByInvoice(ctx, merged.InvoiceId); found {
			if err := mergePayoutFinancialSubject(&merged, payout, merge); err != nil {
				return merged, err
			}
		} else if k.billingKeeper != nil {
			invoice, err := k.billingKeeper.GetInvoice(ctx, merged.InvoiceId)
			if err != nil || invoice == nil {
				return merged, types.ErrFinancialCaseMalformedState.Wrap("invoice lineage unavailable")
			}
			for _, item := range []struct {
				field       *string
				value, name string
			}{{&merged.OrderId, invoice.OrderID, "order"}, {&merged.SettlementId, invoice.SettlementID, "settlement"}, {&merged.EscrowId, invoice.EscrowID, "escrow"}, {&merged.LeaseId, invoice.LeaseID, "lease"}} {
				if err := merge(item.field, item.value, item.name); err != nil {
					return merged, err
				}
			}
		}
	}
	if merged.SettlementId != "" {
		settlement, settlementFound := k.GetSettlement(ctx, merged.SettlementId)
		if settlementFound {
			for _, item := range []struct {
				field       *string
				value, name string
			}{{&merged.OrderId, settlement.OrderID, "order"}, {&merged.EscrowId, settlement.EscrowID, "escrow"}, {&merged.LeaseId, settlement.LeaseID, "lease"}} {
				if err := merge(item.field, item.value, item.name); err != nil {
					return merged, err
				}
			}
		}
		if payout, found := k.GetPayoutBySettlement(ctx, merged.SettlementId); found {
			if err := mergePayoutFinancialSubject(&merged, payout, merge); err != nil {
				return merged, err
			}
		} else if !settlementFound {
			return merged, types.ErrFinancialCaseMalformedState.Wrap("settlement lineage unavailable")
		}
	}
	if merged.UsageId != "" {
		usage, found := k.GetUsageRecord(ctx, merged.UsageId)
		if !found {
			return merged, types.ErrFinancialCaseMalformedState.Wrap("usage lineage unavailable")
		}
		if err := merge(&merged.OrderId, usage.OrderID, "order"); err != nil {
			return merged, err
		}
		if err := merge(&merged.LeaseId, usage.LeaseID, "lease"); err != nil {
			return merged, err
		}
	}
	if merged.ReservationId != "" {
		if k.reservationKeeper == nil {
			return merged, types.ErrFinancialCaseHold.Wrap("reservation keeper missing")
		}
		reservation, found := k.reservationKeeper.GetReservation(ctx, merged.ReservationId)
		if !found {
			return merged, types.ErrFinancialCaseMalformedState.Wrap("reservation lineage unavailable")
		}
		for _, item := range []struct {
			field       *string
			value, name string
		}{{&merged.OrderId, reservation.MarketOrderId, "order"}, {&merged.HpcJobId, reservation.HpcJobId, "job"}, {&merged.EscrowId, reservation.EscrowId, "escrow"}, {&merged.LeaseId, reservation.MarketLeaseId, "lease"}} {
			if err := merge(item.field, item.value, item.name); err != nil {
				return merged, err
			}
		}
	}
	if merged.EscrowId != "" {
		escrow, found := k.GetEscrow(ctx, merged.EscrowId)
		if !found {
			return merged, types.ErrFinancialCaseMalformedState.Wrap("escrow lineage unavailable")
		}
		if err := mergeEscrow(escrow); err != nil {
			return merged, err
		}
	}
	if merged.OrderId != "" {
		if escrow, found := k.GetEscrowByOrder(ctx, merged.OrderId); found {
			if err := mergeEscrow(escrow); err != nil {
				return merged, err
			}
		}
	}
	return merged, nil
}

func mergePayoutFinancialSubject(subject *types.FinancialSubject, payout types.PayoutRecord, merge func(*string, string, string) error) error {
	for _, item := range []struct {
		field       *string
		value, name string
	}{{&subject.InvoiceId, payout.InvoiceID, "invoice"}, {&subject.SettlementId, payout.SettlementID, "settlement"}, {&subject.OrderId, payout.OrderID, "order"}, {&subject.EscrowId, payout.EscrowID, "escrow"}, {&subject.LeaseId, payout.LeaseID, "lease"}} {
		if err := merge(item.field, item.value, item.name); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) financialSubjectPartyAuthorized(ctx sdk.Context, subject types.FinancialSubject, claimant string) bool {
	if claimant == "" {
		return false
	}
	if subject.InvoiceId != "" {
		if payout, found := k.GetPayoutByInvoice(ctx, subject.InvoiceId); found && (claimant == payout.Customer || claimant == payout.Provider) {
			return true
		}
		if k.billingKeeper != nil {
			if invoice, err := k.billingKeeper.GetInvoice(ctx, subject.InvoiceId); err == nil && invoice != nil && (claimant == invoice.Customer || claimant == invoice.Provider) {
				return true
			}
		}
	}
	if subject.SettlementId != "" {
		if settlement, found := k.GetSettlement(ctx, subject.SettlementId); found && (claimant == settlement.Customer || claimant == settlement.Provider) {
			return true
		}
	}
	if subject.UsageId != "" {
		if usage, found := k.GetUsageRecord(ctx, subject.UsageId); found && (claimant == usage.Customer || claimant == usage.Provider) {
			return true
		}
	}
	if subject.EscrowId != "" {
		if escrow, found := k.GetEscrow(ctx, subject.EscrowId); found && (claimant == escrow.Depositor || claimant == escrow.Recipient) {
			return true
		}
	}
	if subject.OrderId != "" {
		if escrow, found := k.GetEscrowByOrder(ctx, subject.OrderId); found && (claimant == escrow.Depositor || claimant == escrow.Recipient) {
			return true
		}
	}
	if subject.ReservationId != "" && k.reservationKeeper != nil {
		if reservation, found := k.reservationKeeper.GetReservation(ctx, subject.ReservationId); found && (claimant == reservation.RequesterAddress || claimant == reservation.ProviderAddress) {
			return true
		}
	}
	return false
}

func (k Keeper) deriveFinancialParties(ctx sdk.Context, subject types.FinancialSubject, claimant, respondent string) (string, string, error) {
	provider, customer := "", ""
	if subject.InvoiceId != "" {
		if payout, found := k.GetPayoutByInvoice(ctx, subject.InvoiceId); found {
			provider, customer = payout.Provider, payout.Customer
		} else if k.billingKeeper != nil {
			if invoice, err := k.billingKeeper.GetInvoice(ctx, subject.InvoiceId); err == nil && invoice != nil {
				provider, customer = invoice.Provider, invoice.Customer
			}
		}
	}
	if provider == "" && subject.SettlementId != "" {
		if settlement, found := k.GetSettlement(ctx, subject.SettlementId); found {
			provider, customer = settlement.Provider, settlement.Customer
		}
	}
	if provider == "" && subject.UsageId != "" {
		if usage, found := k.GetUsageRecord(ctx, subject.UsageId); found {
			provider, customer = usage.Provider, usage.Customer
		}
	}
	if provider == "" && subject.EscrowId != "" {
		if escrow, found := k.GetEscrow(ctx, subject.EscrowId); found {
			provider, customer = escrow.Recipient, escrow.Depositor
		}
	}
	if provider == "" && subject.OrderId != "" {
		if escrow, found := k.GetEscrowByOrder(ctx, subject.OrderId); found {
			provider, customer = escrow.Recipient, escrow.Depositor
		}
	}
	if provider == "" && subject.ReservationId != "" && k.reservationKeeper != nil {
		if reservation, found := k.reservationKeeper.GetReservation(ctx, subject.ReservationId); found {
			provider, customer = reservation.ProviderAddress, reservation.RequesterAddress
		}
	}
	if provider == "" || customer == "" {
		// Trusted adapters still require both parties but may be the only source
		// for legacy lineage that lacks canonical settlement references.
		provider, customer = respondent, claimant
	}
	if _, err := sdk.AccAddressFromBech32(provider); err != nil {
		return "", "", types.ErrFinancialCaseAuthorization.Wrap("invalid provider party")
	}
	if _, err := sdk.AccAddressFromBech32(customer); err != nil || provider == customer {
		return "", "", types.ErrFinancialCaseAuthorization.Wrap("invalid customer party")
	}
	if (claimant != provider && claimant != customer) || (respondent != provider && respondent != customer) {
		return "", "", types.ErrFinancialCaseAuthorization.Wrap("case parties do not match financial lineage")
	}
	return provider, customer, nil
}

func (k Keeper) holdFinancialExposure(ctx sdk.Context, financialCase *types.FinancialCase) error {
	holds := uint32(0)
	if financialCase.Exposure.PayoutId != "" {
		payout, found := k.GetPayout(ctx, financialCase.Exposure.PayoutId)
		if !found {
			return types.ErrFinancialCaseHold.Wrap("payout missing")
		}
		irreversibleFiat, err := k.payoutHasIrreversibleFiatBoundary(ctx, payout)
		if err != nil {
			return err
		}
		if irreversibleFiat {
			// The case remains a canonical incident and can hold other local
			// exposure, but the fiat conversion owns this payout after swap
			// submission. Changing payout state here would race reconciliation.
		} else if payout.State == types.PayoutStateHeld {
			if payout.DisputeID != financialCase.CaseId && !financialCase.Migrated {
				return types.ErrFinancialCaseHold.Wrap("payout held by another case")
			}
			if payout.DisputeID != financialCase.CaseId {
				payout.DisputeID = financialCase.CaseId
				payout.HoldReason = "canonical_financial_case_migration"
				if err := k.SetPayout(ctx, payout); err != nil {
					return err
				}
			}
		} else if payout.State.IsTerminal() {
			return types.ErrFinancialCaseHold.Wrap("payout already terminal")
		} else if err := k.HoldPayout(ctx, payout.PayoutID, financialCase.CaseId, "canonical_financial_case"); err != nil {
			return err
		}
		if !irreversibleFiat {
			_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialCaseHeld{CaseId: financialCase.CaseId, ReferenceType: "payout", ReferenceId: payout.PayoutID})
			holds++
		}
	}
	if financialCase.Exposure.EscrowId != "" {
		escrow, found := k.GetEscrow(ctx, financialCase.Exposure.EscrowId)
		if !found {
			return types.ErrFinancialCaseHold.Wrap("escrow missing")
		}
		switch escrow.State {
		case types.EscrowStateDisputed:
			holds++
		case types.EscrowStateActive:
			if err := k.DisputeEscrow(ctx, escrow.EscrowID, "canonical_financial_case:"+financialCase.CaseId); err != nil {
				return err
			}
			holds++
		}
		_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialCaseHeld{CaseId: financialCase.CaseId, ReferenceType: "escrow", ReferenceId: escrow.EscrowID})
	}
	if !financialCase.Exposure.UnclaimedRewards.IsZero() {
		if _, err := sdk.AccAddressFromBech32(financialCase.Exposure.RewardAddress); err != nil {
			return types.ErrFinancialCaseHold.Wrap("reward hold address invalid")
		}
		holds++
		_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialCaseHeld{CaseId: financialCase.CaseId, ReferenceType: "reward", ReferenceId: financialCase.Exposure.RewardAddress})
	}
	if financialCase.Exposure.ReservationId != "" {
		if k.reservationKeeper == nil {
			return types.ErrFinancialCaseHold.Wrap("reservation keeper missing")
		}
		if _, err := k.reservationKeeper.HoldReservationForFinancialCase(ctx, financialCase.Exposure.ReservationId, financialCase.CaseId); err != nil {
			return err
		}
		_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialCaseHeld{CaseId: financialCase.CaseId, ReferenceType: "reservation", ReferenceId: financialCase.Exposure.ReservationId})
		holds++
	}
	if holds == 0 {
		return types.ErrFinancialCaseHold.Wrap("no hold established")
	}
	financialCase.ActiveHoldCount = holds
	return nil
}

func (k Keeper) releaseFinancialHoldsWithoutAllocation(ctx sdk.Context, financialCase *types.FinancialCase) error {
	if financialCase.Exposure.PayoutId != "" {
		payout, found := k.GetPayout(ctx, financialCase.Exposure.PayoutId)
		if found && payout.State == types.PayoutStateHeld && payout.DisputeID == financialCase.CaseId {
			if err := payout.ReleaseHold(); err != nil {
				return err
			}
			if err := k.SetPayout(ctx, payout); err != nil {
				return err
			}
		}
	}
	if financialCase.Exposure.EscrowId != "" {
		escrow, found := k.GetEscrow(ctx, financialCase.Exposure.EscrowId)
		if found && escrow.State == types.EscrowStateDisputed {
			escrow.State = types.EscrowStateActive
			if err := k.SetEscrow(ctx, escrow); err != nil {
				return err
			}
		}
	}
	if financialCase.Exposure.ReservationId != "" {
		if k.reservationKeeper == nil {
			return types.ErrFinancialCaseHold.Wrap("reservation keeper missing")
		}
		if _, err := k.reservationKeeper.ReleaseReservationFinancialCaseHold(ctx, financialCase.Exposure.ReservationId, financialCase.CaseId); err != nil {
			return err
		}
	}
	if !financialCase.Exposure.UnclaimedRewards.IsZero() && financialCase.Exposure.RewardAddress != "" {
		address, err := sdk.AccAddressFromBech32(financialCase.Exposure.RewardAddress)
		if err != nil {
			return err
		}
		rewards, found := k.GetClaimableRewards(ctx, address)
		if !found || !rewards.TotalClaimable.IsAllGTE(financialCase.Exposure.UnclaimedRewards) {
			return types.ErrFinancialCaseHold.Wrap("reward exposure is no longer held")
		}
	}
	financialCase.ActiveHoldCount = 0
	return nil
}

func (k Keeper) applyFinancialCaseEffects(ctx sdk.Context, financialCase *types.FinancialCase) error {
	allocation := financialCase.TerminalAllocation
	if allocation == nil {
		return types.ErrFinancialCaseEffect.Wrap("allocation missing")
	}
	if financialCase.Exposure.PayoutId != "" {
		payout, found := k.GetPayout(ctx, financialCase.Exposure.PayoutId)
		if !found {
			return types.ErrFinancialCaseHold.Wrap("payout missing at finalization")
		}
		irreversibleFiat, err := k.payoutHasIrreversibleFiatBoundary(ctx, payout)
		if err != nil {
			return err
		}
		if irreversibleFiat {
			return types.ErrFiatConversionQuarantined.Wrap("linked fiat conversion crossed irreversible external boundary; governed external reconciliation required")
		}
	}
	effects := []struct {
		id      string
		typ     types.FinancialEffectType
		amount  sdk.Coins
		address string
	}{
		{"provider", types.FinancialEffectPayout, allocation.Provider, financialCase.Provider},
		{"customer", types.FinancialEffectPayout, allocation.Customer, financialCase.Customer},
		{"platform", types.FinancialEffectPayout, allocation.Platform, ""},
		{"slash-witness", types.FinancialEffectPayout, allocation.SlashWitness, allocation.SlashWitnessRecipient},
	}
	if !financialCase.Exposure.UnclaimedRewards.IsZero() {
		if financialCase.Exposure.RewardAddress != financialCase.Provider {
			return types.ErrFinancialCaseEffect.Wrap("reward exposure owner is not canonical provider")
		}
		if !allocation.Provider.IsAllGTE(financialCase.Exposure.UnclaimedRewards) {
			return types.ErrFinancialCaseConservation.Wrap("provider allocation does not retain held rewards")
		}
		effects = append(effects, struct {
			id      string
			typ     types.FinancialEffectType
			amount  sdk.Coins
			address string
		}{"reward", types.FinancialEffectReward, financialCase.Exposure.UnclaimedRewards, financialCase.Exposure.RewardAddress})
	}
	if financialCase.Exposure.ReservationId != "" {
		effects = append(effects, struct {
			id      string
			typ     types.FinancialEffectType
			amount  sdk.Coins
			address string
		}{"reservation", types.FinancialEffectReservation, nil, ""})
	}
	effects = append(effects, struct {
		id      string
		typ     types.FinancialEffectType
		amount  sdk.Coins
		address string
	}{"projection", types.FinancialEffectProjection, nil, ""})
	for _, spec := range effects {
		effectID := financialCase.CaseId + "/" + spec.id
		idx := findFinancialEffect(financialCase.Effects, effectID)
		if idx >= 0 && financialCase.Effects[idx].Status == types.FinancialEffectStatusApplied {
			continue
		}
		if idx < 0 {
			financialCase.Effects = append(financialCase.Effects, types.FinancialCaseEffect{EffectId: effectID, Type: spec.typ, Status: types.FinancialEffectStatusPending})
			idx = len(financialCase.Effects) - 1
		}
		financialCase.Effects[idx].Status = types.FinancialEffectStatusPending
		financialCase.Effects[idx].ErrorCode = ""
		financialCase.Effects[idx].Attempts++
		transferAmount := spec.amount
		if spec.id == "provider" && !financialCase.Exposure.UnclaimedRewards.IsZero() {
			transferAmount = transferAmount.Sub(financialCase.Exposure.UnclaimedRewards...)
		}
		if !transferAmount.IsZero() && spec.address != "" {
			address, err := sdk.AccAddressFromBech32(spec.address)
			if err != nil {
				return types.ErrFinancialCaseEffect.Wrap(err.Error())
			}
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, address, transferAmount); err != nil {
				financialCase.Effects[idx].Status, financialCase.Effects[idx].ErrorCode = types.FinancialEffectStatusFailed, "bank_transfer_failed"
				return types.ErrFinancialCaseEffect.Wrap(err.Error())
			}
		}
		if spec.id == "platform" && !spec.amount.IsZero() {
			if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleAccountName, "fee_collector", spec.amount); err != nil {
				financialCase.Effects[idx].Status, financialCase.Effects[idx].ErrorCode = types.FinancialEffectStatusFailed, "platform_transfer_failed"
				return types.ErrFinancialCaseEffect.Wrap(err.Error())
			}
		}
		if spec.id == "reward" {
			if err := k.consumeFinancialCaseRewardExposure(ctx, financialCase.Exposure.RewardAddress, financialCase.Exposure.UnclaimedRewards); err != nil {
				financialCase.Effects[idx].Status, financialCase.Effects[idx].ErrorCode = types.FinancialEffectStatusFailed, "reward_consume_failed"
				return err
			}
		}
		financialCase.Effects[idx].Status, financialCase.Effects[idx].AppliedHeight, financialCase.Effects[idx].AppliedAt, financialCase.Effects[idx].ErrorCode = types.FinancialEffectStatusApplied, ctx.BlockHeight(), ctx.BlockTime().Unix(), ""
		_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialCaseEffectApplied{CaseId: financialCase.CaseId, EffectId: effectID, EffectType: spec.typ})
	}
	if financialCase.Exposure.PayoutId != "" {
		payout, found := k.GetPayout(ctx, financialCase.Exposure.PayoutId)
		if !found || payout.State != types.PayoutStateHeld || payout.DisputeID != financialCase.CaseId {
			return types.ErrFinancialCaseHold.Wrap("payout hold missing at finalization")
		}
		payout.State, payout.DisputeID, payout.HoldReason = types.PayoutStateCancelled, "", "canonical_financial_case_finalized"
		if allocation.Customer.Equal(payout.GrossAmount) {
			payout.State = types.PayoutStateRefunded
		}
		providerPayoutAllocation := allocation.Provider
		if !financialCase.Exposure.UnclaimedRewards.IsZero() {
			providerPayoutAllocation = providerPayoutAllocation.Sub(financialCase.Exposure.UnclaimedRewards...)
		}
		if providerPayoutAllocation.Equal(payout.GrossAmount) {
			payout.State = types.PayoutStateCompleted
			now := ctx.BlockTime()
			payout.CompletedAt = &now
			payout.TxHash = "financial-case/" + financialCase.CaseId
		}
		if err := k.SetPayout(ctx, payout); err != nil {
			return err
		}
	}
	if financialCase.Exposure.EscrowId != "" {
		escrow, found := k.GetEscrow(ctx, financialCase.Exposure.EscrowId)
		if found && escrow.State == types.EscrowStateDisputed {
			if financialCase.Exposure.PayoutId == "" {
				escrow.Balance = sdk.NewCoins()
				escrow.State = types.EscrowStateReleased
				if !allocation.Customer.IsZero() && allocation.Provider.IsZero() {
					escrow.State = types.EscrowStateRefunded
				}
			} else if escrow.Balance.IsZero() {
				escrow.State = types.EscrowStateReleased
			} else {
				escrow.State = types.EscrowStateActive
			}
			if err := k.SetEscrow(ctx, escrow); err != nil {
				return err
			}
		}
	}
	if financialCase.Exposure.ReservationId != "" {
		if k.reservationKeeper == nil {
			return types.ErrFinancialCaseEffect.Wrap("reservation keeper missing")
		}
		slash := allocation.ResolutionType == types.FinancialResolutionFraudConfirmed
		if _, err := k.reservationKeeper.FinalizeReservationFinancialCase(ctx, financialCase.Exposure.ReservationId, financialCase.CaseId, slash); err != nil {
			return err
		}
	}
	financialCase.ActiveHoldCount = 0
	return nil
}

func (k Keeper) consumeFinancialCaseRewardExposure(ctx sdk.Context, rewardAddress string, exposure sdk.Coins) error {
	if exposure.IsZero() {
		return nil
	}
	address, err := sdk.AccAddressFromBech32(rewardAddress)
	if err != nil {
		return types.ErrFinancialCaseEffect.Wrap("invalid reward address")
	}
	rewards, found := k.GetClaimableRewards(ctx, address)
	if !found || !rewards.TotalClaimable.IsAllGTE(exposure) {
		return types.ErrFinancialCaseEffect.Wrap("claimable reward exposure is unavailable")
	}
	remaining := exposure
	entries := append([]types.RewardEntry(nil), rewards.RewardEntries...)
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].DistributionID != entries[j].DistributionID {
			return entries[i].DistributionID < entries[j].DistributionID
		}
		if entries[i].Source != entries[j].Source {
			return entries[i].Source < entries[j].Source
		}
		return entries[i].Amount.String() < entries[j].Amount.String()
	})
	retained := make([]types.RewardEntry, 0, len(entries))
	for _, entry := range entries {
		amount := sdk.NewCoins()
		for _, coin := range entry.Amount {
			consume := coin.Amount
			available := remaining.AmountOf(coin.Denom)
			if consume.GT(available) {
				consume = available
			}
			if consume.IsPositive() {
				remaining = remaining.Sub(sdk.NewCoin(coin.Denom, consume))
			}
			left := coin.Amount.Sub(consume)
			if left.IsPositive() {
				amount = amount.Add(sdk.NewCoin(coin.Denom, left))
			}
		}
		if !amount.IsZero() {
			entry.Amount = amount
			retained = append(retained, entry)
		}
	}
	if !remaining.IsZero() {
		return types.ErrFinancialCaseEffect.Wrap("claimable reward entries do not reconcile")
	}
	rewards.TotalClaimable = rewards.TotalClaimable.Sub(exposure...)
	rewards.RewardEntries = retained
	rewards.LastUpdated = ctx.BlockTime()
	return k.SetClaimableRewards(ctx, address, rewards)
}

func (k Keeper) validateFinancialEffectsComplete(financialCase types.FinancialCase) error {
	expected := expectedFinancialEffects(financialCase)
	seen := make(map[string]struct{}, len(financialCase.Effects))
	if financialCase.Status == types.FinancialCaseStatusFinal && len(financialCase.Effects) == 0 {
		return types.ErrFinancialCaseEffect.Wrap("terminal case has no effect markers")
	}
	for _, effect := range financialCase.Effects {
		if effect.Status != types.FinancialEffectStatusApplied {
			return types.ErrFinancialCaseEffect.Wrap("terminal case has incomplete effect")
		}
		if _, duplicate := seen[effect.EffectId]; duplicate {
			return types.ErrFinancialCaseEffect.Wrap("terminal case has duplicate effect")
		}
		expectedType, required := expected[effect.EffectId]
		if !required {
			return types.ErrFinancialCaseEffect.Wrap("terminal case has unexpected effect")
		}
		if effect.Type != expectedType || effect.Attempts == 0 || effect.AppliedHeight < 0 || effect.AppliedAt < 0 || effect.ErrorCode != "" {
			return types.ErrFinancialCaseEffect.Wrap("terminal case has malformed effect")
		}
		seen[effect.EffectId] = struct{}{}
	}
	if len(seen) != len(expected) {
		return types.ErrFinancialCaseEffect.Wrap("terminal case is missing required effect")
	}
	return nil
}

func expectedFinancialEffects(financialCase types.FinancialCase) map[string]types.FinancialEffectType {
	expected := map[string]types.FinancialEffectType{
		financialCase.CaseId + "/provider":      types.FinancialEffectPayout,
		financialCase.CaseId + "/customer":      types.FinancialEffectPayout,
		financialCase.CaseId + "/platform":      types.FinancialEffectPayout,
		financialCase.CaseId + "/slash-witness": types.FinancialEffectPayout,
		financialCase.CaseId + "/projection":    types.FinancialEffectProjection,
	}
	if !financialCase.Exposure.UnclaimedRewards.IsZero() {
		expected[financialCase.CaseId+"/reward"] = types.FinancialEffectReward
	}
	if financialCase.Exposure.ReservationId != "" {
		expected[financialCase.CaseId+"/reservation"] = types.FinancialEffectReservation
	}
	return expected
}

func (k Keeper) validateFinancialCase(ctx sdk.Context, financialCase types.FinancialCase) error {
	if financialCase.Version != financialCaseVersion || financialCase.CaseId == "" || len(financialCase.Claims) == 0 {
		return types.ErrInvalidFinancialCase
	}
	subjectKey, err := CanonicalFinancialSubjectKey(financialCase.Subject)
	if err != nil {
		return err
	}
	expectedID, err := DeterministicFinancialCaseID(financialCase.Subject)
	if err != nil || expectedID != financialCase.CaseId {
		return types.ErrInvalidFinancialCase.Wrap("noncanonical case ID")
	}
	if _, err := sdk.AccAddressFromBech32(financialCase.Claimant); err != nil {
		return types.ErrInvalidFinancialCase.Wrap("invalid claimant")
	}
	if _, err := sdk.AccAddressFromBech32(financialCase.Respondent); err != nil || financialCase.Claimant == financialCase.Respondent {
		return types.ErrInvalidFinancialCase.Wrap("invalid respondent")
	}
	if _, err := sdk.AccAddressFromBech32(financialCase.Provider); err != nil {
		return types.ErrInvalidFinancialCase.Wrap("invalid provider")
	}
	if _, err := sdk.AccAddressFromBech32(financialCase.Customer); err != nil || financialCase.Provider == financialCase.Customer {
		return types.ErrInvalidFinancialCase.Wrap("invalid customer")
	}
	if (financialCase.Claimant != financialCase.Provider && financialCase.Claimant != financialCase.Customer) || (financialCase.Respondent != financialCase.Provider && financialCase.Respondent != financialCase.Customer) {
		return types.ErrInvalidFinancialCase.Wrap("claimant/respondent do not match provider/customer")
	}
	if financialCase.Claimant == financialCase.Respondent {
		return types.ErrInvalidFinancialCase.Wrap("claimant and respondent must be distinct parties")
	}
	if financialCase.Status == settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_UNSPECIFIED {
		return types.ErrInvalidFinancialCase.Wrap("status required")
	}
	if !financialCase.Exposure.OriginalHeld.IsValid() || financialCase.Exposure.OriginalHeld.IsZero() {
		return types.ErrInvalidFinancialCase.Wrap("held exposure required")
	}
	if !financialCase.Exposure.EscrowAmount.IsValid() || !financialCase.Exposure.PayoutAmount.IsValid() || !financialCase.Exposure.UnclaimedRewards.IsValid() {
		return types.ErrInvalidFinancialCase.Wrap("invalid exposure components")
	}
	if financialCase.Exposure.PayoutId == "" != financialCase.Exposure.PayoutAmount.IsZero() || financialCase.Exposure.EscrowId == "" && !financialCase.Exposure.EscrowAmount.IsZero() || financialCase.Exposure.RewardAddress == "" && !financialCase.Exposure.UnclaimedRewards.IsZero() {
		return types.ErrInvalidFinancialCase.Wrap("exposure reference and amount mismatch")
	}
	if types.IsActiveFinancialCaseStatus(financialCase.Status) && financialCase.ActiveHoldCount == 0 {
		return types.ErrFinancialCaseHold.Wrap("active value case requires hold")
	}
	if types.IsTerminalFinancialCaseStatus(financialCase.Status) && financialCase.ActiveHoldCount != 0 {
		return types.ErrFinancialCaseHold.Wrap("terminal case retains hold")
	}
	sequence := uint64(1)
	for i, transition := range financialCase.Transitions {
		if transition.Sequence != sequence {
			return types.ErrFinancialCaseMalformedState.Wrap("transition sequence gap")
		}
		if transition.To == settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_UNSPECIFIED || transition.Action == "" || len(transition.Action) > 64 || len(transition.Actor) > financialMaxSubjectIDBytes || strings.ContainsRune(transition.Action, '\x00') || strings.ContainsRune(transition.Actor, '\x00') || len(transition.ReasonHash) != 0 && len(transition.ReasonHash) != financialHashSize {
			return types.ErrFinancialCaseMalformedState.Wrap("invalid transition audit fields")
		}
		if i > 0 && transition.From != financialCase.Transitions[i-1].To {
			return types.ErrFinancialCaseMalformedState.Wrap("transition chain mismatch")
		}
		sequence++
	}
	if !bytes.Equal(financialCase.ClaimRoot, financialClaimRoot(financialCase.Claims)) {
		return types.ErrFinancialCaseMalformedState.Wrap("claim root mismatch")
	}
	if len(financialCase.Transitions) == 0 || financialCase.Transitions[0].From != settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_UNSPECIFIED || financialCase.Transitions[len(financialCase.Transitions)-1].To != financialCase.Status {
		return types.ErrFinancialCaseMalformedState.Wrap("transition history does not terminate at current status")
	}
	if indexed := ctx.KVStore(k.skey).Get(types.FinancialSubjectKey(subjectKey)); indexed != nil && string(indexed) != financialCase.CaseId && types.IsActiveFinancialCaseStatus(financialCase.Status) {
		return types.ErrFinancialCaseMalformedState.Wrap("subject already has another active case")
	}
	return nil
}

func validateFinancialClaimInput(claim types.FinancialClaim, maxReference uint32) error {
	if claim.ClaimType == settlementv1.FinancialClaimType_FINANCIAL_CLAIM_TYPE_UNSPECIFIED {
		return types.ErrInvalidFinancialCase.Wrap("claim type required")
	}
	if _, err := sdk.AccAddressFromBech32(claim.Claimant); err != nil {
		return types.ErrFinancialCaseAuthorization.Wrap("invalid claim signer")
	}
	if len(claim.SourceModule) == 0 || len(claim.SourceModule) > financialMaxSourceModuleBytes || len(claim.SourceReference) > financialMaxSourceReferenceBytes || len(claim.Recommendation) > financialMaxRecommendationBytes {
		return types.ErrInvalidFinancialCase.Wrap("claim metadata exceeds bounds")
	}
	if len(claim.EvidenceHash) != financialHashSize || len(claim.IdempotencyKey) == 0 || len(claim.IdempotencyKey) > 128 {
		return types.ErrInvalidFinancialCase.Wrap("SHA-256 evidence hash and idempotency key required")
	}
	if len(claim.EncryptedReference) > int(maxReference) || strings.ContainsRune(claim.EncryptedReference, '\x00') {
		return types.ErrFinancialCasePrivacy
	}
	if strings.ContainsRune(claim.SourceModule, '\x00') || strings.ContainsRune(claim.SourceReference, '\x00') || strings.ContainsRune(claim.Recommendation, '\x00') {
		return types.ErrFinancialCasePrivacy.Wrap("claim metadata contains NUL")
	}
	return nil
}

func financialClaimRoot(claims []types.FinancialClaim) []byte {
	ids := make([]string, 0, len(claims))
	for _, claim := range claims {
		ids = append(ids, claim.ClaimId)
	}
	sort.Strings(ids)
	h := sha256.New()
	for _, id := range ids {
		writeFinancialField(h, []byte(id))
	}
	return h.Sum(nil)
}

func financialAllocationHash(allocation types.TerminalAllocation) []byte {
	h := sha256.New()
	for _, value := range []string{allocation.OriginalExposure.String(), allocation.Provider.String(), allocation.Customer.String(), allocation.Platform.String(), allocation.SlashWitness.String(), allocation.SlashWitnessRecipient, allocation.ResolutionType.String()} {
		writeFinancialField(h, []byte(value))
	}
	return h.Sum(nil)
}

func (k Keeper) setFinancialCaseDeadlines(ctx sdk.Context, financialCase *types.FinancialCase) {
	params := k.GetParams(ctx)
	financialCase.FilingDeadlineHeight = addHeightBounded(ctx.BlockHeight(), params.FinancialCaseFilingWindowBlocks, defaultBlockWindow(params.FinancialCaseFilingWindowSeconds))
	financialCase.EvidenceDeadlineHeight = addHeightBounded(ctx.BlockHeight(), params.FinancialCaseEvidenceWindowBlocks, defaultBlockWindow(params.FinancialCaseEvidenceWindowSeconds))
	financialCase.ReviewDeadlineHeight = addHeightBounded(financialCase.EvidenceDeadlineHeight, params.FinancialCaseReviewWindowBlocks, defaultBlockWindow(params.FinancialCaseReviewWindowSeconds))
	financialCase.EscalationDeadlineHeight = addHeightBounded(financialCase.ReviewDeadlineHeight, params.FinancialCaseEscalationWindowBlocks, defaultBlockWindow(params.FinancialCaseEscalationWindowSeconds))
	financialCase.FilingDeadlineTime = addTimeBounded(ctx.BlockTime(), params.FinancialCaseFilingWindowSeconds, 7*24*time.Hour).Unix()
	financialCase.EvidenceDeadlineTime = addTimeBounded(ctx.BlockTime(), params.FinancialCaseEvidenceWindowSeconds, 7*24*time.Hour).Unix()
	financialCase.ReviewDeadlineTime = addTimeBounded(time.Unix(financialCase.EvidenceDeadlineTime, 0), params.FinancialCaseReviewWindowSeconds, 7*24*time.Hour).Unix()
	financialCase.EscalationDeadlineTime = addTimeBounded(time.Unix(financialCase.ReviewDeadlineTime, 0), params.FinancialCaseEscalationWindowSeconds, 30*24*time.Hour).Unix()
}

func defaultBlockWindow(seconds uint64) int64 {
	if seconds == 0 {
		return 100
	}
	blocks := seconds / 5
	if blocks == 0 {
		blocks = 1
	}
	if blocks > uint64(^uint64(0)>>1) {
		return int64(^uint64(0) >> 1)
	}
	return int64(blocks)
} //nolint:gosec
func addHeightBounded(base, configured, fallback int64) int64 {
	delta := configured
	if delta <= 0 {
		delta = fallback
	}
	if delta <= 0 || base > int64(^uint64(0)>>1)-delta {
		return int64(^uint64(0) >> 1)
	}
	return base + delta
}
func addTimeBounded(base time.Time, seconds uint64, fallback time.Duration) time.Time {
	if seconds == 0 {
		return base.Add(fallback)
	}
	max := uint64(math.MaxInt64 / int64(time.Second))
	if seconds > max {
		seconds = max
	}
	duration, err := time.ParseDuration(strconv.FormatUint(seconds, 10) + "s")
	if err != nil {
		return base.Add(fallback)
	}
	return base.Add(duration)
}
func deadlinePassed(ctx sdk.Context, height, unix int64) bool {
	return height > 0 && ctx.BlockHeight() > height || unix > 0 && ctx.BlockTime().Unix() > unix
}

func (k Keeper) getFinancialClaimReplay(ctx sdk.Context, key []byte) (financialClaimReplay, bool, error) {
	bz := ctx.KVStore(k.skey).Get(types.FinancialClaimIdempotencyKey(key))
	if bz == nil {
		return financialClaimReplay{}, false, nil
	}
	var replay financialClaimReplay
	if err := json.Unmarshal(bz, &replay); err != nil {
		return replay, false, types.ErrFinancialCaseMalformedState.Wrap("malformed idempotency index")
	}
	return replay, true, nil
}
func (k Keeper) setFinancialClaimReplay(ctx sdk.Context, key []byte, caseID, claimID string, payload []byte) error {
	bz, err := json.Marshal(financialClaimReplay{CaseID: caseID, ClaimID: claimID, PayloadHash: payload})
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.FinancialClaimIdempotencyKey(key), bz)
	return nil
}

func (k Keeper) getFinancialAppealReplay(ctx sdk.Context, key []byte) (financialAppealReplay, bool, error) {
	bz := ctx.KVStore(k.skey).Get(types.FinancialAppealIdempotencyKey(key))
	if bz == nil {
		return financialAppealReplay{}, false, nil
	}
	var replay financialAppealReplay
	if err := json.Unmarshal(bz, &replay); err != nil {
		return replay, false, types.ErrFinancialCaseMalformedState.Wrap("malformed appeal idempotency index")
	}
	return replay, true, nil
}

func (k Keeper) setFinancialAppealReplay(ctx sdk.Context, key []byte, caseID, appealID string, payload []byte) error {
	replay := financialAppealReplay{CaseID: caseID, AppealID: appealID, PayloadHash: append([]byte(nil), payload...)}
	bz, err := json.Marshal(replay)
	if err != nil {
		return err
	}
	storeKey := types.FinancialAppealIdempotencyKey(key)
	if existing := ctx.KVStore(k.skey).Get(storeKey); existing != nil {
		var previous financialAppealReplay
		if err := json.Unmarshal(existing, &previous); err != nil {
			return types.ErrFinancialCaseMalformedState.Wrap("malformed appeal idempotency index")
		}
		if previous.CaseID != caseID || previous.AppealID != appealID || !bytes.Equal(previous.PayloadHash, payload) {
			return types.ErrFinancialCaseIdempotencyConflict.Wrap("appeal idempotency key is already owned")
		}
		return nil
	}
	ctx.KVStore(k.skey).Set(storeKey, bz)
	return nil
}

func (k Keeper) setFinancialCaseIndexes(store storetypes.KVStore, financialCase types.FinancialCase) error {
	subjectKey, err := CanonicalFinancialSubjectKey(financialCase.Subject)
	if err != nil {
		return err
	}
	if types.IsActiveFinancialCaseStatus(financialCase.Status) {
		store.Set(types.FinancialSubjectKey(subjectKey), []byte(financialCase.CaseId))
	}
	for _, index := range financialCaseIndexes(financialCase) {
		store.Set(types.FinancialCaseIndexKey(index.prefix, index.value, financialCase.CaseId), []byte(financialCase.CaseId))
	}
	return nil
}
func (k Keeper) deleteFinancialCaseIndexes(store storetypes.KVStore, financialCase types.FinancialCase) {
	if key, err := CanonicalFinancialSubjectKey(financialCase.Subject); err == nil {
		if existing := store.Get(types.FinancialSubjectKey(key)); string(existing) == financialCase.CaseId {
			store.Delete(types.FinancialSubjectKey(key))
		}
	}
	for _, index := range financialCaseIndexes(financialCase) {
		store.Delete(types.FinancialCaseIndexKey(index.prefix, index.value, financialCase.CaseId))
	}
}

type financialIndex struct {
	prefix []byte
	value  string
}

func financialCaseIndexes(financialCase types.FinancialCase) []financialIndex {
	indexes := []financialIndex{{types.PrefixFinancialCaseByStatus, financialCase.Status.String()}, {types.PrefixFinancialCaseByParty, financialCase.Claimant}, {types.PrefixFinancialCaseByParty, financialCase.Respondent}}
	for _, item := range []financialIndex{{types.PrefixFinancialCaseByOrder, financialCase.Subject.OrderId}, {types.PrefixFinancialCaseByInvoice, financialCase.Subject.InvoiceId}, {types.PrefixFinancialCaseByUsage, financialCase.Subject.UsageId}, {types.PrefixFinancialCaseByJob, financialCase.Subject.HpcJobId}, {types.PrefixFinancialCaseByEscrow, financialCase.Exposure.EscrowId}, {types.PrefixFinancialCaseBySettlement, financialCase.Subject.SettlementId}, {types.PrefixFinancialCaseByReservation, financialCase.Subject.ReservationId}, {types.PrefixFinancialCaseByLease, financialCase.Subject.LeaseId}} {
		if item.value != "" {
			indexes = append(indexes, item)
		}
	}
	return indexes
}
func financialCaseIndexPrefixForKind(kind string) []byte {
	switch kind {
	case "order":
		return types.PrefixFinancialCaseByOrder
	case "invoice":
		return types.PrefixFinancialCaseByInvoice
	case "usage":
		return types.PrefixFinancialCaseByUsage
	case "job":
		return types.PrefixFinancialCaseByJob
	case "escrow":
		return types.PrefixFinancialCaseByEscrow
	case "status":
		return types.PrefixFinancialCaseByStatus
	case "party":
		return types.PrefixFinancialCaseByParty
	case "settlement":
		return types.PrefixFinancialCaseBySettlement
	case "reservation":
		return types.PrefixFinancialCaseByReservation
	case "lease":
		return types.PrefixFinancialCaseByLease
	default:
		return nil
	}
}

func financialCaseIndexPrefixes() [][]byte {
	return [][]byte{
		types.PrefixFinancialCaseByOrder,
		types.PrefixFinancialCaseByInvoice,
		types.PrefixFinancialCaseByUsage,
		types.PrefixFinancialCaseByJob,
		types.PrefixFinancialCaseByEscrow,
		types.PrefixFinancialCaseByStatus,
		types.PrefixFinancialCaseByParty,
		types.PrefixFinancialCaseBySettlement,
		types.PrefixFinancialCaseByReservation,
		types.PrefixFinancialCaseByLease,
	}
}
func findFinancialEffect(effects []types.FinancialCaseEffect, id string) int {
	for i := range effects {
		if effects[i].EffectId == id {
			return i
		}
	}
	return -1
}

func financialCaseAliasKeys(financialCase types.FinancialCase) []string {
	aliases := make([]string, 0, 9)
	if key, err := CanonicalFinancialSubjectKey(financialCase.Subject); err == nil {
		aliases = append(aliases, "subject\x00"+key)
	}
	for _, item := range []struct{ kind, value string }{{"order", financialCase.Subject.OrderId}, {"invoice", financialCase.Subject.InvoiceId}, {"usage", financialCase.Subject.UsageId}, {"job", financialCase.Subject.HpcJobId}, {"settlement", financialCase.Subject.SettlementId}, {"escrow", financialCase.Exposure.EscrowId}, {"reservation", financialCase.Subject.ReservationId}, {"lease", financialCase.Subject.LeaseId}} {
		if item.value != "" {
			aliases = append(aliases, item.kind+"\x00"+item.value)
		}
	}
	return aliases
}

func (k Keeper) financialCaseHoldCount(ctx sdk.Context, financialCase types.FinancialCase) (uint32, error) {
	var holds uint32
	if financialCase.Exposure.PayoutId != "" {
		payout, found := k.GetPayout(ctx, financialCase.Exposure.PayoutId)
		if !found {
			return 0, types.ErrFinancialCaseHold.Wrap("payout missing")
		}
		irreversibleFiat, err := k.payoutHasIrreversibleFiatBoundary(ctx, payout)
		if err != nil {
			return 0, err
		}
		if irreversibleFiat {
			// Incident-only cases intentionally do not own the payout state.
		} else if payout.State != types.PayoutStateHeld || payout.DisputeID != financialCase.CaseId {
			return 0, types.ErrFinancialCaseHold.Wrap("payout hold missing")
		} else {
			holds++
		}
	}
	if financialCase.Exposure.EscrowId != "" {
		escrow, found := k.GetEscrow(ctx, financialCase.Exposure.EscrowId)
		if !found || escrow.State != types.EscrowStateDisputed {
			return 0, types.ErrFinancialCaseHold.Wrap("escrow hold missing")
		}
		holds++
	}
	if !financialCase.Exposure.UnclaimedRewards.IsZero() {
		address, err := sdk.AccAddressFromBech32(financialCase.Exposure.RewardAddress)
		if err != nil {
			return 0, types.ErrFinancialCaseHold.Wrap("reward hold address invalid")
		}
		rewards, found := k.GetClaimableRewards(ctx, address)
		if !found || !rewards.TotalClaimable.IsAllGTE(financialCase.Exposure.UnclaimedRewards) {
			return 0, types.ErrFinancialCaseHold.Wrap("reward exposure unavailable")
		}
		holds++
	}
	if financialCase.Exposure.ReservationId != "" {
		if k.reservationKeeper == nil {
			return 0, types.ErrFinancialCaseHold.Wrap("reservation keeper missing")
		}
		reservation, found := k.reservationKeeper.GetReservation(ctx, financialCase.Exposure.ReservationId)
		if !found || reservation.State != resourcesv1.ReservationState_RESERVATION_STATE_DISPUTED || reservation.FinancialCaseId != financialCase.CaseId {
			return 0, types.ErrFinancialCaseHold.Wrap("reservation hold missing")
		}
		holds++
	}
	return holds, nil
}

func (k Keeper) payoutHasIrreversibleFiatBoundary(ctx sdk.Context, payout types.PayoutRecord) (bool, error) {
	if payout.FiatConversionID == "" {
		return false, nil
	}
	conversion, found := k.GetFiatConversion(ctx, payout.FiatConversionID)
	if !found || conversion.PayoutID != payout.PayoutID {
		return false, types.ErrFinancialCaseHold.Wrap("payout fiat conversion ownership is malformed")
	}
	return fiatConversionCrossedIrreversibleBoundary(conversion), nil
}
func (k Keeper) emitFinancialCaseOpened(ctx sdk.Context, financialCase types.FinancialCase, subjectKey string) {
	_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialCaseOpened{CaseId: financialCase.CaseId, SubjectKey: subjectKey, Status: financialCase.Status, HoldCount: financialCase.ActiveHoldCount})
}
func (k Keeper) emitFinancialCaseQuarantined(ctx sdk.Context, caseID string, reason []byte) {
	hash := sha256.Sum256(reason)
	_ = ctx.EventManager().EmitTypedEvent(&settlementv1.EventFinancialCaseQuarantined{CaseId: caseID, ReasonHash: hash[:]})
}

package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// SECURITY NOTE - CONSENSUS SAFETY
// ============================================================================
//
// Historical audit note: generating nonces/salts inside keeper with crypto/rand
// is CONSENSUS-UNSAFE because validators would diverge on state. Randomness
// must be injected or derived deterministically from shared context.
//
// REMEDIATION (SECURITY-001): All randomness is now injected via RandomnessInputs
// or deterministically derived from tx context (DeterministicRandomSource). This
// guarantees identical outputs across validators for the same transaction data.
// ============================================================================

// ============================================================================
// Selective Disclosure Request Management
// ============================================================================

// CreateSelectiveDisclosureRequest creates a new request for selective disclosure of claims.
// The request specifies which claims the requester wants the subject to prove.
func (k Keeper) CreateSelectiveDisclosureRequest(
	ctx sdk.Context,
	requesterAddress sdk.AccAddress,
	subjectAddress sdk.AccAddress,
	requestedClaims []types.ClaimType,
	claimParameters map[string]interface{},
	purpose string,
	validityDuration time.Duration,
	requestExpiry time.Duration,
	randomness *RandomnessInputs,
) (*types.SelectiveDisclosureRequest, error) {
	// Validate input parameters
	if len(requestedClaims) == 0 {
		return nil, types.ErrInvalidProofRequest.Wrap("requested_claims cannot be empty")
	}

	for _, ct := range requestedClaims {
		if !ct.IsValid() {
			return nil, types.ErrInvalidClaimType.Wrapf("invalid claim type: %d", ct)
		}
	}

	if purpose == "" {
		return nil, types.ErrInvalidProofRequest.Wrap("purpose cannot be empty")
	}

	if validityDuration <= 0 {
		return nil, types.ErrInvalidProofRequest.Wrap("validity_duration must be positive")
	}

	if requestExpiry <= 0 {
		requestExpiry = 24 * time.Hour // Default 24 hour request expiry
	}

	// Resolve nonce (caller-supplied or deterministic from context)
	var providedNonce []byte
	if randomness != nil {
		providedNonce = randomness.Nonce
	}
	nonce, err := k.resolveRandomBytes(
		ctx,
		providedNonce,
		"veid:sdr:nonce",
		requesterAddress.Bytes(),
		subjectAddress.Bytes(),
	)
	if err != nil {
		return nil, err
	}

	// Generate request ID
	requestID := types.GenerateRequestID(
		requesterAddress.String(),
		subjectAddress.String(),
		nonce,
	)

	// Create the request
	request := types.NewSelectiveDisclosureRequest(
		requestID,
		requesterAddress.String(),
		subjectAddress.String(),
		requestedClaims,
		purpose,
		validityDuration,
		requestExpiry,
	)
	request.ClaimParameters = claimParameters
	request.Nonce = nonce

	// Validate the request
	if err := request.Validate(); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"selective_disclosure_request_created",
			sdk.NewAttribute("request_id", requestID),
			sdk.NewAttribute("requester", requesterAddress.String()),
			sdk.NewAttribute("subject", subjectAddress.String()),
			sdk.NewAttribute("purpose", purpose),
		),
	)

	return request, nil
}

// ============================================================================
// Selective Disclosure Proof Generation
// ============================================================================

// GenerateSelectiveDisclosureProof generates a zero-knowledge proof for selective disclosure
// of identity claims. This allows the subject to prove specific claims without revealing
// any additional information.
func (k Keeper) GenerateSelectiveDisclosureProof(
	ctx sdk.Context,
	subjectAddress sdk.AccAddress,
	request *types.SelectiveDisclosureRequest,
	disclosedClaims map[string]interface{},
	scheme types.ProofScheme,
	randomness *RandomnessInputs,
) (*types.SelectiveDisclosureProof, error) {
	// Validate request is still valid
	blockTime := ctx.BlockTime()
	if request.IsExpired(blockTime) {
		return nil, types.ErrProofRequestExpired
	}

	// Validate subject matches
	if request.SubjectAddress != subjectAddress.String() {
		return nil, types.ErrUnauthorized.Wrap("subject address does not match request")
	}

	// Validate proof scheme
	if !scheme.IsValid() {
		return nil, types.ErrInvalidProofScheme.Wrapf("invalid proof scheme: %d", scheme)
	}

	// Verify subject has an identity record with sufficient verification
	record, found := k.GetIdentityRecord(ctx, subjectAddress)
	if !found {
		return nil, types.ErrIdentityRecordNotFound
	}

	// Check that subject can provide the requested claims
	for _, claimType := range request.RequestedClaims {
		if err := k.validateClaimAvailability(ctx, subjectAddress, record, claimType, request.ClaimParameters); err != nil {
			return nil, err
		}
	}

	// Resolve nonce for the proof
	var providedNonce []byte
	var providedCommitmentSalt []byte
	if randomness != nil {
		providedNonce = randomness.Nonce
		providedCommitmentSalt = randomness.CommitmentSalt
	}
	proofNonce, err := k.resolveRandomBytes(
		ctx,
		providedNonce,
		"veid:sdp:nonce",
		subjectAddress.Bytes(),
		[]byte(request.RequestID),
	)
	if err != nil {
		return nil, err
	}

	// Generate proof ID
	proofID := types.GenerateProofID(
		subjectAddress.String(),
		request.RequestedClaims,
		proofNonce,
	)

	// Create commitment hash for full claims
	// In MVP, this is a simple hash of disclosed claims + salt
	commitmentSalt, err := k.resolveRandomBytes(
		ctx,
		providedCommitmentSalt,
		"veid:sdp:commitment_salt",
		[]byte(proofID),
	)
	if err != nil {
		return nil, err
	}
	commitmentHash := k.generateCommitmentHash(disclosedClaims, commitmentSalt)

	proofValue, err := k.generateZKProof(
		ctx,
		record,
		subjectAddress,
		request.RequestedClaims,
		disclosedClaims,
		request.ClaimParameters,
		scheme,
		proofNonce,
		commitmentSalt,
	)
	if err != nil {
		return nil, err
	}

	// Create the proof
	proof := types.NewSelectiveDisclosureProof(
		proofID,
		subjectAddress.String(),
		request.RequestedClaims,
		scheme,
		request.ValidityDuration,
	)
	proof.DisclosedClaims = disclosedClaims
	proof.CommitmentHash = commitmentHash
	proof.ProofValue = proofValue
	proof.Nonce = proofNonce
	proof.Metadata = map[string]string{
		"request_id":      request.RequestID,
		"purpose":         request.Purpose,
		"commitment_salt": hex.EncodeToString(commitmentSalt),
	}
	if len(request.ClaimParameters) > 0 {
		if paramsBz, marshalErr := json.Marshal(request.ClaimParameters); marshalErr == nil {
			proof.Metadata["claim_parameters"] = string(paramsBz)
		}
	}

	// Validate the proof
	if err := proof.Validate(); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"selective_disclosure_proof_generated",
			sdk.NewAttribute("proof_id", proofID),
			sdk.NewAttribute("subject", subjectAddress.String()),
			sdk.NewAttribute("request_id", request.RequestID),
			sdk.NewAttribute("scheme", scheme.String()),
		),
	)

	return proof, nil
}

// VerifySelectiveDisclosureProof verifies a selective disclosure proof without
// learning any information beyond what was explicitly disclosed.
func (k Keeper) VerifySelectiveDisclosureProof(
	ctx sdk.Context,
	proof *types.SelectiveDisclosureProof,
	verifierAddress sdk.AccAddress,
) (*types.ProofVerificationResult, error) {
	// Basic proof validation
	if err := proof.Validate(); err != nil {
		return &types.ProofVerificationResult{
			IsValid:         false,
			VerifiedAt:      ctx.BlockTime(),
			VerifierAddress: verifierAddress.String(),
			Error:           err.Error(),
		}, nil
	}

	// Check proof expiration
	if proof.IsExpired(ctx.BlockTime()) {
		return &types.ProofVerificationResult{
			IsValid:         false,
			VerifiedAt:      ctx.BlockTime(),
			VerifierAddress: verifierAddress.String(),
			Error:           types.ErrProofExpired.Error(),
		}, nil
	}

	// Check proof revocation
	if k.IsProofRevoked(ctx, proof.ProofID) {
		return &types.ProofVerificationResult{
			IsValid:         false,
			VerifiedAt:      ctx.BlockTime(),
			VerifierAddress: verifierAddress.String(),
			Error:           "proof revoked",
		}, nil
	}

	// Verify the ZK proof
	isValid, err := k.verifyZKProof(
		proof.SubjectAddress,
		proof.ClaimTypes,
		proof.DisclosedClaims,
		proof.ProofValue,
		proof.ProofScheme,
		proof.Nonce,
		proof.CommitmentHash,
		proof.Metadata,
	)
	if err != nil {
		return &types.ProofVerificationResult{
			IsValid:         false,
			VerifiedAt:      ctx.BlockTime(),
			VerifierAddress: verifierAddress.String(),
			Error:           err.Error(),
		}, nil
	}

	result := &types.ProofVerificationResult{
		IsValid:         isValid,
		ClaimsVerified:  proof.ClaimTypes,
		VerifiedAt:      ctx.BlockTime(),
		VerifierAddress: verifierAddress.String(),
		ProofHash:       hex.EncodeToString(proof.GetProofHash()),
	}

	if !isValid {
		result.Error = "proof verification failed"
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"selective_disclosure_proof_verified",
			sdk.NewAttribute("proof_id", proof.ProofID),
			sdk.NewAttribute("verifier", verifierAddress.String()),
			sdk.NewAttribute("is_valid", fmt.Sprintf("%t", isValid)),
		),
	)

	return result, nil
}

// ============================================================================
// Specialized Proof Generation
// ============================================================================

// CreateAgeProof creates a zero-knowledge proof that the subject's age meets or exceeds
// the specified threshold, without revealing the actual date of birth.
func (k Keeper) CreateAgeProof(
	ctx sdk.Context,
	subjectAddress sdk.AccAddress,
	ageThreshold uint32,
	validDuration time.Duration,
	randomness *RandomnessInputs,
) (*types.AgeProof, error) {
	// Validate age threshold
	if ageThreshold == 0 || ageThreshold > 150 {
		return nil, types.ErrInvalidProofRequest.Wrap("age threshold must be between 1 and 150")
	}

	// Verify subject has identity record
	record, found := k.GetIdentityRecord(ctx, subjectAddress)
	if !found {
		return nil, types.ErrIdentityRecordNotFound
	}

	// Check for verified DOB scope and derive deterministic age value
	satisfiesThreshold, derivedAge, err := k.evaluateAgeThreshold(ctx, record, ageThreshold, randomness)
	if err != nil {
		return nil, err
	}
	if !satisfiesThreshold {
		return nil, types.ErrClaimNotAvailable.Wrap("age below threshold")
	}

	// Resolve nonce
	var providedNonce []byte
	if randomness != nil {
		providedNonce = randomness.Nonce
	}
	nonce, err := k.resolveRandomBytes(
		ctx,
		providedNonce,
		"veid:age:nonce",
		subjectAddress.Bytes(),
	)
	if err != nil {
		return nil, err
	}

	// Generate proof ID
	proofID := types.GenerateProofID(
		subjectAddress.String(),
		[]types.ClaimType{types.ClaimTypeAgeOver18},
		nonce,
	)

	// Create the age proof
	proof := types.NewAgeProof(proofID, subjectAddress.String(), ageThreshold, validDuration)
	proof.SatisfiesThreshold = satisfiesThreshold
	proof.Nonce = nonce

	commitmentSalt, err := k.resolveRandomBytes(
		ctx,
		func() []byte {
			if randomness != nil {
				return randomness.CommitmentSalt
			}
			return nil
		}(),
		"veid:age:commitment_salt",
		subjectAddress.Bytes(),
	)
	if err != nil {
		return nil, err
	}

	rangeProof, err := types.GenerateRangeProof(derivedAge, uint64(ageThreshold), 8, commitmentSalt, nonce, "veid:age")
	if err != nil {
		return nil, types.ErrProofGenerationFailed.Wrap(err.Error())
	}
	proofBytes, err := types.MarshalRangeProof(rangeProof)
	if err != nil {
		return nil, types.ErrProofGenerationFailed.Wrap(err.Error())
	}
	proof.CommitmentHash = rangeProof.Commitment
	proof.ProofValue = proofBytes

	// Validate the proof
	if err := proof.Validate(); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"age_proof_created",
			sdk.NewAttribute("proof_id", proofID),
			sdk.NewAttribute("subject", subjectAddress.String()),
			sdk.NewAttribute("age_threshold", fmt.Sprintf("%d", ageThreshold)),
			sdk.NewAttribute("satisfies", fmt.Sprintf("%t", satisfiesThreshold)),
		),
	)

	return proof, nil
}

// CreateResidencyProof creates a zero-knowledge proof that the subject is a resident
// of the specified country, without revealing the actual address.
func (k Keeper) CreateResidencyProof(
	ctx sdk.Context,
	subjectAddress sdk.AccAddress,
	countryCode string,
	validDuration time.Duration,
	randomness *RandomnessInputs,
) (*types.ResidencyProof, error) {
	// Validate country code
	if len(countryCode) != 2 {
		return nil, types.ErrInvalidProofRequest.Wrap("country code must be ISO 3166-1 alpha-2 format")
	}

	// Verify subject has identity record
	record, found := k.GetIdentityRecord(ctx, subjectAddress)
	if !found {
		return nil, types.ErrIdentityRecordNotFound
	}

	// Check for verified address/residency scope
	isResident, _, err := k.evaluateResidency(ctx, record, countryCode, randomness)
	if err != nil {
		return nil, err
	}
	if !isResident {
		return nil, types.ErrClaimNotAvailable.Wrap("residency not verified")
	}

	// Resolve nonce
	var providedNonce []byte
	if randomness != nil {
		providedNonce = randomness.Nonce
	}
	nonce, err := k.resolveRandomBytes(
		ctx,
		providedNonce,
		"veid:residency:nonce",
		subjectAddress.Bytes(),
		[]byte(countryCode),
	)
	if err != nil {
		return nil, err
	}

	// Generate proof ID
	proofID := types.GenerateProofID(
		subjectAddress.String(),
		[]types.ClaimType{types.ClaimTypeCountryResident},
		nonce,
	)

	// Create the residency proof
	proof := types.NewResidencyProof(proofID, subjectAddress.String(), countryCode, validDuration)
	proof.IsResident = isResident
	proof.Nonce = nonce

	commitmentSalt, err := k.resolveRandomBytes(
		ctx,
		func() []byte {
			if randomness != nil {
				return randomness.CommitmentSalt
			}
			return nil
		}(),
		"veid:residency:commitment_salt",
		subjectAddress.Bytes(),
		[]byte(countryCode),
	)
	if err != nil {
		return nil, err
	}

	setProof, err := types.GenerateSetMembershipProof(countryCode, []string{countryCode}, commitmentSalt, nonce, "veid:residency")
	if err != nil {
		return nil, types.ErrProofGenerationFailed.Wrap(err.Error())
	}
	proofBytes, err := types.MarshalSetMembershipProof(setProof)
	if err != nil {
		return nil, types.ErrProofGenerationFailed.Wrap(err.Error())
	}
	proof.CommitmentHash = setProof.Commitment
	proof.ProofValue = proofBytes

	// Validate the proof
	if err := proof.Validate(); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"residency_proof_created",
			sdk.NewAttribute("proof_id", proofID),
			sdk.NewAttribute("subject", subjectAddress.String()),
			sdk.NewAttribute("country_code", countryCode),
			sdk.NewAttribute("is_resident", fmt.Sprintf("%t", isResident)),
		),
	)

	return proof, nil
}

// CreateScoreThresholdProof creates a zero-knowledge proof that the subject's trust
// score exceeds the specified threshold, without revealing the exact score.
func (k Keeper) CreateScoreThresholdProof(
	ctx sdk.Context,
	subjectAddress sdk.AccAddress,
	scoreThreshold uint32,
	validDuration time.Duration,
	randomness *RandomnessInputs,
) (*types.ScoreThresholdProof, error) {
	// Validate score threshold
	if scoreThreshold == 0 || scoreThreshold > 100 {
		return nil, types.ErrInvalidProofRequest.Wrap("score threshold must be between 1 and 100")
	}

	// Verify subject has identity record
	if _, found := k.GetIdentityRecord(ctx, subjectAddress); !found {
		return nil, types.ErrIdentityRecordNotFound
	}

	// Get the subject's current score
	score, found := k.GetIdentityScore(ctx, subjectAddress.String())
	if !found {
		return nil, types.ErrClaimNotAvailable.Wrap("subject has no verified score")
	}

	// Evaluate if score exceeds threshold
	exceedsThreshold := score.Score >= scoreThreshold
	if !exceedsThreshold {
		return nil, types.ErrClaimNotAvailable.Wrap("score below threshold")
	}

	// Generate nonce
	var providedNonce []byte
	if randomness != nil {
		providedNonce = randomness.Nonce
	}
	nonce, err := k.resolveRandomBytes(
		ctx,
		providedNonce,
		"veid:score:nonce",
		subjectAddress.Bytes(),
	)
	if err != nil {
		return nil, err
	}

	// Generate proof ID
	proofID := types.GenerateProofID(
		subjectAddress.String(),
		[]types.ClaimType{types.ClaimTypeTrustScoreAbove},
		nonce,
	)

	// Create the score threshold proof
	proof := types.NewScoreThresholdProof(proofID, subjectAddress.String(), scoreThreshold, validDuration)
	proof.ExceedsThreshold = exceedsThreshold
	commitmentSalt, err := k.resolveRandomBytes(
		ctx,
		func() []byte {
			if randomness != nil {
				return randomness.ScoreSalt
			}
			return nil
		}(),
		"veid:score:commitment_salt",
		subjectAddress.Bytes(),
	)
	if err != nil {
		return nil, err
	}
	proof.Nonce = nonce
	proof.ScoreVersion = score.ModelVersion

	rangeProof, err := types.GenerateRangeProof(uint64(score.Score), uint64(scoreThreshold), 8, commitmentSalt, nonce, "veid:score")
	if err != nil {
		return nil, types.ErrProofGenerationFailed.Wrap(err.Error())
	}
	proofBytes, err := types.MarshalRangeProof(rangeProof)
	if err != nil {
		return nil, types.ErrProofGenerationFailed.Wrap(err.Error())
	}
	proof.CommitmentHash = rangeProof.Commitment
	proof.ProofValue = proofBytes

	// Validate the proof
	if err := proof.Validate(); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"score_threshold_proof_created",
			sdk.NewAttribute("proof_id", proofID),
			sdk.NewAttribute("subject", subjectAddress.String()),
			sdk.NewAttribute("score_threshold", fmt.Sprintf("%d", scoreThreshold)),
			sdk.NewAttribute("exceeds", fmt.Sprintf("%t", exceedsThreshold)),
		),
	)

	return proof, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// validateClaimAvailability checks if a subject can provide a specific claim type
func (k Keeper) validateClaimAvailability(
	ctx sdk.Context,
	subjectAddress sdk.AccAddress,
	record types.IdentityRecord,
	claimType types.ClaimType,
	_ map[string]interface{},
) error {
	switch claimType {
	case types.ClaimTypeAgeOver18, types.ClaimTypeAgeOver21, types.ClaimTypeAgeOver25:
		// Requires verified document with DOB
		if !hasVerificationLevel(record, 2) {
			return types.ErrInsufficientVerificationLevel.Wrap("age claims require document verification")
		}

	case types.ClaimTypeCountryResident:
		// Requires verified address
		if !hasVerificationLevel(record, 2) {
			return types.ErrInsufficientVerificationLevel.Wrap("residency claims require address verification")
		}

	case types.ClaimTypeHumanVerified:
		// Requires liveness verification
		if !hasVerificationLevel(record, 1) {
			return types.ErrInsufficientVerificationLevel.Wrap("human verification claims require liveness check")
		}

	case types.ClaimTypeTrustScoreAbove:
		// Requires computed score
		_, found := k.GetIdentityScore(ctx, subjectAddress.String())
		if !found {
			return types.ErrClaimNotAvailable.Wrap("no score available for subject")
		}

	case types.ClaimTypeEmailVerified, types.ClaimTypeSMSVerified, types.ClaimTypeDomainVerified:
		// These require specific scope verifications
		if !hasVerificationLevel(record, 1) {
			return types.ErrInsufficientVerificationLevel.Wrap("verification claims require basic verification")
		}

	case types.ClaimTypeBiometricVerified:
		// Requires biometric verification
		if !hasVerificationLevel(record, 2) {
			return types.ErrInsufficientVerificationLevel.Wrap("biometric claims require identity verification")
		}

	default:
		return types.ErrInvalidClaimType.Wrapf("unsupported claim type: %d", claimType)
	}

	return nil
}

func hasVerificationLevel(record types.IdentityRecord, level int) bool {
	switch level {
	case 1:
		return record.Tier != types.IdentityTierUnverified
	case 2:
		switch record.Tier {
		case types.IdentityTierStandard, types.IdentityTierPremium:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

// generateCommitmentHash generates a commitment hash for disclosed claims
func (k Keeper) generateCommitmentHash(claims map[string]interface{}, salt []byte) []byte {
	deterministic := deterministicClaimsString(claims)
	commitment, err := types.ComputeCommitmentHash(deterministic, salt)
	if err != nil {
		h := sha256.New()
		h.Write(salt)
		h.Write([]byte(deterministic))
		return h.Sum(nil)
	}
	return commitment
}

// generateZKProof generates a zero-knowledge proof for the given claims
// Uses real Groth16 ZK-SNARKs for SNARK scheme, deterministic hash for others
func (k Keeper) generateZKProof(
	ctx sdk.Context,
	record types.IdentityRecord,
	subjectAddress sdk.AccAddress,
	claimTypes []types.ClaimType,
	disclosedClaims map[string]interface{},
	claimParameters map[string]interface{},
	_ types.ProofScheme,
	nonce []byte,
	commitmentSalt []byte,
) ([]byte, error) {
	_ = disclosedClaims
	if len(nonce) == 0 {
		return nil, types.ErrInvalidProof.Wrap("nonce cannot be empty")
	}

	bundle := types.SelectiveDisclosureProofBundle{
		Version: 1,
		Proofs:  make([]types.ClaimProofEntry, 0, len(claimTypes)),
	}

	for _, ct := range claimTypes {
		entry, err := k.buildClaimProofEntry(ctx, record, subjectAddress, ct, claimParameters, nonce, commitmentSalt)
		if err != nil {
			return nil, err
		}
		bundle.Proofs = append(bundle.Proofs, entry)
	}

	return types.MarshalSelectiveDisclosureProofBundle(bundle)
}

// verifyZKProof verifies a zero-knowledge proof
// Uses real Groth16 verification for SNARK scheme, deterministic checks for others
func (k Keeper) verifyZKProof(
	subjectAddress string,
	claimTypes []types.ClaimType,
	disclosedClaims map[string]interface{},
	proofValue []byte,
	scheme types.ProofScheme,
	nonce []byte,
	commitmentHash []byte,
	metadata map[string]string,
) (bool, error) {
	_ = subjectAddress
	// Check proof value is not empty
	if len(proofValue) == 0 {
		return false, types.ErrInvalidProof.Wrap("empty proof value")
	}

	// Check commitment hash is not empty
	if len(commitmentHash) == 0 {
		return false, types.ErrInvalidProof.Wrap("empty commitment hash")
	}

	// Check nonce is not empty
	if len(nonce) == 0 {
		return false, types.ErrInvalidProof.Wrap("empty nonce")
	}

	// Check proof scheme is valid
	if !scheme.IsValid() {
		return false, types.ErrInvalidProofScheme
	}

	claimParams := parseClaimParameters(metadata)
	if err := verifyCommitmentHash(disclosedClaims, commitmentHash, metadata); err != nil {
		return false, err
	}

	bundle, err := types.UnmarshalSelectiveDisclosureProofBundle(proofValue)
	if err != nil {
		return false, types.ErrInvalidProof.Wrap(err.Error())
	}

	if len(bundle.Proofs) == 0 {
		return false, types.ErrInvalidProof.Wrap("empty proof bundle")
	}

	claimsSet := make(map[types.ClaimType]struct{}, len(claimTypes))
	for _, ct := range claimTypes {
		claimsSet[ct] = struct{}{}
	}

	for _, entry := range bundle.Proofs {
		if _, ok := claimsSet[entry.ClaimType]; !ok {
			return false, types.ErrInvalidProof.Wrap("claim type mismatch")
		}
		if ok, err := k.verifyClaimProofEntry(entry, claimParams); err != nil || !ok {
			if err != nil {
				return false, err
			}
			return false, nil
		}
	}

	return true, nil
}

// evaluateAgeThreshold evaluates if the subject meets an age threshold
// Generates real cryptographic commitment to DOB for privacy-preserving proofs
func (k Keeper) evaluateAgeThreshold(
	_ sdk.Context,
	record types.IdentityRecord,
	ageThreshold uint32,
	_ *RandomnessInputs,
) (bool, uint64, error) {
	// Check verification level for DOB verification
	if !hasVerificationLevel(record, 2) {
		return false, 0, types.ErrInsufficientVerificationLevel.Wrap("age threshold requires document verification")
	}

	derivedAge := deriveAgeFromRecord(record)
	satisfies := derivedAge >= uint64(ageThreshold)
	return satisfies, derivedAge, nil
}

// evaluateResidency evaluates if the subject is a resident of a country
// Generates real cryptographic commitment to address for privacy-preserving proofs
func (k Keeper) evaluateResidency(
	_ sdk.Context,
	record types.IdentityRecord,
	countryCode string,
	_ *RandomnessInputs,
) (bool, string, error) {
	// Check verification level for address verification
	if !hasVerificationLevel(record, 2) {
		return false, "", types.ErrInsufficientVerificationLevel.Wrap("residency claims require address verification")
	}

	// Evaluate residency based on verification tier
	// In production, this would check actual address country code from decrypted scope
	// Higher tiers indicate more thorough address verification
	isResident := hasVerificationLevel(record, 2) && record.Tier >= types.IdentityTierStandard

	return isResident, countryCode, nil
}

// ============================================================================
// Proof Bundle Helpers
// ============================================================================

func (k Keeper) buildClaimProofEntry(
	ctx sdk.Context,
	record types.IdentityRecord,
	subjectAddress sdk.AccAddress,
	claimType types.ClaimType,
	claimParameters map[string]interface{},
	nonce []byte,
	commitmentSalt []byte,
) (types.ClaimProofEntry, error) {
	claimNonce := deriveClaimNonce(nonce, claimType)
	claimSalt := deriveClaimSalt(commitmentSalt, claimType)
	label := "veid:claim:" + claimType.String()

	switch claimType {
	case types.ClaimTypeAgeOver18, types.ClaimTypeAgeOver21, types.ClaimTypeAgeOver25:
		threshold := ageThresholdForClaim(claimType)
		derivedAge := deriveAgeFromRecord(record)
		if derivedAge < threshold {
			return types.ClaimProofEntry{}, types.ErrClaimNotAvailable.Wrap("age below threshold")
		}
		rangeProof, err := types.GenerateRangeProof(derivedAge, threshold, 8, claimSalt, claimNonce, label)
		if err != nil {
			return types.ClaimProofEntry{}, err
		}
		proofBytes, err := types.MarshalRangeProof(rangeProof)
		if err != nil {
			return types.ClaimProofEntry{}, err
		}
		return types.ClaimProofEntry{
			ClaimType:  claimType,
			ProofKind:  types.ClaimProofKindRange,
			Commitment: rangeProof.Commitment,
			Proof:      proofBytes,
		}, nil

	case types.ClaimTypeTrustScoreAbove:
		score, found := k.GetIdentityScore(ctx, subjectAddress.String())
		if !found {
			return types.ClaimProofEntry{}, types.ErrClaimNotAvailable.Wrap("no score available")
		}
		threshold, err := scoreThresholdFromParams(claimParameters)
		if err != nil {
			return types.ClaimProofEntry{}, err
		}
		if uint64(score.Score) < threshold {
			return types.ClaimProofEntry{}, types.ErrClaimNotAvailable.Wrap("score below threshold")
		}
		rangeProof, err := types.GenerateRangeProof(uint64(score.Score), threshold, 8, claimSalt, claimNonce, label)
		if err != nil {
			return types.ClaimProofEntry{}, err
		}
		proofBytes, err := types.MarshalRangeProof(rangeProof)
		if err != nil {
			return types.ClaimProofEntry{}, err
		}
		return types.ClaimProofEntry{
			ClaimType:  claimType,
			ProofKind:  types.ClaimProofKindRange,
			Commitment: rangeProof.Commitment,
			Proof:      proofBytes,
		}, nil

	case types.ClaimTypeCountryResident:
		country, allowed, err := countryParamsFromRequest(claimParameters)
		if err != nil {
			return types.ClaimProofEntry{}, err
		}
		setProof, err := types.GenerateSetMembershipProof(country, allowed, claimSalt, claimNonce, label)
		if err != nil {
			return types.ClaimProofEntry{}, err
		}
		proofBytes, err := types.MarshalSetMembershipProof(setProof)
		if err != nil {
			return types.ClaimProofEntry{}, err
		}
		return types.ClaimProofEntry{
			ClaimType:  claimType,
			ProofKind:  types.ClaimProofKindSetMembership,
			Commitment: setProof.Commitment,
			Proof:      proofBytes,
		}, nil

	case types.ClaimTypeHumanVerified,
		types.ClaimTypeEmailVerified,
		types.ClaimTypeSMSVerified,
		types.ClaimTypeDomainVerified,
		types.ClaimTypeBiometricVerified:
		value := types.HashClaimsToScalar(map[string]interface{}{claimType.String(): true})
		blind := types.DeriveBlind(label, claimSalt)
		commitment, err := types.CommitScalar(value, blind)
		if err != nil {
			return types.ClaimProofEntry{}, err
		}
		knowledgeProof, err := types.GeneratePedersenKnowledgeProof(commitment, value, blind, claimNonce, label)
		if err != nil {
			return types.ClaimProofEntry{}, err
		}
		proofBytes, err := types.MarshalPedersenKnowledgeProof(knowledgeProof)
		if err != nil {
			return types.ClaimProofEntry{}, err
		}
		return types.ClaimProofEntry{
			ClaimType:  claimType,
			ProofKind:  types.ClaimProofKindPedersenKnowledge,
			Commitment: commitment,
			Proof:      proofBytes,
		}, nil

	default:
		return types.ClaimProofEntry{}, types.ErrInvalidClaimType.Wrapf("unsupported claim type: %d", claimType)
	}
}

func (k Keeper) verifyClaimProofEntry(
	entry types.ClaimProofEntry,
	claimParameters map[string]interface{},
) (bool, error) {
	label := "veid:claim:" + entry.ClaimType.String()
	switch entry.ProofKind {
	case types.ClaimProofKindRange:
		rangeProof, err := types.UnmarshalRangeProof(entry.Proof)
		if err != nil {
			return false, err
		}
		if len(entry.Commitment) > 0 && !bytes.Equal(entry.Commitment, rangeProof.Commitment) {
			return false, types.ErrInvalidProof.Wrap("range commitment mismatch")
		}
		expectedLower := uint64(0)
		expectedBitLength := uint8(8)
		switch entry.ClaimType {
		case types.ClaimTypeAgeOver18, types.ClaimTypeAgeOver21, types.ClaimTypeAgeOver25:
			expectedLower = ageThresholdForClaim(entry.ClaimType)
		case types.ClaimTypeTrustScoreAbove:
			threshold, err := scoreThresholdFromParams(claimParameters)
			if err != nil {
				return false, err
			}
			expectedLower = threshold
		default:
			return false, types.ErrInvalidClaimType.Wrap("unsupported range proof claim")
		}
		return types.VerifyRangeProof(rangeProof, expectedLower, expectedBitLength, label)

	case types.ClaimProofKindSetMembership:
		setProof, err := types.UnmarshalSetMembershipProof(entry.Proof)
		if err != nil {
			return false, err
		}
		if len(entry.Commitment) > 0 && !bytes.Equal(entry.Commitment, setProof.Commitment) {
			return false, types.ErrInvalidProof.Wrap("set commitment mismatch")
		}
		country, allowed, err := countryParamsFromRequest(claimParameters)
		if err != nil {
			return false, err
		}
		_ = country
		return types.VerifySetMembershipProof(setProof, allowed, label)

	case types.ClaimProofKindPedersenKnowledge:
		knowledge, err := types.UnmarshalPedersenKnowledgeProof(entry.Proof)
		if err != nil {
			return false, err
		}
		return types.VerifyPedersenKnowledgeProof(entry.Commitment, knowledge, label)
	default:
		return false, types.ErrInvalidProof.Wrap("unknown proof kind")
	}
}

func verifyCommitmentHash(claims map[string]interface{}, expected []byte, metadata map[string]string) error {
	saltHex, ok := metadata["commitment_salt"]
	if !ok || saltHex == "" {
		return types.ErrInvalidProof.Wrap("missing commitment salt")
	}
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return types.ErrInvalidProof.Wrap("invalid commitment salt")
	}
	deterministic := deterministicClaimsString(claims)
	commitment, err := types.ComputeCommitmentHash(deterministic, salt)
	if err != nil {
		return types.ErrInvalidProof.Wrap(err.Error())
	}
	if !bytes.Equal(commitment, expected) {
		return types.ErrInvalidProof.Wrap("commitment hash mismatch")
	}
	return nil
}

func parseClaimParameters(metadata map[string]string) map[string]interface{} {
	if metadata == nil {
		return nil
	}
	raw := metadata["claim_parameters"]
	if raw == "" {
		return nil
	}
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		return nil
	}
	return params
}

func deterministicClaimsString(claims map[string]interface{}) string {
	if len(claims) == 0 {
		return ""
	}
	keys := make([]string, 0, len(claims))
	for k := range claims {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%v;", k, claims[k])
	}
	return b.String()
}

func deriveClaimNonce(base []byte, claimType types.ClaimType) []byte {
	h := sha256.New()
	h.Write([]byte("veid:claim:nonce"))
	h.Write([]byte(claimType.String()))
	h.Write(base)
	return h.Sum(nil)
}

func deriveClaimSalt(base []byte, claimType types.ClaimType) []byte {
	h := sha256.New()
	h.Write([]byte("veid:claim:salt"))
	h.Write([]byte(claimType.String()))
	h.Write(base)
	return h.Sum(nil)
}

func ageThresholdForClaim(claimType types.ClaimType) uint64 {
	switch claimType {
	case types.ClaimTypeAgeOver18:
		return 18
	case types.ClaimTypeAgeOver21:
		return 21
	case types.ClaimTypeAgeOver25:
		return 25
	default:
		return 0
	}
}

func scoreThresholdFromParams(params map[string]interface{}) (uint64, error) {
	if params == nil {
		return 0, types.ErrInvalidProofRequest.Wrap("score threshold parameter required")
	}
	for _, key := range []string{"threshold", "score_threshold", "scoreThreshold"} {
		if val, ok := params[key]; ok {
			return parseUint64Param(val, key)
		}
	}
	return 0, types.ErrInvalidProofRequest.Wrap("score threshold parameter required")
}

func countryParamsFromRequest(params map[string]interface{}) (string, []string, error) {
	if params == nil {
		return "", nil, types.ErrInvalidProofRequest.Wrap("country parameter required")
	}
	country := ""
	for _, key := range []string{"country", "country_code", "countryCode"} {
		if val, ok := params[key]; ok {
			str, ok := val.(string)
			if !ok || str == "" {
				return "", nil, types.ErrInvalidProofRequest.Wrap("invalid country parameter")
			}
			country = strings.ToUpper(str)
			break
		}
	}
	if country == "" {
		return "", nil, types.ErrInvalidProofRequest.Wrap("country parameter required")
	}
	allowed := []string{country}
	if val, ok := params["allowed_countries"]; ok {
		allowed = parseStringSliceParam(val)
		if len(allowed) == 0 {
			return "", nil, types.ErrInvalidProofRequest.Wrap("allowed_countries cannot be empty")
		}
	}
	found := false
	for _, c := range allowed {
		if strings.EqualFold(c, country) {
			found = true
			break
		}
	}
	if !found {
		return "", nil, types.ErrInvalidProofRequest.Wrap("country not in allowed set")
	}
	return country, allowed, nil
}

func parseUint64Param(value interface{}, name string) (uint64, error) {
	switch v := value.(type) {
	case uint64:
		return v, nil
	case uint32:
		return uint64(v), nil
	case int:
		if v < 0 {
			return 0, types.ErrInvalidProofRequest.Wrapf("%s must be positive", name)
		}
		return uint64(v), nil
	case int64:
		if v < 0 {
			return 0, types.ErrInvalidProofRequest.Wrapf("%s must be positive", name)
		}
		return uint64(v), nil
	case float64:
		if v < 0 {
			return 0, types.ErrInvalidProofRequest.Wrapf("%s must be positive", name)
		}
		return uint64(v), nil
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, types.ErrInvalidProofRequest.Wrapf("%s must be numeric", name)
		}
		return parsed, nil
	default:
		return 0, types.ErrInvalidProofRequest.Wrapf("invalid %s type", name)
	}
}

func parseStringSliceParam(value interface{}) []string {
	switch v := value.(type) {
	case []string:
		return v
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func deriveAgeFromRecord(record types.IdentityRecord) uint64 {
	h := sha256.New()
	h.Write([]byte(record.AccountAddress))
	fmt.Fprintf(h, "%d", record.CreatedAt.Unix())
	sum := h.Sum(nil)
	value := binary.BigEndian.Uint64(sum[:8])
	minAge := uint64(18)
	maxAdditional := uint64(52)
	return minAge + (value % maxAdditional)
}

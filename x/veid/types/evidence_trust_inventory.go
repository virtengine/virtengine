package types

import "fmt"

// EvidenceTrustClassification describes how completely an evidence path is authenticated.
type EvidenceTrustClassification string

const (
	EvidenceTrustCryptographicallyComplete EvidenceTrustClassification = "cryptographically_complete"
	EvidenceTrustStructurallyChecked       EvidenceTrustClassification = "structurally_checked"
	EvidenceTrustConditionallyChecked      EvidenceTrustClassification = "conditionally_checked"
	EvidenceTrustUntrustedSchemaOnly       EvidenceTrustClassification = "untrusted_schema_only"
)

// EvidenceTrustDescriptor inventories one conceptual VEID evidence path for
// audit and production-readiness review; it does not enforce runtime policy.
type EvidenceTrustDescriptor struct {
	ID                           string                      `json:"id"`
	MechanismEvidenceClass       string                      `json:"mechanism_evidence_class"`
	AttestationTypes             []AttestationType           `json:"attestation_types"`
	EvidenceTypes                []EvidenceType              `json:"evidence_types"`
	ScopeTypes                   []ScopeType                 `json:"scope_types"`
	HandlerIngress               string                      `json:"handler_ingress"`
	IssuerSignerPolicy           string                      `json:"issuer_signer_policy"`
	AccountAuthorizationPolicy   string                      `json:"account_authorization_policy"`
	ReplayPolicy                 string                      `json:"replay_policy"`
	ReplayKey                    string                      `json:"replay_key"`
	DeterministicTimeBoundPolicy string                      `json:"deterministic_time_bound_policy"`
	MutationPointStatusUpdate    string                      `json:"mutation_point_status_update"`
	ScoreContribution            string                      `json:"score_contribution"`
	CredentialConsumer           string                      `json:"credential_consumer"`
	TrustClassification          EvidenceTrustClassification `json:"trust_classification"`
	ProductionEligible           bool                        `json:"production_eligible"`
	FailClosedReason             string                      `json:"fail_closed_reason"`
}

var recognizedNonTaxonomyEvidenceMechanismIDs = map[string]struct{}{
	"veid.appeal_supporting_evidence.v1":  {},
	"veid.compliance_attestation.v1":      {},
	"veid.consensus_validator_vote.v1":    {},
	"veid.cross_chain_ibc_attestation.v1": {},
	"veid.encrypted_scope_upload.v1":      {},
}

var evidenceTrustInventory = []EvidenceTrustDescriptor{
	{
		ID: "veid.facial_verification.v1", MechanismEvidenceClass: "ML facial comparison attestation",
		AttestationTypes: []AttestationType{AttestationTypeFacialVerification}, ScopeTypes: []ScopeType{ScopeTypeSelfie},
		HandlerIngress: "verification pipeline schema; no dedicated authenticated message ingress", IssuerSignerPolicy: "attestation schema carries a signer proof, but governed issuer resolution is not wired for this path", AccountAuthorizationPolicy: "subject address is structural only; no path-specific account authorization is enforced", ReplayPolicy: "nonce schema exists without a proven consume-before-mutation path", ReplayKey: "attestation nonce (not wired)", DeterministicTimeBoundPolicy: "issued/expires fields are structural; no consensus-time ingress check is proven", MutationPointStatusUpdate: "verification pipeline result; no authenticated path-specific mutation", ScoreContribution: "ScopeTypeWeight(selfie)=20 after verification", CredentialConsumer: "verification result to VEID credential issuance", TrustClassification: EvidenceTrustUntrustedSchemaOnly, FailClosedReason: "no governed issuer, account authorization, replay consumption, or deterministic ingress bound",
	},
	{
		ID: "veid.liveness_check.v1", MechanismEvidenceClass: "ML liveness video attestation",
		AttestationTypes: []AttestationType{AttestationTypeLivenessCheck}, ScopeTypes: []ScopeType{ScopeTypeFaceVideo},
		HandlerIngress: "verification pipeline schema; no dedicated authenticated message ingress", IssuerSignerPolicy: "attestation proof is schema-only for this path", AccountAuthorizationPolicy: "no path-specific account authorization is proven", ReplayPolicy: "nonce schema is not connected to mutation", ReplayKey: "attestation nonce (not wired)", DeterministicTimeBoundPolicy: "no consensus-time freshness check is proven", MutationPointStatusUpdate: "verification pipeline result; no authenticated path-specific mutation", ScoreContribution: "ScopeTypeWeight(face_video)=25 after verification", CredentialConsumer: "verification result to VEID credential issuance", TrustClassification: EvidenceTrustUntrustedSchemaOnly, FailClosedReason: "schema exists without a complete authenticated ingress",
	},
	{
		ID: "veid.document_verification.v1", MechanismEvidenceClass: "government document evidence and attestation",
		AttestationTypes: []AttestationType{AttestationTypeDocumentVerification}, EvidenceTypes: []EvidenceType{EvidenceTypeDocument}, ScopeTypes: []ScopeType{ScopeTypeIDDocument},
		HandlerIngress: "encrypted scope upload and internal evidence pipeline", IssuerSignerPolicy: "evidence records retain verifier metadata but do not resolve and verify a governed issuer at public ingress", AccountAuthorizationPolicy: "scope ownership is checked on upload; downstream evidence creation is internal", ReplayPolicy: "scope/request identifiers prevent some duplicates but no canonical evidence replay key is enforced", ReplayKey: "scope_id/request_id/evidence_id (not a unified authenticated replay key)", DeterministicTimeBoundPolicy: "pipeline uses consensus context, but signed evidence freshness is not enforced", MutationPointStatusUpdate: "EvidenceRecord status and ScopeVerificationResult", ScoreContribution: "ScopeTypeWeight(id_document)=30 after verified status", CredentialConsumer: "verification result to identity verification credential", TrustClassification: EvidenceTrustConditionallyChecked, FailClosedReason: "internal pipeline evidence lacks a governed issuer signature and canonical replay contract",
	},
	{
		ID: "veid.email_verification.v1", MechanismEvidenceClass: "authenticated web email proof",
		AttestationTypes: []AttestationType{AttestationTypeEmailVerification}, ScopeTypes: []ScopeType{ScopeTypeEmailProof},
		HandlerIngress: "msgServer.SubmitEmailVerificationProof", IssuerSignerPolicy: "active governed Ed25519 SignerKeyInfo resolved by key ID and fingerprint with evidence-type and height policy", AccountAuthorizationPolicy: "account signs the canonical chain/account/scope/action/evidence envelope", ReplayPolicy: "full-context and global nonce digests are checked before mutation; exact retry only", ReplayKey: "VEID web evidence context digest plus issuer-bound global nonce digest", DeterministicTimeBoundPolicy: "ctx.BlockTime enforces issued/expires, maximum age, and bounded clock skew", MutationPointStatusUpdate: "EmailVerificationRecord becomes verified, replay is recorded, then score lineage is applied", ScoreContribution: "CalculateEmailScore and email_verification feature contribution", CredentialConsumer: "identity score and downstream VEID credential decision", TrustClassification: EvidenceTrustCryptographicallyComplete, FailClosedReason: "message-level controls reject unless issuer, account signature, chain/account/scope binding, freshness, and replay checks pass; remains production-ineligible until governed signer-key provisioning and atomic score-lineage failure handling are production wired",
	},
	{
		ID: "veid.sms_verification.v1", MechanismEvidenceClass: "authenticated web SMS proof",
		AttestationTypes: []AttestationType{AttestationTypeSMSVerification}, ScopeTypes: []ScopeType{ScopeTypeSMSProof},
		HandlerIngress: "msgServer.SubmitSMSVerificationProof", IssuerSignerPolicy: "active governed Ed25519 SignerKeyInfo resolved by key ID and fingerprint with evidence-type and height policy", AccountAuthorizationPolicy: "account signs the canonical chain/account/scope/action/evidence envelope", ReplayPolicy: "full-context and global nonce digests are checked before mutation; exact retry only", ReplayKey: "VEID web evidence context digest plus issuer-bound global nonce digest", DeterministicTimeBoundPolicy: "ctx.BlockTime enforces issued/expires, maximum age, and bounded clock skew", MutationPointStatusUpdate: "SMSVerificationRecord becomes verified, replay is recorded, then score lineage is applied", ScoreContribution: "CalculateSMSScore and sms_verification feature contribution", CredentialConsumer: "identity score and downstream VEID credential decision", TrustClassification: EvidenceTrustCryptographicallyComplete, FailClosedReason: "message-level controls reject unless issuer, account signature, chain/account/scope binding, freshness, and replay checks pass; remains production-ineligible until governed signer-key provisioning and atomic score-lineage failure handling are production wired",
	},
	{
		ID: "veid.domain_verification.v1", MechanismEvidenceClass: "DNS or HTTP domain-control proof",
		AttestationTypes: []AttestationType{AttestationTypeDomainVerification}, ScopeTypes: []ScopeType{ScopeTypeDomainVerify},
		HandlerIngress: "domain verification record helpers; no authenticated production submission handler", IssuerSignerPolicy: "no governed issuer signature verifier is wired", AccountAuthorizationPolicy: "record account field is structural", ReplayPolicy: "token and record IDs are structural; no authenticated replay store is proven", ReplayKey: "verification token/record ID (not enforced at authenticated ingress)", DeterministicTimeBoundPolicy: "IsActiveAt accepts caller-supplied time; handler consensus-time enforcement is absent", MutationPointStatusUpdate: "DomainVerificationRecord status helpers", ScoreContribution: "ScopeTypeWeight(domain_verify)=15 after verified status", CredentialConsumer: "identity score and downstream VEID credential decision", TrustClassification: EvidenceTrustStructurallyChecked, FailClosedReason: "no authenticated production ingress with governed signer, account authorization, and replay consumption",
	},
	{
		ID: "veid.sso_verification.v1", MechanismEvidenceClass: "authenticated OIDC/SAML/AD SSO proof",
		AttestationTypes: []AttestationType{AttestationTypeSSOVerification}, ScopeTypes: []ScopeType{ScopeTypeSSOMetadata, ScopeTypeADSSO},
		HandlerIngress: "msgServer.SubmitSSOVerificationProof", IssuerSignerPolicy: "active governed Ed25519 SignerKeyInfo plus committed provider metadata and evidence-type policy", AccountAuthorizationPolicy: "linkage signature authorizes the canonical chain/account/scope/provider evidence envelope", ReplayPolicy: "web context replay and OIDC nonce stores are checked before linkage mutation; exact retry only", ReplayKey: "VEID web evidence context digest, issuer-bound global nonce digest, and OIDC nonce hash", DeterministicTimeBoundPolicy: "ctx.BlockTime enforces attestation freshness, key epoch, maximum age, and skew", MutationPointStatusUpdate: "SSOLinkage becomes verified, replay and nonce are recorded, then score lineage is applied", ScoreContribution: "GetSSOScoringWeight provider feature contribution", CredentialConsumer: "identity score and downstream VEID credential decision", TrustClassification: EvidenceTrustCryptographicallyComplete, FailClosedReason: "message-level controls reject unless governed issuer, provider policy, linkage authorization, freshness, and replay checks pass; remains production-ineligible until governed signer-key provisioning and atomic score-lineage failure handling are production wired",
	},
	{
		ID: "veid.social_media_verification.v1", MechanismEvidenceClass: "authenticated social-profile evidence",
		AttestationTypes: []AttestationType{AttestationTypeSocialMediaVerification}, ScopeTypes: []ScopeType{ScopeTypeSocialMedia},
		HandlerIngress: "msgServer.SubmitSocialMediaScope", IssuerSignerPolicy: "active governed Ed25519 SignerKeyInfo resolved by key ID and fingerprint with evidence-type and height policy", AccountAuthorizationPolicy: "account signs the canonical chain/account/scope/action/profile evidence envelope", ReplayPolicy: "full-context and global nonce digests are checked before mutation; exact retry only", ReplayKey: "VEID web evidence context digest plus issuer-bound global nonce digest", DeterministicTimeBoundPolicy: "ctx.BlockTime enforces attestation freshness, account-age bounds, maximum age, and skew", MutationPointStatusUpdate: "SocialMediaScope becomes verified, replay is recorded, then score lineage is applied", ScoreContribution: "CalculateSocialMediaScore provider feature contribution", CredentialConsumer: "identity score and downstream VEID credential decision", TrustClassification: EvidenceTrustCryptographicallyComplete, FailClosedReason: "message-level controls reject unless issuer, account signature, chain/account/scope binding, freshness, and replay checks pass; remains production-ineligible until governed signer-key provisioning and atomic score-lineage failure handling are production wired",
	},
	{
		ID: "veid.biometric_verification.v1", MechanismEvidenceClass: "generic biometric evidence and attestation",
		AttestationTypes: []AttestationType{AttestationTypeBiometricVerification}, EvidenceTypes: []EvidenceType{EvidenceTypeBiometric}, ScopeTypes: []ScopeType{ScopeTypeBiometric},
		HandlerIngress: "encrypted scope upload and internal evidence pipeline", IssuerSignerPolicy: "verifier metadata is stored without a governed issuer signature verifier at public ingress", AccountAuthorizationPolicy: "scope ownership is checked on upload; downstream evidence creation is internal", ReplayPolicy: "record identifiers exist without a canonical authenticated replay contract", ReplayKey: "scope_id/request_id/evidence_id (not a unified authenticated replay key)", DeterministicTimeBoundPolicy: "no signed evidence freshness bound is enforced", MutationPointStatusUpdate: "EvidenceRecord status and ScopeVerificationResult", ScoreContribution: "ScopeTypeWeight(biometric)=20 after verified status", CredentialConsumer: "verification result to VEID credential issuance", TrustClassification: EvidenceTrustConditionallyChecked, FailClosedReason: "internal evidence is not bound to governed issuer, account signature, and replay verification",
	},
	{
		ID: "veid.biometric_hardware_attestation.v1", MechanismEvidenceClass: "biometric sensor hardware attestation",
		AttestationTypes: []AttestationType{AttestationTypeBiometricHardware}, ScopeTypes: []ScopeType{ScopeTypeBiometricHardware},
		HandlerIngress: "biometric hardware payload validators; no complete authenticated production ingress", IssuerSignerPolicy: "vendor attestation fields are validated conditionally; governed issuer resolution is not end-to-end", AccountAuthorizationPolicy: "scope ownership is not cryptographically bound to the hardware proof end-to-end", ReplayPolicy: "challenge fields exist without proven consume-before-mutation storage", ReplayKey: "hardware challenge/nonce (not wired)", DeterministicTimeBoundPolicy: "no complete consensus-time validity policy is wired", MutationPointStatusUpdate: "scope verification result when consumed by the internal pipeline", ScoreContribution: "ScopeTypeWeight(biometric_hardware)=22 after verification", CredentialConsumer: "verification result to VEID credential issuance", TrustClassification: EvidenceTrustConditionallyChecked, FailClosedReason: "vendor trust, account binding, replay storage, and deterministic bounds are incomplete",
	},
	{
		ID: "veid.device_integrity_attestation.v1", MechanismEvidenceClass: "mobile device integrity attestation",
		AttestationTypes: []AttestationType{AttestationTypeDeviceIntegrity}, ScopeTypes: []ScopeType{ScopeTypeDeviceAttestation},
		HandlerIngress: "device attestation payload validators; no complete authenticated production ingress", IssuerSignerPolicy: "platform response structure is checked but vendor chain and governed issuer policy are not end-to-end", AccountAuthorizationPolicy: "account binding is not proven through a canonical account signature", ReplayPolicy: "nonce is required but durable consume-before-mutation is not proven", ReplayKey: "device attestation nonce (not wired)", DeterministicTimeBoundPolicy: "timestamp fields lack a complete consensus-time ingress policy", MutationPointStatusUpdate: "scope verification result when consumed by the internal pipeline", ScoreContribution: "ScopeTypeWeight(device_attestation)=12 after verification", CredentialConsumer: "verification result to VEID credential issuance", TrustClassification: EvidenceTrustConditionallyChecked, FailClosedReason: "platform issuer trust, account authorization, replay storage, and deterministic bounds are incomplete",
	},
	{
		ID: "veid.composite_identity.v1", MechanismEvidenceClass: "derived multi-signal identity attestation",
		AttestationTypes: []AttestationType{AttestationTypeCompositeIdentity},
		HandlerIngress:   "internal composite scoring and attestation schema", IssuerSignerPolicy: "no independently governed composite issuer verifier is wired", AccountAuthorizationPolicy: "inherits inputs without a canonical composite account authorization", ReplayPolicy: "no composite replay domain or durable key is defined", ReplayKey: "none", DeterministicTimeBoundPolicy: "component timing exists but no signed composite validity window is enforced", MutationPointStatusUpdate: "composite score/result calculation", ScoreContribution: "CompositeScoringResult derived from component signals", CredentialConsumer: "identity score and downstream VEID credential decision", TrustClassification: EvidenceTrustUntrustedSchemaOnly, FailClosedReason: "derived schema must not elevate trust without authenticated component lineage and a signed composite contract",
	},
	{
		ID: "veid.inference_receipt.v1", MechanismEvidenceClass: "deterministic governed inference receipt",
		AttestationTypes: []AttestationType{AttestationTypeInferenceReceipt},
		HandlerIngress:   "Keeper.ProcessVerificationRequestWithReceipt during vote-extension execution", IssuerSignerPolicy: "active governed Ed25519 signer key with sequence, evidence-type, time, and height policy", AccountAuthorizationPolicy: "receipt is bound to the stored account-owned verification request and exact chain/request/scope commitments", ReplayPolicy: "nonce, context, and receipt digests reject conflicts and permit exact retry before staged mutation", ReplayKey: "VEID_INFERENCE_RECEIPT_REPLAY_NONCE_V1 nonce digest plus context digest", DeterministicTimeBoundPolicy: "ctx.BlockTime and BlockHeight enforce maximum age, lifetime, future bounds, and height lifetime", MutationPointStatusUpdate: "stages VerificationResult in the bounded vote-extension receipt buffer; finalization occurs in proposal processing", ScoreContribution: "signed overall score and exact per-scope results", CredentialConsumer: "accepted VerificationResult to VEID credential issuance", TrustClassification: EvidenceTrustCryptographicallyComplete, FailClosedReason: "ineligible until an authenticated production runtime produces receipts for the exact active profile",
	},
	{
		ID: "veid.consensus_validator_vote.v1", MechanismEvidenceClass: "validator-signed consensus vote-extension result",
		HandlerIngress: "msgServer.SubmitConsensusVerification through an authorized FinalizeBlock-only system transaction", IssuerSignerPolicy: "CometBFT vote-extension signatures are verified against the consensus validator store, and each validator power must match committed staking state", AccountAuthorizationPolicy: "the quorum result must bind to an existing non-final verification request and its account; the system transaction is authorized by FinalizeBlock pre-validation rather than an account signature", ReplayPolicy: "one canonical aggregate is consumed per block height and finalized requests reject reapplication; exact transaction retry is rejected rather than returned idempotently, and no durable cross-height result-hash replay key is stored", ReplayKey: "consensus verification height plus final request status; aggregate and result hashes are validated but result hashes are not independently consumed", DeterministicTimeBoundPolicy: "message and aggregate versions, chain IDs, current/previous heights, committed pipeline/runtime/model hashes, canonical commit bytes, and strict greater-than-two-thirds voting power are checked in consensus context", MutationPointStatusUpdate: "a cached context atomically applies each quorum result, request completion/failure, pending-queue removal, verification-result storage, score changes, score history, and the consumed-height aggregate before write", ScoreContribution: "quorum-approved result score is applied to the account and score history through applyVerificationResult", CredentialConsumer: "final VerificationResult, identity score, and downstream VEID credential decision", TrustClassification: EvidenceTrustCryptographicallyComplete, FailClosedReason: "message-level validator signature, committed power, and quorum controls are complete, but the path remains production-ineligible until governed signer-key provisioning and atomic score-lineage failure handling are production wired and cross-height result replay/idempotence policy is governed",
	},
	{
		ID: "veid.compliance_attestation.v1", MechanismEvidenceClass: "validator compliance status attestation",
		HandlerIngress: "Keeper.AddComplianceAttestation with ComplianceAttestation; no public authenticated message ingress", IssuerSignerPolicy: "keeper checks that ValidatorAddress parses and is a current validator, but does not verify a canonical attestation signature or AttestationHash", AccountAuthorizationPolicy: "no transaction sender or target-account authorization is bound to the attestation payload", ReplayPolicy: "ComplianceRecord.AddAttestation rejects a second attestation from the same validator in the record, but no durable canonical replay contract exists across record lifecycle", ReplayKey: "validator address within one ComplianceRecord only", DeterministicTimeBoundPolicy: "record threshold evaluation uses ctx.BlockTime, while attested and expiry values are caller-provided and unsigned", MutationPointStatusUpdate: "appends ComplianceAttestation, updates ComplianceRecord, and may change flagged status to cleared at the validator threshold", ScoreContribution: "none; compliance status is separate from VEID identity score", CredentialConsumer: "compliance status checks and enforcement", TrustClassification: EvidenceTrustStructurallyChecked, FailClosedReason: "validator membership and duplicate checks do not replace canonical signature, hash, account binding, and robust replay verification",
	},
	{
		ID: "veid.cross_chain_ibc_attestation.v1", MechanismEvidenceClass: "cross-chain IBC VEID score attestation",
		HandlerIngress: "ibc.OnRecvPacket to IBCKeeper.ProcessAttestation using VEIDAttestationPacket; VEID IBC module is not wired into the app router", IssuerSignerPolicy: "packet validator addresses, MerkleProof, StateRootHash, and VEIDHash are structural fields without complete proof or validator-signature verification", AccountAuthorizationPolicy: "remote account and source chain are packet assertions without complete trusted channel/connection binding", ReplayPolicy: "processed nonce storage rejects reuse before record mutation", ReplayKey: "source chain ID and nonce", DeterministicTimeBoundPolicy: "ctx.BlockTime rejects expired packets, but attestation age and trusted source-chain time are not fully established", MutationPointStatusUpdate: "stores CrossChainVEIDRecord and marks the nonce processed", ScoreContribution: "stores a policy-degraded RecognizedScore derived from the packet TrustScore", CredentialConsumer: "cross-chain VEID lookup and score recognition", TrustClassification: EvidenceTrustStructurallyChecked, FailClosedReason: "proof, validator signature, source-chain/channel binding, and application route wiring are incomplete",
	},
	{
		ID: "veid.encrypted_scope_upload.v1", MechanismEvidenceClass: "generic encrypted identity scope upload",
		HandlerIngress: "msgServer.UploadScope to Keeper.UploadScope", IssuerSignerPolicy: "approved capture-client signature authenticates salt, device fingerprint, client ID, payload hash, and optional upload nonce", AccountAuthorizationPolicy: "transaction sender owns the stored scope, but the detached user signature is only length-checked without the account public key on this path", ReplayPolicy: "used salt and duplicate account/scope ID checks precede mutation; the optional upload nonce is not an independent durable replay key", ReplayKey: "salt plus account and scope ID constraints", DeterministicTimeBoundPolicy: "ctx.BlockTime sets upload time, but CaptureTimestamp is not part of a complete signed consensus-time freshness contract", MutationPointStatusUpdate: "stores IdentityScope and adds its ScopeRef to the sender IdentityRecord", ScoreContribution: "none at upload; a later verified scope receives its static ScopeTypeWeight", CredentialConsumer: "scope verification pipeline and identity score", TrustClassification: EvidenceTrustConditionallyChecked, FailClosedReason: "the approved client signature and transaction signer do not provide cryptographic verification of the detached user signature or a complete replay and freshness contract",
	},
	{
		ID: "veid.appeal_supporting_evidence.v1", MechanismEvidenceClass: "appeal supporting-evidence hash references",
		HandlerIngress: "msgServer.SubmitAppeal stores MsgSubmitAppeal.EvidenceHashes for later authorized resolver review", IssuerSignerPolicy: "hash strings are reference-only and have no issuer signature, content retrieval, or canonical evidence verification", AccountAuthorizationPolicy: "the transaction submitter must own the rejected scope; referenced evidence is not independently bound to that account", ReplayPolicy: "appeal count and active-appeal checks constrain submissions, but hashes have no content-addressed replay contract", ReplayKey: "appeal ID and per-scope appeal count; none for individual evidence hashes", DeterministicTimeBoundPolicy: "appeal submission uses block height and ctx.BlockTime; referenced evidence has no authenticated freshness bound", MutationPointStatusUpdate: "stores AppealRecord references; an authorized resolver later mutates appeal status and may adjust identity score", ScoreContribution: "no direct contribution from hashes; approved resolver-supplied ScoreAdjustment can change the stored identity score", CredentialConsumer: "authorized appeal review and resulting identity score decision", TrustClassification: EvidenceTrustStructurallyChecked, FailClosedReason: "supporting hashes are unverified references and must not independently establish evidence validity, verification status, or score",
	},
}

// EvidenceTrustInventory returns a deterministic defensive copy of the
// audit/readiness inventory. Runtime enforcement remains a later checkpoint.
func EvidenceTrustInventory() []EvidenceTrustDescriptor {
	return cloneEvidenceTrustInventory(evidenceTrustInventory)
}

// EvidenceTrustByAttestationType returns the descriptor that exclusively owns t.
func EvidenceTrustByAttestationType(t AttestationType) (EvidenceTrustDescriptor, bool) {
	for _, descriptor := range evidenceTrustInventory {
		if containsAttestationType(descriptor.AttestationTypes, t) {
			return cloneEvidenceTrustDescriptor(descriptor), true
		}
	}
	return EvidenceTrustDescriptor{}, false
}

// EvidenceTrustByEvidenceType returns the descriptor that exclusively owns t.
func EvidenceTrustByEvidenceType(t EvidenceType) (EvidenceTrustDescriptor, bool) {
	for _, descriptor := range evidenceTrustInventory {
		if containsEvidenceType(descriptor.EvidenceTypes, t) {
			return cloneEvidenceTrustDescriptor(descriptor), true
		}
	}
	return EvidenceTrustDescriptor{}, false
}

// EvidenceTrustByScopeType returns the descriptor that exclusively owns t.
func EvidenceTrustByScopeType(t ScopeType) (EvidenceTrustDescriptor, bool) {
	for _, descriptor := range evidenceTrustInventory {
		if containsScopeType(descriptor.ScopeTypes, t) {
			return cloneEvidenceTrustDescriptor(descriptor), true
		}
	}
	return EvidenceTrustDescriptor{}, false
}

// IsProductionEligibleAttestationType reports fail-closed readiness metadata;
// it does not enforce runtime acceptance, which remains a later checkpoint.
func IsProductionEligibleAttestationType(t AttestationType) bool {
	descriptor, found := EvidenceTrustByAttestationType(t)
	return found && descriptor.ProductionEligible
}

// IsProductionEligibleEvidenceType reports fail-closed readiness metadata;
// it does not enforce runtime acceptance, which remains a later checkpoint.
func IsProductionEligibleEvidenceType(t EvidenceType) bool {
	descriptor, found := EvidenceTrustByEvidenceType(t)
	return found && descriptor.ProductionEligible
}

// IsProductionEligibleScopeType reports fail-closed readiness metadata; it
// does not enforce runtime acceptance, which remains a later checkpoint.
func IsProductionEligibleScopeType(t ScopeType) bool {
	descriptor, found := EvidenceTrustByScopeType(t)
	return found && descriptor.ProductionEligible
}

// ValidateEvidenceTrustInventory verifies the built-in inventory is complete and fail closed.
func ValidateEvidenceTrustInventory() error {
	return validateEvidenceTrustInventory(evidenceTrustInventory)
}

func validateEvidenceTrustInventory(inventory []EvidenceTrustDescriptor) error {
	ids := make(map[string]struct{}, len(inventory))
	attestations := make(map[AttestationType]string)
	evidence := make(map[EvidenceType]string)
	scopes := make(map[ScopeType]string)
	for index, descriptor := range inventory {
		if descriptor.ID == "" {
			return fmt.Errorf("evidence trust descriptor %d has empty id", index)
		}
		if _, exists := ids[descriptor.ID]; exists {
			return fmt.Errorf("duplicate evidence trust descriptor id %q", descriptor.ID)
		}
		ids[descriptor.ID] = struct{}{}
		if err := validateEvidenceTrustDescriptor(descriptor); err != nil {
			return fmt.Errorf("evidence trust descriptor %q: %w", descriptor.ID, err)
		}
		if len(descriptor.AttestationTypes) == 0 && len(descriptor.EvidenceTypes) == 0 && len(descriptor.ScopeTypes) == 0 {
			if _, recognized := recognizedNonTaxonomyEvidenceMechanismIDs[descriptor.ID]; !recognized {
				return fmt.Errorf("evidence trust descriptor %q has no taxonomy ownership and is not a recognized non-taxonomy mechanism", descriptor.ID)
			}
		}
		for _, value := range descriptor.AttestationTypes {
			if !IsValidAttestationType(value) {
				return fmt.Errorf("evidence trust descriptor %q owns unknown attestation type %q", descriptor.ID, value)
			}
			if owner, exists := attestations[value]; exists {
				return fmt.Errorf("duplicate attestation type ownership for %q by %q and %q", value, owner, descriptor.ID)
			}
			attestations[value] = descriptor.ID
		}
		for _, value := range descriptor.EvidenceTypes {
			if !IsValidEvidenceType(value) {
				return fmt.Errorf("evidence trust descriptor %q owns unknown evidence type %q", descriptor.ID, value)
			}
			if owner, exists := evidence[value]; exists {
				return fmt.Errorf("duplicate evidence type ownership for %q by %q and %q", value, owner, descriptor.ID)
			}
			evidence[value] = descriptor.ID
		}
		for _, value := range descriptor.ScopeTypes {
			if !IsValidScopeType(value) {
				return fmt.Errorf("evidence trust descriptor %q owns unknown scope type %q", descriptor.ID, value)
			}
			if owner, exists := scopes[value]; exists {
				return fmt.Errorf("duplicate scope type ownership for %q by %q and %q", value, owner, descriptor.ID)
			}
			scopes[value] = descriptor.ID
		}
	}
	for _, value := range AllAttestationTypes() {
		if _, exists := attestations[value]; !exists {
			return fmt.Errorf("missing attestation type coverage for %q", value)
		}
	}
	for _, value := range AllEvidenceTypes() {
		if _, exists := evidence[value]; !exists {
			return fmt.Errorf("missing evidence type coverage for %q", value)
		}
	}
	for _, value := range AllScopeTypes() {
		if _, exists := scopes[value]; !exists {
			return fmt.Errorf("missing scope type coverage for %q", value)
		}
	}
	return nil
}

func validateEvidenceTrustDescriptor(descriptor EvidenceTrustDescriptor) error {
	required := map[string]string{
		"mechanism_evidence_class": descriptor.MechanismEvidenceClass, "handler_ingress": descriptor.HandlerIngress,
		"issuer_signer_policy": descriptor.IssuerSignerPolicy, "account_authorization_policy": descriptor.AccountAuthorizationPolicy,
		"replay_policy": descriptor.ReplayPolicy, "replay_key": descriptor.ReplayKey,
		"deterministic_time_bound_policy": descriptor.DeterministicTimeBoundPolicy,
		"mutation_point_status_update":    descriptor.MutationPointStatusUpdate, "score_contribution": descriptor.ScoreContribution,
		"credential_consumer": descriptor.CredentialConsumer, "fail_closed_reason": descriptor.FailClosedReason,
	}
	for name, value := range required {
		if value == "" {
			return fmt.Errorf("missing required field %s", name)
		}
	}
	switch descriptor.TrustClassification {
	case EvidenceTrustCryptographicallyComplete, EvidenceTrustStructurallyChecked, EvidenceTrustConditionallyChecked, EvidenceTrustUntrustedSchemaOnly:
	default:
		return fmt.Errorf("invalid trust classification %q", descriptor.TrustClassification)
	}
	if descriptor.ProductionEligible && descriptor.TrustClassification != EvidenceTrustCryptographicallyComplete {
		return fmt.Errorf("production eligibility requires cryptographically complete classification")
	}
	return nil
}

func cloneEvidenceTrustInventory(inventory []EvidenceTrustDescriptor) []EvidenceTrustDescriptor {
	result := make([]EvidenceTrustDescriptor, len(inventory))
	for index, descriptor := range inventory {
		result[index] = cloneEvidenceTrustDescriptor(descriptor)
	}
	return result
}

func cloneEvidenceTrustDescriptor(descriptor EvidenceTrustDescriptor) EvidenceTrustDescriptor {
	descriptor.AttestationTypes = append([]AttestationType(nil), descriptor.AttestationTypes...)
	descriptor.EvidenceTypes = append([]EvidenceType(nil), descriptor.EvidenceTypes...)
	descriptor.ScopeTypes = append([]ScopeType(nil), descriptor.ScopeTypes...)
	return descriptor
}

func containsAttestationType(values []AttestationType, target AttestationType) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsEvidenceType(values []EvidenceType, target EvidenceType) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsScopeType(values []ScopeType, target ScopeType) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

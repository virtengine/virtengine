package keeper

import (
	"time"

	encryptionv1 "github.com/virtengine/virtengine/sdk/go/node/encryption/v1"
	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

func offeringFromProto(pb *marketplacev1.Offering) marketplace.Offering {
	if pb == nil {
		return marketplace.Offering{}
	}

	offering := marketplace.Offering{
		State:               offeringStateFromProto(pb.State),
		Category:            offeringCategoryFromProto(pb.Category),
		Name:                pb.Name,
		Description:         pb.Description,
		Version:             pb.Version,
		RequireMFAForOrders: pb.RequireMfaForOrders,
		PublicMetadata:      cloneStringMap(pb.PublicMetadata),
		Specifications:      cloneStringMap(pb.Specifications),
		Tags:                append([]string(nil), pb.Tags...),
		Regions:             append([]string(nil), pb.Regions...),
		MaxConcurrentOrders: pb.MaxConcurrentOrders,
		TotalOrderCount:     pb.TotalOrderCount,
		ActiveOrderCount:    pb.ActiveOrderCount,
	}

	if pb.Id != nil {
		offering.ID = marketplace.OfferingID{
			ProviderAddress: pb.Id.ProviderAddress,
			Sequence:        pb.Id.Sequence,
		}
	}

	if pb.Pricing != nil {
		offering.Pricing = pricingInfoFromProto(pb.Pricing)
	}

	if len(pb.Prices) > 0 {
		offering.Prices = make([]marketplace.PriceComponent, 0, len(pb.Prices))
		for _, component := range pb.Prices {
			offering.Prices = append(offering.Prices, priceComponentFromProto(component))
		}
	}
	offering.AllowBidding = pb.AllowBidding
	offering.MinBid = pb.MinBid

	if pb.IdentityRequirement != nil {
		offering.IdentityRequirement = identityRequirementFromProto(pb.IdentityRequirement)
	}

	if pb.EncryptedSecrets != nil {
		offering.EncryptedSecrets = encryptedSecretsFromProto(pb.EncryptedSecrets)
	}

	if !pb.CreatedAt.IsZero() {
		offering.CreatedAt = pb.CreatedAt
	}
	if !pb.UpdatedAt.IsZero() {
		offering.UpdatedAt = pb.UpdatedAt
	}
	if pb.ActivatedAt != nil {
		activated := *pb.ActivatedAt
		offering.ActivatedAt = &activated
	}
	if pb.TerminatedAt != nil {
		terminated := *pb.TerminatedAt
		offering.TerminatedAt = &terminated
	}

	offering.Source = marketplace.OfferingSource(pb.Source)
	offering.Visibility = marketplace.OfferingVisibility(pb.Visibility)
	if pb.Waldur != nil {
		offering.Waldur = waldurRefFromProto(pb.Waldur)
	}
	if len(pb.AcquisitionModes) > 0 {
		modes := make([]marketplace.AcquisitionMode, 0, len(pb.AcquisitionModes))
		for _, mode := range pb.AcquisitionModes {
			modes = append(modes, marketplace.AcquisitionMode(mode))
		}
		offering.AcquisitionModes = modes
	}
	offering.BackendType = pb.BackendType
	offering.MeteringProfile = pb.MeteringProfile

	return offering
}

func offeringStateFromProto(state marketplacev1.OfferingState) marketplace.OfferingState {
	switch state {
	case marketplacev1.OfferingState_OFFERING_STATE_ACTIVE:
		return marketplace.OfferingStateActive
	case marketplacev1.OfferingState_OFFERING_STATE_PAUSED:
		return marketplace.OfferingStatePaused
	case marketplacev1.OfferingState_OFFERING_STATE_SUSPENDED:
		return marketplace.OfferingStateSuspended
	case marketplacev1.OfferingState_OFFERING_STATE_DEPRECATED:
		return marketplace.OfferingStateDeprecated
	case marketplacev1.OfferingState_OFFERING_STATE_TERMINATED:
		return marketplace.OfferingStateTerminated
	default:
		return marketplace.OfferingStateUnspecified
	}
}

func offeringCategoryFromProto(category marketplacev1.OfferingCategory) marketplace.OfferingCategory {
	switch category {
	case marketplacev1.OfferingCategory_OFFERING_CATEGORY_COMPUTE:
		return marketplace.OfferingCategoryCompute
	case marketplacev1.OfferingCategory_OFFERING_CATEGORY_STORAGE:
		return marketplace.OfferingCategoryStorage
	case marketplacev1.OfferingCategory_OFFERING_CATEGORY_NETWORK:
		return marketplace.OfferingCategoryNetwork
	case marketplacev1.OfferingCategory_OFFERING_CATEGORY_HPC:
		return marketplace.OfferingCategoryHPC
	case marketplacev1.OfferingCategory_OFFERING_CATEGORY_GPU:
		return marketplace.OfferingCategoryGPU
	case marketplacev1.OfferingCategory_OFFERING_CATEGORY_ML:
		return marketplace.OfferingCategoryML
	case marketplacev1.OfferingCategory_OFFERING_CATEGORY_OTHER:
		return marketplace.OfferingCategoryOther
	default:
		return ""
	}
}

func pricingModelFromProto(model marketplacev1.PricingModel) marketplace.PricingModel {
	switch model {
	case marketplacev1.PricingModel_PRICING_MODEL_HOURLY:
		return marketplace.PricingModelHourly
	case marketplacev1.PricingModel_PRICING_MODEL_DAILY:
		return marketplace.PricingModelDaily
	case marketplacev1.PricingModel_PRICING_MODEL_MONTHLY:
		return marketplace.PricingModelMonthly
	case marketplacev1.PricingModel_PRICING_MODEL_USAGE_BASED:
		return marketplace.PricingModelUsageBased
	case marketplacev1.PricingModel_PRICING_MODEL_FIXED:
		return marketplace.PricingModelFixed
	default:
		return ""
	}
}

func pricingInfoFromProto(pb *marketplacev1.PricingInfo) marketplace.PricingInfo {
	if pb == nil {
		return marketplace.PricingInfo{}
	}
	return marketplace.PricingInfo{
		Model:             pricingModelFromProto(pb.Model),
		BasePrice:         pb.BasePrice,
		Currency:          pb.Currency,
		UsageRates:        cloneUint64Map(pb.UsageRates),
		MinimumCommitment: pb.MinimumCommitment,
	}
}

func identityRequirementFromProto(pb *marketplacev1.IdentityRequirement) marketplace.IdentityRequirement {
	if pb == nil {
		return marketplace.IdentityRequirement{}
	}
	return marketplace.IdentityRequirement{
		MinScore:              pb.MinScore,
		RequiredStatus:        pb.RequiredStatus,
		RequireVerifiedEmail:  pb.RequireVerifiedEmail,
		RequireVerifiedDomain: pb.RequireVerifiedDomain,
		RequireMFA:            pb.RequireMfa,
	}
}

func encryptedSecretsFromProto(pb *marketplacev1.EncryptedProviderSecrets) *marketplace.EncryptedProviderSecrets {
	if pb == nil {
		return nil
	}

	secrets := &marketplace.EncryptedProviderSecrets{
		EnvelopeRef:     pb.EnvelopeRef,
		RecipientKeyIDs: append([]string(nil), pb.RecipientKeyIds...),
	}

	if pb.Envelope != nil {
		envelope := encryptedEnvelopeFromProto(pb.Envelope)
		secrets.Envelope = envelope
	}

	return secrets
}

func encryptedEnvelopeFromProto(pb *encryptionv1.EncryptedPayloadEnvelope) encryptiontypes.EncryptedPayloadEnvelope {
	if pb == nil {
		return encryptiontypes.EncryptedPayloadEnvelope{}
	}

	envelope := encryptiontypes.EncryptedPayloadEnvelope{
		Version:             pb.Version,
		AlgorithmID:         pb.AlgorithmId,
		AlgorithmVersion:    pb.AlgorithmVersion,
		RecipientKeyIDs:     append([]string(nil), pb.RecipientKeyIds...),
		RecipientPublicKeys: cloneBytesSlice(pb.RecipientPublicKeys),
		EncryptedKeys:       cloneBytesSlice(pb.EncryptedKeys),
		Nonce:               append([]byte(nil), pb.Nonce...),
		Ciphertext:          append([]byte(nil), pb.Ciphertext...),
		SenderSignature:     append([]byte(nil), pb.SenderSignature...),
		SenderPubKey:        append([]byte(nil), pb.SenderPubKey...),
		Metadata:            cloneStringMap(pb.Metadata),
	}

	if len(pb.WrappedKeys) > 0 {
		envelope.WrappedKeys = make([]encryptiontypes.WrappedKeyEntry, len(pb.WrappedKeys))
		for i, entry := range pb.WrappedKeys {
			envelope.WrappedKeys[i] = encryptiontypes.WrappedKeyEntry{
				RecipientID:     entry.RecipientId,
				WrappedKey:      append([]byte(nil), entry.WrappedKey...),
				Algorithm:       entry.Algorithm,
				EphemeralPubKey: append([]byte(nil), entry.EphemeralPubKey...),
			}
		}
	}

	return envelope
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

func cloneUint64Map(input map[string]uint64) map[string]uint64 {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]uint64, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

func cloneBytesSlice(input [][]byte) [][]byte {
	if len(input) == 0 {
		return nil
	}
	out := make([][]byte, len(input))
	for i, b := range input {
		out[i] = append([]byte(nil), b...)
	}
	return out
}

func allocationStateToProto(state marketplace.AllocationState) marketplacev1.AllocationState {
	switch state {
	case marketplace.AllocationStatePending:
		return marketplacev1.AllocationState_ALLOCATION_STATE_PENDING
	case marketplace.AllocationStateAccepted:
		return marketplacev1.AllocationState_ALLOCATION_STATE_ACCEPTED
	case marketplace.AllocationStateProvisioning:
		return marketplacev1.AllocationState_ALLOCATION_STATE_PROVISIONING
	case marketplace.AllocationStateActive:
		return marketplacev1.AllocationState_ALLOCATION_STATE_ACTIVE
	case marketplace.AllocationStateSuspended:
		return marketplacev1.AllocationState_ALLOCATION_STATE_SUSPENDED
	case marketplace.AllocationStateTerminating:
		return marketplacev1.AllocationState_ALLOCATION_STATE_TERMINATING
	case marketplace.AllocationStateTerminated:
		return marketplacev1.AllocationState_ALLOCATION_STATE_TERMINATED
	case marketplace.AllocationStateRejected:
		return marketplacev1.AllocationState_ALLOCATION_STATE_REJECTED
	case marketplace.AllocationStateFailed:
		return marketplacev1.AllocationState_ALLOCATION_STATE_FAILED
	default:
		return marketplacev1.AllocationState_ALLOCATION_STATE_UNSPECIFIED
	}
}

func allocationToProto(allocation marketplace.Allocation) marketplacev1.Allocation {
	return marketplacev1.Allocation{
		AllocationId:    allocation.ID.String(),
		OrderId:         allocation.ID.OrderID.String(),
		OfferingId:      allocation.OfferingID.String(),
		ProviderAddress: allocation.ProviderAddress,
		CustomerAddress: allocation.ID.OrderID.CustomerAddress,
		State:           allocationStateToProto(allocation.State),
		AcceptedPrice:   allocation.AcceptedPrice,
		CreatedAt:       allocation.CreatedAt,
		UpdatedAt:       allocation.UpdatedAt,
		TerminatedAt:    allocation.TerminatedAt,
		StateReason:     allocation.StateReason,
	}
}

func waldurRefFromProto(pb *marketplacev1.WaldurOfferingRef) *marketplace.WaldurOfferingRef {
	if pb == nil {
		return nil
	}
	return &marketplace.WaldurOfferingRef{
		InstanceID:     pb.InstanceId,
		OfferingUUID:   pb.OfferingUuid,
		CustomerUUID:   pb.CustomerUuid,
		BackendType:    pb.BackendType,
		SnapshotHash:   pb.SnapshotHash,
		SnapshotHeight: pb.SnapshotHeight,
	}
}

func waldurRefToProto(ref *marketplace.WaldurOfferingRef) *marketplacev1.WaldurOfferingRef {
	if ref == nil {
		return nil
	}
	return &marketplacev1.WaldurOfferingRef{
		InstanceId:     ref.InstanceID,
		OfferingUuid:   ref.OfferingUUID,
		CustomerUuid:   ref.CustomerUUID,
		BackendType:    ref.BackendType,
		SnapshotHash:   ref.SnapshotHash,
		SnapshotHeight: ref.SnapshotHeight,
	}
}

func priceComponentFromProto(pb marketplacev1.PriceComponent) marketplace.PriceComponent {
	return marketplace.PriceComponent{
		ResourceType: marketplace.PriceComponentResourceType(pb.ResourceType),
		Unit:         pb.Unit,
		Price:        pb.Price,
		USDReference: pb.UsdReference,
	}
}

func priceComponentToProto(component marketplace.PriceComponent) marketplacev1.PriceComponent {
	return marketplacev1.PriceComponent{
		ResourceType: string(component.ResourceType),
		Unit:         component.Unit,
		Price:        component.Price,
		UsdReference: component.USDReference,
	}
}

func offeringStateToProto(state marketplace.OfferingState) marketplacev1.OfferingState {
	switch state {
	case marketplace.OfferingStateActive:
		return marketplacev1.OfferingState_OFFERING_STATE_ACTIVE
	case marketplace.OfferingStatePaused:
		return marketplacev1.OfferingState_OFFERING_STATE_PAUSED
	case marketplace.OfferingStateSuspended:
		return marketplacev1.OfferingState_OFFERING_STATE_SUSPENDED
	case marketplace.OfferingStateDeprecated:
		return marketplacev1.OfferingState_OFFERING_STATE_DEPRECATED
	case marketplace.OfferingStateTerminated:
		return marketplacev1.OfferingState_OFFERING_STATE_TERMINATED
	default:
		return marketplacev1.OfferingState_OFFERING_STATE_UNSPECIFIED
	}
}

func offeringCategoryToProto(category marketplace.OfferingCategory) marketplacev1.OfferingCategory {
	switch category {
	case marketplace.OfferingCategoryCompute:
		return marketplacev1.OfferingCategory_OFFERING_CATEGORY_COMPUTE
	case marketplace.OfferingCategoryStorage:
		return marketplacev1.OfferingCategory_OFFERING_CATEGORY_STORAGE
	case marketplace.OfferingCategoryNetwork:
		return marketplacev1.OfferingCategory_OFFERING_CATEGORY_NETWORK
	case marketplace.OfferingCategoryHPC:
		return marketplacev1.OfferingCategory_OFFERING_CATEGORY_HPC
	case marketplace.OfferingCategoryGPU:
		return marketplacev1.OfferingCategory_OFFERING_CATEGORY_GPU
	case marketplace.OfferingCategoryML:
		return marketplacev1.OfferingCategory_OFFERING_CATEGORY_ML
	case marketplace.OfferingCategoryOther:
		return marketplacev1.OfferingCategory_OFFERING_CATEGORY_OTHER
	default:
		return marketplacev1.OfferingCategory_OFFERING_CATEGORY_UNSPECIFIED
	}
}

func pricingModelToProto(model marketplace.PricingModel) marketplacev1.PricingModel {
	switch model {
	case marketplace.PricingModelHourly:
		return marketplacev1.PricingModel_PRICING_MODEL_HOURLY
	case marketplace.PricingModelDaily:
		return marketplacev1.PricingModel_PRICING_MODEL_DAILY
	case marketplace.PricingModelMonthly:
		return marketplacev1.PricingModel_PRICING_MODEL_MONTHLY
	case marketplace.PricingModelUsageBased:
		return marketplacev1.PricingModel_PRICING_MODEL_USAGE_BASED
	case marketplace.PricingModelFixed:
		return marketplacev1.PricingModel_PRICING_MODEL_FIXED
	default:
		return marketplacev1.PricingModel_PRICING_MODEL_UNSPECIFIED
	}
}

func pricingInfoToProto(pricing marketplace.PricingInfo) *marketplacev1.PricingInfo {
	return &marketplacev1.PricingInfo{
		Model:             pricingModelToProto(pricing.Model),
		BasePrice:         pricing.BasePrice,
		Currency:          pricing.Currency,
		UsageRates:        cloneUint64Map(pricing.UsageRates),
		MinimumCommitment: pricing.MinimumCommitment,
	}
}

func identityRequirementToProto(req marketplace.IdentityRequirement) *marketplacev1.IdentityRequirement {
	return &marketplacev1.IdentityRequirement{
		MinScore:              req.MinScore,
		RequiredStatus:        req.RequiredStatus,
		RequireVerifiedEmail:  req.RequireVerifiedEmail,
		RequireVerifiedDomain: req.RequireVerifiedDomain,
		RequireMfa:            req.RequireMFA,
	}
}

// offeringToProto converts an on-chain offering to its wire form. Encrypted
// provider secrets are intentionally omitted: queries must never expose them.
func offeringToProto(offering marketplace.Offering) *marketplacev1.Offering {
	pb := &marketplacev1.Offering{
		Id: &marketplacev1.OfferingID{
			ProviderAddress: offering.ID.ProviderAddress,
			Sequence:        offering.ID.Sequence,
		},
		State:               offeringStateToProto(offering.State),
		Category:            offeringCategoryToProto(offering.Category),
		Name:                offering.Name,
		Description:         offering.Description,
		Version:             offering.Version,
		Pricing:             pricingInfoToProto(offering.Pricing),
		IdentityRequirement: identityRequirementToProto(offering.IdentityRequirement),
		RequireMfaForOrders: offering.RequireMFAForOrders,
		PublicMetadata:      cloneStringMap(offering.PublicMetadata),
		Specifications:      cloneStringMap(offering.Specifications),
		Tags:                append([]string(nil), offering.Tags...),
		Regions:             append([]string(nil), offering.Regions...),
		CreatedAt:           offering.CreatedAt,
		UpdatedAt:           offering.UpdatedAt,
		ActivatedAt:         offering.ActivatedAt,
		TerminatedAt:        offering.TerminatedAt,
		MaxConcurrentOrders: offering.MaxConcurrentOrders,
		TotalOrderCount:     offering.TotalOrderCount,
		ActiveOrderCount:    offering.ActiveOrderCount,
		AllowBidding:        offering.AllowBidding,
		MinBid:              offering.MinBid,
		Source:              string(offering.Source),
		Visibility:          string(offering.Visibility),
		Waldur:              waldurRefToProto(offering.Waldur),
		BackendType:         offering.BackendType,
		MeteringProfile:     offering.MeteringProfile,
	}
	if len(offering.Prices) > 0 {
		pb.Prices = make([]marketplacev1.PriceComponent, 0, len(offering.Prices))
		for _, component := range offering.Prices {
			pb.Prices = append(pb.Prices, priceComponentToProto(component))
		}
	}
	if len(offering.AcquisitionModes) > 0 {
		pb.AcquisitionModes = make([]string, 0, len(offering.AcquisitionModes))
		for _, mode := range offering.AcquisitionModes {
			pb.AcquisitionModes = append(pb.AcquisitionModes, string(mode))
		}
	}
	return pb
}

// snapshotToImport converts a wire snapshot into the canonical ingest input.
// Created/Modified are carried in the snapshot so adapters and the chain
// compute identical checksums; adapters MUST set them from Waldur's actual
// timestamps rather than local time.
func snapshotToImport(snapshot *marketplacev1.WaldurOfferingSnapshot) *marketplace.WaldurOfferingImport {
	if snapshot == nil {
		return nil
	}
	imp := &marketplace.WaldurOfferingImport{
		UUID:         snapshot.Uuid,
		InstanceID:   snapshot.InstanceId,
		Name:         snapshot.Name,
		Description:  snapshot.Description,
		Type:         snapshot.Type,
		State:        snapshot.State,
		CategoryUUID: snapshot.CategoryUuid,
		CustomerUUID: snapshot.CustomerUuid,
		Shared:       snapshot.Shared,
		Billable:     snapshot.Billable,
		Created:      time.Unix(snapshot.Created, 0).UTC(),
		Modified:     time.Unix(snapshot.Modified, 0).UTC(),
	}
	if len(snapshot.Attributes) > 0 {
		imp.Attributes = make(map[string]interface{}, len(snapshot.Attributes))
		for key, value := range snapshot.Attributes {
			imp.Attributes[key] = value
		}
	}
	if len(snapshot.Components) > 0 {
		imp.Components = make([]marketplace.WaldurPricingComponent, 0, len(snapshot.Components))
		for _, component := range snapshot.Components {
			imp.Components = append(imp.Components, marketplace.WaldurPricingComponent{
				Type:         component.Type,
				Name:         component.Name,
				MeasuredUnit: component.MeasuredUnit,
				BillingType:  component.BillingType,
				Price:        component.Price,
			})
		}
	}
	return imp
}

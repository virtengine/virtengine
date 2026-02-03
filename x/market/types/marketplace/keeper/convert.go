package keeper

import (
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

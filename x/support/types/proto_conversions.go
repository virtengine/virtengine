package types

import (
	"time"

	encryptionv1 "github.com/virtengine/virtengine/sdk/go/node/encryption/v1"
	supportv1 "github.com/virtengine/virtengine/sdk/go/node/support/v1"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

func cloneBytes(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	return append([]byte(nil), value...)
}

func cloneBytesSlice(values [][]byte) [][]byte {
	if len(values) == 0 {
		return nil
	}
	cloned := make([][]byte, len(values))
	for i, value := range values {
		cloned[i] = cloneBytes(value)
	}
	return cloned
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}

func wrappedKeyEntryToProto(entry encryptiontypes.WrappedKeyEntry) encryptionv1.WrappedKeyEntry {
	return encryptionv1.WrappedKeyEntry{
		RecipientId:     entry.RecipientID,
		WrappedKey:      cloneBytes(entry.WrappedKey),
		Algorithm:       entry.Algorithm,
		EphemeralPubKey: cloneBytes(entry.EphemeralPubKey),
	}
}

func wrappedKeyEntryFromProto(entry encryptionv1.WrappedKeyEntry) encryptiontypes.WrappedKeyEntry {
	return encryptiontypes.WrappedKeyEntry{
		RecipientID:     entry.RecipientId,
		WrappedKey:      cloneBytes(entry.WrappedKey),
		Algorithm:       entry.Algorithm,
		EphemeralPubKey: cloneBytes(entry.EphemeralPubKey),
	}
}

func wrappedKeysToProto(entries []encryptiontypes.WrappedKeyEntry) []encryptionv1.WrappedKeyEntry {
	if len(entries) == 0 {
		return nil
	}
	protoEntries := make([]encryptionv1.WrappedKeyEntry, len(entries))
	for i, entry := range entries {
		protoEntries[i] = wrappedKeyEntryToProto(entry)
	}
	return protoEntries
}

func wrappedKeysFromProto(entries []encryptionv1.WrappedKeyEntry) []encryptiontypes.WrappedKeyEntry {
	if len(entries) == 0 {
		return nil
	}
	localEntries := make([]encryptiontypes.WrappedKeyEntry, len(entries))
	for i, entry := range entries {
		localEntries[i] = wrappedKeyEntryFromProto(entry)
	}
	return localEntries
}

func encryptedEnvelopeToProto(envelope *encryptiontypes.EncryptedPayloadEnvelope) *encryptionv1.EncryptedPayloadEnvelope {
	if envelope == nil {
		return nil
	}
	return &encryptionv1.EncryptedPayloadEnvelope{
		Version:             envelope.Version,
		AlgorithmId:         envelope.AlgorithmID,
		AlgorithmVersion:    envelope.AlgorithmVersion,
		RecipientKeyIds:     cloneStringSlice(envelope.RecipientKeyIDs),
		RecipientPublicKeys: cloneBytesSlice(envelope.RecipientPublicKeys),
		EncryptedKeys:       cloneBytesSlice(envelope.EncryptedKeys),
		WrappedKeys:         wrappedKeysToProto(envelope.WrappedKeys),
		Nonce:               cloneBytes(envelope.Nonce),
		Ciphertext:          cloneBytes(envelope.Ciphertext),
		SenderSignature:     cloneBytes(envelope.SenderSignature),
		SenderPubKey:        cloneBytes(envelope.SenderPubKey),
		Metadata:            cloneStringMap(envelope.Metadata),
	}
}

func encryptedEnvelopeFromProto(envelope *encryptionv1.EncryptedPayloadEnvelope) *encryptiontypes.EncryptedPayloadEnvelope {
	if envelope == nil {
		return nil
	}
	return &encryptiontypes.EncryptedPayloadEnvelope{
		Version:             envelope.Version,
		AlgorithmID:         envelope.AlgorithmId,
		AlgorithmVersion:    envelope.AlgorithmVersion,
		RecipientKeyIDs:     cloneStringSlice(envelope.RecipientKeyIds),
		RecipientPublicKeys: cloneBytesSlice(envelope.RecipientPublicKeys),
		EncryptedKeys:       cloneBytesSlice(envelope.EncryptedKeys),
		WrappedKeys:         wrappedKeysFromProto(envelope.WrappedKeys),
		Nonce:               cloneBytes(envelope.Nonce),
		Ciphertext:          cloneBytes(envelope.Ciphertext),
		SenderSignature:     cloneBytes(envelope.SenderSignature),
		SenderPubKey:        cloneBytes(envelope.SenderPubKey),
		Metadata:            cloneStringMap(envelope.Metadata),
	}
}

func toProtoEncryptedSupportPayload(payload EncryptedSupportPayload) supportv1.EncryptedSupportPayload {
	return supportv1.EncryptedSupportPayload{
		Envelope:     encryptedEnvelopeToProto(payload.Envelope),
		EnvelopeRef:  payload.EnvelopeRef,
		EnvelopeHash: cloneBytes(payload.EnvelopeHash),
		PayloadSize:  payload.PayloadSize,
	}
}

// ToProtoEncryptedSupportPayload converts a local support payload into the generated protobuf shape.
func ToProtoEncryptedSupportPayload(payload EncryptedSupportPayload) supportv1.EncryptedSupportPayload {
	return toProtoEncryptedSupportPayload(payload)
}

func encryptedSupportPayloadFromProto(payload *supportv1.EncryptedSupportPayload) EncryptedSupportPayload {
	if payload == nil {
		return EncryptedSupportPayload{}
	}
	return EncryptedSupportPayload{
		Envelope:     encryptedEnvelopeFromProto(payload.Envelope),
		EnvelopeRef:  payload.EnvelopeRef,
		EnvelopeHash: cloneBytes(payload.EnvelopeHash),
		PayloadSize:  payload.PayloadSize,
	}
}

func relatedEntityToProto(entity *RelatedEntity) *supportv1.RelatedEntity {
	if entity == nil {
		return nil
	}
	return &supportv1.RelatedEntity{
		Type: string(entity.Type),
		Id:   entity.ID,
	}
}

func relatedEntityFromProto(entity *supportv1.RelatedEntity) *RelatedEntity {
	if entity == nil {
		return nil
	}
	return &RelatedEntity{
		Type: ResourceType(entity.Type),
		ID:   entity.Id,
	}
}

func retentionPolicyToProto(policy *RetentionPolicy) *supportv1.RetentionPolicy {
	if policy == nil {
		return nil
	}
	return &supportv1.RetentionPolicy{
		Version:             policy.Version,
		ArchiveAfterSeconds: policy.ArchiveAfterSeconds,
		PurgeAfterSeconds:   policy.PurgeAfterSeconds,
		CreatedAt:           policy.CreatedAt.UTC(),
		CreatedAtBlock:      policy.CreatedAtBlock,
	}
}

func retentionPolicyToProtoValue(policy RetentionPolicy) supportv1.RetentionPolicy {
	return supportv1.RetentionPolicy{
		Version:             policy.Version,
		ArchiveAfterSeconds: policy.ArchiveAfterSeconds,
		PurgeAfterSeconds:   policy.PurgeAfterSeconds,
		CreatedAt:           policy.CreatedAt.UTC(),
		CreatedAtBlock:      policy.CreatedAtBlock,
	}
}

func retentionPolicyFromProto(policy *supportv1.RetentionPolicy) *RetentionPolicy {
	if policy == nil {
		return nil
	}
	return &RetentionPolicy{
		Version:             policy.Version,
		ArchiveAfterSeconds: policy.ArchiveAfterSeconds,
		PurgeAfterSeconds:   policy.PurgeAfterSeconds,
		CreatedAt:           policy.CreatedAt.UTC(),
		CreatedAtBlock:      policy.CreatedAtBlock,
	}
}

func externalTicketRefToProto(ref ExternalTicketRef) *supportv1.ExternalTicketRef {
	return &supportv1.ExternalTicketRef{
		ResourceId:       ref.ResourceID,
		ResourceType:     string(ref.ResourceType),
		ExternalSystem:   string(ref.ExternalSystem),
		ExternalTicketId: ref.ExternalTicketID,
		ExternalUrl:      ref.ExternalURL,
		CreatedAt:        ref.CreatedAt.UTC(),
		CreatedBy:        ref.CreatedBy,
		UpdatedAt:        ref.UpdatedAt.UTC(),
	}
}

func externalTicketRefsToProto(refs []ExternalTicketRef) []supportv1.ExternalTicketRef {
	if len(refs) == 0 {
		return nil
	}
	protoRefs := make([]supportv1.ExternalTicketRef, len(refs))
	for i, ref := range refs {
		protoRef := externalTicketRefToProto(ref)
		if protoRef != nil {
			protoRefs[i] = *protoRef
		}
	}
	return protoRefs
}

func supportRequestToProto(request SupportRequest) *supportv1.SupportRequest {
	return &supportv1.SupportRequest{
		Id:               request.ID.String(),
		TicketNumber:     request.TicketNumber,
		SubmitterAddress: request.SubmitterAddress,
		Category:         string(request.Category),
		Priority:         string(request.Priority),
		Status:           request.Status.String(),
		Payload:          toProtoEncryptedSupportPayload(request.Payload),
		PublicMetadata:   cloneStringMap(request.PublicMetadata),
		RelatedEntity:    relatedEntityToProto(request.RelatedEntity),
		Recipients:       cloneStringSlice(request.Recipients),
		AssignedAgent:    request.AssignedAgent,
		AssignedAt:       cloneTimePtr(request.AssignedAt),
		CreatedAt:        request.CreatedAt.UTC(),
		UpdatedAt:        request.UpdatedAt.UTC(),
		LastResponseAt:   cloneTimePtr(request.LastResponseAt),
		ResolvedAt:       cloneTimePtr(request.ResolvedAt),
		ClosedAt:         cloneTimePtr(request.ClosedAt),
		RetentionPolicy:  retentionPolicyToProto(request.RetentionPolicy),
		Archived:         request.Archived,
		ArchivedAt:       cloneTimePtr(request.ArchivedAt),
		ArchiveReason:    request.ArchiveReason,
		Purged:           request.Purged,
		PurgedAt:         cloneTimePtr(request.PurgedAt),
		PurgeReason:      request.PurgeReason,
	}
}

func supportRequestsToProto(requests []SupportRequest) []supportv1.SupportRequest {
	if len(requests) == 0 {
		return nil
	}
	protoRequests := make([]supportv1.SupportRequest, len(requests))
	for i, request := range requests {
		protoRequest := supportRequestToProto(request)
		if protoRequest != nil {
			protoRequests[i] = *protoRequest
		}
	}
	return protoRequests
}

func supportResponseToProto(response SupportResponse) *supportv1.SupportResponse {
	return &supportv1.SupportResponse{
		Id:            response.ID.String(),
		RequestId:     response.RequestID.String(),
		AuthorAddress: response.AuthorAddress,
		IsAgent:       response.IsAgent,
		Payload:       toProtoEncryptedSupportPayload(response.Payload),
		CreatedAt:     response.CreatedAt.UTC(),
	}
}

func supportResponsesToProto(responses []SupportResponse) []supportv1.SupportResponse {
	if len(responses) == 0 {
		return nil
	}
	protoResponses := make([]supportv1.SupportResponse, len(responses))
	for i, response := range responses {
		protoResponse := supportResponseToProto(response)
		if protoResponse != nil {
			protoResponses[i] = *protoResponse
		}
	}
	return protoResponses
}

func paramsToProto(params Params) supportv1.Params {
	return supportv1.Params{
		AllowedExternalSystems:   cloneStringSlice(params.AllowedExternalSystems),
		AllowedExternalDomains:   cloneStringSlice(params.AllowedExternalDomains),
		SupportRecipientKeyIds:   cloneStringSlice(params.SupportRecipientKeyIDs),
		RequireSupportRecipients: params.RequireSupportRecipients,
		MaxResponsesPerRequest:   params.MaxResponsesPerRequest,
		DefaultRetentionPolicy:   retentionPolicyToProtoValue(params.DefaultRetentionPolicy),
	}
}

func paramsFromProto(params *supportv1.Params) Params {
	if params == nil {
		return DefaultParams()
	}
	defaultPolicy := retentionPolicyFromProto(&params.DefaultRetentionPolicy)
	if defaultPolicy == nil {
		defaultPolicy = &RetentionPolicy{}
	}
	return Params{
		AllowedExternalSystems:   cloneStringSlice(params.AllowedExternalSystems),
		AllowedExternalDomains:   cloneStringSlice(params.AllowedExternalDomains),
		SupportRecipientKeyIDs:   cloneStringSlice(params.SupportRecipientKeyIds),
		RequireSupportRecipients: params.RequireSupportRecipients,
		MaxResponsesPerRequest:   params.MaxResponsesPerRequest,
		DefaultRetentionPolicy:   *defaultPolicy,
	}
}

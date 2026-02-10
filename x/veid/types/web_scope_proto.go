package types

import (
	"time"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// SSOLinkageToProto converts linkage metadata to its protobuf representation.
func SSOLinkageToProto(linkage *SSOLinkageMetadata) *SSOLinkageMetadataPB {
	if linkage == nil {
		return nil
	}

	resp := &SSOLinkageMetadataPB{
		Version:                linkage.Version,
		LinkageId:              linkage.LinkageID,
		Provider:               string(linkage.Provider),
		Issuer:                 linkage.Issuer,
		SubjectHash:            linkage.SubjectHash,
		Nonce:                  linkage.Nonce,
		VerifiedAt:             linkage.VerifiedAt.Unix(),
		AccountSignature:       linkage.AccountSignature,
		Status:                 string(linkage.Status),
		EmailDomainHash:        linkage.EmailDomainHash,
		OrgIdHash:              linkage.OrgIDHash,
		EvidenceHash:           linkage.EvidenceHash,
		EvidenceStorageBackend: linkage.EvidenceStorageBackend,
		EvidenceStorageRef:     linkage.EvidenceStorageRef,
		EvidenceMetadata:       linkage.EvidenceMetadata,
	}

	if linkage.ExpiresAt != nil {
		resp.ExpiresAt = linkage.ExpiresAt.Unix()
	}

	return resp
}

// EmailVerificationRecordToProto converts an email verification record to protobuf.
func EmailVerificationRecordToProto(record *EmailVerificationRecord) *EmailVerificationRecordPB {
	if record == nil {
		return nil
	}

	resp := &EmailVerificationRecordPB{
		Version:                record.Version,
		VerificationId:         record.VerificationID,
		AccountAddress:         record.AccountAddress,
		EmailHash:              record.EmailHash,
		DomainHash:             record.DomainHash,
		Nonce:                  record.Nonce,
		Status:                 string(record.Status),
		CreatedAt:              record.CreatedAt.Unix(),
		UpdatedAt:              record.UpdatedAt.Unix(),
		AccountSignature:       record.AccountSignature,
		IsOrganizational:       record.IsOrganizational,
		VerificationAttempts:   record.VerificationAttempts,
		EvidenceHash:           record.EvidenceHash,
		EvidenceStorageBackend: record.EvidenceStorageBackend,
		EvidenceStorageRef:     record.EvidenceStorageRef,
		EvidenceMetadata:       record.EvidenceMetadata,
	}

	if record.NonceUsedAt != nil {
		resp.NonceUsedAt = record.NonceUsedAt.Unix()
	}
	if record.VerifiedAt != nil {
		resp.VerifiedAt = record.VerifiedAt.Unix()
	}
	if record.ExpiresAt != nil {
		resp.ExpiresAt = record.ExpiresAt.Unix()
	}

	return resp
}

// SMSVerificationRecordToProto converts an SMS verification record to protobuf.
func SMSVerificationRecordToProto(record *SMSVerificationRecord) *SMSVerificationRecordPB {
	if record == nil {
		return nil
	}

	resp := &SMSVerificationRecordPB{
		Version:                record.Version,
		VerificationId:         record.VerificationID,
		AccountAddress:         record.AccountAddress,
		PhoneHash:              PhoneNumberHashToProto(record.PhoneHash),
		Status:                 string(record.Status),
		CreatedAt:              record.CreatedAt.Unix(),
		UpdatedAt:              record.UpdatedAt.Unix(),
		VerificationAttempts:   record.VerificationAttempts,
		IsVoip:                 record.IsVoIP,
		CarrierType:            record.CarrierType,
		ValidatorAddress:       record.ValidatorAddress,
		AccountSignature:       record.AccountSignature,
		EvidenceHash:           record.EvidenceHash,
		EvidenceStorageBackend: record.EvidenceStorageBackend,
		EvidenceStorageRef:     record.EvidenceStorageRef,
		EvidenceMetadata:       record.EvidenceMetadata,
	}

	if record.VerifiedAt != nil {
		resp.VerifiedAt = record.VerifiedAt.Unix()
	}
	if record.ExpiresAt != nil {
		resp.ExpiresAt = record.ExpiresAt.Unix()
	}

	return resp
}

// SocialMediaScopeToProto converts a social media scope to protobuf.
func SocialMediaScopeToProto(scope *SocialMediaScope) *SocialMediaScopePB {
	if scope == nil {
		return nil
	}

	resp := &SocialMediaScopePB{
		Version:                scope.Version,
		ScopeId:                scope.ScopeID,
		AccountAddress:         scope.AccountAddress,
		Provider:               SocialMediaProviderToProto(scope.Provider),
		ProfileNameHash:        scope.ProfileNameHash,
		EmailHash:              scope.EmailHash,
		UsernameHash:           scope.UsernameHash,
		OrgHash:                scope.OrgHash,
		AccountCreatedAt:       toUnixPointer(scope.AccountCreatedAt),
		AccountAgeDays:         scope.AccountAgeDays,
		IsVerified:             scope.IsVerified,
		FriendCountRange:       scope.FriendCountRange,
		Status:                 string(scope.Status),
		CreatedAt:              scope.CreatedAt.Unix(),
		UpdatedAt:              scope.UpdatedAt.Unix(),
		EncryptedPayload:       encryptedPayloadToProto(scope.EncryptedPayload),
		EvidenceHash:           scope.EvidenceHash,
		EvidenceStorageBackend: scope.EvidenceStorageBackend,
		EvidenceStorageRef:     scope.EvidenceStorageRef,
		EvidenceMetadata:       scope.EvidenceMetadata,
	}

	return resp
}

// PhoneNumberHashToProto converts a phone hash to protobuf.
func PhoneNumberHashToProto(hash PhoneNumberHash) *PhoneNumberHashPB {
	return &PhoneNumberHashPB{
		Hash:            hash.Hash,
		Salt:            hash.Salt,
		CountryCodeHash: hash.CountryCodeHash,
		CreatedAt:       toUnix(hash.CreatedAt),
	}
}

func toUnix(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}

func toUnixPointer(t *time.Time) int64 {
	if t == nil || t.IsZero() {
		return 0
	}
	return t.Unix()
}

func encryptedPayloadToProto(payload encryptiontypes.EncryptedPayloadEnvelope) EncryptedPayloadEnvelope {
	return EncryptedPayloadEnvelope{
		Version:             payload.Version,
		AlgorithmId:         payload.AlgorithmID,
		AlgorithmVersion:    payload.AlgorithmVersion,
		RecipientKeyIds:     payload.RecipientKeyIDs,
		RecipientPublicKeys: payload.RecipientPublicKeys,
		EncryptedKeys:       payload.EncryptedKeys,
		Nonce:               payload.Nonce,
		Ciphertext:          payload.Ciphertext,
		SenderSignature:     payload.SenderSignature,
		SenderPubKey:        payload.SenderPubKey,
		Metadata:            payload.Metadata,
	}
}

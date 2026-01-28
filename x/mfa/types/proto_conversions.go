// Package types provides conversion functions between local and proto types.
package types

import (
	mfav1 "github.com/virtengine/virtengine/sdk/go/node/mfa/v1"
)

// =============================================================================
// Msg Conversion Functions (Proto -> Local)
// =============================================================================

func convertMsgEnrollFactorFromProto(req *mfav1.MsgEnrollFactor) *MsgEnrollFactor {
	return &MsgEnrollFactor{
		Sender:                   req.Sender,
		FactorType:               FactorType(req.FactorType),
		Label:                    req.Label,
		PublicIdentifier:         req.PublicIdentifier,
		Metadata:                 convertFactorMetadataFromProto(req.Metadata),
		InitialVerificationProof: req.InitialVerificationProof,
	}
}

func convertMsgRevokeFactorFromProto(req *mfav1.MsgRevokeFactor) *MsgRevokeFactor {
	return &MsgRevokeFactor{
		Sender:     req.Sender,
		FactorType: FactorType(req.FactorType),
		FactorID:   req.FactorId,
		MFAProof:   convertMFAProofFromProto(req.MfaProof),
	}
}

func convertMsgSetMFAPolicyFromProto(req *mfav1.MsgSetMFAPolicy) *MsgSetMFAPolicy {
	return &MsgSetMFAPolicy{
		Sender:   req.Sender,
		Policy:   convertMFAPolicyFromProto(&req.Policy),
		MFAProof: convertMFAProofFromProto(req.MfaProof),
	}
}

func convertMsgCreateChallengeFromProto(req *mfav1.MsgCreateChallenge) *MsgCreateChallenge {
	return &MsgCreateChallenge{
		Sender:          req.Sender,
		FactorType:      FactorType(req.FactorType),
		FactorID:        req.FactorId,
		TransactionType: SensitiveTransactionType(req.TransactionType),
		ClientInfo:      convertClientInfoFromProto(req.ClientInfo),
	}
}

func convertMsgVerifyChallengeFromProto(req *mfav1.MsgVerifyChallenge) *MsgVerifyChallenge {
	return &MsgVerifyChallenge{
		Sender:      req.Sender,
		ChallengeID: req.ChallengeId,
		Response:    convertChallengeResponseFromProto(&req.Response),
	}
}

func convertMsgAddTrustedDeviceFromProto(req *mfav1.MsgAddTrustedDevice) *MsgAddTrustedDevice {
	return &MsgAddTrustedDevice{
		Sender:     req.Sender,
		DeviceInfo: convertDeviceInfoFromProto(&req.DeviceInfo),
		MFAProof:   convertMFAProofFromProtoDirect(&req.MfaProof),
	}
}

func convertMsgRemoveTrustedDeviceFromProto(req *mfav1.MsgRemoveTrustedDevice) *MsgRemoveTrustedDevice {
	return &MsgRemoveTrustedDevice{
		Sender:            req.Sender,
		DeviceFingerprint: req.DeviceFingerprint,
		MFAProof:          convertMFAProofFromProto(req.MfaProof),
	}
}

func convertMsgUpdateSensitiveTxConfigFromProto(req *mfav1.MsgUpdateSensitiveTxConfig) *MsgUpdateSensitiveTxConfig {
	return &MsgUpdateSensitiveTxConfig{
		Authority: req.Authority,
		Config:    convertSensitiveTxConfigFromProto(&req.Config),
	}
}

// =============================================================================
// Msg Conversion Functions (Local -> Proto)
// =============================================================================

func convertMsgEnrollFactorResponseToProto(resp *MsgEnrollFactorResponse) *mfav1.MsgEnrollFactorResponse {
	return &mfav1.MsgEnrollFactorResponse{
		FactorId: resp.FactorID,
		Status:   mfav1.FactorEnrollmentStatus(resp.Status),
	}
}

func convertMsgRevokeFactorResponseToProto(resp *MsgRevokeFactorResponse) *mfav1.MsgRevokeFactorResponse {
	return &mfav1.MsgRevokeFactorResponse{
		Success: resp.Success,
	}
}

func convertMsgSetMFAPolicyResponseToProto(resp *MsgSetMFAPolicyResponse) *mfav1.MsgSetMFAPolicyResponse {
	return &mfav1.MsgSetMFAPolicyResponse{
		Success: resp.Success,
	}
}

func convertMsgCreateChallengeResponseToProto(resp *MsgCreateChallengeResponse) *mfav1.MsgCreateChallengeResponse {
	return &mfav1.MsgCreateChallengeResponse{
		ChallengeId:   resp.ChallengeID,
		ChallengeData: resp.ChallengeData,
		ExpiresAt:     resp.ExpiresAt,
	}
}

func convertMsgVerifyChallengeResponseToProto(resp *MsgVerifyChallengeResponse) *mfav1.MsgVerifyChallengeResponse {
	factors := make([]mfav1.FactorType, len(resp.RemainingFactors))
	for i, f := range resp.RemainingFactors {
		factors[i] = mfav1.FactorType(f)
	}
	return &mfav1.MsgVerifyChallengeResponse{
		Verified:         resp.Verified,
		SessionId:        resp.SessionID,
		SessionExpiresAt: resp.SessionExpiresAt,
		RemainingFactors: factors,
	}
}

func convertMsgAddTrustedDeviceResponseToProto(resp *MsgAddTrustedDeviceResponse) *mfav1.MsgAddTrustedDeviceResponse {
	return &mfav1.MsgAddTrustedDeviceResponse{
		Success:        resp.Success,
		TrustExpiresAt: resp.TrustExpiresAt,
	}
}

func convertMsgRemoveTrustedDeviceResponseToProto(resp *MsgRemoveTrustedDeviceResponse) *mfav1.MsgRemoveTrustedDeviceResponse {
	return &mfav1.MsgRemoveTrustedDeviceResponse{
		Success: resp.Success,
	}
}

func convertMsgUpdateSensitiveTxConfigResponseToProto(resp *MsgUpdateSensitiveTxConfigResponse) *mfav1.MsgUpdateSensitiveTxConfigResponse {
	return &mfav1.MsgUpdateSensitiveTxConfigResponse{
		Success: resp.Success,
	}
}

// =============================================================================
// Query Response Conversion Functions (Local -> Proto)
// =============================================================================

func convertQueryMFAPolicyResponseToProto(resp *QueryMFAPolicyResponse) *mfav1.QueryMFAPolicyResponse {
	if resp.Policy == nil {
		return &mfav1.QueryMFAPolicyResponse{Policy: nil}
	}
	policy := convertMFAPolicyToProto(resp.Policy)
	return &mfav1.QueryMFAPolicyResponse{Policy: policy}
}

func convertQueryFactorEnrollmentsResponseToProto(resp *QueryFactorEnrollmentsResponse) *mfav1.QueryFactorEnrollmentsResponse {
	enrollments := make([]mfav1.FactorEnrollment, len(resp.Enrollments))
	for i, e := range resp.Enrollments {
		enrollments[i] = *convertFactorEnrollmentToProto(&e)
	}
	return &mfav1.QueryFactorEnrollmentsResponse{Enrollments: enrollments}
}

func convertQueryFactorEnrollmentResponseToProto(resp *QueryFactorEnrollmentResponse) *mfav1.QueryFactorEnrollmentResponse {
	if resp.Enrollment == nil {
		return &mfav1.QueryFactorEnrollmentResponse{Enrollment: nil}
	}
	return &mfav1.QueryFactorEnrollmentResponse{Enrollment: convertFactorEnrollmentToProto(resp.Enrollment)}
}

func convertQueryChallengeResponseToProto(resp *QueryChallengeResponse) *mfav1.QueryChallengeResponse {
	if resp.Challenge == nil {
		return &mfav1.QueryChallengeResponse{Challenge: nil}
	}
	return &mfav1.QueryChallengeResponse{Challenge: convertChallengeToProto(resp.Challenge)}
}

func convertQueryPendingChallengesResponseToProto(resp *QueryPendingChallengesResponse) *mfav1.QueryPendingChallengesResponse {
	challenges := make([]mfav1.Challenge, len(resp.Challenges))
	for i, c := range resp.Challenges {
		challenges[i] = *convertChallengeToProto(&c)
	}
	return &mfav1.QueryPendingChallengesResponse{Challenges: challenges}
}

func convertQueryAuthorizationSessionResponseToProto(resp *QueryAuthorizationSessionResponse) *mfav1.QueryAuthorizationSessionResponse {
	if resp.Session == nil {
		return &mfav1.QueryAuthorizationSessionResponse{Session: nil}
	}
	return &mfav1.QueryAuthorizationSessionResponse{Session: convertAuthorizationSessionToProto(resp.Session)}
}

func convertQueryTrustedDevicesResponseToProto(resp *QueryTrustedDevicesResponse) *mfav1.QueryTrustedDevicesResponse {
	devices := make([]mfav1.TrustedDevice, len(resp.Devices))
	for i, d := range resp.Devices {
		devices[i] = *convertTrustedDeviceToProto(&d)
	}
	return &mfav1.QueryTrustedDevicesResponse{Devices: devices}
}

func convertQuerySensitiveTxConfigResponseToProto(resp *QuerySensitiveTxConfigResponse) *mfav1.QuerySensitiveTxConfigResponse {
	if resp.Config == nil {
		return &mfav1.QuerySensitiveTxConfigResponse{Config: nil}
	}
	return &mfav1.QuerySensitiveTxConfigResponse{Config: convertSensitiveTxConfigToProto(resp.Config)}
}

func convertQueryAllSensitiveTxConfigsResponseToProto(resp *QueryAllSensitiveTxConfigsResponse) *mfav1.QueryAllSensitiveTxConfigsResponse {
	configs := make([]mfav1.SensitiveTxConfig, len(resp.Configs))
	for i, c := range resp.Configs {
		configs[i] = *convertSensitiveTxConfigToProto(&c)
	}
	return &mfav1.QueryAllSensitiveTxConfigsResponse{Configs: configs}
}

func convertQueryParamsResponseToProto(resp *QueryParamsResponse) *mfav1.QueryParamsResponse {
	return &mfav1.QueryParamsResponse{Params: convertParamsToProto(&resp.Params)}
}

func convertQueryMFARequiredResponseToProto(resp *QueryMFARequiredResponse) *mfav1.QueryMFARequiredResponse {
	combinations := make([]mfav1.FactorCombination, len(resp.FactorCombinations))
	for i, c := range resp.FactorCombinations {
		combinations[i] = *convertFactorCombinationToProto(&c)
	}
	return &mfav1.QueryMFARequiredResponse{
		Required:           resp.Required,
		FactorCombinations: combinations,
		MinVeidScore:       resp.MinVEIDScore,
	}
}

// =============================================================================
// Helper Conversion Functions (Proto -> Local)
// =============================================================================

func convertMFAProofFromProto(proof *mfav1.MFAProof) *MFAProof {
	if proof == nil {
		return nil
	}
	factors := make([]FactorType, len(proof.VerifiedFactors))
	for i, f := range proof.VerifiedFactors {
		factors[i] = FactorType(f)
	}
	return &MFAProof{
		SessionID:       proof.SessionId,
		VerifiedFactors: factors,
		Timestamp:       proof.Timestamp,
		Signature:       proof.Signature,
	}
}

func convertMFAProofFromProtoDirect(proof *mfav1.MFAProof) *MFAProof {
	if proof == nil {
		return nil
	}
	factors := make([]FactorType, len(proof.VerifiedFactors))
	for i, f := range proof.VerifiedFactors {
		factors[i] = FactorType(f)
	}
	return &MFAProof{
		SessionID:       proof.SessionId,
		VerifiedFactors: factors,
		Timestamp:       proof.Timestamp,
		Signature:       proof.Signature,
	}
}

func convertMFAPolicyFromProto(policy *mfav1.MFAPolicy) MFAPolicy {
	if policy == nil {
		return MFAPolicy{}
	}
	// Map RequiredFactors from proto to local type
	combinations := make([]FactorCombination, len(policy.RequiredFactors))
	for i, c := range policy.RequiredFactors {
		combinations[i] = convertFactorCombinationFromProto(&c)
	}
	// Map RecoveryFactors from proto
	recovery := make([]FactorCombination, len(policy.RecoveryFactors))
	for i, c := range policy.RecoveryFactors {
		recovery[i] = convertFactorCombinationFromProto(&c)
	}
	// Map KeyRotationFactors from proto
	keyRotation := make([]FactorCombination, len(policy.KeyRotationFactors))
	for i, c := range policy.KeyRotationFactors {
		keyRotation[i] = convertFactorCombinationFromProto(&c)
	}
	return MFAPolicy{
		AccountAddress:     policy.AccountAddress,
		Enabled:            policy.Enabled,
		RequiredFactors:    combinations,
		TrustedDeviceRule:  convertTrustedDevicePolicyFromProto(policy.TrustedDeviceRule),
		RecoveryFactors:    recovery,
		KeyRotationFactors: keyRotation,
		SessionDuration:    policy.SessionDuration,
		VEIDThreshold:      policy.VeidThreshold,
		CreatedAt:          policy.CreatedAt,
		UpdatedAt:          policy.UpdatedAt,
	}
}

func convertFactorCombinationFromProto(c *mfav1.FactorCombination) FactorCombination {
	factors := make([]FactorType, len(c.Factors))
	for i, f := range c.Factors {
		factors[i] = FactorType(f)
	}
	return FactorCombination{
		Factors:          factors,
		MinSecurityLevel: FactorSecurityLevel(c.MinSecurityLevel),
	}
}

func convertTrustedDevicePolicyFromProto(p *mfav1.TrustedDevicePolicy) *TrustedDevicePolicy {
	if p == nil {
		return nil
	}
	var reduced *FactorCombination
	if p.ReducedFactors != nil {
		r := convertFactorCombinationFromProto(p.ReducedFactors)
		reduced = &r
	}
	return &TrustedDevicePolicy{
		Enabled:                   p.Enabled,
		TrustDuration:             p.TrustDuration,
		ReducedFactors:            reduced,
		MaxTrustedDevices:         p.MaxTrustedDevices,
		RequireReauthForSensitive: p.RequireReauthForSensitive,
	}
}

func convertFactorMetadataFromProto(m *mfav1.FactorMetadata) *FactorMetadata {
	if m == nil {
		return nil
	}
	result := &FactorMetadata{
		VEIDThreshold: m.VeidThreshold,
		ContactHash:   m.ContactHash,
	}
	if m.DeviceInfo != nil {
		result.DeviceInfo = convertDeviceInfoFromProtoPtr(m.DeviceInfo)
	}
	if m.Fido2Info != nil {
		result.FIDO2Info = convertFIDO2InfoFromProto(m.Fido2Info)
	}
	if m.HardwareKeyInfo != nil {
		result.HardwareKeyInfo = convertHardwareKeyInfoFromProto(m.HardwareKeyInfo)
	}
	return result
}

func convertDeviceInfoFromProtoPtr(d *mfav1.DeviceInfo) *DeviceInfo {
	if d == nil {
		return nil
	}
	return &DeviceInfo{
		Fingerprint:    d.Fingerprint,
		UserAgent:      d.UserAgent,
		FirstSeenAt:    d.FirstSeenAt,
		LastSeenAt:     d.LastSeenAt,
		IPHash:         d.IpHash,
		TrustExpiresAt: d.TrustExpiresAt,
	}
}

func convertFIDO2InfoFromProto(f *mfav1.FIDO2CredentialInfo) *FIDO2CredentialInfo {
	if f == nil {
		return nil
	}
	return &FIDO2CredentialInfo{
		CredentialID:    f.CredentialId,
		PublicKey:       f.PublicKey,
		AAGUID:          f.Aaguid,
		SignCount:       f.SignCount,
		AttestationType: f.AttestationType,
	}
}

func convertHardwareKeyInfoFromProto(h *mfav1.HardwareKeyEnrollment) *HardwareKeyEnrollment {
	if h == nil {
		return nil
	}
	return &HardwareKeyEnrollment{
		KeyType:                HardwareKeyType(h.KeyType),
		KeyID:                  h.KeyId,
		SubjectDN:              h.SubjectDn,
		IssuerDN:               h.IssuerDn,
		SerialNumber:           h.SerialNumber,
		PublicKeyFingerprint:   h.PublicKeyFingerprint,
		NotBefore:              h.NotBefore,
		NotAfter:               h.NotAfter,
		KeyUsage:               h.KeyUsage,
		ExtendedKeyUsage:       h.ExtendedKeyUsage,
		RevocationCheckEnabled: h.RevocationCheckEnabled,
		LastRevocationCheck:    h.LastRevocationCheck,
		RevocationStatus:       RevocationStatus(h.RevocationStatus),
	}
}

func convertClientInfoFromProto(c *mfav1.ClientInfo) *ClientInfo {
	if c == nil {
		return nil
	}
	return &ClientInfo{
		DeviceFingerprint: c.DeviceFingerprint,
		IPHash:            c.IpHash,
		UserAgent:         c.UserAgent,
		RequestedAt:       c.RequestedAt,
	}
}

func convertChallengeResponseFromProto(r *mfav1.ChallengeResponse) *ChallengeResponse {
	if r == nil {
		return nil
	}
	return &ChallengeResponse{
		ChallengeID:  r.ChallengeId,
		FactorType:   FactorType(r.FactorType),
		ResponseData: r.ResponseData,
		ClientInfo:   convertClientInfoFromProto(r.ClientInfo),
		Timestamp:    r.Timestamp,
	}
}

func convertDeviceInfoFromProto(d *mfav1.DeviceInfo) DeviceInfo {
	return DeviceInfo{
		Fingerprint:    d.Fingerprint,
		UserAgent:      d.UserAgent,
		FirstSeenAt:    d.FirstSeenAt,
		LastSeenAt:     d.LastSeenAt,
		IPHash:         d.IpHash,
		TrustExpiresAt: d.TrustExpiresAt,
	}
}

func convertSensitiveTxConfigFromProto(c *mfav1.SensitiveTxConfig) SensitiveTxConfig {
	combinations := make([]FactorCombination, len(c.RequiredFactorCombinations))
	for i, fc := range c.RequiredFactorCombinations {
		combinations[i] = convertFactorCombinationFromProto(&fc)
	}
	return SensitiveTxConfig{
		TransactionType:             SensitiveTransactionType(c.TransactionType),
		Enabled:                     c.Enabled,
		MinVEIDScore:                c.MinVeidScore,
		RequiredFactorCombinations:  combinations,
		SessionDuration:             c.SessionDuration,
		IsSingleUse:                 c.IsSingleUse,
		AllowTrustedDeviceReduction: c.AllowTrustedDeviceReduction,
		ValueThreshold:              c.ValueThreshold,
		CooldownPeriod:              c.CooldownPeriod,
		Description:                 c.Description,
	}
}

// =============================================================================
// Helper Conversion Functions (Local -> Proto)
// =============================================================================

func convertMFAPolicyToProto(policy *MFAPolicy) *mfav1.MFAPolicy {
	if policy == nil {
		return nil
	}
	// Map RequiredFactors from local to proto
	combinations := make([]mfav1.FactorCombination, len(policy.RequiredFactors))
	for i, c := range policy.RequiredFactors {
		combinations[i] = *convertFactorCombinationToProto(&c)
	}
	// Map RecoveryFactors
	recovery := make([]mfav1.FactorCombination, len(policy.RecoveryFactors))
	for i, c := range policy.RecoveryFactors {
		recovery[i] = *convertFactorCombinationToProto(&c)
	}
	// Map KeyRotationFactors
	keyRotation := make([]mfav1.FactorCombination, len(policy.KeyRotationFactors))
	for i, c := range policy.KeyRotationFactors {
		keyRotation[i] = *convertFactorCombinationToProto(&c)
	}
	return &mfav1.MFAPolicy{
		AccountAddress:     policy.AccountAddress,
		Enabled:            policy.Enabled,
		RequiredFactors:    combinations,
		TrustedDeviceRule:  convertTrustedDevicePolicyToProto(policy.TrustedDeviceRule),
		RecoveryFactors:    recovery,
		KeyRotationFactors: keyRotation,
		SessionDuration:    policy.SessionDuration,
		VeidThreshold:      policy.VEIDThreshold,
		CreatedAt:          policy.CreatedAt,
		UpdatedAt:          policy.UpdatedAt,
	}
}

func convertFactorCombinationToProto(c *FactorCombination) *mfav1.FactorCombination {
	factors := make([]mfav1.FactorType, len(c.Factors))
	for i, f := range c.Factors {
		factors[i] = mfav1.FactorType(f)
	}
	return &mfav1.FactorCombination{
		Factors:          factors,
		MinSecurityLevel: mfav1.FactorSecurityLevel(c.MinSecurityLevel),
	}
}

func convertTrustedDevicePolicyToProto(p *TrustedDevicePolicy) *mfav1.TrustedDevicePolicy {
	if p == nil {
		return nil
	}
	var reduced *mfav1.FactorCombination
	if p.ReducedFactors != nil {
		reduced = convertFactorCombinationToProto(p.ReducedFactors)
	}
	return &mfav1.TrustedDevicePolicy{
		Enabled:                   p.Enabled,
		TrustDuration:             p.TrustDuration,
		ReducedFactors:            reduced,
		MaxTrustedDevices:         p.MaxTrustedDevices,
		RequireReauthForSensitive: p.RequireReauthForSensitive,
	}
}

func convertFactorEnrollmentToProto(e *FactorEnrollment) *mfav1.FactorEnrollment {
	return &mfav1.FactorEnrollment{
		AccountAddress:   e.AccountAddress,
		FactorType:       mfav1.FactorType(e.FactorType),
		FactorId:         e.FactorID,
		PublicIdentifier: e.PublicIdentifier,
		Label:            e.Label,
		Status:           mfav1.FactorEnrollmentStatus(e.Status),
		EnrolledAt:       e.EnrolledAt,
		VerifiedAt:       e.VerifiedAt,
		LastUsedAt:       e.LastUsedAt,
		UseCount:         e.UseCount,
		Metadata:         convertFactorMetadataToProto(e.Metadata),
	}
}

func convertFactorMetadataToProto(m *FactorMetadata) *mfav1.FactorMetadata {
	if m == nil {
		return nil
	}
	result := &mfav1.FactorMetadata{
		VeidThreshold: m.VEIDThreshold,
		ContactHash:   m.ContactHash,
	}
	if m.DeviceInfo != nil {
		result.DeviceInfo = convertDeviceInfoToProtoFromLocal(m.DeviceInfo)
	}
	if m.FIDO2Info != nil {
		result.Fido2Info = convertFIDO2InfoToProto(m.FIDO2Info)
	}
	if m.HardwareKeyInfo != nil {
		result.HardwareKeyInfo = convertHardwareKeyInfoToProto(m.HardwareKeyInfo)
	}
	return result
}

func convertDeviceInfoToProtoFromLocal(d *DeviceInfo) *mfav1.DeviceInfo {
	if d == nil {
		return nil
	}
	return &mfav1.DeviceInfo{
		Fingerprint:    d.Fingerprint,
		UserAgent:      d.UserAgent,
		FirstSeenAt:    d.FirstSeenAt,
		LastSeenAt:     d.LastSeenAt,
		IpHash:         d.IPHash,
		TrustExpiresAt: d.TrustExpiresAt,
	}
}

func convertFIDO2InfoToProto(f *FIDO2CredentialInfo) *mfav1.FIDO2CredentialInfo {
	if f == nil {
		return nil
	}
	return &mfav1.FIDO2CredentialInfo{
		CredentialId:    f.CredentialID,
		PublicKey:       f.PublicKey,
		Aaguid:          f.AAGUID,
		SignCount:       f.SignCount,
		AttestationType: f.AttestationType,
	}
}

func convertHardwareKeyInfoToProto(h *HardwareKeyEnrollment) *mfav1.HardwareKeyEnrollment {
	if h == nil {
		return nil
	}
	return &mfav1.HardwareKeyEnrollment{
		KeyType:                mfav1.HardwareKeyType(h.KeyType),
		KeyId:                  h.KeyID,
		SubjectDn:              h.SubjectDN,
		IssuerDn:               h.IssuerDN,
		SerialNumber:           h.SerialNumber,
		PublicKeyFingerprint:   h.PublicKeyFingerprint,
		NotBefore:              h.NotBefore,
		NotAfter:               h.NotAfter,
		KeyUsage:               h.KeyUsage,
		ExtendedKeyUsage:       h.ExtendedKeyUsage,
		RevocationCheckEnabled: h.RevocationCheckEnabled,
		LastRevocationCheck:    h.LastRevocationCheck,
		RevocationStatus:       mfav1.RevocationStatus(h.RevocationStatus),
	}
}

func convertChallengeToProto(c *Challenge) *mfav1.Challenge {
	return &mfav1.Challenge{
		ChallengeId:     c.ChallengeID,
		AccountAddress:  c.AccountAddress,
		FactorType:      mfav1.FactorType(c.FactorType),
		FactorId:        c.FactorID,
		TransactionType: mfav1.SensitiveTransactionType(c.TransactionType),
		Status:          mfav1.ChallengeStatus(c.Status),
		ChallengeData:   c.ChallengeData,
		CreatedAt:       c.CreatedAt,
		ExpiresAt:       c.ExpiresAt,
		AttemptCount:    c.AttemptCount,
		MaxAttempts:     c.MaxAttempts,
		Metadata:        convertChallengeMetadataToProto(c.Metadata),
	}
}

// convertChallengeMetadataToProto converts challenge metadata to proto
func convertChallengeMetadataToProto(m *ChallengeMetadata) *mfav1.ChallengeMetadata {
	if m == nil {
		return nil
	}
	return &mfav1.ChallengeMetadata{
		ClientInfo: convertClientInfoToProto(m.ClientInfo),
	}
}

func convertClientInfoToProto(c *ClientInfo) *mfav1.ClientInfo {
	if c == nil {
		return nil
	}
	return &mfav1.ClientInfo{
		DeviceFingerprint: c.DeviceFingerprint,
		IpHash:            c.IPHash,
		UserAgent:         c.UserAgent,
		RequestedAt:       c.RequestedAt,
	}
}

func convertAuthorizationSessionToProto(s *AuthorizationSession) *mfav1.AuthorizationSession {
	factors := make([]mfav1.FactorType, len(s.VerifiedFactors))
	for i, f := range s.VerifiedFactors {
		factors[i] = mfav1.FactorType(f)
	}
	return &mfav1.AuthorizationSession{
		SessionId:         s.SessionID,
		AccountAddress:    s.AccountAddress,
		TransactionType:   mfav1.SensitiveTransactionType(s.TransactionType),
		VerifiedFactors:   factors,
		CreatedAt:         s.CreatedAt,
		ExpiresAt:         s.ExpiresAt,
		UsedAt:            s.UsedAt,
		IsSingleUse:       s.IsSingleUse,
		DeviceFingerprint: s.DeviceFingerprint,
	}
}

// Remove unused function


func convertTrustedDeviceToProto(d *TrustedDevice) *mfav1.TrustedDevice {
	return &mfav1.TrustedDevice{
		AccountAddress: d.AccountAddress,
		DeviceInfo:     *convertDeviceInfoToProto(&d.DeviceInfo),
		AddedAt:        d.AddedAt,
		LastUsedAt:     d.LastUsedAt,
	}
}

func convertDeviceInfoToProto(d *DeviceInfo) *mfav1.DeviceInfo {
	return &mfav1.DeviceInfo{
		Fingerprint:    d.Fingerprint,
		UserAgent:      d.UserAgent,
		FirstSeenAt:    d.FirstSeenAt,
		LastSeenAt:     d.LastSeenAt,
		IpHash:         d.IPHash,
		TrustExpiresAt: d.TrustExpiresAt,
	}
}

func convertSensitiveTxConfigToProto(c *SensitiveTxConfig) *mfav1.SensitiveTxConfig {
	combinations := make([]mfav1.FactorCombination, len(c.RequiredFactorCombinations))
	for i, fc := range c.RequiredFactorCombinations {
		combinations[i] = *convertFactorCombinationToProto(&fc)
	}
	return &mfav1.SensitiveTxConfig{
		TransactionType:             mfav1.SensitiveTransactionType(c.TransactionType),
		Enabled:                     c.Enabled,
		MinVeidScore:                c.MinVEIDScore,
		RequiredFactorCombinations:  combinations,
		SessionDuration:             c.SessionDuration,
		IsSingleUse:                 c.IsSingleUse,
		AllowTrustedDeviceReduction: c.AllowTrustedDeviceReduction,
		ValueThreshold:              c.ValueThreshold,
		CooldownPeriod:              c.CooldownPeriod,
		Description:                 c.Description,
	}
}

func convertParamsToProto(p *Params) mfav1.Params {
	factors := make([]mfav1.FactorType, len(p.AllowedFactorTypes))
	for i, f := range p.AllowedFactorTypes {
		factors[i] = mfav1.FactorType(f)
	}
	return mfav1.Params{
		DefaultSessionDuration:  p.DefaultSessionDuration,
		MaxFactorsPerAccount:    p.MaxFactorsPerAccount,
		MaxChallengeAttempts:    p.MaxChallengeAttempts,
		ChallengeTtl:            p.ChallengeTTL,
		MaxTrustedDevices:       p.MaxTrustedDevices,
		TrustedDeviceTtl:        p.TrustedDeviceTTL,
		MinVeidScoreForMfa:      p.MinVEIDScoreForMFA,
		RequireAtLeastOneFactor: p.RequireAtLeastOneFactor,
		AllowedFactorTypes:      factors,
	}
}

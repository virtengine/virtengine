package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/support/types"
)

const testInvalidValue = "bogus"

// --- SupportCategory ---

func TestSupportCategory_IsValid(t *testing.T) {
	valid := []types.SupportCategory{
		types.SupportCategoryAccount,
		types.SupportCategoryIdentity,
		types.SupportCategoryBilling,
		types.SupportCategoryProvider,
		types.SupportCategoryMarketplace,
		types.SupportCategoryTechnical,
		types.SupportCategorySecurity,
		types.SupportCategoryOther,
	}
	for _, c := range valid {
		assert.True(t, c.IsValid(), "expected valid: %s", c)
	}
	assert.False(t, types.SupportCategory("unknown").IsValid())
	assert.False(t, types.SupportCategory("").IsValid())
}

// --- SupportPriority ---

func TestSupportPriority_IsValid(t *testing.T) {
	valid := []types.SupportPriority{
		types.SupportPriorityLow,
		types.SupportPriorityNormal,
		types.SupportPriorityHigh,
		types.SupportPriorityUrgent,
	}
	for _, p := range valid {
		assert.True(t, p.IsValid(), "expected valid: %s", p)
	}
	assert.False(t, types.SupportPriority("critical").IsValid())
	assert.False(t, types.SupportPriority("").IsValid())
}

// --- SupportStatus ---

func TestSupportStatus_IsValid(t *testing.T) {
	assert.False(t, types.SupportStatusUnspecified.IsValid())
	assert.True(t, types.SupportStatusOpen.IsValid())
	assert.True(t, types.SupportStatusAssigned.IsValid())
	assert.True(t, types.SupportStatusInProgress.IsValid())
	assert.True(t, types.SupportStatusWaitingCustomer.IsValid())
	assert.True(t, types.SupportStatusWaitingSupport.IsValid())
	assert.True(t, types.SupportStatusResolved.IsValid())
	assert.True(t, types.SupportStatusClosed.IsValid())
	assert.True(t, types.SupportStatusArchived.IsValid())
	assert.False(t, types.SupportStatus(99).IsValid())
}

func TestSupportStatus_String(t *testing.T) {
	assert.Equal(t, "open", types.SupportStatusOpen.String())
	assert.Equal(t, "closed", types.SupportStatusClosed.String())
	assert.Equal(t, "archived", types.SupportStatusArchived.String())
	assert.Contains(t, types.SupportStatus(99).String(), "unknown")
}

func TestSupportStatus_IsTerminal(t *testing.T) {
	assert.False(t, types.SupportStatusOpen.IsTerminal())
	assert.False(t, types.SupportStatusInProgress.IsTerminal())
	assert.True(t, types.SupportStatusClosed.IsTerminal())
	assert.True(t, types.SupportStatusArchived.IsTerminal())
}

func TestSupportStatusFromString(t *testing.T) {
	assert.Equal(t, types.SupportStatusOpen, types.SupportStatusFromString("open"))
	assert.Equal(t, types.SupportStatusClosed, types.SupportStatusFromString("closed"))
	assert.Equal(t, types.SupportStatusUnspecified, types.SupportStatusFromString(testInvalidValue))
	assert.Equal(t, types.SupportStatusUnspecified, types.SupportStatusFromString(""))
}

func TestSupportStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from    types.SupportStatus
		to      types.SupportStatus
		allowed bool
	}{
		{types.SupportStatusOpen, types.SupportStatusAssigned, true},
		{types.SupportStatusOpen, types.SupportStatusInProgress, true},
		{types.SupportStatusOpen, types.SupportStatusClosed, true},
		{types.SupportStatusOpen, types.SupportStatusArchived, false},
		{types.SupportStatusAssigned, types.SupportStatusInProgress, true},
		{types.SupportStatusAssigned, types.SupportStatusResolved, true},
		{types.SupportStatusInProgress, types.SupportStatusWaitingCustomer, true},
		{types.SupportStatusInProgress, types.SupportStatusResolved, true},
		{types.SupportStatusWaitingCustomer, types.SupportStatusInProgress, true},
		{types.SupportStatusWaitingSupport, types.SupportStatusInProgress, true},
		{types.SupportStatusResolved, types.SupportStatusClosed, true},
		{types.SupportStatusResolved, types.SupportStatusInProgress, true},
		{types.SupportStatusClosed, types.SupportStatusArchived, true},
		{types.SupportStatusClosed, types.SupportStatusOpen, false},
		{types.SupportStatusArchived, types.SupportStatusOpen, false},
		{types.SupportStatusArchived, types.SupportStatusClosed, false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.allowed, tt.from.CanTransitionTo(tt.to),
			"%s -> %s", tt.from, tt.to)
	}
}

// --- SupportRequestID ---

func TestSupportRequestID_Validate(t *testing.T) {
	valid := types.SupportRequestID{SubmitterAddress: "addr1", Sequence: 1}
	require.NoError(t, valid.Validate())

	require.Error(t, types.SupportRequestID{SubmitterAddress: "", Sequence: 1}.Validate())
	require.Error(t, types.SupportRequestID{SubmitterAddress: "addr1", Sequence: 0}.Validate())
}

func TestSupportRequestID_String(t *testing.T) {
	id := types.SupportRequestID{SubmitterAddress: "addr1", Sequence: 42}
	assert.Equal(t, "addr1/support/42", id.String())
}

func TestParseSupportRequestID(t *testing.T) {
	id, err := types.ParseSupportRequestID("addr1/support/42")
	require.NoError(t, err)
	assert.Equal(t, "addr1", id.SubmitterAddress)
	assert.Equal(t, uint64(42), id.Sequence)

	_, err = types.ParseSupportRequestID("bad")
	require.Error(t, err)

	_, err = types.ParseSupportRequestID("addr1/wrong/42")
	require.Error(t, err)

	_, err = types.ParseSupportRequestID("addr1/support/abc")
	require.Error(t, err)
}

// --- SupportResponseID ---

func TestSupportResponseID_Validate(t *testing.T) {
	valid := types.SupportResponseID{
		RequestID: types.SupportRequestID{SubmitterAddress: "addr1", Sequence: 1},
		Sequence:  1,
	}
	require.NoError(t, valid.Validate())

	invalid := types.SupportResponseID{
		RequestID: types.SupportRequestID{SubmitterAddress: "", Sequence: 1},
		Sequence:  1,
	}
	require.Error(t, invalid.Validate())

	zeroSeq := types.SupportResponseID{
		RequestID: types.SupportRequestID{SubmitterAddress: "addr1", Sequence: 1},
		Sequence:  0,
	}
	require.Error(t, zeroSeq.Validate())
}

func TestParseSupportResponseID(t *testing.T) {
	id, err := types.ParseSupportResponseID("addr1/support/1/response/5")
	require.NoError(t, err)
	assert.Equal(t, "addr1", id.RequestID.SubmitterAddress)
	assert.Equal(t, uint64(1), id.RequestID.Sequence)
	assert.Equal(t, uint64(5), id.Sequence)

	_, err = types.ParseSupportResponseID("too/short")
	require.Error(t, err)

	_, err = types.ParseSupportResponseID("addr1/support/1/notresponse/5")
	require.Error(t, err)
}

// --- ExternalSystem ---

func TestExternalSystem_IsValid(t *testing.T) {
	assert.True(t, types.ExternalSystemWaldur.IsValid())
	assert.True(t, types.ExternalSystemJira.IsValid())
	assert.False(t, types.ExternalSystem("slack").IsValid())
}

// --- ResourceType ---

func TestResourceType_IsValid(t *testing.T) {
	valid := []types.ResourceType{
		types.ResourceTypeDeployment,
		types.ResourceTypeLease,
		types.ResourceTypeOrder,
		types.ResourceTypeProvider,
		types.ResourceTypeSupportRequest,
	}
	for _, rt := range valid {
		assert.True(t, rt.IsValid(), "expected valid: %s", rt)
	}
	assert.False(t, types.ResourceType("invalid").IsValid())
}

// --- ExternalTicketRef ---

func TestExternalTicketRef_Validate(t *testing.T) {
	valid := types.ExternalTicketRef{
		ResourceID:       "res-1",
		ResourceType:     types.ResourceTypeDeployment,
		ExternalSystem:   types.ExternalSystemJira,
		ExternalTicketID: "JIRA-100",
		CreatedBy:        "creator",
	}
	require.NoError(t, valid.Validate())

	missing := valid
	missing.ResourceID = ""
	require.Error(t, missing.Validate())

	badType := valid
	badType.ResourceType = testInvalidValue
	require.Error(t, badType.Validate())

	badSys := valid
	badSys.ExternalSystem = "slack"
	require.Error(t, badSys.Validate())

	noTicket := valid
	noTicket.ExternalTicketID = ""
	require.Error(t, noTicket.Validate())

	noCreator := valid
	noCreator.CreatedBy = ""
	require.Error(t, noCreator.Validate())
}

func TestExternalTicketRef_Key(t *testing.T) {
	ref := types.ExternalTicketRef{
		ResourceID:   "res-1",
		ResourceType: types.ResourceTypeDeployment,
	}
	assert.Equal(t, "deployment/res-1", ref.Key())
}

// --- RetentionPolicy ---

func TestRetentionPolicy_Validate(t *testing.T) {
	require.NoError(t, (*types.RetentionPolicy)(nil).Validate(), "nil policy should be valid")

	valid := &types.RetentionPolicy{
		Version:             1,
		ArchiveAfterSeconds: 100,
		PurgeAfterSeconds:   200,
		CreatedAt:           time.Now(),
	}
	require.NoError(t, valid.Validate())

	badVersion := &types.RetentionPolicy{Version: 0}
	require.Error(t, badVersion.Validate())

	badVersion2 := &types.RetentionPolicy{Version: 999}
	require.Error(t, badVersion2.Validate())

	negArchive := &types.RetentionPolicy{Version: 1, ArchiveAfterSeconds: -1}
	require.Error(t, negArchive.Validate())

	negPurge := &types.RetentionPolicy{Version: 1, PurgeAfterSeconds: -1}
	require.Error(t, negPurge.Validate())

	purgeBeforeArchive := &types.RetentionPolicy{
		Version:             1,
		ArchiveAfterSeconds: 200,
		PurgeAfterSeconds:   100,
	}
	require.Error(t, purgeBeforeArchive.Validate())
}

func TestRetentionPolicy_ShouldArchiveAndPurge(t *testing.T) {
	now := time.Now().UTC()
	p := &types.RetentionPolicy{
		Version:             1,
		ArchiveAfterSeconds: 3600,
		PurgeAfterSeconds:   7200,
		CreatedAt:           now,
	}

	assert.False(t, p.ShouldArchive(now))
	assert.True(t, p.ShouldArchive(now.Add(2*time.Hour)))
	assert.False(t, p.ShouldPurge(now))
	assert.True(t, p.ShouldPurge(now.Add(3*time.Hour)))

	assert.False(t, (*types.RetentionPolicy)(nil).ShouldArchive(now))
	assert.False(t, (*types.RetentionPolicy)(nil).ShouldPurge(now))
}

func TestRetentionPolicy_ArchiveAtPurgeAt(t *testing.T) {
	now := time.Now().UTC()
	p := &types.RetentionPolicy{
		Version:             1,
		ArchiveAfterSeconds: 60,
		PurgeAfterSeconds:   120,
		CreatedAt:           now,
	}

	archiveAt, ok := p.ArchiveAt()
	assert.True(t, ok)
	assert.Equal(t, now.Add(60*time.Second).UTC(), archiveAt)

	purgeAt, ok := p.PurgeAt()
	assert.True(t, ok)
	assert.Equal(t, now.Add(120*time.Second).UTC(), purgeAt)

	_, ok = (*types.RetentionPolicy)(nil).ArchiveAt()
	assert.False(t, ok)
}

func TestRetentionPolicy_CopyWithTimestamps(t *testing.T) {
	now := time.Now().UTC()
	p := &types.RetentionPolicy{
		ArchiveAfterSeconds: 100,
	}
	clone := p.CopyWithTimestamps(now, 42)
	assert.Equal(t, types.RetentionPolicyVersion, clone.Version)
	assert.Equal(t, now.UTC(), clone.CreatedAt)
	assert.Equal(t, int64(42), clone.CreatedAtBlock)

	assert.Nil(t, (*types.RetentionPolicy)(nil).CopyWithTimestamps(now, 1))
}

func TestDefaultRetentionPolicy(t *testing.T) {
	now := time.Now().UTC()
	p := types.DefaultRetentionPolicy(now, 100)
	require.NotNil(t, p)
	assert.Equal(t, types.RetentionPolicyVersion, p.Version)
	assert.Greater(t, p.ArchiveAfterSeconds, int64(0))
	assert.Greater(t, p.PurgeAfterSeconds, p.ArchiveAfterSeconds)
	assert.Equal(t, now.UTC(), p.CreatedAt)
	assert.Equal(t, int64(100), p.CreatedAtBlock)
	require.NoError(t, p.Validate())
}

// --- SupportRequest ---

func validPayload() types.EncryptedSupportPayload {
	alg := encryptiontypes.DefaultAlgorithm()
	info, _ := encryptiontypes.GetAlgorithmInfo(alg)
	return types.EncryptedSupportPayload{
		Envelope: &encryptiontypes.EncryptedPayloadEnvelope{
			Version:          encryptiontypes.EnvelopeVersion,
			AlgorithmID:      alg,
			AlgorithmVersion: info.Version,
			RecipientKeyIDs:  []string{"key-1"},
			Nonce:            make([]byte, info.NonceSize),
			Ciphertext:       []byte{0x01, 0x02},
			SenderSignature:  []byte{0x01},
			SenderPubKey:     make([]byte, info.KeySize),
		},
	}
}

func TestNewSupportRequest(t *testing.T) {
	now := time.Now()
	id := types.SupportRequestID{SubmitterAddress: "addr1", Sequence: 1}
	req := types.NewSupportRequest(id, "SUP-000001", "addr1",
		types.SupportCategoryTechnical, types.SupportPriorityHigh,
		validPayload(), now)

	assert.Equal(t, types.SupportStatusOpen, req.Status)
	assert.Equal(t, "SUP-000001", req.TicketNumber)
	assert.NotNil(t, req.PublicMetadata)
}

func TestSupportRequest_Validate(t *testing.T) {
	now := time.Now()
	id := types.SupportRequestID{SubmitterAddress: "addr1", Sequence: 1}
	req := types.NewSupportRequest(id, "SUP-000001", "addr1",
		types.SupportCategoryTechnical, types.SupportPriorityHigh,
		validPayload(), now)
	require.NoError(t, req.Validate())

	// nil request
	require.Error(t, (*types.SupportRequest)(nil).Validate())

	// address mismatch
	bad := *req
	bad.SubmitterAddress = "different"
	require.Error(t, bad.Validate())

	// invalid category
	badCat := *req
	badCat.Category = testInvalidValue
	require.Error(t, badCat.Validate())

	// invalid priority
	badPri := *req
	badPri.Priority = testInvalidValue
	require.Error(t, badPri.Validate())
}

func TestSupportRequest_SetStatus(t *testing.T) {
	now := time.Now()
	id := types.SupportRequestID{SubmitterAddress: "addr1", Sequence: 1}
	req := types.NewSupportRequest(id, "SUP-000001", "addr1",
		types.SupportCategoryTechnical, types.SupportPriorityNormal,
		validPayload(), now)

	require.NoError(t, req.SetStatus(types.SupportStatusAssigned, now))
	assert.Equal(t, types.SupportStatusAssigned, req.Status)

	require.NoError(t, req.SetStatus(types.SupportStatusInProgress, now))
	require.NoError(t, req.SetStatus(types.SupportStatusResolved, now))
	assert.NotNil(t, req.ResolvedAt)

	require.NoError(t, req.SetStatus(types.SupportStatusClosed, now))
	assert.NotNil(t, req.ClosedAt)

	require.NoError(t, req.SetStatus(types.SupportStatusArchived, now))
	assert.True(t, req.Archived)
	assert.NotNil(t, req.ArchivedAt)

	// invalid transition from archived
	require.Error(t, req.SetStatus(types.SupportStatusOpen, now))
}

func TestSupportRequest_MarkArchivedAndPurged(t *testing.T) {
	now := time.Now()
	id := types.SupportRequestID{SubmitterAddress: "addr1", Sequence: 1}
	req := types.NewSupportRequest(id, "SUP-000001", "addr1",
		types.SupportCategoryTechnical, types.SupportPriorityNormal,
		validPayload(), now)

	req.MarkArchived("retention expired", now)
	assert.True(t, req.Archived)
	assert.Equal(t, "retention expired", req.ArchiveReason)
	assert.Equal(t, types.SupportStatusArchived, req.Status)

	req.MarkPurged("data retention", now)
	assert.True(t, req.Purged)
	assert.Equal(t, "data retention", req.PurgeReason)
}

// --- SupportResponse ---

func TestSupportResponse_Validate(t *testing.T) {
	reqID := types.SupportRequestID{SubmitterAddress: "addr1", Sequence: 1}
	respID := types.SupportResponseID{RequestID: reqID, Sequence: 1}
	resp := types.NewSupportResponse(respID, "agent1", true, validPayload(), time.Now())
	require.NoError(t, resp.Validate())

	require.Error(t, (*types.SupportResponse)(nil).Validate())

	noAuthor := *resp
	noAuthor.AuthorAddress = ""
	require.Error(t, noAuthor.Validate())
}

// --- RelatedEntity ---

func TestRelatedEntity_Validate(t *testing.T) {
	require.NoError(t, (*types.RelatedEntity)(nil).Validate())

	valid := &types.RelatedEntity{Type: types.ResourceTypeDeployment, ID: "dep-1"}
	require.NoError(t, valid.Validate())

	noID := &types.RelatedEntity{Type: types.ResourceTypeDeployment, ID: ""}
	require.Error(t, noID.Validate())

	badType := &types.RelatedEntity{Type: testInvalidValue, ID: "dep-1"}
	require.Error(t, badType.Validate())
}

// --- EncryptedSupportPayload ---

func TestEncryptedSupportPayload_Validate(t *testing.T) {
	require.Error(t, (*types.EncryptedSupportPayload)(nil).Validate())

	noEnvelope := &types.EncryptedSupportPayload{}
	require.Error(t, noEnvelope.Validate())

	valid := validPayload()
	require.NoError(t, valid.Validate())

	badHash := validPayload()
	badHash.EnvelopeHash = []byte("short")
	require.Error(t, badHash.Validate())
}

func TestEncryptedSupportPayload_HasEnvelope(t *testing.T) {
	assert.False(t, (*types.EncryptedSupportPayload)(nil).HasEnvelope())
	assert.False(t, (&types.EncryptedSupportPayload{}).HasEnvelope())
	p := validPayload()
	assert.True(t, p.HasEnvelope())
}

func TestEncryptedSupportPayload_CloneWithoutEnvelope(t *testing.T) {
	p := validPayload()
	p.EnvelopeHash = make([]byte, 32)
	p.EnvelopeRef = "ref-1"
	p.PayloadSize = 100

	clone := p.CloneWithoutEnvelope()
	assert.Nil(t, clone.Envelope)
	assert.Equal(t, "ref-1", clone.EnvelopeRef)
	assert.Equal(t, uint32(100), clone.PayloadSize)
	assert.Len(t, clone.EnvelopeHash, 32)
}

// --- Genesis ---

func TestDefaultGenesisState(t *testing.T) {
	gs := types.DefaultGenesisState()
	require.NotNil(t, gs)
	assert.Empty(t, gs.ExternalRefs)
	assert.Empty(t, gs.SupportRequests)
	assert.Empty(t, gs.SupportResponses)
	assert.Equal(t, uint64(0), gs.EventSequence)
	require.NoError(t, gs.Validate())
}

func TestGenesisState_Validate_DuplicateRefs(t *testing.T) {
	ref := types.ExternalTicketRef{
		ResourceID:       "res-1",
		ResourceType:     types.ResourceTypeDeployment,
		ExternalSystem:   types.ExternalSystemWaldur,
		ExternalTicketID: "W-1",
		CreatedBy:        "addr1",
	}
	gs := types.DefaultGenesisState()
	gs.ExternalRefs = []types.ExternalTicketRef{ref, ref}
	require.Error(t, gs.Validate())
}

func TestGenesisState_Validate_DuplicateRequests(t *testing.T) {
	now := time.Now()
	id := types.SupportRequestID{SubmitterAddress: "addr1", Sequence: 1}
	req := types.NewSupportRequest(id, "SUP-000001", "addr1",
		types.SupportCategoryTechnical, types.SupportPriorityNormal,
		validPayload(), now)
	gs := types.DefaultGenesisState()
	gs.SupportRequests = []types.SupportRequest{*req, *req}
	require.Error(t, gs.Validate())
}

// --- Params ---

func TestParams_Validate(t *testing.T) {
	p := types.DefaultParams()
	require.NoError(t, p.Validate())

	noSystems := p
	noSystems.AllowedExternalSystems = []string{}
	require.Error(t, noSystems.Validate())

	badSystem := p
	badSystem.AllowedExternalSystems = []string{"slack"}
	require.Error(t, badSystem.Validate())

	zeroMax := p
	zeroMax.MaxResponsesPerRequest = 0
	require.Error(t, zeroMax.Validate())
}

func TestParams_IsSystemAllowed(t *testing.T) {
	p := types.DefaultParams()
	assert.True(t, p.IsSystemAllowed(types.ExternalSystemWaldur))
	assert.True(t, p.IsSystemAllowed(types.ExternalSystemJira))
	assert.False(t, p.IsSystemAllowed(types.ExternalSystem("slack")))
}

// --- SupportRequestPayload ---

func TestSupportRequestPayload_Validate(t *testing.T) {
	require.Error(t, (*types.SupportRequestPayload)(nil).Validate())

	noSubject := &types.SupportRequestPayload{Description: "desc"}
	require.Error(t, noSubject.Validate())

	noDesc := &types.SupportRequestPayload{Subject: "sub"}
	require.Error(t, noDesc.Validate())

	valid := &types.SupportRequestPayload{Subject: "sub", Description: "desc"}
	require.NoError(t, valid.Validate())
}

// --- SupportResponsePayload ---

func TestSupportResponsePayload_Validate(t *testing.T) {
	require.Error(t, (*types.SupportResponsePayload)(nil).Validate())

	empty := &types.SupportResponsePayload{}
	require.Error(t, empty.Validate())

	valid := &types.SupportResponsePayload{Message: "hello"}
	require.NoError(t, valid.Validate())
}

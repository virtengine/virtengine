package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/fraud/keeper"
	"github.com/virtengine/virtengine/x/fraud/types"
	rolestypes "github.com/virtengine/virtengine/x/roles/types"
)

// MockRolesKeeper implements the RolesKeeper interface for testing
type MockRolesKeeper struct {
	mock.Mock
}

func (m *MockRolesKeeper) HasRole(ctx sdk.Context, address sdk.AccAddress, role rolestypes.Role) bool {
	args := m.Called(ctx, address, role)
	return args.Bool(0)
}

func (m *MockRolesKeeper) IsModerator(ctx sdk.Context, addr sdk.AccAddress) bool {
	args := m.Called(ctx, addr)
	return args.Bool(0)
}

func (m *MockRolesKeeper) IsAdmin(ctx sdk.Context, addr sdk.AccAddress) bool {
	args := m.Called(ctx, addr)
	return args.Bool(0)
}

// MockProviderKeeper implements the ProviderKeeper interface for testing
type MockProviderKeeper struct {
	mock.Mock
}

func (m *MockProviderKeeper) IsProvider(ctx sdk.Context, addr sdk.AccAddress) bool {
	args := m.Called(ctx, addr)
	return args.Bool(0)
}

type MsgServerTestSuite struct {
	suite.Suite
	ctx            sdk.Context
	keeper         keeper.Keeper
	msgServer      types.MsgServer
	cdc            codec.Codec
	rolesKeeper    *MockRolesKeeper
	providerKeeper *MockProviderKeeper
	authority      string
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) SetupTest() {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	s.ctx = s.createContextWithStore(storeKey)

	s.rolesKeeper = new(MockRolesKeeper)
	s.providerKeeper = new(MockProviderKeeper)
	s.authority = "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"

	s.keeper = keeper.NewKeeper(
		s.cdc,
		storeKey,
		s.rolesKeeper,
		s.providerKeeper,
		s.authority,
	)
	s.msgServer = keeper.NewMsgServerImpl(s.keeper)

	// Set default params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func (s *MsgServerTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	if err != nil {
		s.T().Fatalf("failed to load latest version: %v", err)
	}

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())
	return ctx
}

// validEvidence returns a valid evidence slice for fraud reports
func validEvidence() []types.EncryptedEvidencePB {
	return []types.EncryptedEvidencePB{{
		AlgorithmId:     "X25519-XSalsa20-Poly1305",
		RecipientKeyIds: []string{"moderator-key-1"},
		Nonce:           []byte("123456789012345678901234"), // 24 bytes
		Ciphertext:      []byte("encrypted evidence data ciphertext"),
		SenderPubKey:    []byte("sender-public-key-32-bytes-long!"),
		EvidenceHash:    "sha256:abcdef123456",
	}}
}

// Test: SubmitFraudReport - success
func (s *MsgServerTestSuite) TestSubmitFraudReport_Success() {
	reporterAddr := sdk.AccAddress([]byte("reporter-address123"))

	s.providerKeeper.On("IsProvider", mock.Anything, reporterAddr).Return(true)
	s.rolesKeeper.On("HasRole", mock.Anything, reporterAddr, mock.Anything).Return(true)

	msg := &types.MsgSubmitFraudReport{
		Reporter:        reporterAddr.String(),
		ReportedParty:   "cosmos1reportedparty",
		Category:        types.FraudCategoryPBFakeIdentity,
		Description:     "This is a test fraud report",
		RelatedOrderIds: []string{"order-1", "order-2"},
		Evidence:        validEvidence(),
	}

	resp, err := s.msgServer.SubmitFraudReport(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotEmpty(resp.ReportId)

	// Verify report was stored
	report, found := s.keeper.GetFraudReport(s.ctx, resp.ReportId)
	s.Require().True(found)
	s.Require().Equal(reporterAddr.String(), report.Reporter)
	s.Require().Equal(types.FraudReportStatusSubmitted, report.Status)
}

// Test: SubmitFraudReport - invalid reporter address
func (s *MsgServerTestSuite) TestSubmitFraudReport_InvalidAddress() {
	msg := &types.MsgSubmitFraudReport{
		Reporter:      "invalid-address",
		ReportedParty: "cosmos1reportedparty",
		Category:      types.FraudCategoryPBFakeIdentity,
		Description:   "Test description that is valid",
		Evidence:      validEvidence(),
	}

	_, err := s.msgServer.SubmitFraudReport(s.ctx, msg)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrInvalidReporter)
}

// Test: SubmitFraudReport - unauthorized reporter (not a provider)
func (s *MsgServerTestSuite) TestSubmitFraudReport_UnauthorizedReporter() {
	reporterAddr := sdk.AccAddress([]byte("non-provider-addr"))

	s.providerKeeper.On("IsProvider", mock.Anything, reporterAddr).Return(false)

	msg := &types.MsgSubmitFraudReport{
		Reporter:      reporterAddr.String(),
		ReportedParty: "cosmos1reportedparty",
		Category:      types.FraudCategoryPBFakeIdentity,
		Description:   "Test description that is valid",
		Evidence:      validEvidence(),
	}

	_, err := s.msgServer.SubmitFraudReport(s.ctx, msg)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrUnauthorizedReporter)
}

// Test: AssignModerator - success
func (s *MsgServerTestSuite) TestAssignModerator_Success() {
	reporterAddr := sdk.AccAddress([]byte("reporter-for-assign"))
	moderatorAddr := sdk.AccAddress([]byte("moderator-address"))
	assignToAddr := sdk.AccAddress([]byte("assign-to-address"))

	// Setup provider for report submission
	s.providerKeeper.On("IsProvider", mock.Anything, reporterAddr).Return(true)
	s.rolesKeeper.On("HasRole", mock.Anything, reporterAddr, mock.Anything).Return(true)

	// Setup moderator permissions
	s.rolesKeeper.On("IsModerator", mock.Anything, moderatorAddr).Return(true)
	s.rolesKeeper.On("IsModerator", mock.Anything, assignToAddr).Return(true)

	// First submit a report
	submitMsg := &types.MsgSubmitFraudReport{
		Reporter:      reporterAddr.String(),
		ReportedParty: "cosmos1reported",
		Category:      types.FraudCategoryPBPaymentFraud,
		Description:   "Payment fraud description that is at least 10 characters long",
		Evidence:      validEvidence(),
	}

	submitResp, err := s.msgServer.SubmitFraudReport(s.ctx, submitMsg)
	s.Require().NoError(err)

	// Now assign moderator
	msg := &types.MsgAssignModerator{
		ReportId:  submitResp.ReportId,
		Moderator: moderatorAddr.String(),
		AssignTo:  assignToAddr.String(),
	}

	resp, err := s.msgServer.AssignModerator(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// Verify report was updated
	report, found := s.keeper.GetFraudReport(s.ctx, submitResp.ReportId)
	s.Require().True(found)
	s.Require().Equal(assignToAddr.String(), report.AssignedModerator)
}

// Test: AssignModerator - invalid moderator address
func (s *MsgServerTestSuite) TestAssignModerator_InvalidModeratorAddress() {
	msg := &types.MsgAssignModerator{
		ReportId:  "report-1",
		Moderator: "invalid-address",
		AssignTo:  "cosmos1assignee",
	}

	_, err := s.msgServer.AssignModerator(s.ctx, msg)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrUnauthorizedModerator)
}

// Test: AssignModerator - unauthorized (not a moderator)
func (s *MsgServerTestSuite) TestAssignModerator_Unauthorized() {
	moderatorAddr := sdk.AccAddress([]byte("not-a-moderator"))

	s.rolesKeeper.On("IsModerator", mock.Anything, moderatorAddr).Return(false)

	msg := &types.MsgAssignModerator{
		ReportId:  "report-1",
		Moderator: moderatorAddr.String(),
		AssignTo:  "cosmos1assignee",
	}

	_, err := s.msgServer.AssignModerator(s.ctx, msg)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrUnauthorizedModerator)
}

// Test: UpdateReportStatus - success
func (s *MsgServerTestSuite) TestUpdateReportStatus_Success() {
	reporterAddr := sdk.AccAddress([]byte("reporter-status-up"))
	moderatorAddr := sdk.AccAddress([]byte("moderator-status"))

	s.providerKeeper.On("IsProvider", mock.Anything, reporterAddr).Return(true)
	s.rolesKeeper.On("HasRole", mock.Anything, reporterAddr, mock.Anything).Return(true)
	s.rolesKeeper.On("IsModerator", mock.Anything, moderatorAddr).Return(true)

	// Submit a report first
	submitMsg := &types.MsgSubmitFraudReport{
		Reporter:      reporterAddr.String(),
		ReportedParty: "cosmos1reported",
		Category:      types.FraudCategoryPBSybilAttack,
		Description:   "Sybil attack description",
		Evidence:      validEvidence(),
	}

	submitResp, err := s.msgServer.SubmitFraudReport(s.ctx, submitMsg)
	s.Require().NoError(err)

	// Update status
	msg := &types.MsgUpdateReportStatus{
		ReportId:  submitResp.ReportId,
		Moderator: moderatorAddr.String(),
		NewStatus: types.FraudReportStatusPBReviewing,
		Notes:     "Starting review",
	}

	resp, err := s.msgServer.UpdateReportStatus(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// Verify status was updated
	report, found := s.keeper.GetFraudReport(s.ctx, submitResp.ReportId)
	s.Require().True(found)
	s.Require().Equal(types.FraudReportStatusReviewing, report.Status)
}

// Test: ResolveFraudReport - success
func (s *MsgServerTestSuite) TestResolveFraudReport_Success() {
	reporterAddr := sdk.AccAddress([]byte("reporter-resolve"))
	moderatorAddr := sdk.AccAddress([]byte("moderator-resolve"))

	s.providerKeeper.On("IsProvider", mock.Anything, reporterAddr).Return(true)
	s.rolesKeeper.On("HasRole", mock.Anything, reporterAddr, mock.Anything).Return(true)
	s.rolesKeeper.On("IsModerator", mock.Anything, moderatorAddr).Return(true)

	// Submit a report
	submitMsg := &types.MsgSubmitFraudReport{
		Reporter:      reporterAddr.String(),
		ReportedParty: "cosmos1reported",
		Category:      types.FraudCategoryPBPaymentFraud,
		Description:   "Payment fraud description",
		Evidence:      validEvidence(),
	}

	submitResp, err := s.msgServer.SubmitFraudReport(s.ctx, submitMsg)
	s.Require().NoError(err)

	// Resolve the report
	msg := &types.MsgResolveFraudReport{
		ReportId:   submitResp.ReportId,
		Moderator:  moderatorAddr.String(),
		Resolution: types.ResolutionTypePBWarning,
		Notes:      "Warning issued to the reported party",
	}

	resp, err := s.msgServer.ResolveFraudReport(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// Verify report was resolved
	report, found := s.keeper.GetFraudReport(s.ctx, submitResp.ReportId)
	s.Require().True(found)
	s.Require().Equal(types.FraudReportStatusResolved, report.Status)
}

// Test: RejectFraudReport - success
func (s *MsgServerTestSuite) TestRejectFraudReport_Success() {
	reporterAddr := sdk.AccAddress([]byte("reporter-reject"))
	moderatorAddr := sdk.AccAddress([]byte("moderator-reject"))

	s.providerKeeper.On("IsProvider", mock.Anything, reporterAddr).Return(true)
	s.rolesKeeper.On("HasRole", mock.Anything, reporterAddr, mock.Anything).Return(true)
	s.rolesKeeper.On("IsModerator", mock.Anything, moderatorAddr).Return(true)

	// Submit a report
	submitMsg := &types.MsgSubmitFraudReport{
		Reporter:      reporterAddr.String(),
		ReportedParty: "cosmos1reported",
		Category:      types.FraudCategoryPBOther,
		Description:   "False report description",
		Evidence:      validEvidence(),
	}

	submitResp, err := s.msgServer.SubmitFraudReport(s.ctx, submitMsg)
	s.Require().NoError(err)

	// Reject the report
	msg := &types.MsgRejectFraudReport{
		ReportId:  submitResp.ReportId,
		Moderator: moderatorAddr.String(),
		Notes:     "Insufficient evidence",
	}

	resp, err := s.msgServer.RejectFraudReport(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// Verify report was rejected
	report, found := s.keeper.GetFraudReport(s.ctx, submitResp.ReportId)
	s.Require().True(found)
	s.Require().Equal(types.FraudReportStatusRejected, report.Status)
}

// Test: EscalateFraudReport - success
func (s *MsgServerTestSuite) TestEscalateFraudReport_Success() {
	reporterAddr := sdk.AccAddress([]byte("reporter-escalate"))
	moderatorAddr := sdk.AccAddress([]byte("moderator-escalate"))

	s.providerKeeper.On("IsProvider", mock.Anything, reporterAddr).Return(true)
	s.rolesKeeper.On("HasRole", mock.Anything, reporterAddr, mock.Anything).Return(true)
	s.rolesKeeper.On("IsModerator", mock.Anything, moderatorAddr).Return(true)

	// Submit a report
	submitMsg := &types.MsgSubmitFraudReport{
		Reporter:      reporterAddr.String(),
		ReportedParty: "cosmos1reported",
		Category:      types.FraudCategoryPBSybilAttack,
		Description:   "Complex sybil attack description",
		Evidence:      validEvidence(),
	}

	submitResp, err := s.msgServer.SubmitFraudReport(s.ctx, submitMsg)
	s.Require().NoError(err)

	// First transition to reviewing status (required before escalation)
	updateMsg := &types.MsgUpdateReportStatus{
		ReportId:  submitResp.ReportId,
		Moderator: moderatorAddr.String(),
		NewStatus: types.FraudReportStatusPBReviewing,
		Notes:     "Starting review",
	}
	_, err = s.msgServer.UpdateReportStatus(s.ctx, updateMsg)
	s.Require().NoError(err)

	// Escalate the report
	msg := &types.MsgEscalateFraudReport{
		ReportId:  submitResp.ReportId,
		Moderator: moderatorAddr.String(),
		Reason:    "Complex case requiring admin review",
	}

	resp, err := s.msgServer.EscalateFraudReport(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// Verify report was escalated
	report, found := s.keeper.GetFraudReport(s.ctx, submitResp.ReportId)
	s.Require().True(found)
	s.Require().Equal(types.FraudReportStatusEscalated, report.Status)
}

// Test: UpdateParams - success
func (s *MsgServerTestSuite) TestUpdateParams_Success() {
	params := types.DefaultParams()
	params.AutoAssignEnabled = true
	paramsPB := types.ParamsToProto(&params)

	msg := &types.MsgUpdateParams{
		Authority: s.authority,
		Params:    *paramsPB,
	}

	resp, err := s.msgServer.UpdateParams(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// Verify params were updated
	storedParams := s.keeper.GetParams(s.ctx)
	s.Require().Equal(params.AutoAssignEnabled, storedParams.AutoAssignEnabled)
}

// Test: UpdateParams - unauthorized
func (s *MsgServerTestSuite) TestUpdateParams_Unauthorized() {
	params := types.DefaultParams()
	paramsPB := types.ParamsToProto(&params)

	msg := &types.MsgUpdateParams{
		Authority: "cosmos1wrongauthority",
		Params:    *paramsPB,
	}

	_, err := s.msgServer.UpdateParams(s.ctx, msg)
	s.Require().Error(err)
}

// Test: UpdateParams - invalid authority address
func (s *MsgServerTestSuite) TestUpdateParams_InvalidAuthority() {
	params := types.DefaultParams()
	paramsPB := types.ParamsToProto(&params)

	msg := &types.MsgUpdateParams{
		Authority: "invalid-address",
		Params:    *paramsPB,
	}

	_, err := s.msgServer.UpdateParams(s.ctx, msg)
	s.Require().Error(err)
}

// Test: Report not found errors
func (s *MsgServerTestSuite) TestReportNotFound() {
	moderatorAddr := sdk.AccAddress([]byte("moderator-notfound"))
	s.rolesKeeper.On("IsModerator", mock.Anything, moderatorAddr).Return(true)

	// Try to assign moderator to non-existent report
	msg := &types.MsgAssignModerator{
		ReportId:  "non-existent-report",
		Moderator: moderatorAddr.String(),
		AssignTo:  "cosmos1assignee",
	}

	// First set up the assignee as moderator too
	assigneeAddr, _ := sdk.AccAddressFromBech32("cosmos1assignee")
	s.rolesKeeper.On("IsModerator", mock.Anything, assigneeAddr).Return(true)

	_, err := s.msgServer.AssignModerator(s.ctx, msg)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrReportNotFound)
}

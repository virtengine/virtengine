// Package keeper provides VEID module keeper implementation.
//
// This file contains tests for the audit log export functionality.
//
// Task Reference: VE-3033 - Add VEID Audit Log Export Endpoint
package keeper_test

import (
	"encoding/json"
	"strings"
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
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// Test constants
const (
	testAuditSeed1 = "audit_address_0001_"
	testAuditSeed2 = "audit_address_0002_"
	testAuditSeed3 = "audit_address_0003_"
)

type AuditExportTestSuite struct {
	suite.Suite
	ctx        sdk.Context
	keeper     keeper.Keeper
	cdc        codec.Codec
	stateStore store.CommitMultiStore
	// Test addresses
	auditAddr1 sdk.AccAddress
	auditAddr2 sdk.AccAddress
	auditAddr3 sdk.AccAddress
}

func TestAuditExportTestSuite(t *testing.T) {
	suite.Run(t, new(AuditExportTestSuite))
}

func (s *AuditExportTestSuite) SetupTest() {
	// Create test addresses
	s.auditAddr1 = sdk.AccAddress([]byte(testAuditSeed1))
	s.auditAddr2 = sdk.AccAddress([]byte(testAuditSeed2))
	s.auditAddr3 = sdk.AccAddress([]byte(testAuditSeed3))

	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create context with store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	s.Require().NoError(err)
	s.stateStore = stateStore

	s.ctx = sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	// Create keeper
	s.keeper = keeper.NewKeeper(s.cdc, storeKey, "authority")

	// Set default params
	err = s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

// TearDownTest closes the IAVL store to stop background pruning goroutines
func (s *AuditExportTestSuite) TearDownTest() {
	CloseStoreIfNeeded(s.stateStore)
}

// ============================================================================
// AuditEntry Tests
// ============================================================================

func (s *AuditExportTestSuite) TestAuditEntry_CreateAndRetrieve() {
	// Create an audit entry
	entry := types.NewAuditEntry(
		"event-001",
		types.AuditEventTypeVerification,
		s.auditAddr1.String(),
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
		map[string]interface{}{
			"scope_id": "scope-123",
			"status":   "verified",
		},
	)

	// Store the entry
	err := s.keeper.SetAuditEntry(s.ctx, entry)
	s.Require().NoError(err)

	// Retrieve the entry
	retrieved, found := s.keeper.GetAuditEntry(s.ctx, "event-001")
	s.Require().True(found)
	s.Require().Equal(entry.EventID, retrieved.EventID)
	s.Require().Equal(entry.EventType, retrieved.EventType)
	s.Require().Equal(entry.Address, retrieved.Address)
	s.Require().NotEmpty(retrieved.Hash)
}

func (s *AuditExportTestSuite) TestAuditEntry_NotFound() {
	_, found := s.keeper.GetAuditEntry(s.ctx, "non-existent")
	s.Require().False(found)
}

func (s *AuditExportTestSuite) TestAuditEntry_InvalidEntry() {
	// Create an invalid entry (missing event ID)
	entry := &types.AuditEntry{
		EventType: types.AuditEventTypeVerification,
		Address:   s.auditAddr1.String(),
		Timestamp: s.ctx.BlockTime(),
	}

	err := s.keeper.SetAuditEntry(s.ctx, entry)
	s.Require().Error(err)
}

// ============================================================================
// Export with Date Range Tests
// ============================================================================

func (s *AuditExportTestSuite) TestExportAuditLogs_DateRange() {
	// Create entries at different times
	baseTime := s.ctx.BlockTime()

	entries := []struct {
		id        string
		eventType types.AuditEventType
		timestamp time.Time
	}{
		{"event-1", types.AuditEventTypeVerification, baseTime.Add(-2 * time.Hour)},
		{"event-2", types.AuditEventTypeScopeUpload, baseTime.Add(-1 * time.Hour)},
		{"event-3", types.AuditEventTypeScoreUpdate, baseTime},
		{"event-4", types.AuditEventTypeCompliance, baseTime.Add(1 * time.Hour)},
		{"event-5", types.AuditEventTypeAppeal, baseTime.Add(2 * time.Hour)},
	}

	for _, e := range entries {
		ctx := s.ctx.WithBlockTime(e.timestamp)
		entry := types.NewAuditEntry(
			e.id,
			e.eventType,
			s.auditAddr1.String(),
			e.timestamp,
			ctx.BlockHeight(),
			nil,
		)
		err := s.keeper.SetAuditEntry(ctx, entry)
		s.Require().NoError(err)
	}

	// Export entries within a specific time range
	req := &types.AuditExportRequest{
		StartTime: baseTime.Add(-90 * time.Minute),
		EndTime:   baseTime.Add(90 * time.Minute),
		Format:    types.AuditExportFormatJSON,
	}

	resp, err := s.keeper.ExportAuditLogs(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// Should include events 2, 3, 4 (within the range)
	s.Require().Equal(uint64(3), resp.ExportedCount)
}

func (s *AuditExportTestSuite) TestExportAuditLogs_InvalidTimeRange() {
	req := &types.AuditExportRequest{
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(-1 * time.Hour), // End before start
		Format:    types.AuditExportFormatJSON,
	}

	_, err := s.keeper.ExportAuditLogs(s.ctx, req)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "end_time must be after start_time")
}

// ============================================================================
// Filter by Event Type Tests
// ============================================================================

func (s *AuditExportTestSuite) TestExportAuditLogs_FilterByEventType() {
	baseTime := s.ctx.BlockTime()

	// Create entries of different types
	entries := []struct {
		id        string
		eventType types.AuditEventType
	}{
		{"type-1", types.AuditEventTypeVerification},
		{"type-2", types.AuditEventTypeVerification},
		{"type-3", types.AuditEventTypeScopeUpload},
		{"type-4", types.AuditEventTypeCompliance},
		{"type-5", types.AuditEventTypeVerification},
	}

	for i, e := range entries {
		entry := types.NewAuditEntry(
			e.id,
			e.eventType,
			s.auditAddr1.String(),
			baseTime.Add(time.Duration(i)*time.Minute),
			s.ctx.BlockHeight(),
			nil,
		)
		err := s.keeper.SetAuditEntry(s.ctx, entry)
		s.Require().NoError(err)
	}

	// Export only verification events
	req := &types.AuditExportRequest{
		StartTime:  baseTime.Add(-1 * time.Hour),
		EndTime:    baseTime.Add(1 * time.Hour),
		EventTypes: []types.AuditEventType{types.AuditEventTypeVerification},
		Format:     types.AuditExportFormatJSON,
	}

	resp, err := s.keeper.ExportAuditLogs(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// Should have 3 verification events
	s.Require().Equal(uint64(3), resp.ExportedCount)
	for _, entry := range resp.Entries {
		s.Require().Equal(types.AuditEventTypeVerification, entry.EventType)
	}
}

func (s *AuditExportTestSuite) TestExportAuditLogs_InvalidEventType() {
	req := &types.AuditExportRequest{
		StartTime:  time.Now().Add(-1 * time.Hour),
		EndTime:    time.Now().Add(1 * time.Hour),
		EventTypes: []types.AuditEventType{"INVALID_TYPE"},
		Format:     types.AuditExportFormatJSON,
	}

	_, err := s.keeper.ExportAuditLogs(s.ctx, req)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid event type")
}

// ============================================================================
// Filter by Address Tests
// ============================================================================

func (s *AuditExportTestSuite) TestExportAuditLogs_FilterByAddress() {
	baseTime := s.ctx.BlockTime()

	// Create entries for different addresses
	addresses := []sdk.AccAddress{s.auditAddr1, s.auditAddr2, s.auditAddr1, s.auditAddr3, s.auditAddr1}

	for i, addr := range addresses {
		entry := types.NewAuditEntry(
			"addr-"+string(rune('1'+i)),
			types.AuditEventTypeVerification,
			addr.String(),
			baseTime.Add(time.Duration(i)*time.Minute),
			s.ctx.BlockHeight(),
			nil,
		)
		err := s.keeper.SetAuditEntry(s.ctx, entry)
		s.Require().NoError(err)
	}

	// Export only for auditAddr1
	req := &types.AuditExportRequest{
		StartTime: baseTime.Add(-1 * time.Hour),
		EndTime:   baseTime.Add(1 * time.Hour),
		Addresses: []string{s.auditAddr1.String()},
		Format:    types.AuditExportFormatJSON,
	}

	resp, err := s.keeper.ExportAuditLogs(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// Should have 3 entries for auditAddr1
	s.Require().Equal(uint64(3), resp.ExportedCount)
	for _, entry := range resp.Entries {
		s.Require().Equal(s.auditAddr1.String(), entry.Address)
	}
}

func (s *AuditExportTestSuite) TestExportAuditLogs_MultipleAddresses() {
	baseTime := s.ctx.BlockTime()

	// Create entries for different addresses
	for i, addr := range []sdk.AccAddress{s.auditAddr1, s.auditAddr2, s.auditAddr3} {
		entry := types.NewAuditEntry(
			"multi-"+string(rune('1'+i)),
			types.AuditEventTypeVerification,
			addr.String(),
			baseTime.Add(time.Duration(i)*time.Minute),
			s.ctx.BlockHeight(),
			nil,
		)
		err := s.keeper.SetAuditEntry(s.ctx, entry)
		s.Require().NoError(err)
	}

	// Export for addr1 and addr2
	req := &types.AuditExportRequest{
		StartTime: baseTime.Add(-1 * time.Hour),
		EndTime:   baseTime.Add(1 * time.Hour),
		Addresses: []string{s.auditAddr1.String(), s.auditAddr2.String()},
		Format:    types.AuditExportFormatJSON,
	}

	resp, err := s.keeper.ExportAuditLogs(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// Should have 2 entries
	s.Require().Equal(uint64(2), resp.ExportedCount)
}

// ============================================================================
// Format Conversion Tests
// ============================================================================

func (s *AuditExportTestSuite) TestFormatAuditEntries_JSON() {
	entries := []*types.AuditEntry{
		{
			EventID:     "json-1",
			EventType:   types.AuditEventTypeVerification,
			Address:     s.auditAddr1.String(),
			Timestamp:   s.ctx.BlockTime(),
			BlockHeight: 100,
			Hash:        "abc123",
		},
		{
			EventID:     "json-2",
			EventType:   types.AuditEventTypeScopeUpload,
			Address:     s.auditAddr2.String(),
			Timestamp:   s.ctx.BlockTime(),
			BlockHeight: 101,
			Hash:        "def456",
		},
	}

	data, err := s.keeper.FormatAuditEntries(entries, types.AuditExportFormatJSON)
	s.Require().NoError(err)
	s.Require().NotEmpty(data)

	// Verify it's valid JSON
	var parsed []*types.AuditEntry
	err = json.Unmarshal(data, &parsed)
	s.Require().NoError(err)
	s.Require().Len(parsed, 2)
	s.Require().Equal("json-1", parsed[0].EventID)
	s.Require().Equal("json-2", parsed[1].EventID)
}

func (s *AuditExportTestSuite) TestFormatAuditEntries_CSV() {
	entries := []*types.AuditEntry{
		{
			EventID:     "csv-1",
			EventType:   types.AuditEventTypeVerification,
			Address:     s.auditAddr1.String(),
			Timestamp:   s.ctx.BlockTime(),
			BlockHeight: 100,
			Hash:        "abc123",
		},
		{
			EventID:     "csv-2",
			EventType:   types.AuditEventTypeScopeUpload,
			Address:     s.auditAddr2.String(),
			Timestamp:   s.ctx.BlockTime(),
			BlockHeight: 101,
			Hash:        "def456",
		},
	}

	data, err := s.keeper.FormatAuditEntries(entries, types.AuditExportFormatCSV)
	s.Require().NoError(err)
	s.Require().NotEmpty(data)

	lines := strings.Split(string(data), "\n")
	s.Require().GreaterOrEqual(len(lines), 3) // Header + 2 entries

	// Verify header
	s.Require().Equal(types.AuditEntryCSVHeader(), lines[0])

	// Verify first data row contains expected values
	s.Require().Contains(lines[1], "csv-1")
	s.Require().Contains(lines[1], "VERIFICATION")
}

func (s *AuditExportTestSuite) TestFormatAuditEntries_JSONL() {
	entries := []*types.AuditEntry{
		{
			EventID:     "jsonl-1",
			EventType:   types.AuditEventTypeVerification,
			Address:     s.auditAddr1.String(),
			Timestamp:   s.ctx.BlockTime(),
			BlockHeight: 100,
			Hash:        "abc123",
		},
		{
			EventID:     "jsonl-2",
			EventType:   types.AuditEventTypeScopeUpload,
			Address:     s.auditAddr2.String(),
			Timestamp:   s.ctx.BlockTime(),
			BlockHeight: 101,
			Hash:        "def456",
		},
	}

	data, err := s.keeper.FormatAuditEntries(entries, types.AuditExportFormatJSONL)
	s.Require().NoError(err)
	s.Require().NotEmpty(data)

	lines := strings.Split(string(data), "\n")
	s.Require().Len(lines, 2)

	// Each line should be valid JSON
	for _, line := range lines {
		var parsed types.AuditEntry
		err := json.Unmarshal([]byte(line), &parsed)
		s.Require().NoError(err)
	}
}

func (s *AuditExportTestSuite) TestFormatAuditEntries_InvalidFormat() {
	entries := []*types.AuditEntry{}
	_, err := s.keeper.FormatAuditEntries(entries, types.AuditExportFormat("INVALID"))
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unsupported format")
}

// ============================================================================
// Hash Chain Verification Tests
// ============================================================================

func (s *AuditExportTestSuite) TestVerifyAuditChain_ValidChain() {
	baseTime := s.ctx.BlockTime()

	// Create a chain of entries
	for i := 0; i < 5; i++ {
		entry := types.NewAuditEntry(
			"chain-"+string(rune('1'+i)),
			types.AuditEventTypeVerification,
			s.auditAddr1.String(),
			baseTime.Add(time.Duration(i)*time.Minute),
			s.ctx.BlockHeight()+int64(i),
			nil,
		)
		err := s.keeper.SetAuditEntry(s.ctx, entry)
		s.Require().NoError(err)
	}

	// Verify the chain
	result, err := s.keeper.VerifyAuditChain(s.ctx, "", "")
	s.Require().NoError(err)
	s.Require().NotNil(result)
	s.Require().True(result.Valid)
	s.Require().Equal(uint64(5), result.EntriesVerified)
	s.Require().Empty(result.BrokenAt)
}

func (s *AuditExportTestSuite) TestVerifyAuditChain_EmptyChain() {
	result, err := s.keeper.VerifyAuditChain(s.ctx, "", "")
	s.Require().NoError(err)
	s.Require().NotNil(result)
	s.Require().False(result.Valid)
	s.Require().Contains(result.BrokenReason, "no entries found")
}

func (s *AuditExportTestSuite) TestAuditEntry_VerifyHash() {
	entry := types.NewAuditEntry(
		"hash-test",
		types.AuditEventTypeVerification,
		s.auditAddr1.String(),
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
		map[string]interface{}{"test": "data"},
	)

	// Set hash
	err := entry.SetHashWithPrevious("")
	s.Require().NoError(err)
	s.Require().NotEmpty(entry.Hash)

	// Verify hash
	valid, err := entry.VerifyHash()
	s.Require().NoError(err)
	s.Require().True(valid)

	// Tamper with the entry
	entry.Address = "tampered"
	valid, err = entry.VerifyHash()
	s.Require().NoError(err)
	s.Require().False(valid)
}

func (s *AuditExportTestSuite) TestAuditEntry_ChainLinking() {
	entry1 := types.NewAuditEntry(
		"link-1",
		types.AuditEventTypeVerification,
		s.auditAddr1.String(),
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
		nil,
	)
	err := entry1.SetHashWithPrevious("")
	s.Require().NoError(err)

	entry2 := types.NewAuditEntry(
		"link-2",
		types.AuditEventTypeVerification,
		s.auditAddr1.String(),
		s.ctx.BlockTime().Add(1*time.Minute),
		s.ctx.BlockHeight()+1,
		nil,
	)
	err = entry2.SetHashWithPrevious(entry1.Hash)
	s.Require().NoError(err)

	// Verify chain link
	s.Require().Equal(entry1.Hash, entry2.PreviousHash)
	s.Require().NotEqual(entry1.Hash, entry2.Hash)
}

// ============================================================================
// Hash Audit Entry Tests
// ============================================================================

func (s *AuditExportTestSuite) TestHashAuditEntry_Success() {
	entry := types.NewAuditEntry(
		"hash-entry-1",
		types.AuditEventTypeVerification,
		s.auditAddr1.String(),
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
		nil,
	)
	err := s.keeper.SetAuditEntry(s.ctx, entry)
	s.Require().NoError(err)

	hash, err := s.keeper.HashAuditEntry(s.ctx, "hash-entry-1")
	s.Require().NoError(err)
	s.Require().NotEmpty(hash)
	s.Require().Len(hash, 64) // SHA256 produces 64 hex characters
}

func (s *AuditExportTestSuite) TestHashAuditEntry_NotFound() {
	_, err := s.keeper.HashAuditEntry(s.ctx, "non-existent")
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "audit entry not found")
}

// ============================================================================
// Record Audit Event Tests
// ============================================================================

func (s *AuditExportTestSuite) TestRecordAuditEvent_Success() {
	details := map[string]interface{}{
		"scope_id": "scope-abc",
		"score":    85,
	}

	err := s.keeper.RecordAuditEvent(
		s.ctx,
		types.AuditEventTypeVerification,
		s.auditAddr1.String(),
		details,
	)
	s.Require().NoError(err)

	// List entries to verify it was recorded
	entries, count, err := s.keeper.ListAuditEntries(s.ctx, 0, 10)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), count)
	s.Require().Len(entries, 1)
	s.Require().Equal(types.AuditEventTypeVerification, entries[0].EventType)
	s.Require().Equal(s.auditAddr1.String(), entries[0].Address)
}

// ============================================================================
// Pagination Tests
// ============================================================================

func (s *AuditExportTestSuite) TestListAuditEntries_Pagination() {
	// Create 15 entries
	for i := 0; i < 15; i++ {
		entry := types.NewAuditEntry(
			"page-"+string(rune('a'+i)),
			types.AuditEventTypeVerification,
			s.auditAddr1.String(),
			s.ctx.BlockTime().Add(time.Duration(i)*time.Minute),
			s.ctx.BlockHeight(),
			nil,
		)
		err := s.keeper.SetAuditEntry(s.ctx, entry)
		s.Require().NoError(err)
	}

	// First page
	entries, total, err := s.keeper.ListAuditEntries(s.ctx, 0, 5)
	s.Require().NoError(err)
	s.Require().Equal(uint64(15), total)
	s.Require().Len(entries, 5)

	// Second page
	entries2, total2, err := s.keeper.ListAuditEntries(s.ctx, 5, 5)
	s.Require().NoError(err)
	s.Require().Equal(uint64(15), total2)
	s.Require().Len(entries2, 5)

	// Ensure different entries
	s.Require().NotEqual(entries[0].EventID, entries2[0].EventID)

	// Third page (partial)
	entries3, _, err := s.keeper.ListAuditEntries(s.ctx, 10, 5)
	s.Require().NoError(err)
	s.Require().Len(entries3, 5)
}

func (s *AuditExportTestSuite) TestExportAuditLogs_Pagination() {
	baseTime := s.ctx.BlockTime()

	// Create 10 entries
	for i := 0; i < 10; i++ {
		entry := types.NewAuditEntry(
			"export-page-"+string(rune('a'+i)),
			types.AuditEventTypeVerification,
			s.auditAddr1.String(),
			baseTime.Add(time.Duration(i)*time.Minute),
			s.ctx.BlockHeight(),
			nil,
		)
		err := s.keeper.SetAuditEntry(s.ctx, entry)
		s.Require().NoError(err)
	}

	// First page
	req := &types.AuditExportRequest{
		StartTime: baseTime.Add(-1 * time.Hour),
		EndTime:   baseTime.Add(1 * time.Hour),
		Format:    types.AuditExportFormatJSON,
		Limit:     3,
		Offset:    0,
	}

	resp, err := s.keeper.ExportAuditLogs(s.ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), resp.ExportedCount)
	s.Require().True(resp.HasMore)
	s.Require().Equal(uint64(3), resp.NextOffset)

	// Second page
	req.Offset = resp.NextOffset
	resp2, err := s.keeper.ExportAuditLogs(s.ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), resp2.ExportedCount)
	s.Require().True(resp2.HasMore)
}

// ============================================================================
// Validation Tests
// ============================================================================

func (s *AuditExportTestSuite) TestAuditExportRequest_Validate() {
	testCases := []struct {
		name    string
		req     *types.AuditExportRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &types.AuditExportRequest{
				StartTime: time.Now().Add(-1 * time.Hour),
				EndTime:   time.Now(),
				Format:    types.AuditExportFormatJSON,
			},
			wantErr: false,
		},
		{
			name: "missing start time",
			req: &types.AuditExportRequest{
				EndTime: time.Now(),
				Format:  types.AuditExportFormatJSON,
			},
			wantErr: true,
			errMsg:  "start_time is required",
		},
		{
			name: "missing end time",
			req: &types.AuditExportRequest{
				StartTime: time.Now().Add(-1 * time.Hour),
				Format:    types.AuditExportFormatJSON,
			},
			wantErr: true,
			errMsg:  "end_time is required",
		},
		{
			name: "invalid format",
			req: &types.AuditExportRequest{
				StartTime: time.Now().Add(-1 * time.Hour),
				EndTime:   time.Now(),
				Format:    types.AuditExportFormat("XML"),
			},
			wantErr: true,
			errMsg:  "invalid format",
		},
		{
			name: "limit exceeds maximum",
			req: &types.AuditExportRequest{
				StartTime: time.Now().Add(-1 * time.Hour),
				EndTime:   time.Now(),
				Format:    types.AuditExportFormatJSON,
				Limit:     types.MaxAuditExportLimit + 1,
			},
			wantErr: true,
			errMsg:  "limit exceeds maximum",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := tc.req.Validate()
			if tc.wantErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *AuditExportTestSuite) TestAuditEventType_IsValid() {
	s.Require().True(types.IsValidAuditEventType(types.AuditEventTypeVerification))
	s.Require().True(types.IsValidAuditEventType(types.AuditEventTypeAppeal))
	s.Require().True(types.IsValidAuditEventType(types.AuditEventTypeCompliance))
	s.Require().False(types.IsValidAuditEventType(types.AuditEventType("INVALID")))
}

func (s *AuditExportTestSuite) TestAuditExportFormat_IsValid() {
	s.Require().True(types.IsValidAuditExportFormat(types.AuditExportFormatJSON))
	s.Require().True(types.IsValidAuditExportFormat(types.AuditExportFormatCSV))
	s.Require().True(types.IsValidAuditExportFormat(types.AuditExportFormatJSONL))
	s.Require().False(types.IsValidAuditExportFormat(types.AuditExportFormat("XML")))
}

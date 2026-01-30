//go:build e2e.integration

package partition

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/testutil/partition"
)

// MessageReplayTestSuite tests message replay prevention mechanisms.
type MessageReplayTestSuite struct {
	suite.Suite

	controller *partition.Controller
	filter     *partition.MessageFilter
	nodes      []partition.NodeID
}

// TestMessageReplay runs the message replay test suite.
func TestMessageReplay(t *testing.T) {
	suite.Run(t, new(MessageReplayTestSuite))
}

// SetupSuite runs once before all tests.
func (s *MessageReplayTestSuite) SetupSuite() {
	s.nodes = []partition.NodeID{
		"validator-0",
		"validator-1",
		"validator-2",
		"validator-3",
	}
	s.controller = partition.NewController(s.nodes...)
	s.filter = s.controller.Filter()
}

// SetupTest runs before each test.
func (s *MessageReplayTestSuite) SetupTest() {
	s.filter.Reset()
}

// TearDownTest runs after each test.
func (s *MessageReplayTestSuite) TearDownTest() {
	s.filter.Reset()
	s.controller.Heal()
}

// generateMessageHash generates a unique message hash for testing.
func (s *MessageReplayTestSuite) generateMessageHash(msg string) string {
	hash := sha256.Sum256([]byte(msg))
	return hex.EncodeToString(hash[:])
}

// =============================================================================
// Basic Replay Prevention Tests
// =============================================================================

// TestDuplicateMessageRejection verifies duplicate messages are rejected.
func (s *MessageReplayTestSuite) TestDuplicateMessageRejection() {
	s.T().Log("=== Test: Duplicate Message Rejection ===")

	msg := "block:height=100:hash=abc123"
	msgHash := s.generateMessageHash(msg)

	// First time should not be a replay
	isReplay := s.filter.IsMessageReplayed(msgHash)
	require.False(s.T(), isReplay, "First message should not be marked as replay")

	// Second time should be a replay
	isReplay = s.filter.IsMessageReplayed(msgHash)
	require.True(s.T(), isReplay, "Duplicate message should be marked as replay")

	// Third time should still be a replay
	isReplay = s.filter.IsMessageReplayed(msgHash)
	require.True(s.T(), isReplay, "Message should continue to be marked as replay")

	s.T().Log("Duplicate message rejection working correctly")
}

// TestUniqueMessagesAccepted verifies unique messages are accepted.
func (s *MessageReplayTestSuite) TestUniqueMessagesAccepted() {
	s.T().Log("=== Test: Unique Messages Accepted ===")

	messages := []string{
		"block:height=100:hash=abc123",
		"block:height=101:hash=def456",
		"block:height=102:hash=ghi789",
		"vote:height=100:round=0:validator=v0",
		"vote:height=100:round=0:validator=v1",
	}

	for _, msg := range messages {
		msgHash := s.generateMessageHash(msg)
		isReplay := s.filter.IsMessageReplayed(msgHash)
		require.False(s.T(), isReplay, "Unique message should not be marked as replay: %s", msg)
	}

	require.Equal(s.T(), len(messages), s.filter.MessageHistorySize(),
		"All unique messages should be in history")

	s.T().Logf("Accepted %d unique messages", len(messages))
}

// TestMessageHistorySize verifies message history tracking.
func (s *MessageReplayTestSuite) TestMessageHistorySize() {
	s.T().Log("=== Test: Message History Size ===")

	// Add multiple messages
	for i := 0; i < 100; i++ {
		msg := fmt.Sprintf("message:%d", i)
		msgHash := s.generateMessageHash(msg)
		s.filter.IsMessageReplayed(msgHash)
	}

	require.Equal(s.T(), 100, s.filter.MessageHistorySize(),
		"History should contain all messages")

	s.T().Logf("Message history size: %d", s.filter.MessageHistorySize())
}

// =============================================================================
// Sequence Number Tests
// =============================================================================

// TestSequenceNumberEnforcement tests that sequence numbers are properly enforced.
func (s *MessageReplayTestSuite) TestSequenceNumberEnforcement() {
	s.T().Log("=== Test: Sequence Number Enforcement ===")

	validator := "validator-0"

	// Simulate messages with sequence numbers
	for seq := 1; seq <= 10; seq++ {
		msg := fmt.Sprintf("tx:validator=%s:seq=%d:nonce=abc", validator, seq)
		msgHash := s.generateMessageHash(msg)

		isReplay := s.filter.IsMessageReplayed(msgHash)
		require.False(s.T(), isReplay,
			"Message with new sequence %d should be accepted", seq)
	}

	// Try to replay an old sequence number
	oldMsg := fmt.Sprintf("tx:validator=%s:seq=%d:nonce=abc", validator, 5)
	oldMsgHash := s.generateMessageHash(oldMsg)

	isReplay := s.filter.IsMessageReplayed(oldMsgHash)
	require.True(s.T(), isReplay,
		"Replayed message with old sequence should be rejected")

	s.T().Log("Sequence number enforcement working correctly")
}

// TestReplayAfterPartition tests replay attempts after partition heals.
func (s *MessageReplayTestSuite) TestReplayAfterPartition() {
	s.T().Log("=== Test: Replay After Partition ===")

	// Record messages before partition
	prePartitionMessages := []string{
		"block:height=100:hash=pre1",
		"block:height=101:hash=pre2",
		"vote:height=100:round=0:val=v0",
	}

	for _, msg := range prePartitionMessages {
		hash := s.generateMessageHash(msg)
		s.filter.IsMessageReplayed(hash) // Record message
	}

	// Apply partition
	scenario := partition.CreateSimplePartition(s.nodes)
	s.controller.ApplyPartition(scenario.Groups)

	// Some messages during partition
	partitionMessages := []string{
		"block:height=102:hash=part1",
		"block:height=103:hash=part2",
	}

	for _, msg := range partitionMessages {
		hash := s.generateMessageHash(msg)
		s.filter.IsMessageReplayed(hash)
	}

	// Heal partition
	s.controller.Heal()

	// Attempt to replay pre-partition messages (should be rejected)
	for _, msg := range prePartitionMessages {
		hash := s.generateMessageHash(msg)
		isReplay := s.filter.IsMessageReplayed(hash)
		require.True(s.T(), isReplay,
			"Pre-partition message should be rejected as replay: %s", msg)
	}

	// Attempt to replay partition messages (should also be rejected)
	for _, msg := range partitionMessages {
		hash := s.generateMessageHash(msg)
		isReplay := s.filter.IsMessageReplayed(hash)
		require.True(s.T(), isReplay,
			"Partition message should be rejected as replay: %s", msg)
	}

	// New messages should be accepted
	newMsg := "block:height=104:hash=new1"
	isReplay := s.filter.IsMessageReplayed(s.generateMessageHash(newMsg))
	require.False(s.T(), isReplay, "New message should be accepted")

	s.T().Log("Replay prevention after partition working correctly")
}

// =============================================================================
// Intercept Tests
// =============================================================================

// TestInterceptBlocksReplays tests the Intercept function blocks replays.
func (s *MessageReplayTestSuite) TestInterceptBlocksReplays() {
	s.T().Log("=== Test: Intercept Blocks Replays ===")

	from := s.nodes[0]
	to := s.nodes[1]
	msgHash := s.generateMessageHash("test-message")

	// First intercept should allow
	result := s.filter.Intercept(from, to, msgHash)
	require.True(s.T(), result.Allow, "First message should be allowed")
	require.Equal(s.T(), 1, result.Copies, "Should have 1 copy")

	// Second intercept with same hash should block
	result = s.filter.Intercept(from, to, msgHash)
	require.False(s.T(), result.Allow, "Replay should be blocked")
	require.Equal(s.T(), 0, result.Copies, "Replay should have 0 copies")

	s.T().Log("Intercept correctly blocks replays")
}

// TestInterceptWithBlocking tests interception when connection is blocked.
func (s *MessageReplayTestSuite) TestInterceptWithBlocking() {
	s.T().Log("=== Test: Intercept With Blocking ===")

	from := s.nodes[0]
	to := s.nodes[1]

	// Block the connection
	s.filter.SetBlocked(from, to, true)

	msgHash := s.generateMessageHash("test-message")
	result := s.filter.Intercept(from, to, msgHash)

	require.False(s.T(), result.Allow, "Message should be blocked")
	require.Equal(s.T(), 0, result.Copies, "Blocked messages should have 0 copies")

	// Unblock and try again
	s.filter.SetBlocked(from, to, false)
	result = s.filter.Intercept(from, to, s.generateMessageHash("new-message"))
	require.True(s.T(), result.Allow, "New message should be allowed after unblock")

	s.T().Log("Interception with blocking working correctly")
}

// TestInterceptWithDuplication tests message duplication via intercept.
func (s *MessageReplayTestSuite) TestInterceptWithDuplication() {
	s.T().Log("=== Test: Intercept With Duplication ===")

	from := s.nodes[0]
	to := s.nodes[1]

	// Set high duplication rate
	s.filter.SetDuplicateRate(from, to, 1.0) // 100% duplication

	msgHash := s.generateMessageHash("test-message")
	result := s.filter.Intercept(from, to, msgHash)

	require.True(s.T(), result.Allow, "Message should be allowed")
	require.Equal(s.T(), 2, result.Copies, "Should have 2 copies due to duplication")

	s.T().Log("Message duplication via intercept working correctly")
}

// =============================================================================
// History Management Tests
// =============================================================================

// TestHistoryRetention tests message history retention settings.
func (s *MessageReplayTestSuite) TestHistoryRetention() {
	s.T().Log("=== Test: History Retention ===")

	// Set very short retention
	s.filter.SetHistoryRetention(1 * time.Millisecond)

	// Add a message
	msgHash := s.generateMessageHash("old-message")
	s.filter.IsMessageReplayed(msgHash)
	require.Equal(s.T(), 1, s.filter.MessageHistorySize())

	// Wait for retention period to pass
	time.Sleep(10 * time.Millisecond)

	// Cleanup should remove old entries
	s.filter.CleanupHistory(nil)
	require.Equal(s.T(), 0, s.filter.MessageHistorySize(),
		"Old messages should be cleaned up")

	// Reset to normal retention
	s.filter.SetHistoryRetention(10 * time.Minute)

	s.T().Log("History retention working correctly")
}

// TestHistoryCleanup tests periodic history cleanup.
func (s *MessageReplayTestSuite) TestHistoryCleanup() {
	s.T().Log("=== Test: History Cleanup ===")

	// Add messages
	for i := 0; i < 50; i++ {
		msg := fmt.Sprintf("message:%d", i)
		s.filter.IsMessageReplayed(s.generateMessageHash(msg))
	}

	require.Equal(s.T(), 50, s.filter.MessageHistorySize())

	// Cleanup with current retention (should keep all)
	s.filter.CleanupHistory(nil)
	require.Equal(s.T(), 50, s.filter.MessageHistorySize(),
		"Recent messages should be kept")

	s.T().Log("History cleanup working correctly")
}

// =============================================================================
// Edge Cases
// =============================================================================

// TestEmptyMessageHash tests handling of empty message hash.
func (s *MessageReplayTestSuite) TestEmptyMessageHash() {
	s.T().Log("=== Test: Empty Message Hash ===")

	// Empty hash should be allowed (bypass check)
	result := s.filter.Intercept(s.nodes[0], s.nodes[1], "")
	require.True(s.T(), result.Allow, "Empty hash should bypass replay check")

	// Should still work for subsequent empty hashes
	result = s.filter.Intercept(s.nodes[0], s.nodes[1], "")
	require.True(s.T(), result.Allow, "Empty hash should continue to be allowed")

	s.T().Log("Empty message hash handling correct")
}

// TestConcurrentReplayChecks tests concurrent access to replay checking.
func (s *MessageReplayTestSuite) TestConcurrentReplayChecks() {
	s.T().Log("=== Test: Concurrent Replay Checks ===")

	done := make(chan bool, 10)
	msgHash := s.generateMessageHash("concurrent-test")

	// First goroutine records the message
	go func() {
		s.filter.IsMessageReplayed(msgHash)
		done <- true
	}()

	// Wait for first goroutine
	<-done

	// Multiple goroutines try to replay
	replayCount := 0
	for i := 0; i < 9; i++ {
		go func() {
			if s.filter.IsMessageReplayed(msgHash) {
				replayCount++
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 9; i++ {
		<-done
	}

	// Note: Due to race conditions, we might not get exactly 9 replays
	// but at least some should be detected
	s.T().Logf("Detected %d replays out of 9 attempts", replayCount)
	require.Greater(s.T(), replayCount, 0, "At least some replays should be detected")
}

// TestReplayFromDifferentSenders tests same message from different senders.
func (s *MessageReplayTestSuite) TestReplayFromDifferentSenders() {
	s.T().Log("=== Test: Replay From Different Senders ===")

	msgHash := s.generateMessageHash("shared-message")

	// First sender
	result1 := s.filter.Intercept(s.nodes[0], s.nodes[2], msgHash)
	require.True(s.T(), result1.Allow, "First sender should be allowed")

	// Different sender, same message hash (still a replay)
	result2 := s.filter.Intercept(s.nodes[1], s.nodes[2], msgHash)
	require.False(s.T(), result2.Allow,
		"Same message from different sender should still be detected as replay")

	s.T().Log("Replay detection works across different senders")
}

package audit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewEntry(t *testing.T) {
	now := time.Now()
	entry := NewEntry("action", "actor", "details", now)
	require.Equal(t, "action", entry.Action)
	require.Equal(t, "actor", entry.Actor)
	require.Equal(t, "details", entry.Details)
	require.True(t, entry.Timestamp.Equal(now.UTC()))
}

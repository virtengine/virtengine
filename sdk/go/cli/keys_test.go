package cli

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommands(t *testing.T) {
	rootCommands := KeysCmds()
	assert.Assert(t, rootCommands != nil)

	// Commands are registered
	assert.Equal(t, 12, len(rootCommands.Commands()))
}

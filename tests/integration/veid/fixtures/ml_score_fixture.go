package fixtures

import (
	"encoding/hex"
	"time"

	"github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// MLScoreFixture provides deterministic inputs and expected outputs for ML scoring tests.
type MLScoreFixture struct {
	AccountAddress      string
	BlockHeight         int64
	RequestTime         time.Time
	Scopes              []keeper.DecryptedScope
	ExpectedScore       uint32
	ExpectedScopeScores map[string]uint32
	ExpectedModel       string
	ExpectedInputHash   []byte
	ExpectedInputHex    string
}

const (
	fixtureAccountAddress  = "veid1fixture"
	fixtureBlockHeight     = int64(42)
	fixtureTimestamp       = int64(1_700_000_000)
	fixtureModelVersion    = "v1.0.0"
	fixtureScore           = uint32(80)
	fixtureInputHashHex    = "612452e4c2b57dfa230a8b513cca730c7da89afe314358fe8119f91d1a7d4ffa"
	fixtureIDPlaintext     = "fixture-id-document-v1"
	fixtureSelfiePlaintext = "fixture-selfie-v1"
)

var fixtureScopeScores = map[string]uint32{
	"scope-id-001":     78,
	"scope-selfie-001": 84,
}

// DeterministicMLScoreFixture returns a fixture aligned with the stub ML scorer.
func DeterministicMLScoreFixture() MLScoreFixture {
	scopes := []keeper.DecryptedScope{
		*keeper.NewDecryptedScope("scope-id-001", veidtypes.ScopeTypeIDDocument, []byte(fixtureIDPlaintext)),
		*keeper.NewDecryptedScope("scope-selfie-001", veidtypes.ScopeTypeSelfie, []byte(fixtureSelfiePlaintext)),
	}

	inputHash, _ := hex.DecodeString(fixtureInputHashHex)

	return MLScoreFixture{
		AccountAddress:      fixtureAccountAddress,
		BlockHeight:         fixtureBlockHeight,
		RequestTime:         time.Unix(fixtureTimestamp, 0).UTC(),
		Scopes:              scopes,
		ExpectedScore:       fixtureScore,
		ExpectedScopeScores: fixtureScopeScores,
		ExpectedModel:       fixtureModelVersion,
		ExpectedInputHash:   inputHash,
		ExpectedInputHex:    fixtureInputHashHex,
	}
}

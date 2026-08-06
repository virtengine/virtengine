package participant

import (
	"fmt"
	"strings"

	"github.com/virtengine/virtengine/tools/trusted-setup/transcript"
)

func BuildContributionSigningMessage(phase, participantID, publicKey, attestation, inputHash, outputHash string) []byte {
	return []byte(strings.Join([]string{
		"virtengine-trusted-setup",
		phase,
		participantID,
		publicKey,
		attestation,
		inputHash,
		outputHash,
	}, "\n"))
}

func (c *Client) SignContribution(phase string, inputPayload, outputPayload []byte) (string, error) {
	if c == nil || c.Identity == nil {
		return "", fmt.Errorf("participant identity is required")
	}
	inputHash := transcript.HashBytes(inputPayload)
	outputHash := transcript.HashBytes(outputPayload)
	return c.Identity.Sign(BuildContributionSigningMessage(
		phase,
		c.Identity.ID,
		c.Identity.PublicKey,
		c.Attestation,
		inputHash,
		outputHash,
	))
}

func VerifyContributionSignature(phase, participantID, publicKey, attestation, signature string, inputPayload, outputPayload []byte) error {
	if participantID == "" {
		return fmt.Errorf("participant id is required")
	}
	if publicKey == "" {
		return fmt.Errorf("public key is required")
	}
	if signature == "" {
		return fmt.Errorf("signature is required")
	}
	inputHash := transcript.HashBytes(inputPayload)
	outputHash := transcript.HashBytes(outputPayload)
	return VerifySignature(publicKey, BuildContributionSigningMessage(
		phase,
		participantID,
		publicKey,
		attestation,
		inputHash,
		outputHash,
	), signature)
}

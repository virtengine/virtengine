package participant

import (
	"bytes"
	"fmt"

	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"
)

type Client struct {
	Identity    *Identity
	Attestation string
}

func NewClient(identity *Identity, attestation string) *Client {
	return &Client{
		Identity:    identity,
		Attestation: attestation,
	}
}

func (c *Client) ContributePhase1(payload []byte) ([]byte, string, error) {
	phase := new(mpcsetup.Phase1)
	if _, err := phase.ReadFrom(bytes.NewReader(payload)); err != nil {
		return nil, "", fmt.Errorf("read phase1: %w", err)
	}
	phase.Contribute()

	var buf bytes.Buffer
	if _, err := phase.WriteTo(&buf); err != nil {
		return nil, "", fmt.Errorf("serialize phase1: %w", err)
	}
	signature, err := c.SignContribution("phase1", payload, buf.Bytes())
	if err != nil {
		return nil, "", err
	}
	return buf.Bytes(), signature, nil
}

func (c *Client) ContributePhase2(payload []byte) ([]byte, string, error) {
	phase := new(mpcsetup.Phase2)
	if _, err := phase.ReadFrom(bytes.NewReader(payload)); err != nil {
		return nil, "", fmt.Errorf("read phase2: %w", err)
	}
	phase.Contribute()

	var buf bytes.Buffer
	if _, err := phase.WriteTo(&buf); err != nil {
		return nil, "", fmt.Errorf("serialize phase2: %w", err)
	}
	signature, err := c.SignContribution("phase2", payload, buf.Bytes())
	if err != nil {
		return nil, "", err
	}
	return buf.Bytes(), signature, nil
}

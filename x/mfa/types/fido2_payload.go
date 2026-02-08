// Copyright (c) VirtEngine, Inc.
// SPDX-License-Identifier: BSL-1.1

package types

import "encoding/json"

// FIDO2RegistrationPayload carries a WebAuthn registration response.
type FIDO2RegistrationPayload struct {
	// ChallengeID is the on-chain challenge ID issued for registration.
	ChallengeID string `json:"challenge_id"`

	// ClientDataJSON is the raw clientDataJSON from the browser.
	ClientDataJSON []byte `json:"client_data_json"`

	// AttestationObject is the raw attestationObject from the browser.
	AttestationObject []byte `json:"attestation_object"`

	// Transports lists authenticator transports.
	Transports []string `json:"transports,omitempty"`
}

// Validate validates the registration payload.
func (p *FIDO2RegistrationPayload) Validate() error {
	if p.ChallengeID == "" {
		return ErrInvalidChallengeResponse.Wrap("challenge_id cannot be empty")
	}
	if len(p.ClientDataJSON) == 0 {
		return ErrInvalidChallengeResponse.Wrap("client_data_json cannot be empty")
	}
	if len(p.AttestationObject) == 0 {
		return ErrInvalidChallengeResponse.Wrap("attestation_object cannot be empty")
	}
	return nil
}

// ParseFIDO2RegistrationPayload parses a JSON-encoded FIDO2RegistrationPayload.
func ParseFIDO2RegistrationPayload(data []byte) (*FIDO2RegistrationPayload, error) {
	if len(data) == 0 {
		return nil, ErrInvalidChallengeResponse.Wrap("empty registration payload")
	}

	var payload FIDO2RegistrationPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, ErrInvalidChallengeResponse.Wrapf("invalid registration payload: %v", err)
	}

	if err := payload.Validate(); err != nil {
		return nil, err
	}

	return &payload, nil
}

// FIDO2AssertionPayload carries a WebAuthn assertion response.
type FIDO2AssertionPayload struct {
	// CredentialID is the credential ID used for assertion.
	CredentialID []byte `json:"credential_id"`

	// ClientDataJSON is the raw clientDataJSON from the browser.
	ClientDataJSON []byte `json:"client_data_json"`

	// AuthenticatorData is the raw authenticatorData from the browser.
	AuthenticatorData []byte `json:"authenticator_data"`

	// Signature is the assertion signature.
	Signature []byte `json:"signature"`

	// UserHandle is the optional user handle.
	UserHandle []byte `json:"user_handle,omitempty"`
}

// Validate validates the assertion payload.
func (p *FIDO2AssertionPayload) Validate() error {
	if len(p.CredentialID) == 0 {
		return ErrInvalidChallengeResponse.Wrap("credential_id cannot be empty")
	}
	if len(p.ClientDataJSON) == 0 {
		return ErrInvalidChallengeResponse.Wrap("client_data_json cannot be empty")
	}
	if len(p.AuthenticatorData) == 0 {
		return ErrInvalidChallengeResponse.Wrap("authenticator_data cannot be empty")
	}
	if len(p.Signature) == 0 {
		return ErrInvalidChallengeResponse.Wrap("signature cannot be empty")
	}
	return nil
}

// ParseFIDO2AssertionPayload parses a JSON-encoded FIDO2AssertionPayload.
func ParseFIDO2AssertionPayload(data []byte) (*FIDO2AssertionPayload, error) {
	if len(data) == 0 {
		return nil, ErrInvalidChallengeResponse.Wrap("empty assertion payload")
	}

	var payload FIDO2AssertionPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, ErrInvalidChallengeResponse.Wrapf("invalid assertion payload: %v", err)
	}

	if err := payload.Validate(); err != nil {
		return nil, err
	}

	return &payload, nil
}

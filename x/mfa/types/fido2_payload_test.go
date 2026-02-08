// Copyright (c) VirtEngine, Inc.
// SPDX-License-Identifier: BSL-1.1

package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFIDO2RegistrationPayload(t *testing.T) {
	payload := FIDO2RegistrationPayload{
		ChallengeID:       "challenge-123",
		ClientDataJSON:    []byte("client-data"),
		AttestationObject: []byte("attestation"),
		Transports:        []string{"usb"},
	}

	bz, err := json.Marshal(payload)
	require.NoError(t, err)

	parsed, err := ParseFIDO2RegistrationPayload(bz)
	require.NoError(t, err)
	require.Equal(t, payload.ChallengeID, parsed.ChallengeID)
	require.Equal(t, payload.ClientDataJSON, parsed.ClientDataJSON)
	require.Equal(t, payload.AttestationObject, parsed.AttestationObject)
	require.Equal(t, payload.Transports, parsed.Transports)
}

func TestParseFIDO2RegistrationPayload_Invalid(t *testing.T) {
	_, err := ParseFIDO2RegistrationPayload(nil)
	require.Error(t, err)

	_, err = ParseFIDO2RegistrationPayload([]byte("{invalid-json"))
	require.Error(t, err)

	_, err = ParseFIDO2RegistrationPayload([]byte(`{"challenge_id":""}`))
	require.Error(t, err)
}

func TestParseFIDO2AssertionPayload(t *testing.T) {
	payload := FIDO2AssertionPayload{
		CredentialID:      []byte("cred-id"),
		ClientDataJSON:    []byte("client-data"),
		AuthenticatorData: []byte("auth-data"),
		Signature:         []byte("sig"),
		UserHandle:        []byte("user"),
	}

	bz, err := json.Marshal(payload)
	require.NoError(t, err)

	parsed, err := ParseFIDO2AssertionPayload(bz)
	require.NoError(t, err)
	require.Equal(t, payload.CredentialID, parsed.CredentialID)
	require.Equal(t, payload.ClientDataJSON, parsed.ClientDataJSON)
	require.Equal(t, payload.AuthenticatorData, parsed.AuthenticatorData)
	require.Equal(t, payload.Signature, parsed.Signature)
	require.Equal(t, payload.UserHandle, parsed.UserHandle)
}

func TestParseFIDO2AssertionPayload_Invalid(t *testing.T) {
	_, err := ParseFIDO2AssertionPayload(nil)
	require.Error(t, err)

	_, err = ParseFIDO2AssertionPayload([]byte("{invalid-json"))
	require.Error(t, err)

	_, err = ParseFIDO2AssertionPayload([]byte(`{"credential_id":""}`))
	require.Error(t, err)
}

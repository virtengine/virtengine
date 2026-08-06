// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

func TestValidateGenesisRejectsOrphanAndInvalidEpochLineage(t *testing.T) {
	owner := sdk.AccAddress([]byte("provider-genesis-key")).String()
	publicOne, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	publicTwo, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	first := types.NewProviderPublicKeyRecord(publicOne, types.PublicKeyTypeEd25519, 10)
	first.ActivatedAtUnix = time.Unix(1_700_000_000, 0).Unix()
	second := types.NewProviderPublicKeyRecord(publicTwo, types.PublicKeyTypeEd25519, 20)
	second.Epoch = 2
	second.ActivatedAtUnix = time.Unix(1_700_000_100, 0).Unix()
	second.PreviousKeyID = "wrong-predecessor"

	orphan := &types.GenesisState{SigningKeys: []types.ProviderSigningKeyGenesisRecord{{Owner: owner, Key: ProviderKeyRecordToProto(first), Current: true}}}
	require.Error(t, ValidateGenesis(orphan))

	providerRecord := types.Provider{Owner: owner, HostURI: "https://provider.example.com"}
	brokenLineage := &types.GenesisState{
		Providers: types.Providers{providerRecord},
		SigningKeys: []types.ProviderSigningKeyGenesisRecord{
			{Owner: owner, Key: ProviderKeyRecordToProto(first)},
			{Owner: owner, Key: ProviderKeyRecordToProto(second), Current: true},
		},
	}
	require.Error(t, ValidateGenesis(brokenLineage))
}

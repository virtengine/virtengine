package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	taketype "github.com/virtengine/virtengine/sdk/go/node/take/v1"
)

func TestDefaultParams(t *testing.T) {
	params := taketype.DefaultParams()

	require.Equal(t, uint32(20), params.DefaultTakeRate)
	require.Len(t, params.DenomTakeRates, 1)
	require.Equal(t, "uve", params.DenomTakeRates[0].Denom)
	require.Equal(t, uint32(2), params.DenomTakeRates[0].Rate)
}

func TestParamsValidate(t *testing.T) {
	tests := []struct {
		name      string
		params    taketype.Params
		expectErr bool
		errMsg    string
	}{
		{
			name:      "default params valid",
			params:    taketype.DefaultParams(),
			expectErr: false,
		},
		{
			name: "default rate too high",
			params: taketype.Params{
				DefaultTakeRate: 120,
				DenomTakeRates: taketype.DenomTakeRates{
					{Denom: "uve", Rate: 2},
				},
			},
			expectErr: true,
			errMsg:    "invalid Take Rate",
		},
		{
			name: "denom rate too high",
			params: taketype.Params{
				DefaultTakeRate: 20,
				DenomTakeRates: taketype.DenomTakeRates{
					{Denom: "uve", Rate: 200},
				},
			},
			expectErr: true,
			errMsg:    "invalid Denom Take Rate",
		},
		{
			name: "missing uve denom",
			params: taketype.Params{
				DefaultTakeRate: 20,
				DenomTakeRates: taketype.DenomTakeRates{
					{Denom: "ufoo", Rate: 2},
				},
			},
			expectErr: true,
			errMsg:    "uve must be present",
		},
		{
			name: "duplicate denom",
			params: taketype.Params{
				DefaultTakeRate: 20,
				DenomTakeRates: taketype.DenomTakeRates{
					{Denom: "uve", Rate: 2},
					{Denom: "uve", Rate: 3},
				},
			},
			expectErr: true,
			errMsg:    "duplicate Denom Take Rate",
		},
		{
			name: "zero default rate valid",
			params: taketype.Params{
				DefaultTakeRate: 0,
				DenomTakeRates: taketype.DenomTakeRates{
					{Denom: "uve", Rate: 0},
				},
			},
			expectErr: false,
		},
		{
			name: "max valid default rate",
			params: taketype.Params{
				DefaultTakeRate: 100,
				DenomTakeRates: taketype.DenomTakeRates{
					{Denom: "uve", Rate: 100},
				},
			},
			expectErr: false,
		},
		{
			name: "empty denom rates",
			params: taketype.Params{
				DefaultTakeRate: 20,
				DenomTakeRates:  taketype.DenomTakeRates{},
			},
			expectErr: true,
			errMsg:    "uve must be present",
		},
		{
			name: "multiple valid denoms",
			params: taketype.Params{
				DefaultTakeRate: 20,
				DenomTakeRates: taketype.DenomTakeRates{
					{Denom: "uve", Rate: 2},
					{Denom: "ufoo", Rate: 5},
					{Denom: "ubar", Rate: 10},
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}

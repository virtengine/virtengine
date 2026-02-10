package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
)

const (
	DefaultHeartbeatTimeoutSeconds   uint64 = 120
	DefaultReservationTimeoutSeconds uint64 = 300
	DefaultMaxCandidates             uint64 = 25
	DefaultLocalityWeight                   = "600000"
	DefaultCapacityWeight                   = "400000"
	DefaultSlashingGraceSeconds      uint64 = 60
	DefaultSlashingPenalty                  = "0.01"
)

// Params defines the module parameters.
type Params struct {
	HeartbeatTimeoutSeconds   uint64
	ReservationTimeoutSeconds uint64
	MaxCandidates             uint64
	LocalityWeight            string
	CapacityWeight            string
	SlashingGraceSeconds      uint64
	SlashingPenalty           string
}

var (
	KeyHeartbeatTimeoutSeconds   = []byte("HeartbeatTimeoutSeconds")
	KeyReservationTimeoutSeconds = []byte("ReservationTimeoutSeconds")
	KeyMaxCandidates             = []byte("MaxCandidates")
	KeyLocalityWeight            = []byte("LocalityWeight")
	KeyCapacityWeight            = []byte("CapacityWeight")
	KeySlashingGraceSeconds      = []byte("SlashingGraceSeconds")
	KeySlashingPenalty           = []byte("SlashingPenalty")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns module default params.
func DefaultParams() Params {
	return Params{
		HeartbeatTimeoutSeconds:   DefaultHeartbeatTimeoutSeconds,
		ReservationTimeoutSeconds: DefaultReservationTimeoutSeconds,
		MaxCandidates:             DefaultMaxCandidates,
		LocalityWeight:            DefaultLocalityWeight,
		CapacityWeight:            DefaultCapacityWeight,
		SlashingGraceSeconds:      DefaultSlashingGraceSeconds,
		SlashingPenalty:           DefaultSlashingPenalty,
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyHeartbeatTimeoutSeconds, &p.HeartbeatTimeoutSeconds, validatePositiveUint64),
		paramtypes.NewParamSetPair(KeyReservationTimeoutSeconds, &p.ReservationTimeoutSeconds, validatePositiveUint64),
		paramtypes.NewParamSetPair(KeyMaxCandidates, &p.MaxCandidates, validatePositiveUint64),
		paramtypes.NewParamSetPair(KeyLocalityWeight, &p.LocalityWeight, validateWeightString),
		paramtypes.NewParamSetPair(KeyCapacityWeight, &p.CapacityWeight, validateWeightString),
		paramtypes.NewParamSetPair(KeySlashingGraceSeconds, &p.SlashingGraceSeconds, validatePositiveUint64),
		paramtypes.NewParamSetPair(KeySlashingPenalty, &p.SlashingPenalty, validatePenaltyString),
	}
}

// Validate validates params.
func (p Params) Validate() error {
	if p.HeartbeatTimeoutSeconds == 0 {
		return fmt.Errorf("heartbeat timeout must be positive")
	}
	if p.ReservationTimeoutSeconds == 0 {
		return fmt.Errorf("reservation timeout must be positive")
	}
	if p.MaxCandidates == 0 {
		return fmt.Errorf("max candidates must be positive")
	}
	if err := validateWeightString(p.LocalityWeight); err != nil {
		return err
	}
	if err := validateWeightString(p.CapacityWeight); err != nil {
		return err
	}
	if p.SlashingGraceSeconds == 0 {
		return fmt.Errorf("slashing grace seconds must be positive")
	}
	if err := validatePenaltyString(p.SlashingPenalty); err != nil {
		return err
	}
	return nil
}

func validatePositiveUint64(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type")
	}
	if v == 0 {
		return fmt.Errorf("value must be positive")
	}
	return nil
}

func validateWeightString(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid weight type")
	}
	if v == "" {
		return fmt.Errorf("weight cannot be empty")
	}
	return nil
}

func validatePenaltyString(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid penalty type")
	}
	if v == "" {
		return fmt.Errorf("penalty cannot be empty")
	}
	return nil
}

// ParamsFromProto converts a proto params into local params.
func ParamsFromProto(p resourcesv1.Params) Params {
	return Params{
		HeartbeatTimeoutSeconds:   p.HeartbeatTimeoutSeconds,
		ReservationTimeoutSeconds: p.ReservationTimeoutSeconds,
		MaxCandidates:             p.MaxCandidates,
		LocalityWeight:            p.LocalityWeight,
		CapacityWeight:            p.CapacityWeight,
		SlashingGraceSeconds:      p.SlashingGraceSeconds,
		SlashingPenalty:           p.SlashingPenalty,
	}
}

// ToProto converts params to proto.
func (p Params) ToProto() resourcesv1.Params {
	return resourcesv1.Params{
		HeartbeatTimeoutSeconds:   p.HeartbeatTimeoutSeconds,
		ReservationTimeoutSeconds: p.ReservationTimeoutSeconds,
		MaxCandidates:             p.MaxCandidates,
		LocalityWeight:            p.LocalityWeight,
		CapacityWeight:            p.CapacityWeight,
		SlashingGraceSeconds:      p.SlashingGraceSeconds,
		SlashingPenalty:           p.SlashingPenalty,
	}
}

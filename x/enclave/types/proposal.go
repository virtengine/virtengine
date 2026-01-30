package types

import (
	"encoding/hex"
	"fmt"

	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeAddMeasurement adds a measurement to the allowlist.
	ProposalTypeAddMeasurement = "AddMeasurement"
	// ProposalTypeRevokeMeasurement revokes a measurement from the allowlist.
	ProposalTypeRevokeMeasurement = "RevokeMeasurement"
)

func init() {
	govv1beta1.RegisterProposalType(ProposalTypeAddMeasurement)
	govv1beta1.RegisterProposalType(ProposalTypeRevokeMeasurement)
}

// AddMeasurementProposal defines a governance proposal to add an enclave measurement.
type AddMeasurementProposal struct {
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	MeasurementHash []byte  `json:"measurement_hash"`
	TEEType         TEEType `json:"tee_type"`
	MinISVSVN       uint16  `json:"min_isv_svn"`
	ExpiryBlocks    int64   `json:"expiry_blocks,omitempty"`
}

// ProtoMessage implements proto.Message
func (*AddMeasurementProposal) ProtoMessage() {}

// Reset implements proto.Message
func (p *AddMeasurementProposal) Reset() { *p = AddMeasurementProposal{} }

// NewAddMeasurementProposal creates a new AddMeasurementProposal.
func NewAddMeasurementProposal(title, description string, measurementHash []byte, teeType TEEType, minISVSVN uint16, expiryBlocks int64) *AddMeasurementProposal {
	return &AddMeasurementProposal{
		Title:           title,
		Description:     description,
		MeasurementHash: measurementHash,
		TEEType:         teeType,
		MinISVSVN:       minISVSVN,
		ExpiryBlocks:    expiryBlocks,
	}
}

// GetTitle returns the proposal title.
func (p *AddMeasurementProposal) GetTitle() string {
	return p.Title
}

// GetDescription returns the proposal description.
func (p *AddMeasurementProposal) GetDescription() string {
	return p.Description
}

// ProposalRoute returns the proposal route.
func (p *AddMeasurementProposal) ProposalRoute() string {
	return RouterKey
}

// ProposalType returns the proposal type.
func (p *AddMeasurementProposal) ProposalType() string {
	return ProposalTypeAddMeasurement
}

// ValidateBasic validates the proposal.
func (p *AddMeasurementProposal) ValidateBasic() error {
	if err := govv1beta1.ValidateAbstract(p); err != nil {
		return err
	}

	if len(p.MeasurementHash) != 32 {
		return ErrInvalidMeasurement.Wrapf("measurement hash must be 32 bytes, got %d", len(p.MeasurementHash))
	}

	if !IsValidTEEType(p.TEEType) {
		return ErrInvalidMeasurement.Wrapf("invalid TEE type: %s", p.TEEType)
	}

	if p.ExpiryBlocks < 0 {
		return ErrInvalidMeasurement.Wrap("expiry blocks cannot be negative")
	}

	return nil
}

// String returns a string representation of the proposal.
func (p *AddMeasurementProposal) String() string {
	hash := ""
	if len(p.MeasurementHash) > 0 {
		hash = hex.EncodeToString(p.MeasurementHash)
	}
	return fmt.Sprintf("AddMeasurementProposal{title:%q hash:%s tee:%s}", p.Title, hash, p.TEEType)
}

// RevokeMeasurementProposal defines a governance proposal to revoke a measurement.
type RevokeMeasurementProposal struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	MeasurementHash []byte `json:"measurement_hash"`
	Reason          string `json:"reason"`
}

// ProtoMessage implements proto.Message
func (*RevokeMeasurementProposal) ProtoMessage() {}

// Reset implements proto.Message
func (p *RevokeMeasurementProposal) Reset() { *p = RevokeMeasurementProposal{} }

// NewRevokeMeasurementProposal creates a new RevokeMeasurementProposal.
func NewRevokeMeasurementProposal(title, description string, measurementHash []byte, reason string) *RevokeMeasurementProposal {
	return &RevokeMeasurementProposal{
		Title:           title,
		Description:     description,
		MeasurementHash: measurementHash,
		Reason:          reason,
	}
}

// GetTitle returns the proposal title.
func (p *RevokeMeasurementProposal) GetTitle() string {
	return p.Title
}

// GetDescription returns the proposal description.
func (p *RevokeMeasurementProposal) GetDescription() string {
	return p.Description
}

// ProposalRoute returns the proposal route.
func (p *RevokeMeasurementProposal) ProposalRoute() string {
	return RouterKey
}

// ProposalType returns the proposal type.
func (p *RevokeMeasurementProposal) ProposalType() string {
	return ProposalTypeRevokeMeasurement
}

// ValidateBasic validates the proposal.
func (p *RevokeMeasurementProposal) ValidateBasic() error {
	if err := govv1beta1.ValidateAbstract(p); err != nil {
		return err
	}

	if len(p.MeasurementHash) != 32 {
		return ErrInvalidMeasurement.Wrapf("measurement hash must be 32 bytes, got %d", len(p.MeasurementHash))
	}

	if p.Reason == "" {
		return ErrInvalidMeasurement.Wrap("reason cannot be empty")
	}

	return nil
}

// String returns a string representation of the proposal.
func (p *RevokeMeasurementProposal) String() string {
	hash := ""
	if len(p.MeasurementHash) > 0 {
		hash = hex.EncodeToString(p.MeasurementHash)
	}
	return fmt.Sprintf("RevokeMeasurementProposal{title:%q hash:%s}", p.Title, hash)
}

package types

import (
	"fmt"

	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeAddVerifierVersion   = "AddVerifierVersion"
	ProposalTypeActivateVerifier     = "ActivateVerifier"
	ProposalTypeUpdateRegistryParams = "UpdateRegistryParams"
)

func init() {
	govv1beta1.RegisterProposalType(ProposalTypeAddVerifierVersion)
	govv1beta1.RegisterProposalType(ProposalTypeActivateVerifier)
	govv1beta1.RegisterProposalType(ProposalTypeUpdateRegistryParams)
}

type AddVerifierVersionProposal struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Verifier    VerifierVersion `json:"verifier"`
}

func (*AddVerifierVersionProposal) ProtoMessage()            {}
func (p *AddVerifierVersionProposal) Reset()                 { *p = AddVerifierVersionProposal{} }
func (p *AddVerifierVersionProposal) GetTitle() string       { return p.Title }
func (p *AddVerifierVersionProposal) GetDescription() string { return p.Description }
func (p *AddVerifierVersionProposal) ProposalRoute() string  { return RouterKey }
func (p *AddVerifierVersionProposal) ProposalType() string   { return ProposalTypeAddVerifierVersion }
func (p *AddVerifierVersionProposal) ValidateBasic() error {
	if err := govv1beta1.ValidateAbstract(p); err != nil {
		return err
	}
	return p.Verifier.Validate()
}
func (p *AddVerifierVersionProposal) String() string {
	return fmt.Sprintf("AddVerifierVersionProposal{title:%q verifier:%s version:%s}", p.Title, p.Verifier.VerifierID, p.Verifier.SpecVersion)
}

type ActivateVerifierProposal struct {
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Active      ActiveVerifierPointer `json:"active"`
}

func (*ActivateVerifierProposal) ProtoMessage()            {}
func (p *ActivateVerifierProposal) Reset()                 { *p = ActivateVerifierProposal{} }
func (p *ActivateVerifierProposal) GetTitle() string       { return p.Title }
func (p *ActivateVerifierProposal) GetDescription() string { return p.Description }
func (p *ActivateVerifierProposal) ProposalRoute() string  { return RouterKey }
func (p *ActivateVerifierProposal) ProposalType() string   { return ProposalTypeActivateVerifier }
func (p *ActivateVerifierProposal) ValidateBasic() error {
	if err := govv1beta1.ValidateAbstract(p); err != nil {
		return err
	}
	return p.Active.Validate()
}
func (p *ActivateVerifierProposal) String() string {
	return fmt.Sprintf("ActivateVerifierProposal{title:%q verifier:%s version:%s}", p.Title, p.Active.VerifierID, p.Active.SpecVersion)
}

type UpdateParamsProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Params      Params `json:"params"`
}

func (*UpdateParamsProposal) ProtoMessage()            {}
func (p *UpdateParamsProposal) Reset()                 { *p = UpdateParamsProposal{} }
func (p *UpdateParamsProposal) GetTitle() string       { return p.Title }
func (p *UpdateParamsProposal) GetDescription() string { return p.Description }
func (p *UpdateParamsProposal) ProposalRoute() string  { return RouterKey }
func (p *UpdateParamsProposal) ProposalType() string   { return ProposalTypeUpdateRegistryParams }
func (p *UpdateParamsProposal) ValidateBasic() error {
	if err := govv1beta1.ValidateAbstract(p); err != nil {
		return err
	}
	return p.Params.Validate()
}
func (p *UpdateParamsProposal) String() string {
	return fmt.Sprintf("UpdateParamsProposal{title:%q min_ready:%d}", p.Title, p.Params.MinimumReadyValidators)
}

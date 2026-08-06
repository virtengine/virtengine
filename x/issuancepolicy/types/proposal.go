package types

import (
	"fmt"

	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeUpsertPolicy       = "UpsertPolicy"
	ProposalTypeSetActivePolicy    = "SetActivePolicy"
	ProposalTypeUpdatePolicyParams = "UpdatePolicyParams"
)

func init() {
	govv1beta1.RegisterProposalType(ProposalTypeUpsertPolicy)
	govv1beta1.RegisterProposalType(ProposalTypeSetActivePolicy)
	govv1beta1.RegisterProposalType(ProposalTypeUpdatePolicyParams)
}

type UpsertPolicyProposal struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Policy      IssuancePolicy `json:"policy"`
}

func (*UpsertPolicyProposal) ProtoMessage()            {}
func (p *UpsertPolicyProposal) Reset()                 { *p = UpsertPolicyProposal{} }
func (p *UpsertPolicyProposal) GetTitle() string       { return p.Title }
func (p *UpsertPolicyProposal) GetDescription() string { return p.Description }
func (p *UpsertPolicyProposal) ProposalRoute() string  { return RouterKey }
func (p *UpsertPolicyProposal) ProposalType() string   { return ProposalTypeUpsertPolicy }
func (p *UpsertPolicyProposal) ValidateBasic() error {
	if err := govv1beta1.ValidateAbstract(p); err != nil {
		return err
	}
	return p.Policy.Validate()
}
func (p *UpsertPolicyProposal) String() string {
	return fmt.Sprintf("UpsertPolicyProposal{title:%q policy:%s}", p.Title, p.Policy.PolicyID)
}

type SetActivePolicyProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PolicyID    string `json:"policy_id"`
}

func (*SetActivePolicyProposal) ProtoMessage()            {}
func (p *SetActivePolicyProposal) Reset()                 { *p = SetActivePolicyProposal{} }
func (p *SetActivePolicyProposal) GetTitle() string       { return p.Title }
func (p *SetActivePolicyProposal) GetDescription() string { return p.Description }
func (p *SetActivePolicyProposal) ProposalRoute() string  { return RouterKey }
func (p *SetActivePolicyProposal) ProposalType() string   { return ProposalTypeSetActivePolicy }
func (p *SetActivePolicyProposal) ValidateBasic() error {
	if err := govv1beta1.ValidateAbstract(p); err != nil {
		return err
	}
	if p.PolicyID == "" {
		return fmt.Errorf("policy_id cannot be empty")
	}
	return nil
}
func (p *SetActivePolicyProposal) String() string {
	return fmt.Sprintf("SetActivePolicyProposal{title:%q policy:%s}", p.Title, p.PolicyID)
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
func (p *UpdateParamsProposal) ProposalType() string   { return ProposalTypeUpdatePolicyParams }
func (p *UpdateParamsProposal) ValidateBasic() error {
	if err := govv1beta1.ValidateAbstract(p); err != nil {
		return err
	}
	return p.Params.Validate()
}
func (p *UpdateParamsProposal) String() string {
	return fmt.Sprintf("UpdateParamsProposal{title:%q epoch_length:%d}", p.Title, p.Params.EpochLengthBlocks)
}

import { MsgCreateGroup, MsgCreateGroupPolicy, MsgCreateGroupPolicyResponse, MsgCreateGroupResponse, MsgCreateGroupWithPolicy, MsgCreateGroupWithPolicyResponse, MsgExec, MsgExecResponse, MsgLeaveGroup, MsgLeaveGroupResponse, MsgSubmitProposal, MsgSubmitProposalResponse, MsgUpdateGroupAdmin, MsgUpdateGroupAdminResponse, MsgUpdateGroupMembers, MsgUpdateGroupMembersResponse, MsgUpdateGroupMetadata, MsgUpdateGroupMetadataResponse, MsgUpdateGroupPolicyAdmin, MsgUpdateGroupPolicyAdminResponse, MsgUpdateGroupPolicyDecisionPolicy, MsgUpdateGroupPolicyDecisionPolicyResponse, MsgUpdateGroupPolicyMetadata, MsgUpdateGroupPolicyMetadataResponse, MsgVote, MsgVoteResponse, MsgWithdrawProposal, MsgWithdrawProposalResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.group.v1.Msg",
  methods: {
    createGroup: {
      name: "CreateGroup",
      input: MsgCreateGroup,
      output: MsgCreateGroupResponse,
      get parent() { return Msg; },
    },
    updateGroupMembers: {
      name: "UpdateGroupMembers",
      input: MsgUpdateGroupMembers,
      output: MsgUpdateGroupMembersResponse,
      get parent() { return Msg; },
    },
    updateGroupAdmin: {
      name: "UpdateGroupAdmin",
      input: MsgUpdateGroupAdmin,
      output: MsgUpdateGroupAdminResponse,
      get parent() { return Msg; },
    },
    updateGroupMetadata: {
      name: "UpdateGroupMetadata",
      input: MsgUpdateGroupMetadata,
      output: MsgUpdateGroupMetadataResponse,
      get parent() { return Msg; },
    },
    createGroupPolicy: {
      name: "CreateGroupPolicy",
      input: MsgCreateGroupPolicy,
      output: MsgCreateGroupPolicyResponse,
      get parent() { return Msg; },
    },
    createGroupWithPolicy: {
      name: "CreateGroupWithPolicy",
      input: MsgCreateGroupWithPolicy,
      output: MsgCreateGroupWithPolicyResponse,
      get parent() { return Msg; },
    },
    updateGroupPolicyAdmin: {
      name: "UpdateGroupPolicyAdmin",
      input: MsgUpdateGroupPolicyAdmin,
      output: MsgUpdateGroupPolicyAdminResponse,
      get parent() { return Msg; },
    },
    updateGroupPolicyDecisionPolicy: {
      name: "UpdateGroupPolicyDecisionPolicy",
      input: MsgUpdateGroupPolicyDecisionPolicy,
      output: MsgUpdateGroupPolicyDecisionPolicyResponse,
      get parent() { return Msg; },
    },
    updateGroupPolicyMetadata: {
      name: "UpdateGroupPolicyMetadata",
      input: MsgUpdateGroupPolicyMetadata,
      output: MsgUpdateGroupPolicyMetadataResponse,
      get parent() { return Msg; },
    },
    submitProposal: {
      name: "SubmitProposal",
      input: MsgSubmitProposal,
      output: MsgSubmitProposalResponse,
      get parent() { return Msg; },
    },
    withdrawProposal: {
      name: "WithdrawProposal",
      input: MsgWithdrawProposal,
      output: MsgWithdrawProposalResponse,
      get parent() { return Msg; },
    },
    vote: {
      name: "Vote",
      input: MsgVote,
      output: MsgVoteResponse,
      get parent() { return Msg; },
    },
    exec: {
      name: "Exec",
      input: MsgExec,
      output: MsgExecResponse,
      get parent() { return Msg; },
    },
    leaveGroup: {
      name: "LeaveGroup",
      input: MsgLeaveGroup,
      output: MsgLeaveGroupResponse,
      get parent() { return Msg; },
    },
  },
} as const;

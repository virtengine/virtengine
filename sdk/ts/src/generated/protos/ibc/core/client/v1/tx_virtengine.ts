import { MsgCreateClient, MsgCreateClientResponse, MsgDeleteClientCreator, MsgDeleteClientCreatorResponse, MsgIBCSoftwareUpgrade, MsgIBCSoftwareUpgradeResponse, MsgRecoverClient, MsgRecoverClientResponse, MsgSubmitMisbehaviour, MsgSubmitMisbehaviourResponse, MsgUpdateClient, MsgUpdateClientResponse, MsgUpdateParams, MsgUpdateParamsResponse, MsgUpgradeClient, MsgUpgradeClientResponse } from "./tx.ts";

export const Msg = {
  typeName: "ibc.core.client.v1.Msg",
  methods: {
    createClient: {
      name: "CreateClient",
      input: MsgCreateClient,
      output: MsgCreateClientResponse,
      get parent() { return Msg; },
    },
    updateClient: {
      name: "UpdateClient",
      input: MsgUpdateClient,
      output: MsgUpdateClientResponse,
      get parent() { return Msg; },
    },
    upgradeClient: {
      name: "UpgradeClient",
      input: MsgUpgradeClient,
      output: MsgUpgradeClientResponse,
      get parent() { return Msg; },
    },
    submitMisbehaviour: {
      name: "SubmitMisbehaviour",
      input: MsgSubmitMisbehaviour,
      output: MsgSubmitMisbehaviourResponse,
      get parent() { return Msg; },
    },
    recoverClient: {
      name: "RecoverClient",
      input: MsgRecoverClient,
      output: MsgRecoverClientResponse,
      get parent() { return Msg; },
    },
    iBCSoftwareUpgrade: {
      name: "IBCSoftwareUpgrade",
      input: MsgIBCSoftwareUpgrade,
      output: MsgIBCSoftwareUpgradeResponse,
      get parent() { return Msg; },
    },
    updateClientParams: {
      name: "UpdateClientParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return Msg; },
    },
    deleteClientCreator: {
      name: "DeleteClientCreator",
      input: MsgDeleteClientCreator,
      output: MsgDeleteClientCreatorResponse,
      get parent() { return Msg; },
    },
  },
} as const;

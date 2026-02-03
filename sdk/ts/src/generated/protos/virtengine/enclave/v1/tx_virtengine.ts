import { MsgProposeMeasurement, MsgProposeMeasurementResponse, MsgRegisterEnclaveIdentity, MsgRegisterEnclaveIdentityResponse, MsgRevokeMeasurement, MsgRevokeMeasurementResponse, MsgRotateEnclaveIdentity, MsgRotateEnclaveIdentityResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.enclave.v1.Msg",
  methods: {
    registerEnclaveIdentity: {
      name: "RegisterEnclaveIdentity",
      input: MsgRegisterEnclaveIdentity,
      output: MsgRegisterEnclaveIdentityResponse,
      get parent() { return Msg; },
    },
    rotateEnclaveIdentity: {
      name: "RotateEnclaveIdentity",
      input: MsgRotateEnclaveIdentity,
      output: MsgRotateEnclaveIdentityResponse,
      get parent() { return Msg; },
    },
    proposeMeasurement: {
      name: "ProposeMeasurement",
      input: MsgProposeMeasurement,
      output: MsgProposeMeasurementResponse,
      get parent() { return Msg; },
    },
    revokeMeasurement: {
      name: "RevokeMeasurement",
      input: MsgRevokeMeasurement,
      output: MsgRevokeMeasurementResponse,
      get parent() { return Msg; },
    },
    updateParams: {
      name: "UpdateParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return Msg; },
    },
  },
} as const;

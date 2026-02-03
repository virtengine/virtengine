import { MsgAddTrustedDevice, MsgAddTrustedDeviceResponse, MsgCreateChallenge, MsgCreateChallengeResponse, MsgEnrollFactor, MsgEnrollFactorResponse, MsgRemoveTrustedDevice, MsgRemoveTrustedDeviceResponse, MsgRevokeFactor, MsgRevokeFactorResponse, MsgSetMFAPolicy, MsgSetMFAPolicyResponse, MsgUpdateParams, MsgUpdateParamsResponse, MsgUpdateSensitiveTxConfig, MsgUpdateSensitiveTxConfigResponse, MsgVerifyChallenge, MsgVerifyChallengeResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.mfa.v1.Msg",
  methods: {
    enrollFactor: {
      name: "EnrollFactor",
      input: MsgEnrollFactor,
      output: MsgEnrollFactorResponse,
      get parent() { return Msg; },
    },
    revokeFactor: {
      name: "RevokeFactor",
      input: MsgRevokeFactor,
      output: MsgRevokeFactorResponse,
      get parent() { return Msg; },
    },
    setMFAPolicy: {
      name: "SetMFAPolicy",
      input: MsgSetMFAPolicy,
      output: MsgSetMFAPolicyResponse,
      get parent() { return Msg; },
    },
    createChallenge: {
      name: "CreateChallenge",
      input: MsgCreateChallenge,
      output: MsgCreateChallengeResponse,
      get parent() { return Msg; },
    },
    verifyChallenge: {
      name: "VerifyChallenge",
      input: MsgVerifyChallenge,
      output: MsgVerifyChallengeResponse,
      get parent() { return Msg; },
    },
    addTrustedDevice: {
      name: "AddTrustedDevice",
      input: MsgAddTrustedDevice,
      output: MsgAddTrustedDeviceResponse,
      get parent() { return Msg; },
    },
    removeTrustedDevice: {
      name: "RemoveTrustedDevice",
      input: MsgRemoveTrustedDevice,
      output: MsgRemoveTrustedDeviceResponse,
      get parent() { return Msg; },
    },
    updateSensitiveTxConfig: {
      name: "UpdateSensitiveTxConfig",
      input: MsgUpdateSensitiveTxConfig,
      output: MsgUpdateSensitiveTxConfigResponse,
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

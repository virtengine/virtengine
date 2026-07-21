import { MsgRegisterRecipientKey, MsgRegisterRecipientKeyResponse, MsgRevokeRecipientKey, MsgRevokeRecipientKeyResponse, MsgRotateKey, MsgRotateKeyResponse, MsgUpdateKeyLabel, MsgUpdateKeyLabelResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.encryption.v1.Msg",
  methods: {
    registerRecipientKey: {
      name: "RegisterRecipientKey",
      input: MsgRegisterRecipientKey,
      output: MsgRegisterRecipientKeyResponse,
      get parent() { return Msg; },
    },
    revokeRecipientKey: {
      name: "RevokeRecipientKey",
      input: MsgRevokeRecipientKey,
      output: MsgRevokeRecipientKeyResponse,
      get parent() { return Msg; },
    },
    updateKeyLabel: {
      name: "UpdateKeyLabel",
      input: MsgUpdateKeyLabel,
      output: MsgUpdateKeyLabelResponse,
      get parent() { return Msg; },
    },
    rotateKey: {
      name: "RotateKey",
      input: MsgRotateKey,
      output: MsgRotateKeyResponse,
      get parent() { return Msg; },
    },
  },
} as const;

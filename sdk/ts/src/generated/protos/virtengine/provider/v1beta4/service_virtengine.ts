import { MsgConfirmDomainVerification, MsgConfirmDomainVerificationResponse, MsgCreateProvider, MsgCreateProviderResponse, MsgDeleteProvider, MsgDeleteProviderResponse, MsgGenerateDomainVerificationToken, MsgGenerateDomainVerificationTokenResponse, MsgRequestDomainVerification, MsgRequestDomainVerificationResponse, MsgRevokeDomainVerification, MsgRevokeDomainVerificationResponse, MsgRevokeProviderSigningKey, MsgRevokeProviderSigningKeyResponse, MsgRotateProviderSigningKey, MsgRotateProviderSigningKeyResponse, MsgSetProviderSigningKey, MsgSetProviderSigningKeyResponse, MsgUpdateProvider, MsgUpdateProviderResponse, MsgVerifyProviderDomain, MsgVerifyProviderDomainResponse } from "./msg.ts";

export const Msg = {
  typeName: "virtengine.provider.v1beta4.Msg",
  methods: {
    createProvider: {
      name: "CreateProvider",
      input: MsgCreateProvider,
      output: MsgCreateProviderResponse,
      get parent() { return Msg; },
    },
    updateProvider: {
      name: "UpdateProvider",
      input: MsgUpdateProvider,
      output: MsgUpdateProviderResponse,
      get parent() { return Msg; },
    },
    deleteProvider: {
      name: "DeleteProvider",
      input: MsgDeleteProvider,
      output: MsgDeleteProviderResponse,
      get parent() { return Msg; },
    },
    generateDomainVerificationToken: {
      name: "GenerateDomainVerificationToken",
      input: MsgGenerateDomainVerificationToken,
      output: MsgGenerateDomainVerificationTokenResponse,
      get parent() { return Msg; },
    },
    verifyProviderDomain: {
      name: "VerifyProviderDomain",
      input: MsgVerifyProviderDomain,
      output: MsgVerifyProviderDomainResponse,
      get parent() { return Msg; },
    },
    requestDomainVerification: {
      name: "RequestDomainVerification",
      input: MsgRequestDomainVerification,
      output: MsgRequestDomainVerificationResponse,
      get parent() { return Msg; },
    },
    confirmDomainVerification: {
      name: "ConfirmDomainVerification",
      input: MsgConfirmDomainVerification,
      output: MsgConfirmDomainVerificationResponse,
      get parent() { return Msg; },
    },
    revokeDomainVerification: {
      name: "RevokeDomainVerification",
      input: MsgRevokeDomainVerification,
      output: MsgRevokeDomainVerificationResponse,
      get parent() { return Msg; },
    },
    setProviderSigningKey: {
      name: "SetProviderSigningKey",
      input: MsgSetProviderSigningKey,
      output: MsgSetProviderSigningKeyResponse,
      get parent() { return Msg; },
    },
    rotateProviderSigningKey: {
      name: "RotateProviderSigningKey",
      input: MsgRotateProviderSigningKey,
      output: MsgRotateProviderSigningKeyResponse,
      get parent() { return Msg; },
    },
    revokeProviderSigningKey: {
      name: "RevokeProviderSigningKey",
      input: MsgRevokeProviderSigningKey,
      output: MsgRevokeProviderSigningKeyResponse,
      get parent() { return Msg; },
    },
  },
} as const;

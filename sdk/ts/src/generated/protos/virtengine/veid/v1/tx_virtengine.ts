import { MsgAddScopeToWallet, MsgAddScopeToWalletResponse, MsgCompleteBorderlineFallback, MsgCompleteBorderlineFallbackResponse, MsgCreateIdentityWallet, MsgCreateIdentityWalletResponse, MsgRebindWallet, MsgRebindWalletResponse, MsgRequestVerification, MsgRequestVerificationResponse, MsgRevokeScope, MsgRevokeScopeFromWallet, MsgRevokeScopeFromWalletResponse, MsgRevokeScopeResponse, MsgUpdateBorderlineParams, MsgUpdateBorderlineParamsResponse, MsgUpdateConsentSettings, MsgUpdateConsentSettingsResponse, MsgUpdateDerivedFeatures, MsgUpdateDerivedFeaturesResponse, MsgUpdateParams, MsgUpdateParamsResponse, MsgUpdateScore, MsgUpdateScoreResponse, MsgUpdateVerificationStatus, MsgUpdateVerificationStatusResponse, MsgUploadScope, MsgUploadScopeResponse } from "./tx.ts";
import { MsgClaimAppeal, MsgClaimAppealResponse, MsgResolveAppeal, MsgResolveAppealResponse, MsgSubmitAppeal, MsgSubmitAppealResponse, MsgWithdrawAppeal, MsgWithdrawAppealResponse } from "./appeal.ts";
import { MsgAttestCompliance, MsgAttestComplianceResponse, MsgDeactivateComplianceProvider, MsgDeactivateComplianceProviderResponse, MsgRegisterComplianceProvider, MsgRegisterComplianceProviderResponse, MsgSubmitComplianceCheck, MsgSubmitComplianceCheckResponse, MsgUpdateComplianceParams, MsgUpdateComplianceParamsResponse } from "./compliance.ts";
import { MsgActivateModel, MsgActivateModelResponse, MsgDeprecateModel, MsgDeprecateModelResponse, MsgProposeModelUpdate, MsgProposeModelUpdateResponse, MsgRegisterModel, MsgRegisterModelResponse, MsgReportModelVersion, MsgReportModelVersionResponse, MsgRevokeModel, MsgRevokeModelResponse } from "./model.ts";

export const Msg = {
  typeName: "virtengine.veid.v1.Msg",
  methods: {
    uploadScope: {
      name: "UploadScope",
      input: MsgUploadScope,
      output: MsgUploadScopeResponse,
      get parent() { return Msg; },
    },
    revokeScope: {
      name: "RevokeScope",
      input: MsgRevokeScope,
      output: MsgRevokeScopeResponse,
      get parent() { return Msg; },
    },
    requestVerification: {
      name: "RequestVerification",
      input: MsgRequestVerification,
      output: MsgRequestVerificationResponse,
      get parent() { return Msg; },
    },
    updateVerificationStatus: {
      name: "UpdateVerificationStatus",
      input: MsgUpdateVerificationStatus,
      output: MsgUpdateVerificationStatusResponse,
      get parent() { return Msg; },
    },
    updateScore: {
      name: "UpdateScore",
      input: MsgUpdateScore,
      output: MsgUpdateScoreResponse,
      get parent() { return Msg; },
    },
    createIdentityWallet: {
      name: "CreateIdentityWallet",
      input: MsgCreateIdentityWallet,
      output: MsgCreateIdentityWalletResponse,
      get parent() { return Msg; },
    },
    addScopeToWallet: {
      name: "AddScopeToWallet",
      input: MsgAddScopeToWallet,
      output: MsgAddScopeToWalletResponse,
      get parent() { return Msg; },
    },
    revokeScopeFromWallet: {
      name: "RevokeScopeFromWallet",
      input: MsgRevokeScopeFromWallet,
      output: MsgRevokeScopeFromWalletResponse,
      get parent() { return Msg; },
    },
    updateConsentSettings: {
      name: "UpdateConsentSettings",
      input: MsgUpdateConsentSettings,
      output: MsgUpdateConsentSettingsResponse,
      get parent() { return Msg; },
    },
    rebindWallet: {
      name: "RebindWallet",
      input: MsgRebindWallet,
      output: MsgRebindWalletResponse,
      get parent() { return Msg; },
    },
    updateDerivedFeatures: {
      name: "UpdateDerivedFeatures",
      input: MsgUpdateDerivedFeatures,
      output: MsgUpdateDerivedFeaturesResponse,
      get parent() { return Msg; },
    },
    completeBorderlineFallback: {
      name: "CompleteBorderlineFallback",
      input: MsgCompleteBorderlineFallback,
      output: MsgCompleteBorderlineFallbackResponse,
      get parent() { return Msg; },
    },
    updateBorderlineParams: {
      name: "UpdateBorderlineParams",
      input: MsgUpdateBorderlineParams,
      output: MsgUpdateBorderlineParamsResponse,
      get parent() { return Msg; },
    },
    updateParams: {
      name: "UpdateParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return Msg; },
    },
    submitAppeal: {
      name: "SubmitAppeal",
      input: MsgSubmitAppeal,
      output: MsgSubmitAppealResponse,
      get parent() { return Msg; },
    },
    claimAppeal: {
      name: "ClaimAppeal",
      input: MsgClaimAppeal,
      output: MsgClaimAppealResponse,
      get parent() { return Msg; },
    },
    resolveAppeal: {
      name: "ResolveAppeal",
      input: MsgResolveAppeal,
      output: MsgResolveAppealResponse,
      get parent() { return Msg; },
    },
    withdrawAppeal: {
      name: "WithdrawAppeal",
      input: MsgWithdrawAppeal,
      output: MsgWithdrawAppealResponse,
      get parent() { return Msg; },
    },
    submitComplianceCheck: {
      name: "SubmitComplianceCheck",
      input: MsgSubmitComplianceCheck,
      output: MsgSubmitComplianceCheckResponse,
      get parent() { return Msg; },
    },
    attestCompliance: {
      name: "AttestCompliance",
      input: MsgAttestCompliance,
      output: MsgAttestComplianceResponse,
      get parent() { return Msg; },
    },
    updateComplianceParams: {
      name: "UpdateComplianceParams",
      input: MsgUpdateComplianceParams,
      output: MsgUpdateComplianceParamsResponse,
      get parent() { return Msg; },
    },
    registerComplianceProvider: {
      name: "RegisterComplianceProvider",
      input: MsgRegisterComplianceProvider,
      output: MsgRegisterComplianceProviderResponse,
      get parent() { return Msg; },
    },
    deactivateComplianceProvider: {
      name: "DeactivateComplianceProvider",
      input: MsgDeactivateComplianceProvider,
      output: MsgDeactivateComplianceProviderResponse,
      get parent() { return Msg; },
    },
    registerModel: {
      name: "RegisterModel",
      input: MsgRegisterModel,
      output: MsgRegisterModelResponse,
      get parent() { return Msg; },
    },
    proposeModelUpdate: {
      name: "ProposeModelUpdate",
      input: MsgProposeModelUpdate,
      output: MsgProposeModelUpdateResponse,
      get parent() { return Msg; },
    },
    reportModelVersion: {
      name: "ReportModelVersion",
      input: MsgReportModelVersion,
      output: MsgReportModelVersionResponse,
      get parent() { return Msg; },
    },
    activateModel: {
      name: "ActivateModel",
      input: MsgActivateModel,
      output: MsgActivateModelResponse,
      get parent() { return Msg; },
    },
    deprecateModel: {
      name: "DeprecateModel",
      input: MsgDeprecateModel,
      output: MsgDeprecateModelResponse,
      get parent() { return Msg; },
    },
    revokeModel: {
      name: "RevokeModel",
      input: MsgRevokeModel,
      output: MsgRevokeModelResponse,
      get parent() { return Msg; },
    },
  },
} as const;

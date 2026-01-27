import { GetAuthnDescriptorRequest, GetAuthnDescriptorResponse, GetChainDescriptorRequest, GetChainDescriptorResponse, GetCodecDescriptorRequest, GetCodecDescriptorResponse, GetConfigurationDescriptorRequest, GetConfigurationDescriptorResponse, GetQueryServicesDescriptorRequest, GetQueryServicesDescriptorResponse, GetTxDescriptorRequest, GetTxDescriptorResponse } from "./reflection.ts";

export const ReflectionService = {
  typeName: "cosmos.base.reflection.v2alpha1.ReflectionService",
  methods: {
    getAuthnDescriptor: {
      name: "GetAuthnDescriptor",
      httpPath: "/cosmos/base/reflection/v1beta1/app_descriptor/authn",
      input: GetAuthnDescriptorRequest,
      output: GetAuthnDescriptorResponse,
      get parent() { return ReflectionService; },
    },
    getChainDescriptor: {
      name: "GetChainDescriptor",
      httpPath: "/cosmos/base/reflection/v1beta1/app_descriptor/chain",
      input: GetChainDescriptorRequest,
      output: GetChainDescriptorResponse,
      get parent() { return ReflectionService; },
    },
    getCodecDescriptor: {
      name: "GetCodecDescriptor",
      httpPath: "/cosmos/base/reflection/v1beta1/app_descriptor/codec",
      input: GetCodecDescriptorRequest,
      output: GetCodecDescriptorResponse,
      get parent() { return ReflectionService; },
    },
    getConfigurationDescriptor: {
      name: "GetConfigurationDescriptor",
      httpPath: "/cosmos/base/reflection/v1beta1/app_descriptor/configuration",
      input: GetConfigurationDescriptorRequest,
      output: GetConfigurationDescriptorResponse,
      get parent() { return ReflectionService; },
    },
    getQueryServicesDescriptor: {
      name: "GetQueryServicesDescriptor",
      httpPath: "/cosmos/base/reflection/v1beta1/app_descriptor/query_services",
      input: GetQueryServicesDescriptorRequest,
      output: GetQueryServicesDescriptorResponse,
      get parent() { return ReflectionService; },
    },
    getTxDescriptor: {
      name: "GetTxDescriptor",
      httpPath: "/cosmos/base/reflection/v1beta1/app_descriptor/tx_descriptor",
      input: GetTxDescriptorRequest,
      output: GetTxDescriptorResponse,
      get parent() { return ReflectionService; },
    },
  },
} as const;

import { MsgCreateProvider, MsgCreateProviderResponse, MsgDeleteProvider, MsgDeleteProviderResponse, MsgUpdateProvider, MsgUpdateProviderResponse } from "./msg.ts";

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
  },
} as const;

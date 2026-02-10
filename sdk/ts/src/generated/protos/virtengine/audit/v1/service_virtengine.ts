import { MsgDeleteProviderAttributes, MsgDeleteProviderAttributesResponse, MsgSignProviderAttributes, MsgSignProviderAttributesResponse } from "./msg.ts";

export const Msg = {
  typeName: "virtengine.audit.v1.Msg",
  methods: {
    signProviderAttributes: {
      name: "SignProviderAttributes",
      input: MsgSignProviderAttributes,
      output: MsgSignProviderAttributesResponse,
      get parent() { return Msg; },
    },
    deleteProviderAttributes: {
      name: "DeleteProviderAttributes",
      input: MsgDeleteProviderAttributes,
      output: MsgDeleteProviderAttributesResponse,
      get parent() { return Msg; },
    },
  },
} as const;

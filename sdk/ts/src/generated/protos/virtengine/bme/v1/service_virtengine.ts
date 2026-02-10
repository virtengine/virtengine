import { MsgBurnACT, MsgBurnACTResponse, MsgBurnMint, MsgBurnMintResponse, MsgMintACT, MsgMintACTResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./msgs.ts";

export const Msg = {
  typeName: "virtengine.bme.v1.Msg",
  methods: {
    updateParams: {
      name: "UpdateParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return Msg; },
    },
    burnMint: {
      name: "BurnMint",
      input: MsgBurnMint,
      output: MsgBurnMintResponse,
      get parent() { return Msg; },
    },
    mintACT: {
      name: "MintACT",
      input: MsgMintACT,
      output: MsgMintACTResponse,
      get parent() { return Msg; },
    },
    burnACT: {
      name: "BurnACT",
      input: MsgBurnACT,
      output: MsgBurnACTResponse,
      get parent() { return Msg; },
    },
  },
} as const;

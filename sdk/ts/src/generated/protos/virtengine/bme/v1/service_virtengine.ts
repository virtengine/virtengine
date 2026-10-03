import { MsgBurnMint, MsgBurnMintResponse, MsgBurnVCC, MsgBurnVCCResponse, MsgMintVCC, MsgMintVCCResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./msgs.ts";

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
    mintVCC: {
      name: "MintVCC",
      input: MsgMintVCC,
      output: MsgMintVCCResponse,
      get parent() { return Msg; },
    },
    burnVCC: {
      name: "BurnVCC",
      input: MsgBurnVCC,
      output: MsgBurnVCCResponse,
      get parent() { return Msg; },
    },
  },
} as const;

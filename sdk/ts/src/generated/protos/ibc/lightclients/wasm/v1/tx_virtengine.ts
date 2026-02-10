import { MsgMigrateContract, MsgMigrateContractResponse, MsgRemoveChecksum, MsgRemoveChecksumResponse, MsgStoreCode, MsgStoreCodeResponse } from "./tx.ts";

export const Msg = {
  typeName: "ibc.lightclients.wasm.v1.Msg",
  methods: {
    storeCode: {
      name: "StoreCode",
      input: MsgStoreCode,
      output: MsgStoreCodeResponse,
      get parent() { return Msg; },
    },
    removeChecksum: {
      name: "RemoveChecksum",
      input: MsgRemoveChecksum,
      output: MsgRemoveChecksumResponse,
      get parent() { return Msg; },
    },
    migrateContract: {
      name: "MigrateContract",
      input: MsgMigrateContract,
      output: MsgMigrateContractResponse,
      get parent() { return Msg; },
    },
  },
} as const;

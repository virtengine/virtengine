import { MsgRecordPerformance, MsgRecordPerformanceResponse, MsgSlashValidator, MsgSlashValidatorResponse, MsgUnjailValidator, MsgUnjailValidatorResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.staking.v1.Msg",
  methods: {
    updateParams: {
      name: "UpdateParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return Msg; },
    },
    slashValidator: {
      name: "SlashValidator",
      input: MsgSlashValidator,
      output: MsgSlashValidatorResponse,
      get parent() { return Msg; },
    },
    unjailValidator: {
      name: "UnjailValidator",
      input: MsgUnjailValidator,
      output: MsgUnjailValidatorResponse,
      get parent() { return Msg; },
    },
    recordPerformance: {
      name: "RecordPerformance",
      input: MsgRecordPerformance,
      output: MsgRecordPerformanceResponse,
      get parent() { return Msg; },
    },
  },
} as const;

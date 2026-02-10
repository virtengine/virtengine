import { MsgSubmitEvidence, MsgSubmitEvidenceResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.evidence.v1beta1.Msg",
  methods: {
    submitEvidence: {
      name: "SubmitEvidence",
      input: MsgSubmitEvidence,
      output: MsgSubmitEvidenceResponse,
      get parent() { return Msg; },
    },
  },
} as const;

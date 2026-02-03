import { MsgDeleteReview, MsgDeleteReviewResponse, MsgSubmitReview, MsgSubmitReviewResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.review.v1.Msg",
  methods: {
    submitReview: {
      name: "SubmitReview",
      input: MsgSubmitReview,
      output: MsgSubmitReviewResponse,
      get parent() { return Msg; },
    },
    deleteReview: {
      name: "DeleteReview",
      input: MsgDeleteReview,
      output: MsgDeleteReviewResponse,
      get parent() { return Msg; },
    },
    updateParams: {
      name: "UpdateParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return Msg; },
    },
  },
} as const;

import { QueryParamsRequest, QueryParamsResponse, QueryReviewRequest, QueryReviewResponse, QueryReviewsByUserRequest, QueryReviewsByUserResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.review.v1.Query",
  methods: {
    review: {
      name: "Review",
      httpPath: "/virtengine/review/v1/reviews/{review_id}",
      input: QueryReviewRequest,
      output: QueryReviewResponse,
      get parent() { return Query; },
    },
    reviewsByUser: {
      name: "ReviewsByUser",
      httpPath: "/virtengine/review/v1/reviews/by-user/{reviewer}",
      input: QueryReviewsByUserRequest,
      output: QueryReviewsByUserResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/virtengine/review/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
  },
} as const;

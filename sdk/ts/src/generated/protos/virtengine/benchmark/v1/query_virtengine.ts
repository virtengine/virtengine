import { QueryBenchmarkRequest, QueryBenchmarkResponse, QueryBenchmarksByProviderRequest, QueryBenchmarksByProviderResponse, QueryParamsRequest, QueryParamsResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.benchmark.v1.Query",
  methods: {
    benchmark: {
      name: "Benchmark",
      httpPath: "/virtengine/benchmark/v1/benchmarks/{report_id}",
      input: QueryBenchmarkRequest,
      output: QueryBenchmarkResponse,
      get parent() { return Query; },
    },
    benchmarksByProvider: {
      name: "BenchmarksByProvider",
      httpPath: "/virtengine/benchmark/v1/benchmarks/by-provider/{provider}",
      input: QueryBenchmarksByProviderRequest,
      output: QueryBenchmarksByProviderResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/virtengine/benchmark/v1/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
  },
} as const;

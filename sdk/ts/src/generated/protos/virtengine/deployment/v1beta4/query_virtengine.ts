import { QueryDeploymentRequest, QueryDeploymentResponse, QueryDeploymentsRequest, QueryDeploymentsResponse, QueryGroupRequest, QueryGroupResponse, QueryParamsRequest, QueryParamsResponse } from "./query.ts";

export const Query = {
  typeName: "virtengine.deployment.v1beta4.Query",
  methods: {
    deployments: {
      name: "Deployments",
      httpPath: "/virtengine/deployment/v1beta4/deployments/list",
      input: QueryDeploymentsRequest,
      output: QueryDeploymentsResponse,
      get parent() { return Query; },
    },
    deployment: {
      name: "Deployment",
      httpPath: "/virtengine/deployment/v1beta4/deployments/info",
      input: QueryDeploymentRequest,
      output: QueryDeploymentResponse,
      get parent() { return Query; },
    },
    group: {
      name: "Group",
      httpPath: "/virtengine/deployment/v1beta4/groups/info",
      input: QueryGroupRequest,
      output: QueryGroupResponse,
      get parent() { return Query; },
    },
    params: {
      name: "Params",
      httpPath: "/virtengine/deployment/v1beta4/params",
      input: QueryParamsRequest,
      output: QueryParamsResponse,
      get parent() { return Query; },
    },
  },
} as const;

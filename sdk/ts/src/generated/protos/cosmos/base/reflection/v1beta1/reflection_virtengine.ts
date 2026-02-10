import { ListAllInterfacesRequest, ListAllInterfacesResponse, ListImplementationsRequest, ListImplementationsResponse } from "./reflection.ts";

export const ReflectionService = {
  typeName: "cosmos.base.reflection.v1beta1.ReflectionService",
  methods: {
    listAllInterfaces: {
      name: "ListAllInterfaces",
      httpPath: "/cosmos/base/reflection/v1beta1/interfaces",
      input: ListAllInterfacesRequest,
      output: ListAllInterfacesResponse,
      get parent() { return ReflectionService; },
    },
    listImplementations: {
      name: "ListImplementations",
      httpPath: "/cosmos/base/reflection/v1beta1/interfaces/{interface_name}/implementations",
      input: ListImplementationsRequest,
      output: ListImplementationsResponse,
      get parent() { return ReflectionService; },
    },
  },
} as const;

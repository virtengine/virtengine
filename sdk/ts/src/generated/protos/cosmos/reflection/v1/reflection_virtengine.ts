import { FileDescriptorsRequest, FileDescriptorsResponse } from "./reflection.ts";

export const ReflectionService = {
  typeName: "cosmos.reflection.v1.ReflectionService",
  methods: {
    fileDescriptors: {
      name: "FileDescriptors",
      input: FileDescriptorsRequest,
      output: FileDescriptorsResponse,
      get parent() { return ReflectionService; },
    },
  },
} as const;

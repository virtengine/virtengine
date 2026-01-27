const SDK_METHOD_METADATA = new Map<SDKMethod, SDKMethodMetadata>();
export function withMetadata<T extends SDKMethod>(fn: T, metadata: SDKMethodMetadata): T {
  SDK_METHOD_METADATA.set(fn, metadata);
  return fn;
}

export function getMetadata(fn: SDKMethod): SDKMethodMetadata | undefined {
  return SDK_METHOD_METADATA.get(fn);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type SDKMethod = (input: any, options?: any) => Promise<any>;

interface SDKMethodMetadata {
  /**
   * The path to the method in the service loader.
   * 1st number is the index of the service and the 2nd is the index of the method.
   */
  path: [number, number];
}

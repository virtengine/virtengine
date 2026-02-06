/* eslint-disable @typescript-eslint/no-explicit-any */
declare module 'prom-client' {
  const promClient: any;
  export = promClient;
}

declare module 'redis' {
  const redis: any;
  export = redis;
}

export {};

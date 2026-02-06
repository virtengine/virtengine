/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

/* eslint-disable @typescript-eslint/no-explicit-any */

declare module 'prom-client' {
  namespace client {
    class Registry {
      constructor(...args: any[]);
      metrics: (...args: any[]) => Promise<string> | string;
      contentType: string;
    }
    class Gauge {
      constructor(...args: any[]);
      set: (...args: any[]) => void;
    }
    class Counter {
      constructor(...args: any[]);
      inc: (...args: any[]) => void;
    }
    function collectDefaultMetrics(...args: any[]): void;
  }

  export = client;
}

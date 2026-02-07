/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

/* eslint-disable @typescript-eslint/no-explicit-any */

declare module 'redis' {
  export const createClient: (...args: any[]) => any;
}

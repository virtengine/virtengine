/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

const { spawnSync } = require('node:child_process');

process.env.VITE_CJS_IGNORE_WARNING = '1';

const result = spawnSync('vitest', ['run'], {
  stdio: 'inherit',
  shell: true,
  env: process.env,
});

process.exit(result.status ?? 1);

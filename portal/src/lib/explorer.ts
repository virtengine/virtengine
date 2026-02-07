/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { getChainInfo } from '@/config/chains';

const DEFAULT_EXPLORER_URL = 'http://localhost:8088';

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '');
}

export function getExplorerBaseUrl(): string {
  const override = process.env.NEXT_PUBLIC_EXPLORER_URL ?? process.env.EXPLORER_URL;
  const base = override || getChainInfo().explorerUrl || DEFAULT_EXPLORER_URL;
  return trimTrailingSlash(base);
}

export function txLink(hash: string): string {
  return `${getExplorerBaseUrl()}/tx/${hash}`;
}

export function blockLink(height: number | string): string {
  return `${getExplorerBaseUrl()}/block/${height}`;
}

export function accountLink(address: string): string {
  return `${getExplorerBaseUrl()}/account/${address}`;
}

export function validatorLink(address: string): string {
  return `${getExplorerBaseUrl()}/validator/${address}`;
}

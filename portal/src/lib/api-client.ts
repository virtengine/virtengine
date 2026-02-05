/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { env } from '@/config/env';

export interface ApiClientConfig {
  baseUrl?: string;
  headers?: Record<string, string>;
  fetcher?: typeof fetch;
}

export interface ApiRequestOptions extends RequestInit {
  query?: Record<string, string | number | boolean | undefined>;
}

export class ApiError extends Error {
  status: number;
  payload?: unknown;

  constructor(message: string, status: number, payload?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.payload = payload;
  }
}

/**
 * ApiClient
 *
 * Thin fetch wrapper for Portal API calls. This is a scaffold and will be
 * expanded in task 23B with typed endpoints and auth wiring.
 */
export class ApiClient {
  private baseUrl: string;
  private headers: Record<string, string>;
  private fetcher: typeof fetch;

  constructor(config: ApiClientConfig = {}) {
    this.baseUrl = config.baseUrl ?? env.apiUrl;
    this.headers = {
      'Content-Type': 'application/json',
      ...config.headers,
    };
    this.fetcher = config.fetcher ?? fetch;
  }

  async request<T>(path: string, options: ApiRequestOptions = {}): Promise<T> {
    const url = new URL(path, this.baseUrl);
    if (options.query) {
      Object.entries(options.query).forEach(([key, value]) => {
        if (value !== undefined) {
          url.searchParams.set(key, String(value));
        }
      });
    }

    const response = await this.fetcher(url.toString(), {
      ...options,
      headers: {
        ...this.headers,
        ...options.headers,
      },
    });

    const contentType = response.headers.get('content-type') ?? '';
    let payload: unknown;
    if (contentType.includes('application/json')) {
      payload = (await response.json()) as unknown;
    } else {
      payload = await response.text();
    }

    if (!response.ok) {
      throw new ApiError(response.statusText, response.status, payload);
    }

    return payload as T;
  }

  get<T>(path: string, options: ApiRequestOptions = {}) {
    return this.request<T>(path, { ...options, method: 'GET' });
  }

  post<T>(path: string, body?: unknown, options: ApiRequestOptions = {}) {
    return this.request<T>(path, {
      ...options,
      method: 'POST',
      body: body ? JSON.stringify(body) : options.body,
    });
  }
}

export const apiClient = new ApiClient();

/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { createClient } from 'redis';

type RedisClient = ReturnType<typeof createClient>;

const globalForRedis = globalThis as unknown as {
  veRedisClient?: RedisClient;
  veRedisPromise?: Promise<RedisClient>;
};

function getRedisUrl(): string | null {
  return process.env.REDIS_URL ?? process.env.VE_REDIS_URL ?? null;
}

export async function getRedisClient(): Promise<RedisClient | null> {
  const redisUrl = getRedisUrl();
  if (!redisUrl) return null;

  if (globalForRedis.veRedisClient) {
    return globalForRedis.veRedisClient;
  }

  if (!globalForRedis.veRedisPromise) {
    const client = createClient({ url: redisUrl });
    client.on('error', (err) => {
      console.error('Redis client error:', err);
    });

    globalForRedis.veRedisPromise = client.connect().then(() => client);
  }

  try {
    const client = await globalForRedis.veRedisPromise;
    globalForRedis.veRedisClient = client;
    return client;
  } catch (err) {
    console.error('Failed to connect to Redis:', err);
    return null;
  }
}

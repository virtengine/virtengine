import { afterEach, describe, expect, it, vi } from 'vitest';

const loadEnv = async () => {
  vi.resetModules();
  return (await import('./env')).env;
};

describe('chat environment flags', () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it('defaults chat and chat mutations to false', async () => {
    vi.stubEnv('NEXT_PUBLIC_ENABLE_CHAT', '');
    vi.stubEnv('NEXT_PUBLIC_ENABLE_CHAT_MUTATIONS', '');

    const env = await loadEnv();

    expect(env.enableChat).toBe(false);
    expect(env.enableChatMutations).toBe(false);
  });

  it('does not imply mutations when chat is explicitly enabled', async () => {
    vi.stubEnv('NEXT_PUBLIC_ENABLE_CHAT', 'true');
    vi.stubEnv('NEXT_PUBLIC_ENABLE_CHAT_MUTATIONS', '');

    const env = await loadEnv();

    expect(env.enableChat).toBe(true);
    expect(env.enableChatMutations).toBe(false);
  });
});

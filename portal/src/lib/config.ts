/**
 * Portal runtime configuration helpers.
 */

import { env } from '@/config/env';

export type LLMProviderName = 'openai' | 'local';

export interface LLMConfig {
  provider: LLMProviderName;
  endpoint: string;
  apiKey?: string;
  model: string;
  localEndpoint: string;
}

export const llmConfig: LLMConfig = {
  provider: env.llmProvider as LLMProviderName,
  endpoint: env.llmEndpoint,
  apiKey: env.llmApiKey || undefined,
  model: env.llmModel,
  localEndpoint: env.llmLocalEndpoint,
};

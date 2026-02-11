import { LocalChatProvider } from "./local";
import { OpenAIChatProvider } from "./openai";
import type { ChatProviderConfig } from "../types";
import type { ChatProvider } from "./base";

export const createChatProvider = (
  config: ChatProviderConfig,
): ChatProvider => {
  if (config.provider === "local") {
    return new LocalChatProvider(config);
  }
  return new OpenAIChatProvider(config);
};

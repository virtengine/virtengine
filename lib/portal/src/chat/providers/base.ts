/**
 * Base utilities for LLM providers.
 */

import type { ChatCompletionTool, ChatToolDefinition } from "../types";

export function toCompletionTools(
  tools: ChatToolDefinition[],
): ChatCompletionTool[] {
  return tools.map((tool) => ({
    type: "function",
    function: {
      name: tool.name,
      description: tool.description,
      parameters: tool.parameters,
    },
  }));
}

export function normalizeEndpoint(endpoint: string): string {
  if (!endpoint) return "";
  const trimmed = endpoint.replace(/\/$/, "");
  if (trimmed.endsWith("/v1")) {
    return trimmed;
  }
  return `${trimmed}/v1`;
}

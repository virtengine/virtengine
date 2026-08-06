import type { ChatAction, ChatActionExecution } from '@/lib/portal-adapter';

interface ExecuteChatActionDependencies {
  mutationsEnabled: boolean;
  estimateFee: (gas: number) => unknown;
  sendTransaction: (msgs: unknown[], options?: { memo?: string }) => Promise<{ txHash: string }>;
  executeHandler: (action: ChatAction) => Promise<ChatActionExecution>;
}

export async function executeChatAction(
  action: ChatAction,
  dependencies: ExecuteChatActionDependencies
): Promise<ChatActionExecution> {
  if (!dependencies.mutationsEnabled) {
    return {
      ok: false,
      code: 'feature_unavailable',
      summary: 'Chat mutations are disabled.',
    };
  }

  if (action.payload.kind === 'transaction') {
    const fee = dependencies.estimateFee(200000);
    const result = await dependencies.sendTransaction(action.payload.msgs, {
      memo: action.payload.memo,
    });
    return {
      ok: true,
      summary: `Transaction submitted: ${result.txHash}.`,
      details: { fee, txHash: result.txHash },
    };
  }

  if (action.payload.kind === 'provider-action') {
    return dependencies.executeHandler(action);
  }

  return {
    ok: false,
    code: 'unsupported_action',
    summary: 'This action cannot be executed.',
  };
}

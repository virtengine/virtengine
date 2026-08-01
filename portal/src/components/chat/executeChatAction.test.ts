import { describe, expect, it, vi } from 'vitest';

import type { ChatAction } from '@/lib/portal-adapter';
import { executeChatAction } from './executeChatAction';

const staleAction: ChatAction = {
  id: 'stale-action',
  toolName: 'transfer-tokens',
  title: 'Transfer tokens',
  summary: 'Persisted model-selected action',
  payload: {
    kind: 'transaction',
    msgs: [{ typeUrl: '/cosmos.bank.v1beta1.MsgSend', value: {} }],
  },
};

describe('executeChatAction', () => {
  it('returns feature_unavailable without fee, signer, transaction, or handler calls', async () => {
    const estimateFee = vi.fn();
    const sendTransaction = vi.fn();
    const executeHandler = vi.fn();

    const result = await executeChatAction(staleAction, {
      mutationsEnabled: false,
      estimateFee,
      sendTransaction,
      executeHandler,
    });

    expect(result).toEqual({
      ok: false,
      code: 'feature_unavailable',
      summary: 'Chat mutations are disabled.',
    });
    expect(estimateFee).not.toHaveBeenCalled();
    expect(sendTransaction).not.toHaveBeenCalled();
    expect(executeHandler).not.toHaveBeenCalled();
  });
});

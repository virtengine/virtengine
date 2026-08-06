import type {
  ProviderDeploymentAction,
  ProviderDeploymentActionReceipt,
} from '@/lib/portal-adapter';

interface ProviderDeploymentActionUIHandlers {
  setPendingAction: (action: ProviderDeploymentAction | null) => void;
  setActionError: (message: string | null) => void;
  setLastReceipt: (receipt: ProviderDeploymentActionReceipt) => void;
  requestWallet: () => void;
}

const errorCode = (error: unknown): string | undefined =>
  typeof error === 'object' && error !== null && 'code' in error
    ? String((error as { code?: unknown }).code)
    : undefined;

export const executeProviderDeploymentAction = async (
  action: ProviderDeploymentAction,
  execute: () => Promise<ProviderDeploymentActionReceipt>,
  handlers: ProviderDeploymentActionUIHandlers
): Promise<ProviderDeploymentActionReceipt | null> => {
  handlers.setActionError(null);
  handlers.setPendingAction(action);
  try {
    const receipt = await execute();
    handlers.setLastReceipt(receipt);
    return receipt;
  } catch (error) {
    if (errorCode(error) === 'chain_signing_required') {
      handlers.requestWallet();
    }
    handlers.setActionError(
      error instanceof Error ? error.message : `Provider rejected the ${action} operation.`
    );
    return null;
  } finally {
    handlers.setPendingAction(null);
  }
};

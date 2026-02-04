import { useWallet } from '../../wallet/context';

export interface NetworkBadgeProps {
  className?: string;
}

export function NetworkBadge({ className }: NetworkBadgeProps) {
  const { state } = useWallet();

  return (
    <span className={className}>
      {state.networkName || state.chainId || 'Unknown network'}
    </span>
  );
}

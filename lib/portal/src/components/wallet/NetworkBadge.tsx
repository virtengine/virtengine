import { useWallet } from '../../wallet';

export interface NetworkBadgeProps {
  className?: string;
}

export function NetworkBadge({ className }: NetworkBadgeProps) {
  const { chainId } = useWallet();

  return (
    <span className={className}>
      {chainId || 'Unknown network'}
    </span>
  );
}

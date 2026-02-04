import { useWallet } from '../../wallet/context';

export interface AccountDisplayProps {
  className?: string;
  showStatus?: boolean;
}

export function AccountDisplay({ className, showStatus = true }: AccountDisplayProps) {
  const { state } = useWallet();

  if (!state.isConnected || !state.address) {
    return <span className={className}>Not connected</span>;
  }

  return (
    <span className={className}>
      {showStatus ? 'Connected: ' : ''}
      {state.address.slice(0, 10)}...{state.address.slice(-4)}
    </span>
  );
}

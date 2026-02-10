import { useWallet } from '../../wallet';

export interface AccountDisplayProps {
  className?: string;
  showStatus?: boolean;
}

export function AccountDisplay({ className, showStatus = true }: AccountDisplayProps) {
  const { status, accounts, activeAccountIndex } = useWallet();
  const account = accounts[activeAccountIndex];

  if (status !== 'connected' || !account) {
    return <span className={className}>Not connected</span>;
  }

  return (
    <span className={className}>
      {showStatus ? 'Connected: ' : ''}
      {account.address.slice(0, 10)}...{account.address.slice(-4)}
    </span>
  );
}

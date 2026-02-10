import { useChain } from "@cosmos-kit/react";
import { WalletStatus } from "@cosmos-kit/core";

import { useBalances } from "../hooks/useBalances";
import { useVeidStatus } from "../hooks/useVeidStatus";
import { chainName, formatCoin } from "../services/chain";

export const AccountSummary = () => {
  const { address, status } = useChain(chainName);
  const balancesQuery = useBalances(address ?? undefined);
  const veidQuery = useVeidStatus(address ?? undefined);

  if (status !== WalletStatus.Connected) {
    return (
      <div className="rounded-2xl border border-dashed border-slate-800 bg-slate-900/40 p-6">
        <p className="text-sm text-slate-400">
          Connect a wallet to see balances and VEID status.
        </p>
      </div>
    );
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[2fr,1fr]">
      <div className="rounded-2xl border border-slate-800 bg-slate-900/60 p-6">
        <h3 className="text-lg font-semibold text-slate-100">Balances</h3>
        <p className="mt-2 text-xs text-slate-400">{address}</p>
        {balancesQuery.isLoading ? (
          <p className="mt-3 text-sm text-slate-400">Loading balances...</p>
        ) : balancesQuery.isError ? (
          <p className="mt-3 text-sm text-rose-300">
            Unable to load balances from RPC.
          </p>
        ) : balancesQuery.data?.length ? (
          <ul className="mt-4 space-y-2 text-sm text-slate-200">
            {balancesQuery.data.map((coin) => (
              <li key={coin.denom} className="flex items-center justify-between">
                <span className="uppercase text-slate-400">{coin.denom}</span>
                <span className="font-semibold text-slate-100">
                  {formatCoin(coin)}
                </span>
              </li>
            ))}
          </ul>
        ) : (
          <p className="mt-3 text-sm text-slate-400">No balances yet.</p>
        )}
      </div>
      <div className="rounded-2xl border border-slate-800 bg-slate-900/60 p-6">
        <h3 className="text-lg font-semibold text-slate-100">VEID status</h3>
        {veidQuery.isLoading ? (
          <p className="mt-3 text-sm text-slate-400">Checking VEID...</p>
        ) : (
          <div className="mt-3 space-y-2">
            <p className="text-2xl font-semibold text-slate-100">
              {veidQuery.data?.status ?? "Unknown"}
            </p>
            {veidQuery.data?.detail && (
              <p className="text-xs text-slate-400">{veidQuery.data.detail}</p>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

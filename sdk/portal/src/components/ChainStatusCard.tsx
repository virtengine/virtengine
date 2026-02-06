import { useChainStatus } from "../hooks/useChainStatus";

export const ChainStatusCard = () => {
  const { data, isLoading, isError } = useChainStatus();

  return (
    <div className="rounded-2xl border border-slate-800 bg-slate-900/60 p-6">
      <h3 className="text-lg font-semibold text-slate-100">Chain status</h3>
      {isLoading ? (
        <p className="mt-3 text-sm text-slate-400">Loading latest block...</p>
      ) : isError || !data ? (
        <p className="mt-3 text-sm text-rose-300">
          Unable to load chain status. Verify REST endpoint.
        </p>
      ) : (
        <div className="mt-4 grid gap-4 md:grid-cols-3">
          <div>
            <p className="text-xs uppercase tracking-wide text-slate-400">Height</p>
            <p className="text-2xl font-semibold text-slate-100">
              {data.latestHeight?.toLocaleString() ?? "-"}
            </p>
          </div>
          <div>
            <p className="text-xs uppercase tracking-wide text-slate-400">
              Validators
            </p>
            <p className="text-2xl font-semibold text-slate-100">
              {data.validatorCount ?? "-"}
            </p>
          </div>
          <div>
            <p className="text-xs uppercase tracking-wide text-slate-400">Chain ID</p>
            <p className="text-sm font-semibold text-slate-200">{data.chainId}</p>
          </div>
        </div>
      )}
    </div>
  );
};

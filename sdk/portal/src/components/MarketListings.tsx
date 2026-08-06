import { useMarketListings } from "../hooks/useMarketListings";

export const MarketListings = () => {
  const { data, isLoading, isError } = useMarketListings();

  return (
    <div className="rounded-2xl border border-slate-800 bg-slate-900/60 p-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h3 className="text-lg font-semibold text-slate-100">Marketplace preview</h3>
          <p className="mt-1 text-sm text-slate-400">
            Live offerings resolved from the same chain/provider surfaces the production portal
            uses.
          </p>
        </div>
      </div>

      {isLoading ? (
        <p className="mt-4 text-sm text-slate-400">Loading live offerings…</p>
      ) : isError ? (
        <p className="mt-4 text-sm text-rose-300">
          Unable to load offerings from the configured chain/provider endpoints.
        </p>
      ) : data?.length ? (
        <ul className="mt-4 space-y-3">
          {data.slice(0, 5).map((listing) => (
            <li
              key={listing.id}
              className="rounded-xl border border-slate-800 bg-slate-950/60 p-4"
            >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p className="font-semibold text-slate-100">{listing.title}</p>
                  <p className="mt-1 text-xs uppercase tracking-wide text-slate-500">
                    {listing.category ?? "other"} · {listing.providerName ?? listing.providerAddress}
                  </p>
                  {listing.description && (
                    <p className="mt-2 text-sm text-slate-400">{listing.description}</p>
                  )}
                </div>
                <div className="text-right">
                  <p className="text-sm font-semibold text-slate-100">{listing.price}</p>
                  {listing.state && (
                    <p className="mt-1 text-xs text-slate-500">State: {listing.state}</p>
                  )}
                </div>
              </div>
            </li>
          ))}
        </ul>
      ) : (
        <p className="mt-4 text-sm text-slate-400">
          No live offerings were returned by the configured endpoints.
        </p>
      )}
    </div>
  );
};

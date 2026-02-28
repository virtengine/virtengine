import { MarketListings } from "../components/MarketListings";

export const Marketplace = () => {
  return (
    <div className="space-y-6">
      <header>
        <h2 className="text-2xl font-semibold text-slate-100">Marketplace</h2>
        <p className="mt-2 text-sm text-slate-400">
          Live offering previews resolved from chain and provider registry data.
        </p>
      </header>
      <MarketListings />
    </div>
  );
};

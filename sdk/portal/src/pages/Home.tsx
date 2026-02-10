import { ChainStatusCard } from "../components/ChainStatusCard";

export const Home = () => {
  return (
    <div className="space-y-8">
      <section className="rounded-3xl border border-slate-800 bg-slate-900/50 p-8">
        <p className="text-xs uppercase tracking-[0.4em] text-slate-500">
          Localnet ready
        </p>
        <h2 className="mt-3 text-3xl font-semibold text-slate-100">
          VirtEngine Portal
        </h2>
        <p className="mt-4 max-w-2xl text-sm text-slate-300">
          Connect a wallet to inspect VE balances, check your VEID status, and
          monitor chain health in real-time. This scaffold is wired for
          localnet-first development and Cosmos SDK queries.
        </p>
      </section>

      <ChainStatusCard />
    </div>
  );
};

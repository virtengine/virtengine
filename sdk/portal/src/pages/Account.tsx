import { AccountSummary } from "../components/AccountSummary";

export const Account = () => {
  return (
    <div className="space-y-6">
      <header>
        <h2 className="text-2xl font-semibold text-slate-100">Account</h2>
        <p className="mt-2 text-sm text-slate-400">
          Wallet balances and VEID verification status.
        </p>
      </header>
      <AccountSummary />
    </div>
  );
};

import { NavLink } from "react-router-dom";

import { WalletConnect } from "./WalletConnect";

const navItems = [
  { to: "/", label: "Overview" },
  { to: "/account", label: "Account" },
];

export const AppShell = ({ children }: { children: React.ReactNode }) => {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950">
      <header className="border-b border-slate-900/60 bg-slate-950/80 backdrop-blur">
        <div className="mx-auto flex w-full max-w-6xl flex-wrap items-center justify-between gap-4 px-6 py-6">
          <div>
            <p className="text-xs uppercase tracking-[0.4em] text-slate-500">
              VirtEngine
            </p>
            <h1 className="text-2xl font-semibold text-slate-100">VE Portal</h1>
          </div>
          <WalletConnect />
        </div>
        <nav className="mx-auto w-full max-w-6xl px-6 pb-4">
          <div className="flex gap-2">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  [
                    "rounded-full border px-4 py-2 text-sm transition",
                    isActive
                      ? "border-slate-500 bg-slate-900 text-slate-100"
                      : "border-slate-800 text-slate-400 hover:border-slate-600 hover:text-slate-100",
                  ].join(" ")
                }
              >
                {item.label}
              </NavLink>
            ))}
          </div>
        </nav>
      </header>
      <main className="mx-auto w-full max-w-6xl px-6 py-10">{children}</main>
    </div>
  );
};

import React from "react";
import ReactDOM from "react-dom/client";
import { ChainProvider } from "@cosmos-kit/react";
import { wallets as cosmostationWallets } from "@cosmos-kit/cosmostation";
import { wallets as keplrWallets } from "@cosmos-kit/keplr";
import { wallets as leapWallets } from "@cosmos-kit/leap";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { App } from "./App";
import "./index.css";
import { virtengineAssets, virtengineChain } from "./services/chain";

const queryClient = new QueryClient();
const wallets = [...keplrWallets, ...leapWallets, ...cosmostationWallets];

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <ChainProvider
        chains={[virtengineChain]}
        assetLists={[virtengineAssets]}
        wallets={wallets}
        throwErrors={false}
      >
        <App />
      </ChainProvider>
    </QueryClientProvider>
  </React.StrictMode>
);

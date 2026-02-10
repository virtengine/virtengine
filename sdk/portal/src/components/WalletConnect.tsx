import { useChain } from "@cosmos-kit/react";
import { WalletStatus } from "@cosmos-kit/core";

import { chainName } from "../services/chain";
import { Button } from "./ui/button";

export const WalletConnect = () => {
  const { connect, disconnect, status, address, wallet, username } =
    useChain(chainName);

  const isConnected = status === WalletStatus.Connected;

  const handleConnect = async () => {
    await connect();
  };

  return (
    <div className="flex flex-wrap items-center gap-3">
      <div className="text-sm text-slate-300">
        {isConnected ? (
          <div className="space-y-1">
            <p className="font-medium text-slate-100">
              {wallet?.prettyName ?? "Wallet"}
            </p>
            <p className="text-xs text-slate-400">{username ?? address}</p>
          </div>
        ) : (
          <p className="text-xs text-slate-400">Not connected</p>
        )}
      </div>

      {isConnected ? (
        <Button variant="outline" onClick={() => disconnect()}>
          Disconnect
        </Button>
      ) : (
        <Button onClick={handleConnect}>
          {status === WalletStatus.Connecting ? "Connecting..." : "Connect Wallet"}
        </Button>
      )}
    </div>
  );
};

import { useQuery } from "@tanstack/react-query";

import { fetchBalances } from "../services/chain";

export const useBalances = (address?: string) =>
  useQuery({
    queryKey: ["balances", address],
    queryFn: () => fetchBalances(address ?? ""),
    enabled: Boolean(address),
    refetchInterval: 15_000,
  });

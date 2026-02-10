import { useQuery } from "@tanstack/react-query";

import { fetchChainStatus } from "../services/chain";

export const useChainStatus = () =>
  useQuery({
    queryKey: ["chain-status"],
    queryFn: fetchChainStatus,
    refetchInterval: 10_000,
  });

import { useQuery } from "@tanstack/react-query";

import { fetchVeidStatus } from "../services/chain";

export const useVeidStatus = (address?: string) =>
  useQuery({
    queryKey: ["veid-status", address],
    queryFn: () => fetchVeidStatus(address ?? ""),
    enabled: Boolean(address),
    refetchInterval: 30_000,
  });

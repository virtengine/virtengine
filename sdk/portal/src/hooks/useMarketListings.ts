import { useQuery } from "@tanstack/react-query";

import { fetchMarketListings } from "../services/market";

export const useMarketListings = () =>
  useQuery({
    queryKey: ["market-listings"],
    queryFn: fetchMarketListings,
    staleTime: 30_000,
    refetchInterval: 60_000,
  });

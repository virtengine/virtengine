export type MarketListing = {
  id: string;
  title: string;
  price: string;
};

export const fetchMarketListings = async (): Promise<MarketListing[]> => {
  return [];
};

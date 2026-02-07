export type ChainStatus = {
  chainId: string;
  latestHeight: number | null;
  validatorCount: number | null;
};

export type VeidStatus = {
  status: string;
  detail?: string;
};

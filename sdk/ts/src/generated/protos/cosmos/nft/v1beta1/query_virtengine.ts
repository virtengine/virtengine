import { QueryBalanceRequest, QueryBalanceResponse, QueryClassesRequest, QueryClassesResponse, QueryClassRequest, QueryClassResponse, QueryNFTRequest, QueryNFTResponse, QueryNFTsRequest, QueryNFTsResponse, QueryOwnerRequest, QueryOwnerResponse, QuerySupplyRequest, QuerySupplyResponse } from "./query.ts";

export const Query = {
  typeName: "cosmos.nft.v1beta1.Query",
  methods: {
    balance: {
      name: "Balance",
      httpPath: "/cosmos/nft/v1beta1/balance/{owner}/{class_id}",
      input: QueryBalanceRequest,
      output: QueryBalanceResponse,
      get parent() { return Query; },
    },
    owner: {
      name: "Owner",
      httpPath: "/cosmos/nft/v1beta1/owner/{class_id}/{id}",
      input: QueryOwnerRequest,
      output: QueryOwnerResponse,
      get parent() { return Query; },
    },
    supply: {
      name: "Supply",
      httpPath: "/cosmos/nft/v1beta1/supply/{class_id}",
      input: QuerySupplyRequest,
      output: QuerySupplyResponse,
      get parent() { return Query; },
    },
    nFTs: {
      name: "NFTs",
      httpPath: "/cosmos/nft/v1beta1/nfts",
      input: QueryNFTsRequest,
      output: QueryNFTsResponse,
      get parent() { return Query; },
    },
    nFT: {
      name: "NFT",
      httpPath: "/cosmos/nft/v1beta1/nfts/{class_id}/{id}",
      input: QueryNFTRequest,
      output: QueryNFTResponse,
      get parent() { return Query; },
    },
    class: {
      name: "Class",
      httpPath: "/cosmos/nft/v1beta1/classes/{class_id}",
      input: QueryClassRequest,
      output: QueryClassResponse,
      get parent() { return Query; },
    },
    classes: {
      name: "Classes",
      httpPath: "/cosmos/nft/v1beta1/classes",
      input: QueryClassesRequest,
      output: QueryClassesResponse,
      get parent() { return Query; },
    },
  },
} as const;

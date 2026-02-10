import type {
  AminoSignDoc,
  AminoSignResponse,
  WalletSignOptions,
} from "../wallet/types";

export interface WalletRequestSigner {
  signAmino: (
    signDoc: AminoSignDoc,
    options?: WalletSignOptions,
  ) => Promise<AminoSignResponse>;
  signArbitrary?: (
    data: string | Uint8Array,
  ) => Promise<{ signature: string; pubKey: Uint8Array }>;
}

export interface SignedRequestHeaders {
  "X-VE-Address": string;
  "X-VE-Timestamp": string;
  "X-VE-Nonce": string;
  "X-VE-Signature": string;
  "X-VE-PubKey": string;
}

export interface SignRequestOptions {
  method: string;
  path: string;
  body?: unknown;
  signer: WalletRequestSigner;
  address: string;
  chainId: string;
  memo?: string;
}

import { type OfflineAminoSigner, type StdSignature } from "@cosmjs/amino";

import { base64Encode } from "./base64.ts";

export interface OfflineDataSigner {
  /**
   * The algorithm used to sign the data.
   * @default "ES256KADR36"
   */
  algorithm?: "ES256KADR36" | "ES256K";
  signArbitrary: (signer: string, data: string | Uint8Array) => Promise<StdSignature>;
}

export function createOfflineDataSigner(wallet: OfflineAminoSigner): OfflineDataSigner {
  return {
    algorithm: "ES256KADR36",
    async signArbitrary(signer, data) {
      const { signature } = await wallet.signAmino(signer, {
        chain_id: "",
        account_number: "0",
        sequence: "0",
        fee: {
          gas: "0",
          amount: [],
        },
        msgs: [
          {
            type: "sign/MsgSignData",
            value: {
              signer,
              data: base64Encode(data),
            },
          },
        ],
        memo: "",
      });
      return signature;
    },
  };
}

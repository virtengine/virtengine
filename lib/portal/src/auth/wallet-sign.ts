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

interface RequestData {
  method: string;
  path: string;
  timestamp: number;
  nonce: string;
  body_hash: string;
}

const textEncoder = new TextEncoder();

const serializeRequestData = (data: RequestData): string => {
  return `{"method":"${data.method}","path":"${data.path}","timestamp":${data.timestamp},"nonce":"${data.nonce}","body_hash":"${data.body_hash}"}`;
};

const stableStringify = (value: unknown): string => {
  if (value === null || typeof value !== "object") {
    return JSON.stringify(value);
  }

  if (Array.isArray(value)) {
    return `[${value.map(stableStringify).join(",")}]`;
  }

  const obj = value as Record<string, unknown>;
  const keys = Object.keys(obj).sort();
  return `{${keys
    .map((key) => `${JSON.stringify(key)}:${stableStringify(obj[key])}`)
    .join(",")}}`;
};

const toBase64 = (value: string): string => {
  if (typeof btoa === "function") {
    return btoa(value);
  }
  return Buffer.from(value, "utf-8").toString("base64");
};

const toHex = (bytes: Uint8Array): string => {
  return Array.from(bytes)
    .map((byte) => byte.toString(16).padStart(2, "0"))
    .join("");
};

const sha256Hex = async (payload: string): Promise<string> => {
  const data = textEncoder.encode(payload);
  if (globalThis.crypto?.subtle) {
    const digest = await globalThis.crypto.subtle.digest("SHA-256", data);
    return toHex(new Uint8Array(digest));
  }

  const nodeCrypto = await import("crypto");
  return nodeCrypto.createHash("sha256").update(data).digest("hex");
};

const randomNonce = (): string => {
  const bytes = new Uint8Array(16);
  if (globalThis.crypto?.getRandomValues) {
    globalThis.crypto.getRandomValues(bytes);
  } else {
    for (let i = 0; i < bytes.length; i += 1) {
      bytes[i] = Math.floor(Math.random() * 256);
    }
  }
  return toHex(bytes);
};

export async function signRequest(
  options: SignRequestOptions,
): Promise<SignedRequestHeaders> {
  const timestamp = Date.now();
  const nonce = randomNonce();
  const bodyPayload = options.body ? stableStringify(options.body) : "";
  const bodyHash = bodyPayload ? await sha256Hex(bodyPayload) : "";

  const requestData: RequestData = {
    method: options.method.toUpperCase(),
    path: options.path,
    timestamp,
    nonce,
    body_hash: bodyHash,
  };

  const dataToSign = serializeRequestData(requestData);
  const dataBase64 = toBase64(dataToSign);

  const signDoc: AminoSignDoc = {
    chain_id: options.chainId,
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
          signer: options.address,
          data: dataBase64,
        },
      },
    ],
    memo: options.memo ?? "",
  };

  const signResponse = await options.signer.signAmino(signDoc);

  return {
    "X-VE-Address": options.address,
    "X-VE-Timestamp": timestamp.toString(),
    "X-VE-Nonce": nonce,
    "X-VE-Signature": signResponse.signature.signature,
    "X-VE-PubKey": signResponse.signature.pub_key.value,
  };
}

import { Decimal } from "@cosmjs/math";

import type { CustomType } from "./CustomType.ts";

/**
 * @see https://github.com/cosmos/cosmos-sdk/blob/25b14c3caa2ecdc99840dbb88fdb3a2d8ac02158/math/dec.go#L21
 */
const PRECISION = 18;

export const LegacyDec = {
  typeName: "cosmossdk.io/math.LegacyDec",
  shortName: "LegacyDec",
  encode(value: string) {
    if (!value.length) return "";
    const { sign, value: positiveValue } = unsignedDecimal(value);
    return sign + Decimal.fromUserInput(positiveValue, PRECISION).atomics;
  },
  decode(value: string) {
    if (!value.length) return "";
    const { sign, value: positiveValue } = unsignedDecimal(value);
    return sign + Decimal.fromAtomics(positiveValue, PRECISION).toString();
  },
} as const satisfies CustomType<string, string>;

// cosmjs Decimal supports only non-negative decimals: https://github.com/cosmos/cosmjs/issues/1897
function unsignedDecimal(value: string) {
  if (value[0] !== "-") return { sign: "", value };
  const positiveValue = value.slice(1);
  return { sign: "-", value: positiveValue };
}

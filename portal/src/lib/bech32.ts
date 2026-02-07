/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

const BECH32_CHARSET = 'qpzry9x8gf2tvdw0s3jn54khce6mua7l';
const BECH32_CHARSET_MAP = new Map(BECH32_CHARSET.split('').map((char, index) => [char, index]));

function bech32Polymod(values: number[]): number {
  const GEN = [0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3];
  let chk = 1;
  for (const v of values) {
    const b = chk >> 25;
    chk = ((chk & 0x1ffffff) << 5) ^ v;
    for (let i = 0; i < 5; i += 1) {
      chk ^= (b >> i) & 1 ? GEN[i] : 0;
    }
  }
  return chk;
}

function bech32HrpExpand(hrp: string): number[] {
  const ret: number[] = [];
  for (let i = 0; i < hrp.length; i += 1) {
    ret.push(hrp.charCodeAt(i) >> 5);
  }
  ret.push(0);
  for (let i = 0; i < hrp.length; i += 1) {
    ret.push(hrp.charCodeAt(i) & 31);
  }
  return ret;
}

function verifyChecksum(hrp: string, data: number[]): boolean {
  return bech32Polymod([...bech32HrpExpand(hrp), ...data]) === 1;
}

function createChecksum(hrp: string, data: number[]): number[] {
  const values = [...bech32HrpExpand(hrp), ...data, 0, 0, 0, 0, 0, 0];
  const polymod = bech32Polymod(values) ^ 1;
  const result: number[] = [];
  for (let i = 0; i < 6; i += 1) {
    result.push((polymod >> (5 * (5 - i))) & 31);
  }
  return result;
}

function bech32Decode(address: string): { hrp: string; data: number[] } {
  const normalized = address.toLowerCase();
  const separatorIndex = normalized.lastIndexOf('1');
  if (separatorIndex <= 0) {
    throw new Error('Invalid bech32 address: missing separator');
  }

  const hrp = normalized.slice(0, separatorIndex);
  const dataPart = normalized.slice(separatorIndex + 1);
  if (dataPart.length < 6) {
    throw new Error('Invalid bech32 address: data too short');
  }

  const data: number[] = [];
  for (const char of dataPart) {
    const value = BECH32_CHARSET_MAP.get(char);
    if (value === undefined) {
      throw new Error('Invalid bech32 address: invalid character');
    }
    data.push(value);
  }

  if (!verifyChecksum(hrp, data)) {
    throw new Error('Invalid bech32 address: checksum failed');
  }

  return {
    hrp,
    data: data.slice(0, -6),
  };
}

function bech32Encode(hrp: string, data: number[]): string {
  const checksum = createChecksum(hrp, data);
  const combined = [...data, ...checksum];
  const encoded = combined.map((value) => BECH32_CHARSET[value]).join('');
  return `${hrp}1${encoded}`;
}

export function convertBech32Prefix(address: string, nextPrefix: string): string {
  const decoded = bech32Decode(address);
  return bech32Encode(nextPrefix, decoded.data);
}

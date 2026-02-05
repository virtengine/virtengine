/**
 * Tests for Transaction Utilities
 * @module wallet/transaction.test
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  calculateFee,
  adjustGas,
  formatFeeAmount,
  validateTransaction,
  estimateGas,
  createTransactionPreview,
  createDefaultGasSettings,
  GAS_TIERS,
  DEFAULT_GAS_ADJUSTMENT,
  DEFAULT_GAS_LIMIT,
  type GasTier,
  type FeeEstimate,
  type TransactionPreview,
} from '../../src/wallet/transaction';
import type { WalletChainInfo } from '../../src/wallet/types';

// Test fixture for chain info
const createTestChainInfo = (overrides: Partial<WalletChainInfo> = {}): WalletChainInfo => ({
  chainId: 'virtengine-1',
  chainName: 'VirtEngine',
  rpcEndpoint: 'https://rpc.virtengine.io',
  restEndpoint: 'https://rest.virtengine.io',
  bech32Config: {
    bech32PrefixAccAddr: 'virtengine',
    bech32PrefixAccPub: 'virtuenginepub',
    bech32PrefixValAddr: 'virtenginevaloper',
    bech32PrefixValPub: 'virtenginevaloperpub',
    bech32PrefixConsAddr: 'virtenginevalcons',
    bech32PrefixConsPub: 'virtenginevalconspub',
  },
  stakeCurrency: {
    coinDenom: 'VIRT',
    coinMinimalDenom: 'uvirt',
    coinDecimals: 6,
  },
  currencies: [
    {
      coinDenom: 'VIRT',
      coinMinimalDenom: 'uvirt',
      coinDecimals: 6,
    },
  ],
  feeCurrencies: [
    {
      coinDenom: 'VIRT',
      coinMinimalDenom: 'uvirt',
      coinDecimals: 6,
      gasPriceStep: {
        low: 0.01,
        average: 0.025,
        high: 0.04,
      },
    },
  ],
  ...overrides,
});

describe('GAS_TIERS constants', () => {
  it('should define LOW tier as 1.0', () => {
    expect(GAS_TIERS.LOW).toBe(1.0);
  });

  it('should define AVERAGE tier as 1.3', () => {
    expect(GAS_TIERS.AVERAGE).toBe(1.3);
  });

  it('should define HIGH tier as 2.0', () => {
    expect(GAS_TIERS.HIGH).toBe(2.0);
  });
});

describe('DEFAULT_GAS_ADJUSTMENT', () => {
  it('should be 1.3', () => {
    expect(DEFAULT_GAS_ADJUSTMENT).toBe(1.3);
  });
});

describe('DEFAULT_GAS_LIMIT', () => {
  it('should be 200000', () => {
    expect(DEFAULT_GAS_LIMIT).toBe(200000);
  });
});

describe('calculateFee', () => {
  const chainInfo = createTestChainInfo();

  it('should calculate fee for average tier', () => {
    const fee = calculateFee(200000, chainInfo, 'average');
    
    expect(fee.gas).toBe('200000');
    expect(fee.amount).toHaveLength(1);
    expect(fee.amount[0].denom).toBe('uvirt');
    // 200000 * 0.025 * 1.3 = 6500
    expect(parseInt(fee.amount[0].amount, 10)).toBe(6500);
  });

  it('should calculate fee for low tier', () => {
    const fee = calculateFee(200000, chainInfo, 'low');
    
    // 200000 * 0.01 * 1.0 = 2000
    expect(parseInt(fee.amount[0].amount, 10)).toBe(2000);
  });

  it('should calculate fee for high tier', () => {
    const fee = calculateFee(200000, chainInfo, 'high');
    
    // 200000 * 0.04 * 2.0 = 16000
    expect(parseInt(fee.amount[0].amount, 10)).toBe(16000);
  });

  it('should default to average tier', () => {
    const fee = calculateFee(200000, chainInfo);
    
    // Same as average tier
    expect(parseInt(fee.amount[0].amount, 10)).toBe(6500);
  });

  it('should handle missing gasPriceStep', () => {
    const chainWithoutGasPrice = createTestChainInfo({
      feeCurrencies: [
        {
          coinDenom: 'VIRT',
          coinMinimalDenom: 'uvirt',
          coinDecimals: 6,
          // No gasPriceStep
        },
      ],
    });
    
    const fee = calculateFee(200000, chainWithoutGasPrice, 'average');
    
    // Uses default: 200000 * 0.025 * 1.3 = 6500
    expect(parseInt(fee.amount[0].amount, 10)).toBe(6500);
  });

  it('should return empty amount if no fee currencies', () => {
    const chainWithoutFees = createTestChainInfo({
      feeCurrencies: [],
    });
    
    const fee = calculateFee(200000, chainWithoutFees);
    
    expect(fee.amount).toEqual([]);
    expect(fee.gas).toBe('200000');
  });

  it('should round up fee amount', () => {
    // 150000 * 0.01 * 1.0 = 1500 (no rounding needed)
    // 150001 * 0.01 * 1.0 = 1500.01 -> should be 1501
    const fee = calculateFee(150001, chainInfo, 'low');
    
    // Math.ceil(150001 * 0.01 * 1.0) = Math.ceil(1500.01) = 1501
    expect(parseInt(fee.amount[0].amount, 10)).toBe(1501);
  });
});

describe('adjustGas', () => {
  it('should apply default adjustment of 1.3', () => {
    const adjusted = adjustGas(100000);
    
    expect(adjusted).toBe(130000);
  });

  it('should apply custom adjustment', () => {
    const adjusted = adjustGas(100000, 1.5);
    
    expect(adjusted).toBe(150000);
  });

  it('should round up adjusted value', () => {
    const adjusted = adjustGas(100001, 1.3);
    
    // 100001 * 1.3 = 130001.3 -> 130002
    expect(adjusted).toBe(130002);
  });

  it('should return DEFAULT_GAS_LIMIT for zero or negative gas', () => {
    expect(adjustGas(0)).toBe(DEFAULT_GAS_LIMIT);
    expect(adjustGas(-100)).toBe(DEFAULT_GAS_LIMIT);
  });

  it('should use default adjustment for invalid adjustment value', () => {
    expect(adjustGas(100000, 0)).toBe(130000);
    expect(adjustGas(100000, -1)).toBe(130000);
  });
});

describe('formatFeeAmount', () => {
  it('should format fee with default 6 decimals', () => {
    const fee: FeeEstimate = {
      amount: [{ denom: 'uvirt', amount: '1000000' }],
      gas: '200000',
    };
    
    const formatted = formatFeeAmount(fee);
    
    expect(formatted).toBe('1.000000 uvirt');
  });

  it('should format fee with custom decimals', () => {
    const fee: FeeEstimate = {
      amount: [{ denom: 'uvirt', amount: '1000000' }],
      gas: '200000',
    };
    
    const formatted = formatFeeAmount(fee, 3);
    
    expect(formatted).toBe('1000.000 uvirt');
  });

  it('should handle multiple denominations', () => {
    const fee: FeeEstimate = {
      amount: [
        { denom: 'uvirt', amount: '1000000' },
        { denom: 'uatom', amount: '500000' },
      ],
      gas: '200000',
    };
    
    const formatted = formatFeeAmount(fee);
    
    expect(formatted).toBe('1.000000 uvirt + 0.500000 uatom');
  });

  it('should return "0" for empty amounts', () => {
    const fee: FeeEstimate = {
      amount: [],
      gas: '200000',
    };
    
    const formatted = formatFeeAmount(fee);
    
    expect(formatted).toBe('0');
  });

  it('should handle NaN amount gracefully', () => {
    const fee: FeeEstimate = {
      amount: [{ denom: 'uvirt', amount: 'invalid' }],
      gas: '200000',
    };
    
    const formatted = formatFeeAmount(fee);
    
    expect(formatted).toBe('0 uvirt');
  });

  it('should format small amounts correctly', () => {
    const fee: FeeEstimate = {
      amount: [{ denom: 'uvirt', amount: '1' }],
      gas: '200000',
    };
    
    const formatted = formatFeeAmount(fee, 6);
    
    expect(formatted).toBe('0.000001 uvirt');
  });
});

describe('validateTransaction', () => {
  const createTestPreview = (overrides: Partial<TransactionPreview> = {}): TransactionPreview => ({
    type: 'MsgSend',
    description: 'Send tokens to another address',
    fee: {
      amount: [{ denom: 'uvirt', amount: '5000' }],
      gas: '200000',
    },
    ...overrides,
  });

  it('should validate a valid transaction', () => {
    const preview = createTestPreview();
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('should error on missing type', () => {
    const preview = createTestPreview({ type: '' });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Transaction type is missing or unknown');
  });

  it('should error on unknown type', () => {
    const preview = createTestPreview({ type: 'unknown' });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Transaction type is missing or unknown');
  });

  it('should error on missing fee', () => {
    const preview = createTestPreview();
    // @ts-expect-error - Testing invalid input
    delete preview.fee;
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Transaction fee is required');
  });

  it('should error on invalid gas limit', () => {
    const preview = createTestPreview({
      fee: { amount: [{ denom: 'uvirt', amount: '5000' }], gas: '0' },
    });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Invalid gas limit in fee');
  });

  it('should error on empty fee amount', () => {
    const preview = createTestPreview({
      fee: { amount: [], gas: '200000' },
    });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Fee amount is required');
  });

  it('should error on missing fee denomination', () => {
    const preview = createTestPreview({
      fee: { amount: [{ denom: '', amount: '5000' }], gas: '200000' },
    });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Fee denomination is required');
  });

  it('should error on negative fee amount', () => {
    const preview = createTestPreview({
      fee: { amount: [{ denom: 'uvirt', amount: '-100' }], gas: '200000' },
    });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Invalid fee amount');
  });

  it('should validate transaction amount if present', () => {
    const preview = createTestPreview({
      amount: { denom: 'uvirt', amount: '1000000' },
      recipient: 'virtengine1abcdefghijklmnopqrstuvwxyz12345678901234',
    });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(true);
  });

  it('should error on invalid amount', () => {
    const preview = createTestPreview({
      amount: { denom: 'uvirt', amount: '0' },
    });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Invalid transaction amount');
  });

  it('should error on missing amount denomination', () => {
    const preview = createTestPreview({
      amount: { denom: '', amount: '1000' },
    });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Amount denomination is required');
  });

  it('should validate recipient address format', () => {
    const preview = createTestPreview({
      recipient: 'virtengine1abcdefghijklmnopqrstuvwxyz12345678901234',
    });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(true);
  });

  it('should error on empty recipient', () => {
    const preview = createTestPreview({
      recipient: '',
    });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Recipient address cannot be empty');
  });

  it('should error on invalid recipient address format', () => {
    const preview = createTestPreview({
      recipient: 'invalid-address',
    });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Invalid recipient address format');
  });

  it('should validate memo length', () => {
    const preview = createTestPreview({
      memo: 'Short memo',
    });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(true);
  });

  it('should error on memo exceeding 256 bytes', () => {
    const preview = createTestPreview({
      memo: 'a'.repeat(300),
    });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Memo exceeds maximum length of 256 bytes');
  });

  it('should collect multiple errors', () => {
    const preview = createTestPreview({
      type: '',
      fee: { amount: [], gas: '0' },
    });
    
    const result = validateTransaction(preview);
    
    expect(result.valid).toBe(false);
    expect(result.errors.length).toBeGreaterThan(1);
  });
});

describe('estimateGas', () => {
  const chainInfo = createTestChainInfo();

  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  it('should return DEFAULT_GAS_LIMIT for empty messages', async () => {
    const gas = await estimateGas([], chainInfo);
    
    expect(gas).toBe(DEFAULT_GAS_LIMIT);
  });

  it('should return DEFAULT_GAS_LIMIT for null messages', async () => {
    const gas = await estimateGas(null as unknown as unknown[], chainInfo);
    
    expect(gas).toBe(DEFAULT_GAS_LIMIT);
  });

  it('should return estimated gas from simulation', async () => {
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        gas_info: { gas_used: '150000' },
      }),
    } as Response);
    
    const msgs = [{ typeUrl: '/cosmos.bank.v1beta1.MsgSend', value: {} }];
    const gas = await estimateGas(msgs, chainInfo);
    
    expect(gas).toBe(150000);
  });

  it('should return DEFAULT_GAS_LIMIT on fetch error', async () => {
    vi.mocked(fetch).mockRejectedValueOnce(new Error('Network error'));
    
    const msgs = [{ typeUrl: '/cosmos.bank.v1beta1.MsgSend', value: {} }];
    const gas = await estimateGas(msgs, chainInfo);
    
    expect(gas).toBe(DEFAULT_GAS_LIMIT);
  });

  it('should return DEFAULT_GAS_LIMIT on non-ok response', async () => {
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: false,
      status: 400,
    } as Response);
    
    const msgs = [{ typeUrl: '/cosmos.bank.v1beta1.MsgSend', value: {} }];
    const gas = await estimateGas(msgs, chainInfo);
    
    expect(gas).toBe(DEFAULT_GAS_LIMIT);
  });

  it('should call correct endpoint', async () => {
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ gas_info: { gas_used: '100000' } }),
    } as Response);
    
    const msgs = [{ typeUrl: '/cosmos.bank.v1beta1.MsgSend', value: {} }];
    await estimateGas(msgs, chainInfo);
    
    expect(fetch).toHaveBeenCalledWith(
      'https://rest.virtengine.io/cosmos/tx/v1beta1/simulate',
      expect.objectContaining({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      })
    );
  });
});

describe('createTransactionPreview', () => {
  const fee: FeeEstimate = {
    amount: [{ denom: 'uvirt', amount: '5000' }],
    gas: '200000',
  };

  it('should return empty array for no messages', () => {
    const previews = createTransactionPreview([], fee);
    
    expect(previews).toEqual([]);
  });

  it('should return empty array for null messages', () => {
    const previews = createTransactionPreview(null as unknown as unknown[], fee);
    
    expect(previews).toEqual([]);
  });

  it('should create preview for MsgSend', () => {
    const msgs = [{
      typeUrl: '/cosmos.bank.v1beta1.MsgSend',
      value: {
        amount: [{ denom: 'uvirt', amount: '1000000' }],
        toAddress: 'virtengine1recipient',
      },
    }];
    
    const previews = createTransactionPreview(msgs, fee);
    
    expect(previews).toHaveLength(1);
    // Type includes the path after the last slash OR full path if no slash
    expect(previews[0].type).toContain('MsgSend');
    expect(previews[0].description).toContain('Send');
    expect(previews[0].amount).toEqual({ denom: 'uvirt', amount: '1000000' });
    expect(previews[0].recipient).toBe('virtengine1recipient');
    expect(previews[0].fee).toBe(fee);
  });

  it('should handle to_address field', () => {
    const msgs = [{
      typeUrl: '/cosmos.bank.v1beta1.MsgSend',
      value: {
        to_address: 'virtengine1recipient',
      },
    }];
    
    const previews = createTransactionPreview(msgs, fee);
    
    expect(previews[0].recipient).toBe('virtengine1recipient');
  });

  it('should handle validatorAddress field', () => {
    const msgs = [{
      typeUrl: '/cosmos.staking.v1beta1.MsgDelegate',
      value: {
        validatorAddress: 'virtenginevaloper1validator',
      },
    }];
    
    const previews = createTransactionPreview(msgs, fee);
    
    expect(previews[0].recipient).toBe('virtenginevaloper1validator');
  });

  it('should create previews for multiple messages', () => {
    const msgs = [
      { typeUrl: '/cosmos.bank.v1beta1.MsgSend', value: {} },
      { typeUrl: '/cosmos.staking.v1beta1.MsgDelegate', value: {} },
    ];
    
    const previews = createTransactionPreview(msgs, fee);
    
    expect(previews).toHaveLength(2);
    expect(previews[0].type).toContain('MsgSend');
    expect(previews[1].type).toContain('MsgDelegate');
  });

  it('should handle unknown message types', () => {
    const msgs = [{
      typeUrl: '/custom.module.v1.MsgCustom',
      value: {},
    }];
    
    const previews = createTransactionPreview(msgs, fee);
    
    expect(previews[0].type).toContain('MsgCustom');
    expect(previews[0].description).toContain('MsgCustom');
  });

  it('should extract memo if present', () => {
    const msgs = [{
      typeUrl: '/cosmos.bank.v1beta1.MsgSend',
      value: {
        memo: 'Test memo',
      },
    }];
    
    const previews = createTransactionPreview(msgs, fee);
    
    expect(previews[0].memo).toBe('Test memo');
  });

  it('should handle single amount object', () => {
    const msgs = [{
      typeUrl: '/cosmos.staking.v1beta1.MsgDelegate',
      value: {
        amount: { denom: 'uvirt', amount: '500000' },
      },
    }];
    
    const previews = createTransactionPreview(msgs, fee);
    
    expect(previews[0].amount).toEqual({ denom: 'uvirt', amount: '500000' });
  });

  it('should get first amount from array', () => {
    const msgs = [{
      typeUrl: '/cosmos.bank.v1beta1.MsgSend',
      value: {
        amount: [
          { denom: 'uvirt', amount: '1000000' },
          { denom: 'uatom', amount: '500000' },
        ],
      },
    }];
    
    const previews = createTransactionPreview(msgs, fee);
    
    expect(previews[0].amount).toEqual({ denom: 'uvirt', amount: '1000000' });
  });
});

describe('createDefaultGasSettings', () => {
  const chainInfo = createTestChainInfo();

  it('should create settings with default gas limit', () => {
    const settings = createDefaultGasSettings(chainInfo);
    
    expect(settings.gasLimit).toBe(DEFAULT_GAS_LIMIT);
  });

  it('should create settings with default adjustment', () => {
    const settings = createDefaultGasSettings(chainInfo);
    
    expect(settings.gasAdjustment).toBe(DEFAULT_GAS_ADJUSTMENT);
  });

  it('should use gas price from chain info for average tier', () => {
    const settings = createDefaultGasSettings(chainInfo, 'average');
    
    expect(settings.gasPrice).toBe(0.025);
  });

  it('should use gas price for low tier', () => {
    const settings = createDefaultGasSettings(chainInfo, 'low');
    
    expect(settings.gasPrice).toBe(0.01);
  });

  it('should use gas price for high tier', () => {
    const settings = createDefaultGasSettings(chainInfo, 'high');
    
    expect(settings.gasPrice).toBe(0.04);
  });

  it('should use default prices when gasPriceStep not set', () => {
    const chainWithoutGasPrice = createTestChainInfo({
      feeCurrencies: [
        {
          coinDenom: 'VIRT',
          coinMinimalDenom: 'uvirt',
          coinDecimals: 6,
        },
      ],
    });
    
    const settings = createDefaultGasSettings(chainWithoutGasPrice, 'average');
    
    expect(settings.gasPrice).toBe(0.025);
  });
});

import { describe, it, expect } from 'vitest';
import {
  calculatePriceBreakdown,
  durationToHours,
  validateResources,
  formatTokenAmount,
  DEFAULT_RESOURCE_CONFIG,
  RESOURCE_LIMITS,
} from '@/features/orders/types';
import type { PriceComponent } from '@/types/offerings';

describe('Order Wizard Types & Utilities', () => {
  describe('durationToHours', () => {
    it('converts hours to hours', () => {
      expect(durationToHours(5, 'hours')).toBe(5);
    });

    it('converts days to hours', () => {
      expect(durationToHours(2, 'days')).toBe(48);
    });

    it('converts months to hours', () => {
      expect(durationToHours(1, 'months')).toBe(720);
    });

    it('handles zero duration', () => {
      expect(durationToHours(0, 'hours')).toBe(0);
    });
  });

  describe('formatTokenAmount', () => {
    it('formats with default 2 decimals', () => {
      expect(formatTokenAmount(1234.5678)).toBe('1,234.57');
    });

    it('formats with custom decimals', () => {
      expect(formatTokenAmount(0.001234, 4)).toBe('0.0012');
    });

    it('formats zero', () => {
      expect(formatTokenAmount(0)).toBe('0.00');
    });

    it('formats large numbers', () => {
      expect(formatTokenAmount(1000000)).toBe('1,000,000.00');
    });
  });

  describe('validateResources', () => {
    it('returns no errors for valid default config', () => {
      const errors = validateResources(DEFAULT_RESOURCE_CONFIG);
      expect(errors).toHaveLength(0);
    });

    it('rejects cpu below minimum', () => {
      const errors = validateResources({ ...DEFAULT_RESOURCE_CONFIG, cpu: 0 });
      expect(errors.length).toBeGreaterThan(0);
      expect(errors[0]).toContain('CPU');
    });

    it('rejects cpu above maximum', () => {
      const errors = validateResources({ ...DEFAULT_RESOURCE_CONFIG, cpu: 999 });
      expect(errors.length).toBeGreaterThan(0);
      expect(errors[0]).toContain('CPU');
    });

    it('rejects memory below minimum', () => {
      const errors = validateResources({ ...DEFAULT_RESOURCE_CONFIG, memory: 0 });
      expect(errors.length).toBeGreaterThan(0);
      expect(errors.some((e) => e.includes('Memory'))).toBe(true);
    });

    it('rejects storage below minimum', () => {
      const errors = validateResources({ ...DEFAULT_RESOURCE_CONFIG, storage: 1 });
      expect(errors.length).toBeGreaterThan(0);
      expect(errors.some((e) => e.includes('Storage'))).toBe(true);
    });

    it('rejects gpu below minimum', () => {
      const errors = validateResources({ ...DEFAULT_RESOURCE_CONFIG, gpu: -1 });
      expect(errors.length).toBeGreaterThan(0);
      expect(errors.some((e) => e.includes('GPU'))).toBe(true);
    });

    it('accepts boundary values', () => {
      const config = {
        ...DEFAULT_RESOURCE_CONFIG,
        cpu: RESOURCE_LIMITS.cpu.min,
        memory: RESOURCE_LIMITS.memory.min,
        storage: RESOURCE_LIMITS.storage.min,
        gpu: RESOURCE_LIMITS.gpu.min,
        duration: RESOURCE_LIMITS.duration.min,
      };
      expect(validateResources(config)).toHaveLength(0);
    });

    it('accepts max boundary values', () => {
      const config = {
        ...DEFAULT_RESOURCE_CONFIG,
        cpu: RESOURCE_LIMITS.cpu.max,
        memory: RESOURCE_LIMITS.memory.max,
        storage: RESOURCE_LIMITS.storage.max,
        gpu: RESOURCE_LIMITS.gpu.max,
        duration: RESOURCE_LIMITS.duration.max,
      };
      expect(validateResources(config)).toHaveLength(0);
    });
  });

  describe('calculatePriceBreakdown', () => {
    const mockPrices: PriceComponent[] = [
      {
        resourceType: 'cpu',
        unit: 'vcpu-hour',
        price: { denom: 'uve', amount: '10000' },
        usdReference: '0.01',
      },
      {
        resourceType: 'ram',
        unit: 'gb-hour',
        price: { denom: 'uve', amount: '5000' },
        usdReference: '0.005',
      },
      {
        resourceType: 'gpu',
        unit: 'hour',
        price: { denom: 'uve', amount: '2500000' },
        usdReference: '2.50',
      },
      {
        resourceType: 'storage',
        unit: 'gb-month',
        price: { denom: 'uve', amount: '100000' },
        usdReference: '0.10',
      },
    ];

    it('calculates breakdown for basic config', () => {
      const resources = {
        ...DEFAULT_RESOURCE_CONFIG,
        cpu: 4,
        memory: 16,
        storage: 100,
        gpu: 0,
        duration: 1,
        durationUnit: 'hours' as const,
        region: 'us-west',
      };

      const result = calculatePriceBreakdown(mockPrices, resources);

      expect(result.items.length).toBeGreaterThan(0);
      expect(result.subtotal).toBeGreaterThan(0);
      expect(result.escrowDeposit).toBeGreaterThan(0);
      expect(result.currency).toBe('VE');
      expect(result.denom).toBe('uve');
    });

    it('includes GPU pricing when GPU is configured', () => {
      const resources = {
        ...DEFAULT_RESOURCE_CONFIG,
        gpu: 2,
        duration: 1,
        durationUnit: 'hours' as const,
        region: 'us-west',
      };

      const result = calculatePriceBreakdown(mockPrices, resources);
      const gpuItem = result.items.find((item) => item.resourceType === 'gpu');

      expect(gpuItem).toBeDefined();
      expect(gpuItem!.quantity).toBe(2); // 2 GPU * 1 hour
      expect(gpuItem!.total).toBeGreaterThan(0);
    });

    it('excludes GPU pricing when GPU is zero', () => {
      const resources = {
        ...DEFAULT_RESOURCE_CONFIG,
        gpu: 0,
        duration: 1,
        durationUnit: 'hours' as const,
        region: 'us-west',
      };

      const result = calculatePriceBreakdown(mockPrices, resources);
      const gpuItem = result.items.find((item) => item.resourceType === 'gpu');

      expect(gpuItem).toBeUndefined();
    });

    it('scales cost with duration', () => {
      const resources1h = {
        ...DEFAULT_RESOURCE_CONFIG,
        cpu: 4,
        memory: 16,
        gpu: 0,
        duration: 1,
        durationUnit: 'hours' as const,
        region: 'us-west',
      };

      const resources24h = {
        ...resources1h,
        duration: 24,
      };

      const result1h = calculatePriceBreakdown(mockPrices, resources1h);
      const result24h = calculatePriceBreakdown(mockPrices, resources24h);

      // CPU+RAM cost should scale linearly with hours
      const cpuRam1h = result1h.items
        .filter((i) => i.resourceType === 'cpu' || i.resourceType === 'ram')
        .reduce((acc, i) => acc + i.total, 0);
      const cpuRam24h = result24h.items
        .filter((i) => i.resourceType === 'cpu' || i.resourceType === 'ram')
        .reduce((acc, i) => acc + i.total, 0);

      expect(cpuRam24h).toBeCloseTo(cpuRam1h * 24, 4);
    });

    it('returns empty items for empty prices array', () => {
      const result = calculatePriceBreakdown([], DEFAULT_RESOURCE_CONFIG);
      expect(result.items).toHaveLength(0);
      expect(result.subtotal).toBe(0);
    });

    it('escrow deposit is at least 10% of subtotal', () => {
      const resources = {
        ...DEFAULT_RESOURCE_CONFIG,
        cpu: 4,
        memory: 16,
        gpu: 0,
        duration: 100,
        durationUnit: 'hours' as const,
        region: 'us-west',
      };

      const result = calculatePriceBreakdown(mockPrices, resources);

      // Escrow is max of (hourly rate, 10% of subtotal)
      expect(result.escrowDeposit).toBeGreaterThanOrEqual(result.subtotal * 0.1);
    });

    it('unit prices are correctly derived from micro denomination', () => {
      const resources = {
        ...DEFAULT_RESOURCE_CONFIG,
        cpu: 1,
        memory: 0,
        storage: 0,
        gpu: 0,
        duration: 1,
        durationUnit: 'hours' as const,
        region: 'us-west',
      };

      const simplePrices: PriceComponent[] = [
        {
          resourceType: 'cpu',
          unit: 'vcpu-hour',
          price: { denom: 'uve', amount: '1000000' }, // 1 VE
          usdReference: '1.00',
        },
      ];

      const result = calculatePriceBreakdown(simplePrices, resources);
      expect(result.items[0].unitPrice).toBe(1); // 1,000,000 micro / 1,000,000 = 1 VE
      expect(result.items[0].total).toBe(1); // 1 vCPU * 1 hour * 1 VE/vcpu-hour
    });
  });

  describe('DEFAULT_RESOURCE_CONFIG', () => {
    it('has valid default values', () => {
      expect(DEFAULT_RESOURCE_CONFIG.cpu).toBeGreaterThanOrEqual(RESOURCE_LIMITS.cpu.min);
      expect(DEFAULT_RESOURCE_CONFIG.memory).toBeGreaterThanOrEqual(RESOURCE_LIMITS.memory.min);
      expect(DEFAULT_RESOURCE_CONFIG.storage).toBeGreaterThanOrEqual(RESOURCE_LIMITS.storage.min);
      expect(DEFAULT_RESOURCE_CONFIG.gpu).toBeGreaterThanOrEqual(RESOURCE_LIMITS.gpu.min);
      expect(DEFAULT_RESOURCE_CONFIG.duration).toBeGreaterThanOrEqual(RESOURCE_LIMITS.duration.min);
    });

    it('has sensible defaults', () => {
      expect(DEFAULT_RESOURCE_CONFIG.cpu).toBe(4);
      expect(DEFAULT_RESOURCE_CONFIG.memory).toBe(16);
      expect(DEFAULT_RESOURCE_CONFIG.storage).toBe(100);
      expect(DEFAULT_RESOURCE_CONFIG.gpu).toBe(0);
      expect(DEFAULT_RESOURCE_CONFIG.durationUnit).toBe('hours');
    });
  });
});

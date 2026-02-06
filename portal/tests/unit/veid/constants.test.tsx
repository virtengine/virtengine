import { describe, it, expect } from 'vitest';
import {
  TIER_INFO,
  STATUS_DISPLAY,
  SCOPE_DISPLAY,
  WIZARD_STEPS,
  FEATURE_THRESHOLDS,
} from '@/features/veid/constants';

describe('VEID Constants', () => {
  describe('TIER_INFO', () => {
    it('has all tier levels defined', () => {
      expect(TIER_INFO).toHaveProperty('unverified');
      expect(TIER_INFO).toHaveProperty('basic');
      expect(TIER_INFO).toHaveProperty('standard');
      expect(TIER_INFO).toHaveProperty('premium');
      expect(TIER_INFO).toHaveProperty('elite');
    });

    it('tiers have correct score ranges', () => {
      expect(TIER_INFO.unverified.minScore).toBe(0);
      expect(TIER_INFO.basic.minScore).toBe(1);
      expect(TIER_INFO.basic.maxScore).toBe(40);
      expect(TIER_INFO.standard.minScore).toBe(41);
      expect(TIER_INFO.standard.maxScore).toBe(70);
      expect(TIER_INFO.premium.minScore).toBe(71);
      expect(TIER_INFO.premium.maxScore).toBe(90);
      expect(TIER_INFO.elite.minScore).toBe(91);
      expect(TIER_INFO.elite.maxScore).toBe(100);
    });

    it('each tier has required display properties', () => {
      Object.values(TIER_INFO).forEach((tier) => {
        expect(tier.label).toBeTruthy();
        expect(tier.description).toBeTruthy();
        expect(tier.color).toBeTruthy();
        expect(tier.bgColor).toBeTruthy();
        expect(tier.icon).toBeTruthy();
      });
    });
  });

  describe('STATUS_DISPLAY', () => {
    it('has all identity statuses defined', () => {
      expect(STATUS_DISPLAY).toHaveProperty('unknown');
      expect(STATUS_DISPLAY).toHaveProperty('pending');
      expect(STATUS_DISPLAY).toHaveProperty('processing');
      expect(STATUS_DISPLAY).toHaveProperty('verified');
      expect(STATUS_DISPLAY).toHaveProperty('rejected');
      expect(STATUS_DISPLAY).toHaveProperty('expired');
    });

    it('pending and processing show progress', () => {
      expect(STATUS_DISPLAY.pending.showProgress).toBe(true);
      expect(STATUS_DISPLAY.processing.showProgress).toBe(true);
    });

    it('verified does not show progress', () => {
      expect(STATUS_DISPLAY.verified.showProgress).toBe(false);
    });
  });

  describe('SCOPE_DISPLAY', () => {
    it('has all scope types defined', () => {
      expect(SCOPE_DISPLAY).toHaveProperty('email');
      expect(SCOPE_DISPLAY).toHaveProperty('id_document');
      expect(SCOPE_DISPLAY).toHaveProperty('selfie');
      expect(SCOPE_DISPLAY).toHaveProperty('sso');
      expect(SCOPE_DISPLAY).toHaveProperty('domain');
      expect(SCOPE_DISPLAY).toHaveProperty('biometric');
    });

    it('each scope has points > 0', () => {
      Object.values(SCOPE_DISPLAY).forEach((scope) => {
        expect(scope.points).toBeGreaterThan(0);
      });
    });
  });

  describe('WIZARD_STEPS', () => {
    it('has steps in order', () => {
      for (let i = 1; i < WIZARD_STEPS.length; i++) {
        expect(WIZARD_STEPS[i]!.order).toBeGreaterThan(WIZARD_STEPS[i - 1]!.order);
      }
    });

    it('includes welcome and complete steps', () => {
      const keys = WIZARD_STEPS.map((s) => s.key);
      expect(keys).toContain('welcome');
      expect(keys).toContain('complete');
    });
  });

  describe('FEATURE_THRESHOLDS', () => {
    it('has thresholds in ascending score order', () => {
      for (let i = 1; i < FEATURE_THRESHOLDS.length; i++) {
        expect(FEATURE_THRESHOLDS[i]!.minScore).toBeGreaterThanOrEqual(
          FEATURE_THRESHOLDS[i - 1]!.minScore
        );
      }
    });

    it('browse_offerings requires score 0', () => {
      const browse = FEATURE_THRESHOLDS.find((t) => t.action === 'browse_offerings');
      expect(browse?.minScore).toBe(0);
    });
  });
});

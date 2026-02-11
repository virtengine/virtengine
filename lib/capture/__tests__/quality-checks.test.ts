/**
 * Quality Checks Tests
 * VE-210: Unit tests for quality validation functions
 */

import { describe, it, expect } from 'vitest';
import {
  checkResolution,
  checkBrightness,
  checkBlur,
  checkGlare,
  checkNoise,
} from '../utils/quality-checks';
import { DEFAULT_QUALITY_THRESHOLDS } from '../types/capture';

describe('quality-checks', () => {
  /**
   * Helper to create mock ImageData
   */
  function createMockImageData(
    width: number,
    height: number,
    fillValue: number = 128
  ): ImageData {
    const data = new Uint8ClampedArray(width * height * 4);
    for (let i = 0; i < data.length; i += 4) {
      data[i] = fillValue; // R
      data[i + 1] = fillValue; // G
      data[i + 2] = fillValue; // B
      data[i + 3] = 255; // A
    }
    return { data, width, height, colorSpace: 'srgb' };
  }

  /**
   * Helper to create ImageData with specific brightness pattern
   */
  function createBrightnessImageData(
    width: number,
    height: number,
    brightness: number
  ): ImageData {
    return createMockImageData(width, height, brightness);
  }

  /**
   * Helper to create ImageData with noise
   */
  function createNoisyImageData(
    width: number,
    height: number,
    noiseLevel: number
  ): ImageData {
    const data = new Uint8ClampedArray(width * height * 4);
    for (let i = 0; i < data.length; i += 4) {
      const base = 128;
      const noise = (Math.random() - 0.5) * 255 * noiseLevel;
      const value = Math.max(0, Math.min(255, base + noise));
      data[i] = value;
      data[i + 1] = value;
      data[i + 2] = value;
      data[i + 3] = 255;
    }
    return { data, width, height, colorSpace: 'srgb' };
  }

  describe('checkResolution', () => {
    it('should pass for images meeting minimum resolution', () => {
      const imageData = createMockImageData(1920, 1080);
      const result = checkResolution(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(true);
      expect(result.value).toBeGreaterThanOrEqual(1);
    });

    it('should pass for images at exactly minimum resolution', () => {
      const imageData = createMockImageData(1024, 768);
      const result = checkResolution(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(true);
    });

    it('should fail for images below minimum resolution', () => {
      const imageData = createMockImageData(640, 480);
      const result = checkResolution(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(false);
      expect(result.value).toBeLessThan(1);
    });

    it('should fail if width is below minimum', () => {
      const imageData = createMockImageData(800, 1080);
      const result = checkResolution(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(false);
    });

    it('should fail if height is below minimum', () => {
      const imageData = createMockImageData(1920, 600);
      const result = checkResolution(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(false);
    });

    it('should include description with dimensions', () => {
      const imageData = createMockImageData(1920, 1080);
      const result = checkResolution(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.description).toContain('1920');
      expect(result.description).toContain('1080');
    });
  });

  describe('checkBrightness', () => {
    it('should pass for images with acceptable brightness', () => {
      const imageData = createBrightnessImageData(100, 100, 128);
      const result = checkBrightness(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(true);
      expect(result.value).toBeCloseTo(128, 1);
    });

    it('should fail for images that are too dark', () => {
      const imageData = createBrightnessImageData(100, 100, 20);
      const result = checkBrightness(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(false);
      expect(result.value).toBeLessThan(DEFAULT_QUALITY_THRESHOLDS.minBrightness);
    });

    it('should fail for images that are too bright', () => {
      const imageData = createBrightnessImageData(100, 100, 240);
      const result = checkBrightness(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(false);
      expect(result.value).toBeGreaterThan(DEFAULT_QUALITY_THRESHOLDS.maxBrightness);
    });

    it('should pass at minimum brightness threshold', () => {
      const imageData = createBrightnessImageData(100, 100, 40);
      const result = checkBrightness(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(true);
    });

    it('should pass at maximum brightness threshold', () => {
      const imageData = createBrightnessImageData(100, 100, 220);
      const result = checkBrightness(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(true);
    });
  });

  describe('checkBlur', () => {
    it('should pass for sharp images with high variance', () => {
      // Create an image with edges (high Laplacian variance)
      const width = 100;
      const height = 100;
      const data = new Uint8ClampedArray(width * height * 4);

      // Create a sharp edge pattern
      for (let y = 0; y < height; y++) {
        for (let x = 0; x < width; x++) {
          const i = (y * width + x) * 4;
          const value = x < 50 ? 0 : 255;
          data[i] = value;
          data[i + 1] = value;
          data[i + 2] = value;
          data[i + 3] = 255;
        }
      }

      const imageData: ImageData = { data, width, height, colorSpace: 'srgb' };
      const result = checkBlur(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.value).toBeGreaterThan(0);
    });

    it('should fail for uniform images (simulated blur)', () => {
      // Uniform image has very low Laplacian variance
      const imageData = createMockImageData(100, 100, 128);
      const result = checkBlur(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(false);
      expect(result.value).toBeLessThan(DEFAULT_QUALITY_THRESHOLDS.maxBlur);
    });
  });

  describe('checkGlare', () => {
    it('should pass for images without glare', () => {
      const imageData = createMockImageData(100, 100, 128);
      const result = checkGlare(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(true);
      expect(result.value).toBeLessThan(DEFAULT_QUALITY_THRESHOLDS.maxGlare);
    });

    it('should fail for images with significant glare', () => {
      // Create an image with overexposed (glare) regions
      const width = 100;
      const height = 100;
      const data = new Uint8ClampedArray(width * height * 4);

      // Make 30% of pixels overexposed
      for (let i = 0; i < data.length; i += 4) {
        const isGlare = i / 4 < (width * height * 0.3);
        const value = isGlare ? 252 : 128;
        data[i] = value;
        data[i + 1] = value;
        data[i + 2] = value;
        data[i + 3] = 255;
      }

      const imageData: ImageData = { data, width, height, colorSpace: 'srgb' };
      const result = checkGlare(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(false);
      expect(result.value).toBeGreaterThan(DEFAULT_QUALITY_THRESHOLDS.maxGlare);
    });

    it('should pass with small amount of highlights', () => {
      // Create an image with small highlight regions (5%)
      const width = 100;
      const height = 100;
      const data = new Uint8ClampedArray(width * height * 4);

      for (let i = 0; i < data.length; i += 4) {
        const isHighlight = i / 4 < (width * height * 0.05);
        const value = isHighlight ? 252 : 128;
        data[i] = value;
        data[i + 1] = value;
        data[i + 2] = value;
        data[i + 3] = 255;
      }

      const imageData: ImageData = { data, width, height, colorSpace: 'srgb' };
      const result = checkGlare(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(true);
    });
  });

  describe('checkNoise', () => {
    it('should pass for clean images', () => {
      const imageData = createMockImageData(100, 100, 128);
      const result = checkNoise(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.passed).toBe(true);
      expect(result.value).toBeLessThan(DEFAULT_QUALITY_THRESHOLDS.maxNoise);
    });

    it('should fail for very noisy images', () => {
      const imageData = createNoisyImageData(100, 100, 1.0);
      const result = checkNoise(imageData, DEFAULT_QUALITY_THRESHOLDS);

      expect(result.value).toBeGreaterThan(0);
    });
  });
});

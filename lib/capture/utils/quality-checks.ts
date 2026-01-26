/**
 * Quality Check Implementations
 * VE-210: Image quality validation for document and selfie capture
 *
 * Implements client-side quality checks:
 * - Resolution check (minimum 1024x768)
 * - Brightness check (histogram analysis)
 * - Blur detection (Laplacian variance)
 * - Skew/perspective detection (edge detection)
 * - Glare detection (highlight analysis)
 * - Noise detection
 */

import type {
  QualityCheckResult,
  QualityCheckDetail,
  QualityIssue,
  QualityThresholds,
} from '../types/capture';
import { DEFAULT_QUALITY_THRESHOLDS } from '../types/capture';

/**
 * Perform all quality checks on an image
 *
 * @param imageData - ImageData from canvas
 * @param thresholds - Quality thresholds to use
 * @returns Quality check result
 */
export async function performQualityChecks(
  imageData: ImageData,
  thresholds: QualityThresholds = DEFAULT_QUALITY_THRESHOLDS
): Promise<QualityCheckResult> {
  const startTime = performance.now();
  const issues: QualityIssue[] = [];

  // Perform individual checks
  const resolutionCheck = checkResolution(imageData, thresholds);
  const brightnessCheck = checkBrightness(imageData, thresholds);
  const blurCheck = checkBlur(imageData, thresholds);
  const skewCheck = await checkSkew(imageData, thresholds);
  const glareCheck = checkGlare(imageData, thresholds);
  const noiseCheck = checkNoise(imageData, thresholds);

  // Collect issues from each check
  if (!resolutionCheck.passed) {
    issues.push({
      type: 'resolution',
      severity: 'error',
      message: `Image resolution is too low (${imageData.width}x${imageData.height})`,
      suggestion: `Move closer to capture at least ${thresholds.minResolution.width}x${thresholds.minResolution.height} pixels`,
      confidence: 1.0,
    });
  }

  if (!brightnessCheck.passed) {
    if (brightnessCheck.value < thresholds.minBrightness) {
      issues.push({
        type: 'dark',
        severity: 'error',
        message: 'Image is too dark',
        suggestion: 'Move to a brighter area or turn on more lights',
        confidence: 0.9,
      });
    } else {
      issues.push({
        type: 'bright',
        severity: 'error',
        message: 'Image is overexposed',
        suggestion: 'Move away from direct light sources or reduce lighting',
        confidence: 0.9,
      });
    }
  }

  if (!blurCheck.passed) {
    issues.push({
      type: 'blur',
      severity: 'error',
      message: 'Image is blurry',
      suggestion: 'Hold the device steady and ensure the document is in focus',
      confidence: blurCheck.value / thresholds.maxBlur,
    });
  }

  if (!skewCheck.passed) {
    issues.push({
      type: 'skew',
      severity: 'warning',
      message: `Document appears tilted (${skewCheck.value.toFixed(1)}°)`,
      suggestion: 'Hold the camera parallel to the document surface',
      confidence: 0.8,
    });
  }

  if (!glareCheck.passed) {
    issues.push({
      type: 'glare',
      severity: 'warning',
      message: 'Glare detected on the document',
      suggestion: 'Tilt the document or camera to reduce reflections',
      confidence: glareCheck.value,
    });
  }

  if (!noiseCheck.passed) {
    issues.push({
      type: 'noise',
      severity: 'warning',
      message: 'Image has excessive noise',
      suggestion: 'Improve lighting conditions for clearer capture',
      confidence: noiseCheck.value,
    });
  }

  // Calculate overall score
  const checkResults = [resolutionCheck, brightnessCheck, blurCheck, skewCheck, glareCheck, noiseCheck];
  const weights = [0.2, 0.2, 0.25, 0.15, 0.1, 0.1];
  let score = 0;

  for (let i = 0; i < checkResults.length; i++) {
    const check = checkResults[i];
    const weight = weights[i];
    // Normalize value to 0-1 score
    let checkScore = check.passed ? 1.0 : Math.max(0, check.value / check.threshold);
    checkScore = Math.min(1.0, checkScore);
    score += checkScore * weight * 100;
  }

  score = Math.round(score);

  const analysisTimeMs = performance.now() - startTime;
  const hasErrors = issues.some((i) => i.severity === 'error');
  const passed = !hasErrors && score >= thresholds.minScore;

  return {
    passed,
    score,
    issues,
    checks: {
      resolution: resolutionCheck,
      brightness: brightnessCheck,
      blur: blurCheck,
      skew: skewCheck,
      glare: glareCheck,
      noise: noiseCheck,
    },
    analysisTimeMs,
  };
}

/**
 * Check image resolution
 */
export function checkResolution(
  imageData: ImageData,
  thresholds: QualityThresholds
): QualityCheckDetail {
  const { width, height } = imageData;
  const minWidth = thresholds.minResolution.width;
  const minHeight = thresholds.minResolution.height;

  const passed = width >= minWidth && height >= minHeight;
  const value = Math.min(width / minWidth, height / minHeight);

  return {
    passed,
    value,
    threshold: 1.0,
    description: `Resolution: ${width}x${height} (min: ${minWidth}x${minHeight})`,
  };
}

/**
 * Check image brightness using histogram analysis
 */
export function checkBrightness(
  imageData: ImageData,
  thresholds: QualityThresholds
): QualityCheckDetail {
  const { data } = imageData;
  let totalBrightness = 0;
  const pixelCount = data.length / 4;

  // Calculate average brightness (luminance)
  for (let i = 0; i < data.length; i += 4) {
    const r = data[i];
    const g = data[i + 1];
    const b = data[i + 2];
    // Use standard luminance formula
    const brightness = 0.299 * r + 0.587 * g + 0.114 * b;
    totalBrightness += brightness;
  }

  const averageBrightness = totalBrightness / pixelCount;
  const passed =
    averageBrightness >= thresholds.minBrightness &&
    averageBrightness <= thresholds.maxBrightness;

  return {
    passed,
    value: averageBrightness,
    threshold: (thresholds.minBrightness + thresholds.maxBrightness) / 2,
    description: `Brightness: ${averageBrightness.toFixed(1)} (range: ${thresholds.minBrightness}-${thresholds.maxBrightness})`,
  };
}

/**
 * Check image blur using Laplacian variance
 * Higher variance = sharper image
 */
export function checkBlur(
  imageData: ImageData,
  thresholds: QualityThresholds
): QualityCheckDetail {
  const { data, width, height } = imageData;

  // Convert to grayscale and apply Laplacian kernel
  const grayscale = new Float32Array(width * height);
  for (let i = 0; i < data.length; i += 4) {
    const idx = i / 4;
    grayscale[idx] = 0.299 * data[i] + 0.587 * data[i + 1] + 0.114 * data[i + 2];
  }

  // Laplacian kernel: [0, 1, 0], [1, -4, 1], [0, 1, 0]
  let sum = 0;
  let sumSq = 0;
  let count = 0;

  for (let y = 1; y < height - 1; y++) {
    for (let x = 1; x < width - 1; x++) {
      const idx = y * width + x;
      const laplacian =
        grayscale[idx - width] +
        grayscale[idx - 1] +
        -4 * grayscale[idx] +
        grayscale[idx + 1] +
        grayscale[idx + width];

      sum += laplacian;
      sumSq += laplacian * laplacian;
      count++;
    }
  }

  const mean = sum / count;
  const variance = sumSq / count - mean * mean;

  // Higher variance means sharper image
  // Low variance indicates blur
  const passed = variance >= thresholds.maxBlur;

  return {
    passed,
    value: variance,
    threshold: thresholds.maxBlur,
    description: `Blur (Laplacian variance): ${variance.toFixed(2)} (min: ${thresholds.maxBlur})`,
  };
}

/**
 * Check document skew using edge detection
 * Uses Sobel operator to find dominant edge angles
 */
export async function checkSkew(
  imageData: ImageData,
  thresholds: QualityThresholds
): Promise<QualityCheckDetail> {
  const { data, width, height } = imageData;

  // Convert to grayscale
  const grayscale = new Float32Array(width * height);
  for (let i = 0; i < data.length; i += 4) {
    const idx = i / 4;
    grayscale[idx] = 0.299 * data[i] + 0.587 * data[i + 1] + 0.114 * data[i + 2];
  }

  // Apply Sobel operator
  const gradientAngles: number[] = [];

  for (let y = 1; y < height - 1; y++) {
    for (let x = 1; x < width - 1; x++) {
      const idx = y * width + x;

      // Sobel X kernel
      const gx =
        -grayscale[idx - width - 1] +
        grayscale[idx - width + 1] +
        -2 * grayscale[idx - 1] +
        2 * grayscale[idx + 1] +
        -grayscale[idx + width - 1] +
        grayscale[idx + width + 1];

      // Sobel Y kernel
      const gy =
        -grayscale[idx - width - 1] +
        -2 * grayscale[idx - width] +
        -grayscale[idx - width + 1] +
        grayscale[idx + width - 1] +
        2 * grayscale[idx + width] +
        grayscale[idx + width + 1];

      const magnitude = Math.sqrt(gx * gx + gy * gy);

      // Only consider significant edges
      if (magnitude > 50) {
        const angle = Math.atan2(gy, gx) * (180 / Math.PI);
        gradientAngles.push(angle);
      }
    }
  }

  // Find dominant angle using histogram
  const angleBins = new Array(180).fill(0);
  for (const angle of gradientAngles) {
    // Normalize to 0-180 (perpendicular edges should group together)
    let normalizedAngle = angle < 0 ? angle + 180 : angle;
    normalizedAngle = normalizedAngle >= 180 ? normalizedAngle - 180 : normalizedAngle;
    const bin = Math.floor(normalizedAngle);
    angleBins[bin]++;
  }

  // Find the peak angle (should be near 0 or 90 for aligned documents)
  let maxBin = 0;
  let maxCount = 0;
  for (let i = 0; i < angleBins.length; i++) {
    if (angleBins[i] > maxCount) {
      maxCount = angleBins[i];
      maxBin = i;
    }
  }

  // Calculate skew from nearest axis (0 or 90 degrees)
  let skewAngle = maxBin;
  if (skewAngle > 45) skewAngle = 90 - skewAngle;

  const passed = Math.abs(skewAngle) <= thresholds.maxSkew;

  return {
    passed,
    value: skewAngle,
    threshold: thresholds.maxSkew,
    description: `Skew angle: ${skewAngle.toFixed(1)}° (max: ${thresholds.maxSkew}°)`,
  };
}

/**
 * Check for glare/specular highlights
 */
export function checkGlare(
  imageData: ImageData,
  thresholds: QualityThresholds
): QualityCheckDetail {
  const { data } = imageData;
  const pixelCount = data.length / 4;
  let highlightCount = 0;

  // Count overexposed pixels (very bright, near white)
  const highlightThreshold = 250;

  for (let i = 0; i < data.length; i += 4) {
    const r = data[i];
    const g = data[i + 1];
    const b = data[i + 2];

    // Check if all channels are very bright (indicates specular highlight)
    if (r > highlightThreshold && g > highlightThreshold && b > highlightThreshold) {
      highlightCount++;
    }
  }

  const glarePercentage = highlightCount / pixelCount;
  const passed = glarePercentage <= thresholds.maxGlare;

  return {
    passed,
    value: glarePercentage,
    threshold: thresholds.maxGlare,
    description: `Glare: ${(glarePercentage * 100).toFixed(2)}% (max: ${thresholds.maxGlare * 100}%)`,
  };
}

/**
 * Check for image noise
 */
export function checkNoise(
  imageData: ImageData,
  thresholds: QualityThresholds
): QualityCheckDetail {
  const { data, width, height } = imageData;

  // Estimate noise using local variance in a 3x3 neighborhood
  let totalVariance = 0;
  let count = 0;

  for (let y = 1; y < height - 1; y += 2) {
    // Sample every other row for performance
    for (let x = 1; x < width - 1; x += 2) {
      const idx = (y * width + x) * 4;

      // Get 3x3 neighborhood luminance values
      const neighbors: number[] = [];
      for (let dy = -1; dy <= 1; dy++) {
        for (let dx = -1; dx <= 1; dx++) {
          const nIdx = ((y + dy) * width + (x + dx)) * 4;
          const lum = 0.299 * data[nIdx] + 0.587 * data[nIdx + 1] + 0.114 * data[nIdx + 2];
          neighbors.push(lum);
        }
      }

      // Calculate local variance
      const mean = neighbors.reduce((a, b) => a + b) / neighbors.length;
      const variance = neighbors.reduce((a, b) => a + (b - mean) ** 2, 0) / neighbors.length;
      totalVariance += variance;
      count++;
    }
  }

  // Normalize variance (high variance in low-contrast areas indicates noise)
  const averageVariance = totalVariance / count;
  const normalizedNoise = Math.min(1, averageVariance / 500); // Normalize to 0-1

  const passed = normalizedNoise <= thresholds.maxNoise;

  return {
    passed,
    value: normalizedNoise,
    threshold: thresholds.maxNoise,
    description: `Noise level: ${(normalizedNoise * 100).toFixed(1)}% (max: ${thresholds.maxNoise * 100}%)`,
  };
}

/**
 * Perform a quick quality assessment for real-time feedback
 * This is a lighter-weight version for the live preview
 */
export function quickQualityAssessment(
  imageData: ImageData,
  thresholds: QualityThresholds = DEFAULT_QUALITY_THRESHOLDS
): { acceptable: boolean; mainIssue: string | null } {
  // Quick brightness check
  const brightnessCheck = checkBrightness(imageData, thresholds);
  if (!brightnessCheck.passed) {
    if (brightnessCheck.value < thresholds.minBrightness) {
      return { acceptable: false, mainIssue: 'Too dark - improve lighting' };
    }
    return { acceptable: false, mainIssue: 'Too bright - reduce lighting' };
  }

  // Quick blur check (simplified)
  const blurCheck = checkBlur(imageData, thresholds);
  if (!blurCheck.passed) {
    return { acceptable: false, mainIssue: 'Hold steady - image is blurry' };
  }

  // Quick glare check
  const glareCheck = checkGlare(imageData, thresholds);
  if (!glareCheck.passed) {
    return { acceptable: false, mainIssue: 'Glare detected - tilt device' };
  }

  return { acceptable: true, mainIssue: null };
}

/**
 * Get image data from a video element
 */
export function getImageDataFromVideo(video: HTMLVideoElement): ImageData {
  const canvas = document.createElement('canvas');
  canvas.width = video.videoWidth;
  canvas.height = video.videoHeight;
  const ctx = canvas.getContext('2d')!;
  ctx.drawImage(video, 0, 0);
  return ctx.getImageData(0, 0, canvas.width, canvas.height);
}

/**
 * Get image data from a Blob
 */
export async function getImageDataFromBlob(blob: Blob): Promise<ImageData> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    const url = URL.createObjectURL(blob);

    img.onload = () => {
      URL.revokeObjectURL(url);
      const canvas = document.createElement('canvas');
      canvas.width = img.naturalWidth;
      canvas.height = img.naturalHeight;
      const ctx = canvas.getContext('2d')!;
      ctx.drawImage(img, 0, 0);
      resolve(ctx.getImageData(0, 0, canvas.width, canvas.height));
    };

    img.onerror = () => {
      URL.revokeObjectURL(url);
      reject(new Error('Failed to load image'));
    };

    img.src = url;
  });
}

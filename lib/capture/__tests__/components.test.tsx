/**
 * Component Tests
 * VE-210: Unit tests for React capture components
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { CaptureGuidance } from '../components/CaptureGuidance';
import { QualityFeedback, QualityIndicator } from '../components/QualityFeedback';
import type { GuidanceState, QualityCheckResult } from '../types/capture';

// Mock guidance state
const mockGuidanceReady: GuidanceState = {
  documentDetected: true,
  alignmentOk: true,
  currentIssues: [],
  readyToCapture: true,
  instruction: 'Ready to capture! Hold steady and tap the button.',
};

const mockGuidanceNotReady: GuidanceState = {
  documentDetected: true,
  alignmentOk: false,
  currentIssues: [
    {
      type: 'blur',
      severity: 'error',
      message: 'Image is blurry',
      suggestion: 'Hold the device steady',
      confidence: 0.9,
    },
    {
      type: 'dark',
      severity: 'warning',
      message: 'Image is too dark',
      suggestion: 'Improve lighting',
      confidence: 0.8,
    },
  ],
  readyToCapture: false,
  instruction: 'Hold the device steady',
};

// Mock quality result
const mockQualityPassed: QualityCheckResult = {
  passed: true,
  score: 85,
  issues: [],
  checks: {
    resolution: { passed: true, value: 1.5, threshold: 1.0, description: 'Resolution OK' },
    brightness: { passed: true, value: 120, threshold: 130, description: 'Brightness OK' },
    blur: { passed: true, value: 150, threshold: 100, description: 'Not blurry' },
    skew: { passed: true, value: 2, threshold: 10, description: 'Aligned' },
    glare: { passed: true, value: 0.05, threshold: 0.15, description: 'No glare' },
    noise: { passed: true, value: 0.1, threshold: 0.2, description: 'Low noise' },
  },
  analysisTimeMs: 45,
};

const mockQualityFailed: QualityCheckResult = {
  passed: false,
  score: 45,
  issues: [
    {
      type: 'blur',
      severity: 'error',
      message: 'Image is blurry',
      suggestion: 'Hold the device steady',
      confidence: 0.9,
    },
    {
      type: 'dark',
      severity: 'warning',
      message: 'Image is too dark',
      suggestion: 'Improve lighting',
      confidence: 0.8,
    },
  ],
  checks: {
    resolution: { passed: true, value: 1.5, threshold: 1.0, description: 'Resolution OK' },
    brightness: { passed: false, value: 30, threshold: 130, description: 'Too dark' },
    blur: { passed: false, value: 50, threshold: 100, description: 'Blurry' },
    skew: { passed: true, value: 5, threshold: 10, description: 'Aligned' },
    glare: { passed: true, value: 0.05, threshold: 0.15, description: 'No glare' },
    noise: { passed: true, value: 0.15, threshold: 0.2, description: 'Acceptable noise' },
  },
  analysisTimeMs: 52,
};

describe('CaptureGuidance', () => {
  describe('document capture', () => {
    it('should render frame label for ID card', () => {
      render(
        <CaptureGuidance
          guidance={mockGuidanceReady}
          captureType="id_card"
        />
      );

      expect(screen.getByText(/Position Front of ID Card/i)).toBeInTheDocument();
    });

    it('should render back side label when isBackSide is true', () => {
      render(
        <CaptureGuidance
          guidance={mockGuidanceReady}
          captureType="id_card"
          isBackSide={true}
        />
      );

      expect(screen.getByText(/Position Back of ID Card/i)).toBeInTheDocument();
    });

    it('should render frame label for passport', () => {
      render(
        <CaptureGuidance
          guidance={mockGuidanceReady}
          captureType="passport"
        />
      );

      expect(screen.getByText(/Position Photo Page of Passport/i)).toBeInTheDocument();
    });

    it('should render frame label for driver license', () => {
      render(
        <CaptureGuidance
          guidance={mockGuidanceReady}
          captureType="drivers_license"
        />
      );

      expect(screen.getByText(/Position Front of Driver's License/i)).toBeInTheDocument();
    });

    it('should show ready instruction when ready to capture', () => {
      render(
        <CaptureGuidance
          guidance={mockGuidanceReady}
          captureType="id_card"
        />
      );

      expect(screen.getByText(/Ready to capture/i)).toBeInTheDocument();
    });

    it('should show issue messages when not ready', () => {
      render(
        <CaptureGuidance
          guidance={mockGuidanceNotReady}
          captureType="id_card"
        />
      );

      expect(screen.getByText(/Image is blurry/i)).toBeInTheDocument();
    });

    it('should show guidance instruction', () => {
      render(
        <CaptureGuidance
          guidance={mockGuidanceNotReady}
          captureType="id_card"
        />
      );

      expect(screen.getByText(/Hold the device steady/i)).toBeInTheDocument();
    });
  });

  describe('selfie capture', () => {
    it('should render selfie guidance text', () => {
      render(
        <CaptureGuidance
          guidance={mockGuidanceReady}
          captureType="selfie"
        />
      );

      expect(screen.getByText(/Position your face within the oval/i)).toBeInTheDocument();
    });
  });

  describe('styling', () => {
    it('should apply custom className', () => {
      const { container } = render(
        <CaptureGuidance
          guidance={mockGuidanceReady}
          captureType="id_card"
          className="custom-class"
        />
      );

      expect(container.querySelector('.custom-class')).toBeInTheDocument();
    });
  });
});

describe('QualityFeedback', () => {
  it('should render null when result is null', () => {
    const { container } = render(<QualityFeedback result={null} />);

    expect(container.firstChild).toBeNull();
  });

  it('should show score when result is provided', () => {
    render(<QualityFeedback result={mockQualityPassed} />);

    expect(screen.getByText('85')).toBeInTheDocument();
  });

  it('should show passed status for passing result', () => {
    render(<QualityFeedback result={mockQualityPassed} />);

    expect(screen.getByText(/Image quality is acceptable/i)).toBeInTheDocument();
  });

  it('should show failed status for failing result', () => {
    render(<QualityFeedback result={mockQualityFailed} />);

    expect(screen.getByText(/does not meet quality requirements/i)).toBeInTheDocument();
  });

  it('should show issues list for failing result', () => {
    render(<QualityFeedback result={mockQualityFailed} />);

    expect(screen.getByText(/Image is blurry/i)).toBeInTheDocument();
    expect(screen.getByText(/Image is too dark/i)).toBeInTheDocument();
  });

  it('should show suggestions for issues', () => {
    render(<QualityFeedback result={mockQualityFailed} />);

    expect(screen.getByText(/Hold the device steady/i)).toBeInTheDocument();
    expect(screen.getByText(/Improve lighting/i)).toBeInTheDocument();
  });

  it('should show no issues message for passing result', () => {
    render(<QualityFeedback result={mockQualityPassed} />);

    expect(screen.getByText(/No quality issues detected/i)).toBeInTheDocument();
  });

  it('should show analysis time', () => {
    render(<QualityFeedback result={mockQualityPassed} />);

    expect(screen.getByText(/Analyzed in/i)).toBeInTheDocument();
  });

  describe('compact mode', () => {
    it('should render compact version', () => {
      render(<QualityFeedback result={mockQualityPassed} compact />);

      expect(screen.getByText('85')).toBeInTheDocument();
      expect(screen.getByText(/Quality OK/i)).toBeInTheDocument();
    });

    it('should show issue count in compact mode', () => {
      render(<QualityFeedback result={mockQualityFailed} compact />);

      expect(screen.getByText(/2 issue\(s\) found/i)).toBeInTheDocument();
    });
  });

  describe('showDetails', () => {
    it('should hide details when showDetails is false', () => {
      render(<QualityFeedback result={mockQualityPassed} showDetails={false} />);

      // Check names should not be visible
      expect(screen.queryByText(/Resolution/i)).not.toBeInTheDocument();
    });
  });

  describe('showScore', () => {
    it('should hide score when showScore is false', () => {
      render(<QualityFeedback result={mockQualityPassed} showScore={false} />);

      expect(screen.queryByText('85')).not.toBeInTheDocument();
    });
  });
});

describe('QualityIndicator', () => {
  it('should render score value', () => {
    render(<QualityIndicator score={85} passed={true} />);

    expect(screen.getByText('85')).toBeInTheDocument();
  });

  it('should render with different sizes', () => {
    const { rerender } = render(<QualityIndicator score={85} passed={true} size="small" />);
    expect(screen.getByText('85')).toBeInTheDocument();

    rerender(<QualityIndicator score={85} passed={true} size="medium" />);
    expect(screen.getByText('85')).toBeInTheDocument();

    rerender(<QualityIndicator score={85} passed={true} size="large" />);
    expect(screen.getByText('85')).toBeInTheDocument();
  });
});

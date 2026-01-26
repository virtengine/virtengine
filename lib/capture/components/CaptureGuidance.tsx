/**
 * CaptureGuidance Component
 * VE-210: Visual guidance overlay for document and selfie capture
 *
 * Provides real-time visual feedback to help users capture
 * high-quality images.
 */

import React from 'react';
import type { GuidanceState, DocumentType, QualityIssue } from '../types/capture';

/**
 * Props for CaptureGuidance component
 */
export interface CaptureGuidanceProps {
  /** Current guidance state */
  guidance: GuidanceState;
  /** Type of capture (document type or 'selfie') */
  captureType: DocumentType | 'selfie';
  /** Whether this is for a document back side */
  isBackSide?: boolean;
  /** Show corner markers */
  showCorners?: boolean;
  /** Show center target */
  showCenterTarget?: boolean;
  /** Custom class name */
  className?: string;
}

/**
 * Get frame aspect ratio for document type
 */
function getFrameAspectRatio(captureType: DocumentType | 'selfie'): number {
  switch (captureType) {
    case 'id_card':
    case 'drivers_license':
      return 1.586; // Standard ID card ratio (85.6mm × 53.98mm)
    case 'passport':
      return 1.42; // Passport ratio (125mm × 88mm)
    case 'selfie':
      return 0.75; // Portrait ratio (3:4)
    default:
      return 1.5;
  }
}

/**
 * Get frame label for document type
 */
function getFrameLabel(captureType: DocumentType | 'selfie', isBackSide?: boolean): string {
  if (captureType === 'selfie') {
    return 'Position your face within the oval';
  }

  const sideLabel = isBackSide ? 'Back' : 'Front';
  switch (captureType) {
    case 'id_card':
      return `Position ${sideLabel} of ID Card`;
    case 'passport':
      return 'Position Photo Page of Passport';
    case 'drivers_license':
      return `Position ${sideLabel} of Driver's License`;
    default:
      return 'Position Document';
  }
}

/**
 * Get icon for quality issue
 */
function getIssueIcon(type: QualityIssue['type']): string {
  switch (type) {
    case 'blur':
      return '📷';
    case 'dark':
      return '🌑';
    case 'bright':
      return '☀️';
    case 'skew':
      return '📐';
    case 'resolution':
      return '🔍';
    case 'glare':
      return '✨';
    case 'noise':
      return '📶';
    case 'partial':
      return '⬜';
    case 'reflection':
      return '🪞';
    default:
      return '⚠️';
  }
}

/**
 * Styles for the guidance overlay
 */
const styles = {
  container: {
    position: 'absolute' as const,
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    display: 'flex',
    flexDirection: 'column' as const,
    alignItems: 'center',
    justifyContent: 'center',
    pointerEvents: 'none' as const,
  },
  frameContainer: {
    position: 'relative' as const,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  documentFrame: {
    position: 'relative' as const,
    border: '3px solid',
    borderRadius: '12px',
    transition: 'border-color 0.3s ease',
  },
  selfieFrame: {
    position: 'relative' as const,
    border: '3px solid',
    borderRadius: '50%',
    transition: 'border-color 0.3s ease',
  },
  corner: {
    position: 'absolute' as const,
    width: '30px',
    height: '30px',
    borderColor: 'inherit',
    borderStyle: 'solid',
    borderWidth: '4px',
  },
  cornerTopLeft: {
    top: '-2px',
    left: '-2px',
    borderRight: 'none',
    borderBottom: 'none',
    borderTopLeftRadius: '12px',
  },
  cornerTopRight: {
    top: '-2px',
    right: '-2px',
    borderLeft: 'none',
    borderBottom: 'none',
    borderTopRightRadius: '12px',
  },
  cornerBottomLeft: {
    bottom: '-2px',
    left: '-2px',
    borderRight: 'none',
    borderTop: 'none',
    borderBottomLeftRadius: '12px',
  },
  cornerBottomRight: {
    bottom: '-2px',
    right: '-2px',
    borderLeft: 'none',
    borderTop: 'none',
    borderBottomRightRadius: '12px',
  },
  label: {
    position: 'absolute' as const,
    top: '-40px',
    left: '50%',
    transform: 'translateX(-50%)',
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    color: 'white',
    padding: '8px 16px',
    borderRadius: '20px',
    fontSize: '14px',
    fontWeight: 500,
    whiteSpace: 'nowrap' as const,
  },
  instruction: {
    position: 'absolute' as const,
    bottom: '-50px',
    left: '50%',
    transform: 'translateX(-50%)',
    backgroundColor: 'rgba(0, 0, 0, 0.8)',
    color: 'white',
    padding: '10px 20px',
    borderRadius: '8px',
    fontSize: '16px',
    fontWeight: 500,
    textAlign: 'center' as const,
    maxWidth: '300px',
    transition: 'background-color 0.3s ease',
  },
  issuesList: {
    position: 'absolute' as const,
    bottom: '-120px',
    left: '50%',
    transform: 'translateX(-50%)',
    display: 'flex',
    flexDirection: 'column' as const,
    gap: '4px',
    alignItems: 'center',
  },
  issueItem: {
    backgroundColor: 'rgba(0, 0, 0, 0.6)',
    color: 'white',
    padding: '4px 12px',
    borderRadius: '4px',
    fontSize: '12px',
    display: 'flex',
    alignItems: 'center',
    gap: '6px',
  },
  readyIndicator: {
    position: 'absolute' as const,
    top: '50%',
    left: '50%',
    transform: 'translate(-50%, -50%)',
    width: '60px',
    height: '60px',
    borderRadius: '50%',
    backgroundColor: 'rgba(34, 197, 94, 0.2)',
    border: '3px solid #22c55e',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    animation: 'pulse 1.5s infinite',
  },
  checkmark: {
    color: '#22c55e',
    fontSize: '30px',
  },
};

/**
 * CaptureGuidance component
 */
export const CaptureGuidance: React.FC<CaptureGuidanceProps> = ({
  guidance,
  captureType,
  isBackSide = false,
  showCorners = true,
  showCenterTarget = true,
  className = '',
}) => {
  const aspectRatio = getFrameAspectRatio(captureType);
  const frameLabel = getFrameLabel(captureType, isBackSide);
  const isSelfie = captureType === 'selfie';

  // Calculate frame dimensions (responsive)
  const frameWidth = isSelfie ? 220 : 320;
  const frameHeight = isSelfie ? frameWidth / aspectRatio : frameWidth / aspectRatio;

  // Determine frame color based on state
  const getFrameColor = () => {
    if (guidance.readyToCapture) return '#22c55e'; // Green
    if (guidance.currentIssues.some((i) => i.severity === 'error')) return '#ef4444'; // Red
    if (guidance.currentIssues.some((i) => i.severity === 'warning')) return '#f59e0b'; // Amber
    return '#3b82f6'; // Blue (default)
  };

  const frameColor = getFrameColor();

  return (
    <div className={`capture-guidance ${className}`} style={styles.container}>
      {/* Keyframe animation for pulse */}
      <style>
        {`
          @keyframes pulse {
            0% { opacity: 1; transform: translate(-50%, -50%) scale(1); }
            50% { opacity: 0.8; transform: translate(-50%, -50%) scale(1.1); }
            100% { opacity: 1; transform: translate(-50%, -50%) scale(1); }
          }
        `}
      </style>

      <div style={styles.frameContainer}>
        {/* Frame label */}
        <div style={styles.label}>{frameLabel}</div>

        {/* Document/Selfie frame */}
        <div
          style={{
            ...(isSelfie ? styles.selfieFrame : styles.documentFrame),
            width: frameWidth,
            height: frameHeight,
            borderColor: frameColor,
          }}
        >
          {/* Corner markers (for documents only) */}
          {showCorners && !isSelfie && (
            <>
              <div style={{ ...styles.corner, ...styles.cornerTopLeft, borderColor: frameColor }} />
              <div style={{ ...styles.corner, ...styles.cornerTopRight, borderColor: frameColor }} />
              <div style={{ ...styles.corner, ...styles.cornerBottomLeft, borderColor: frameColor }} />
              <div style={{ ...styles.corner, ...styles.cornerBottomRight, borderColor: frameColor }} />
            </>
          )}

          {/* Ready indicator */}
          {showCenterTarget && guidance.readyToCapture && (
            <div style={styles.readyIndicator}>
              <span style={styles.checkmark}>✓</span>
            </div>
          )}
        </div>

        {/* Instruction message */}
        <div
          style={{
            ...styles.instruction,
            backgroundColor: guidance.readyToCapture
              ? 'rgba(34, 197, 94, 0.9)'
              : 'rgba(0, 0, 0, 0.8)',
          }}
        >
          {guidance.instruction}
        </div>

        {/* Issues list (show top 2) */}
        {guidance.currentIssues.length > 0 && !guidance.readyToCapture && (
          <div style={styles.issuesList}>
            {guidance.currentIssues.slice(0, 2).map((issue, index) => (
              <div
                key={`${issue.type}-${index}`}
                style={{
                  ...styles.issueItem,
                  backgroundColor:
                    issue.severity === 'error'
                      ? 'rgba(239, 68, 68, 0.8)'
                      : 'rgba(245, 158, 11, 0.8)',
                }}
              >
                <span>{getIssueIcon(issue.type)}</span>
                <span>{issue.message}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default CaptureGuidance;

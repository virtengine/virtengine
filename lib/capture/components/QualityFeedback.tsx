/**
 * QualityFeedback Component
 * VE-210: Quality feedback display for capture components
 *
 * Shows detailed quality check results and provides
 * actionable feedback to users.
 */

import React from 'react';
import type { QualityCheckResult, QualityIssue } from '../types/capture';

/**
 * Props for QualityFeedback component
 */
export interface QualityFeedbackProps {
  /** Quality check result */
  result: QualityCheckResult | null;
  /** Show detailed breakdown */
  showDetails?: boolean;
  /** Show score */
  showScore?: boolean;
  /** Compact mode */
  compact?: boolean;
  /** Custom class name */
  className?: string;
}

/**
 * Get color for score
 */
function getScoreColor(score: number): string {
  if (score >= 80) return '#22c55e'; // Green
  if (score >= 60) return '#f59e0b'; // Amber
  return '#ef4444'; // Red
}

/**
 * Get icon for check type
 */
function getCheckIcon(passed: boolean): string {
  return passed ? '✓' : '✗';
}

/**
 * Format check name
 */
function formatCheckName(name: string): string {
  return name.charAt(0).toUpperCase() + name.slice(1);
}

/**
 * Styles for the quality feedback component
 */
const styles = {
  container: {
    backgroundColor: 'rgba(0, 0, 0, 0.85)',
    borderRadius: '12px',
    padding: '16px',
    color: 'white',
    fontFamily: 'system-ui, -apple-system, sans-serif',
  },
  compactContainer: {
    backgroundColor: 'rgba(0, 0, 0, 0.85)',
    borderRadius: '8px',
    padding: '10px 16px',
    color: 'white',
    fontFamily: 'system-ui, -apple-system, sans-serif',
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
  },
  header: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: '12px',
  },
  title: {
    fontSize: '16px',
    fontWeight: 600,
    margin: 0,
  },
  scoreContainer: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
  },
  scoreCircle: {
    width: '48px',
    height: '48px',
    borderRadius: '50%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '18px',
    fontWeight: 700,
    border: '3px solid',
  },
  compactScoreCircle: {
    width: '36px',
    height: '36px',
    borderRadius: '50%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '14px',
    fontWeight: 700,
    border: '2px solid',
  },
  scoreLabel: {
    fontSize: '12px',
    color: '#9ca3af',
  },
  status: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    padding: '8px 12px',
    borderRadius: '8px',
    marginBottom: '12px',
  },
  statusIcon: {
    fontSize: '20px',
  },
  statusText: {
    fontSize: '14px',
    fontWeight: 500,
  },
  checksGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, 1fr)',
    gap: '8px',
    marginBottom: '12px',
  },
  checkItem: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    padding: '6px 10px',
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
    borderRadius: '6px',
    fontSize: '13px',
  },
  checkIcon: {
    width: '18px',
    height: '18px',
    borderRadius: '50%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '12px',
    fontWeight: 700,
  },
  issuesList: {
    display: 'flex',
    flexDirection: 'column' as const,
    gap: '6px',
  },
  issueItem: {
    display: 'flex',
    alignItems: 'flex-start',
    gap: '10px',
    padding: '10px 12px',
    borderRadius: '8px',
    fontSize: '13px',
  },
  issueIcon: {
    fontSize: '16px',
    flexShrink: 0,
  },
  issueContent: {
    display: 'flex',
    flexDirection: 'column' as const,
    gap: '2px',
  },
  issueMessage: {
    fontWeight: 500,
  },
  issueSuggestion: {
    fontSize: '12px',
    color: '#d1d5db',
  },
  noIssues: {
    textAlign: 'center' as const,
    padding: '16px',
    color: '#22c55e',
    fontSize: '14px',
  },
  analysisTime: {
    fontSize: '11px',
    color: '#6b7280',
    textAlign: 'right' as const,
    marginTop: '8px',
  },
};

/**
 * Get severity icon
 */
function getSeverityIcon(severity: QualityIssue['severity']): string {
  return severity === 'error' ? '🚫' : '⚠️';
}

/**
 * QualityFeedback component
 */
export const QualityFeedback: React.FC<QualityFeedbackProps> = ({
  result,
  showDetails = true,
  showScore = true,
  compact = false,
  className = '',
}) => {
  if (!result) {
    return null;
  }

  const scoreColor = getScoreColor(result.score);

  // Compact mode
  if (compact) {
    return (
      <div className={`quality-feedback ${className}`} style={styles.compactContainer}>
        {showScore && (
          <div
            style={{
              ...styles.compactScoreCircle,
              borderColor: scoreColor,
              color: scoreColor,
            }}
          >
            {result.score}
          </div>
        )}
        <span style={{ fontSize: '14px', fontWeight: 500 }}>
          {result.passed ? 'Quality OK' : `${result.issues.length} issue(s) found`}
        </span>
        <span style={{ fontSize: '20px' }}>{result.passed ? '✅' : '⚠️'}</span>
      </div>
    );
  }

  // Full mode
  return (
    <div className={`quality-feedback ${className}`} style={styles.container}>
      {/* Header with score */}
      <div style={styles.header}>
        <h3 style={styles.title}>Quality Analysis</h3>
        {showScore && (
          <div style={styles.scoreContainer}>
            <div style={{ textAlign: 'right' }}>
              <div style={styles.scoreLabel}>Score</div>
            </div>
            <div
              style={{
                ...styles.scoreCircle,
                borderColor: scoreColor,
                color: scoreColor,
              }}
            >
              {result.score}
            </div>
          </div>
        )}
      </div>

      {/* Pass/Fail status */}
      <div
        style={{
          ...styles.status,
          backgroundColor: result.passed
            ? 'rgba(34, 197, 94, 0.2)'
            : 'rgba(239, 68, 68, 0.2)',
        }}
      >
        <span style={styles.statusIcon}>{result.passed ? '✅' : '❌'}</span>
        <span style={styles.statusText}>
          {result.passed
            ? 'Image quality is acceptable'
            : 'Image does not meet quality requirements'}
        </span>
      </div>

      {/* Detailed checks grid */}
      {showDetails && (
        <div style={styles.checksGrid}>
          {Object.entries(result.checks).map(([name, check]) => (
            <div key={name} style={styles.checkItem}>
              <div
                style={{
                  ...styles.checkIcon,
                  backgroundColor: check.passed
                    ? 'rgba(34, 197, 94, 0.2)'
                    : 'rgba(239, 68, 68, 0.2)',
                  color: check.passed ? '#22c55e' : '#ef4444',
                }}
              >
                {getCheckIcon(check.passed)}
              </div>
              <span>{formatCheckName(name)}</span>
            </div>
          ))}
        </div>
      )}

      {/* Issues list */}
      {result.issues.length > 0 ? (
        <div style={styles.issuesList}>
          {result.issues.map((issue, index) => (
            <div
              key={`${issue.type}-${index}`}
              style={{
                ...styles.issueItem,
                backgroundColor:
                  issue.severity === 'error'
                    ? 'rgba(239, 68, 68, 0.15)'
                    : 'rgba(245, 158, 11, 0.15)',
              }}
            >
              <span style={styles.issueIcon}>{getSeverityIcon(issue.severity)}</span>
              <div style={styles.issueContent}>
                <span style={styles.issueMessage}>{issue.message}</span>
                <span style={styles.issueSuggestion}>{issue.suggestion}</span>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div style={styles.noIssues}>
          ✓ No quality issues detected
        </div>
      )}

      {/* Analysis time */}
      <div style={styles.analysisTime}>
        Analyzed in {result.analysisTimeMs.toFixed(0)}ms
      </div>
    </div>
  );
};

/**
 * Minimal quality indicator for inline use
 */
export interface QualityIndicatorProps {
  score: number;
  passed: boolean;
  size?: 'small' | 'medium' | 'large';
}

export const QualityIndicator: React.FC<QualityIndicatorProps> = ({
  score,
  passed,
  size = 'medium',
}) => {
  const sizes = {
    small: { circle: 24, font: 10, border: 2 },
    medium: { circle: 36, font: 14, border: 2 },
    large: { circle: 48, font: 18, border: 3 },
  };

  const s = sizes[size];
  const color = getScoreColor(score);

  return (
    <div
      style={{
        width: s.circle,
        height: s.circle,
        borderRadius: '50%',
        border: `${s.border}px solid ${color}`,
        color: color,
        fontSize: s.font,
        fontWeight: 700,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: passed ? 'rgba(34, 197, 94, 0.1)' : 'rgba(239, 68, 68, 0.1)',
      }}
    >
      {score}
    </div>
  );
};

export default QualityFeedback;

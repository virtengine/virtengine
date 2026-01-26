/**
 * Identity Score Display Component
 * VE-701: Display identity score with visual indicator
 */
import * as React from 'react';
import { useIdentity } from '../../hooks/useIdentity';
import { formatScore } from '../../utils/format';
import type { IdentityScore } from '../../types/identity';

/**
 * Identity score display props
 */
export interface IdentityScoreDisplayProps {
  /**
   * Optional score to display (uses context if not provided)
   */
  score?: IdentityScore;

  /**
   * Size variant
   */
  size?: 'sm' | 'md' | 'lg';

  /**
   * Whether to show component breakdown
   */
  showComponents?: boolean;

  /**
   * Custom CSS class
   */
  className?: string;
}

/**
 * Get color based on score value
 */
function getScoreColor(overall: number): string {
  if (overall >= 80) return '#22c55e'; // green
  if (overall >= 60) return '#eab308'; // yellow
  if (overall >= 40) return '#f97316'; // orange
  return '#ef4444'; // red
}

/**
 * Get label based on score value
 */
function getScoreLabel(overall: number): string {
  if (overall >= 80) return 'Excellent';
  if (overall >= 60) return 'Good';
  if (overall >= 40) return 'Fair';
  return 'Low';
}

/**
 * Identity score display component
 */
export function IdentityScoreDisplay({
  score: propScore,
  size = 'md',
  showComponents = false,
  className = '',
}: IdentityScoreDisplayProps): JSX.Element {
  const { state } = useIdentity();
  const score = propScore || state.identity?.score;

  if (!score) {
    return (
      <div className={`identity-score identity-score--${size} identity-score--empty ${className}`}>
        <div className="identity-score__circle">
          <span className="identity-score__value">--</span>
        </div>
        <span className="identity-score__label">Not Verified</span>
      </div>
    );
  }

  const color = getScoreColor(score.overall);
  const label = getScoreLabel(score.overall);
  const circumference = 2 * Math.PI * 45; // Circle radius of 45
  const progress = (score.overall / 100) * circumference;

  const sizeMap = {
    sm: { width: 60, height: 60, fontSize: 16 },
    md: { width: 100, height: 100, fontSize: 24 },
    lg: { width: 140, height: 140, fontSize: 32 },
  };

  const { width, height, fontSize } = sizeMap[size];

  return (
    <div className={`identity-score identity-score--${size} ${className}`}>
      <div className="identity-score__circle-container" style={{ width, height }}>
        <svg viewBox="0 0 100 100" className="identity-score__svg">
          {/* Background circle */}
          <circle
            cx="50"
            cy="50"
            r="45"
            fill="none"
            stroke="#e5e7eb"
            strokeWidth="8"
          />
          {/* Progress circle */}
          <circle
            cx="50"
            cy="50"
            r="45"
            fill="none"
            stroke={color}
            strokeWidth="8"
            strokeLinecap="round"
            strokeDasharray={`${progress} ${circumference}`}
            transform="rotate(-90 50 50)"
          />
        </svg>
        <div className="identity-score__value-container">
          <span 
            className="identity-score__value" 
            style={{ fontSize, color }}
          >
            {formatScore(score.overall)}
          </span>
        </div>
      </div>
      <span className="identity-score__label" style={{ color }}>
        {label}
      </span>

      {showComponents && (
        <div className="identity-score__components">
          <ScoreComponent label="Face Match" value={score.faceMatch} />
          <ScoreComponent label="Document" value={score.documentAuthenticity} />
          <ScoreComponent label="Liveness" value={score.liveness} />
          <ScoreComponent label="Consistency" value={score.dataConsistency} />
        </div>
      )}

      <style>{`
        .identity-score {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 8px;
        }

        .identity-score__circle-container {
          position: relative;
        }

        .identity-score__svg {
          width: 100%;
          height: 100%;
        }

        .identity-score__value-container {
          position: absolute;
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
        }

        .identity-score__value {
          font-weight: 700;
        }

        .identity-score__label {
          font-weight: 600;
          text-transform: uppercase;
          font-size: 0.75rem;
          letter-spacing: 0.05em;
        }

        .identity-score__components {
          display: grid;
          grid-template-columns: repeat(2, 1fr);
          gap: 8px;
          margin-top: 16px;
          width: 100%;
        }

        .identity-score--sm .identity-score__label {
          font-size: 0.625rem;
        }

        .identity-score--lg .identity-score__label {
          font-size: 0.875rem;
        }
      `}</style>
    </div>
  );
}

/**
 * Score component item
 */
interface ScoreComponentProps {
  label: string;
  value: number;
}

function ScoreComponent({ label, value }: ScoreComponentProps): JSX.Element {
  const color = getScoreColor(value);

  return (
    <div className="score-component">
      <span className="score-component__label">{label}</span>
      <div className="score-component__bar">
        <div 
          className="score-component__fill" 
          style={{ width: `${value}%`, backgroundColor: color }}
        />
      </div>
      <span className="score-component__value" style={{ color }}>
        {formatScore(value)}
      </span>

      <style>{`
        .score-component {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .score-component__label {
          font-size: 0.75rem;
          color: #6b7280;
          min-width: 80px;
        }

        .score-component__bar {
          flex: 1;
          height: 4px;
          background: #e5e7eb;
          border-radius: 2px;
          overflow: hidden;
        }

        .score-component__fill {
          height: 100%;
          border-radius: 2px;
          transition: width 0.3s ease;
        }

        .score-component__value {
          font-size: 0.75rem;
          font-weight: 600;
          min-width: 24px;
          text-align: right;
        }
      `}</style>
    </div>
  );
}

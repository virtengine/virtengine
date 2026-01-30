/**
 * Accessible Base Components
 * VE-UI-002: WCAG 2.1 AA compliant components
 */
import * as React from 'react';
import { generateA11yId, srOnlyStyles } from '../../utils/a11y';

/**
 * Screen Reader Only - visually hidden but accessible to screen readers
 */
export interface SrOnlyProps {
  children: React.ReactNode;
  as?: keyof JSX.IntrinsicElements;
}

export function SrOnly({ children, as: Tag = 'span' }: SrOnlyProps): JSX.Element {
  return <Tag style={srOnlyStyles}>{children}</Tag>;
}

/**
 * Skip Link - allows keyboard users to skip to main content
 */
export interface SkipLinkProps {
  targetId: string;
  children?: React.ReactNode;
}

export function SkipLink({ targetId, children = 'Skip to main content' }: SkipLinkProps): JSX.Element {
  return (
    <>
      <a
        href={`#${targetId}`}
        className="skip-link"
        onClick={(e) => {
          e.preventDefault();
          const target = document.getElementById(targetId);
          if (target) {
            target.setAttribute('tabindex', '-1');
            target.focus();
            target.removeAttribute('tabindex');
          }
        }}
      >
        {children}
      </a>
      <style>{`
        .skip-link {
          position: absolute;
          top: -100%;
          left: 0;
          background: #111827;
          color: white;
          padding: 12px 24px;
          text-decoration: none;
          font-weight: 600;
          z-index: 9999;
          border-radius: 0 0 8px 0;
          transition: top 0.2s ease;
        }
        .skip-link:focus {
          top: 0;
          outline: 2px solid #3b82f6;
          outline-offset: 2px;
        }
      `}</style>
    </>
  );
}

/**
 * Accessible Button
 */
export interface AccessibleButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
  loadingText?: string;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
}

export const AccessibleButton = React.forwardRef<HTMLButtonElement, AccessibleButtonProps>(
  (
    {
      variant = 'primary',
      size = 'md',
      isLoading = false,
      loadingText,
      leftIcon,
      rightIcon,
      children,
      disabled,
      className = '',
      ...props
    },
    ref
  ): JSX.Element => {
    const isDisabled = disabled || isLoading;

    return (
      <>
        <button
          ref={ref}
          className={`a11y-button a11y-button--${variant} a11y-button--${size} ${className}`}
          disabled={isDisabled}
          aria-disabled={isDisabled}
          aria-busy={isLoading}
          {...props}
        >
          {isLoading && (
            <span className="a11y-button__spinner" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" className="a11y-button__spinner-svg">
                <circle
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                  strokeDasharray="32"
                  strokeLinecap="round"
                />
              </svg>
            </span>
          )}
          {!isLoading && leftIcon && (
            <span className="a11y-button__icon a11y-button__icon--left" aria-hidden="true">
              {leftIcon}
            </span>
          )}
          <span className="a11y-button__text">
            {isLoading && loadingText ? loadingText : children}
          </span>
          {!isLoading && rightIcon && (
            <span className="a11y-button__icon a11y-button__icon--right" aria-hidden="true">
              {rightIcon}
            </span>
          )}
          {isLoading && <SrOnly>Loading</SrOnly>}
        </button>
        <style>{buttonStyles}</style>
      </>
    );
  }
);

AccessibleButton.displayName = 'AccessibleButton';

const buttonStyles = `
  .a11y-button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    border: none;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
    position: relative;
  }

  .a11y-button:focus-visible {
    outline: 2px solid #3b82f6;
    outline-offset: 2px;
  }

  .a11y-button:disabled,
  .a11y-button[aria-disabled="true"] {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Sizes */
  .a11y-button--sm {
    padding: 8px 16px;
    font-size: 0.875rem;
    min-height: 36px;
  }

  .a11y-button--md {
    padding: 10px 20px;
    font-size: 0.875rem;
    min-height: 44px; /* WCAG 2.5.5 target size */
  }

  .a11y-button--lg {
    padding: 12px 24px;
    font-size: 1rem;
    min-height: 48px;
  }

  /* Variants */
  .a11y-button--primary {
    background: #3b82f6;
    color: white;
  }

  .a11y-button--primary:hover:not(:disabled) {
    background: #2563eb;
  }

  .a11y-button--secondary {
    background: #6b7280;
    color: white;
  }

  .a11y-button--secondary:hover:not(:disabled) {
    background: #4b5563;
  }

  .a11y-button--outline {
    background: transparent;
    border: 2px solid #d1d5db;
    color: #374151;
  }

  .a11y-button--outline:hover:not(:disabled) {
    border-color: #3b82f6;
    color: #3b82f6;
  }

  .a11y-button--ghost {
    background: transparent;
    color: #374151;
  }

  .a11y-button--ghost:hover:not(:disabled) {
    background: #f3f4f6;
  }

  .a11y-button--danger {
    background: #dc2626;
    color: white;
  }

  .a11y-button--danger:hover:not(:disabled) {
    background: #b91c1c;
  }

  /* Spinner */
  .a11y-button__spinner {
    display: inline-flex;
  }

  .a11y-button__spinner-svg {
    width: 16px;
    height: 16px;
    animation: a11y-button-spin 1s linear infinite;
  }

  @keyframes a11y-button-spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  /* Reduced motion */
  @media (prefers-reduced-motion: reduce) {
    .a11y-button__spinner-svg {
      animation: none;
    }
  }
`;

/**
 * Accessible Input
 */
export interface AccessibleInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  hint?: string;
  hideLabel?: boolean;
}

export const AccessibleInput = React.forwardRef<HTMLInputElement, AccessibleInputProps>(
  ({ label, error, hint, hideLabel = false, id, className = '', ...props }, ref): JSX.Element => {
    const inputId = id || generateA11yId('input');
    const errorId = error ? `${inputId}-error` : undefined;
    const hintId = hint ? `${inputId}-hint` : undefined;
    const describedBy = [errorId, hintId].filter(Boolean).join(' ') || undefined;

    return (
      <>
        <div className={`a11y-input-wrapper ${className}`}>
          <label
            htmlFor={inputId}
            className={`a11y-input-label ${hideLabel ? 'sr-only' : ''}`}
            style={hideLabel ? srOnlyStyles : undefined}
          >
            {label}
            {props.required && <span aria-hidden="true"> *</span>}
          </label>
          {hint && (
            <p id={hintId} className="a11y-input-hint">
              {hint}
            </p>
          )}
          <input
            ref={ref}
            id={inputId}
            className={`a11y-input ${error ? 'a11y-input--error' : ''}`}
            aria-invalid={error ? 'true' : undefined}
            aria-describedby={describedBy}
            aria-required={props.required}
            {...props}
          />
          {error && (
            <p id={errorId} className="a11y-input-error" role="alert">
              <span aria-hidden="true">⚠ </span>
              {error}
            </p>
          )}
        </div>
        <style>{inputStyles}</style>
      </>
    );
  }
);

AccessibleInput.displayName = 'AccessibleInput';

const inputStyles = `
  .a11y-input-wrapper {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .a11y-input-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
  }

  .a11y-input-hint {
    font-size: 0.75rem;
    color: #6b7280;
    margin: 0;
  }

  .a11y-input {
    width: 100%;
    padding: 10px 12px;
    border: 2px solid #d1d5db;
    border-radius: 6px;
    font-size: 1rem;
    min-height: 44px; /* WCAG 2.5.5 target size */
    transition: border-color 0.2s, box-shadow 0.2s;
  }

  .a11y-input:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
  }

  .a11y-input--error {
    border-color: #dc2626;
  }

  .a11y-input--error:focus {
    border-color: #dc2626;
    box-shadow: 0 0 0 3px rgba(220, 38, 38, 0.2);
  }

  .a11y-input-error {
    font-size: 0.75rem;
    color: #dc2626;
    margin: 0;
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
`;

/**
 * Accessible Select
 */
export interface AccessibleSelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  label: string;
  error?: string;
  hint?: string;
  hideLabel?: boolean;
  options: Array<{ value: string; label: string; disabled?: boolean }>;
}

export const AccessibleSelect = React.forwardRef<HTMLSelectElement, AccessibleSelectProps>(
  ({ label, error, hint, hideLabel = false, options, id, className = '', ...props }, ref): JSX.Element => {
    const selectId = id || generateA11yId('select');
    const errorId = error ? `${selectId}-error` : undefined;
    const hintId = hint ? `${selectId}-hint` : undefined;
    const describedBy = [errorId, hintId].filter(Boolean).join(' ') || undefined;

    return (
      <>
        <div className={`a11y-select-wrapper ${className}`}>
          <label
            htmlFor={selectId}
            className={`a11y-select-label ${hideLabel ? 'sr-only' : ''}`}
            style={hideLabel ? srOnlyStyles : undefined}
          >
            {label}
            {props.required && <span aria-hidden="true"> *</span>}
          </label>
          {hint && (
            <p id={hintId} className="a11y-select-hint">
              {hint}
            </p>
          )}
          <select
            ref={ref}
            id={selectId}
            className={`a11y-select ${error ? 'a11y-select--error' : ''}`}
            aria-invalid={error ? 'true' : undefined}
            aria-describedby={describedBy}
            aria-required={props.required}
            {...props}
          >
            {options.map((option) => (
              <option key={option.value} value={option.value} disabled={option.disabled}>
                {option.label}
              </option>
            ))}
          </select>
          {error && (
            <p id={errorId} className="a11y-select-error" role="alert">
              <span aria-hidden="true">⚠ </span>
              {error}
            </p>
          )}
        </div>
        <style>{selectStyles}</style>
      </>
    );
  }
);

AccessibleSelect.displayName = 'AccessibleSelect';

const selectStyles = `
  .a11y-select-wrapper {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .a11y-select-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
  }

  .a11y-select-hint {
    font-size: 0.75rem;
    color: #6b7280;
    margin: 0;
  }

  .a11y-select {
    width: 100%;
    padding: 10px 12px;
    border: 2px solid #d1d5db;
    border-radius: 6px;
    font-size: 1rem;
    min-height: 44px;
    background: white;
    cursor: pointer;
    transition: border-color 0.2s, box-shadow 0.2s;
  }

  .a11y-select:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
  }

  .a11y-select--error {
    border-color: #dc2626;
  }

  .a11y-select-error {
    font-size: 0.75rem;
    color: #dc2626;
    margin: 0;
    display: flex;
    align-items: center;
    gap: 4px;
  }
`;

/**
 * Accessible Checkbox
 */
export interface AccessibleCheckboxProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label: string;
  description?: string;
}

export const AccessibleCheckbox = React.forwardRef<HTMLInputElement, AccessibleCheckboxProps>(
  ({ label, description, id, className = '', ...props }, ref): JSX.Element => {
    const checkboxId = id || generateA11yId('checkbox');
    const descriptionId = description ? `${checkboxId}-description` : undefined;

    return (
      <>
        <div className={`a11y-checkbox-wrapper ${className}`}>
          <input
            ref={ref}
            type="checkbox"
            id={checkboxId}
            className="a11y-checkbox"
            aria-describedby={descriptionId}
            {...props}
          />
          <label htmlFor={checkboxId} className="a11y-checkbox-label">
            <span className="a11y-checkbox-text">{label}</span>
            {description && (
              <span id={descriptionId} className="a11y-checkbox-description">
                {description}
              </span>
            )}
          </label>
        </div>
        <style>{checkboxStyles}</style>
      </>
    );
  }
);

AccessibleCheckbox.displayName = 'AccessibleCheckbox';

const checkboxStyles = `
  .a11y-checkbox-wrapper {
    display: flex;
    align-items: flex-start;
    gap: 12px;
  }

  .a11y-checkbox {
    width: 20px;
    height: 20px;
    margin: 2px 0 0 0;
    cursor: pointer;
    accent-color: #3b82f6;
  }

  .a11y-checkbox:focus-visible {
    outline: 2px solid #3b82f6;
    outline-offset: 2px;
  }

  .a11y-checkbox-label {
    display: flex;
    flex-direction: column;
    gap: 4px;
    cursor: pointer;
  }

  .a11y-checkbox-text {
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
  }

  .a11y-checkbox-description {
    font-size: 0.75rem;
    color: #6b7280;
  }
`;

/**
 * Alert component for status messages
 */
export interface AccessibleAlertProps {
  type: 'info' | 'success' | 'warning' | 'error';
  title?: string;
  children: React.ReactNode;
  dismissible?: boolean;
  onDismiss?: () => void;
  role?: 'alert' | 'status';
}

export function AccessibleAlert({
  type,
  title,
  children,
  dismissible = false,
  onDismiss,
  role = type === 'error' ? 'alert' : 'status',
}: AccessibleAlertProps): JSX.Element {
  const icons = {
    info: 'ℹ️',
    success: '✓',
    warning: '⚠',
    error: '✕',
  };

  return (
    <>
      <div className={`a11y-alert a11y-alert--${type}`} role={role} aria-live={role === 'alert' ? 'assertive' : 'polite'}>
        <span className="a11y-alert__icon" aria-hidden="true">
          {icons[type]}
        </span>
        <div className="a11y-alert__content">
          {title && <p className="a11y-alert__title">{title}</p>}
          <div className="a11y-alert__message">{children}</div>
        </div>
        {dismissible && onDismiss && (
          <button
            className="a11y-alert__dismiss"
            onClick={onDismiss}
            aria-label="Dismiss alert"
          >
            ✕
          </button>
        )}
      </div>
      <style>{alertStyles}</style>
    </>
  );
}

const alertStyles = `
  .a11y-alert {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 16px;
    border-radius: 8px;
    border-left: 4px solid;
  }

  .a11y-alert--info {
    background: #dbeafe;
    border-color: #3b82f6;
    color: #1e40af;
  }

  .a11y-alert--success {
    background: #dcfce7;
    border-color: #22c55e;
    color: #166534;
  }

  .a11y-alert--warning {
    background: #fef9c3;
    border-color: #eab308;
    color: #854d0e;
  }

  .a11y-alert--error {
    background: #fee2e2;
    border-color: #ef4444;
    color: #991b1b;
  }

  .a11y-alert__icon {
    font-size: 1.25rem;
    flex-shrink: 0;
  }

  .a11y-alert__content {
    flex: 1;
  }

  .a11y-alert__title {
    font-weight: 600;
    margin: 0 0 4px;
  }

  .a11y-alert__message {
    font-size: 0.875rem;
  }

  .a11y-alert__dismiss {
    background: none;
    border: none;
    cursor: pointer;
    padding: 4px;
    opacity: 0.7;
    min-width: 44px;
    min-height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .a11y-alert__dismiss:hover {
    opacity: 1;
  }

  .a11y-alert__dismiss:focus-visible {
    outline: 2px solid currentColor;
    outline-offset: 2px;
    border-radius: 4px;
  }
`;

/**
 * Progress indicator with screen reader support
 */
export interface AccessibleProgressProps {
  value: number;
  max?: number;
  label: string;
  showValue?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

export function AccessibleProgress({
  value,
  max = 100,
  label,
  showValue = true,
  size = 'md',
}: AccessibleProgressProps): JSX.Element {
  const percentage = Math.round((value / max) * 100);

  return (
    <>
      <div className={`a11y-progress a11y-progress--${size}`}>
        <div className="a11y-progress__header">
          <span className="a11y-progress__label">{label}</span>
          {showValue && (
            <span className="a11y-progress__value" aria-hidden="true">
              {percentage}%
            </span>
          )}
        </div>
        <div
          role="progressbar"
          aria-valuenow={value}
          aria-valuemin={0}
          aria-valuemax={max}
          aria-label={`${label}: ${percentage}%`}
          className="a11y-progress__track"
        >
          <div
            className="a11y-progress__fill"
            style={{ width: `${percentage}%` }}
          />
        </div>
      </div>
      <style>{progressStyles}</style>
    </>
  );
}

const progressStyles = `
  .a11y-progress {
    width: 100%;
  }

  .a11y-progress__header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 4px;
  }

  .a11y-progress__label {
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
  }

  .a11y-progress__value {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .a11y-progress__track {
    width: 100%;
    background: #e5e7eb;
    border-radius: 9999px;
    overflow: hidden;
  }

  .a11y-progress--sm .a11y-progress__track {
    height: 4px;
  }

  .a11y-progress--md .a11y-progress__track {
    height: 8px;
  }

  .a11y-progress--lg .a11y-progress__track {
    height: 12px;
  }

  .a11y-progress__fill {
    height: 100%;
    background: #3b82f6;
    border-radius: 9999px;
    transition: width 0.3s ease;
  }

  @media (prefers-reduced-motion: reduce) {
    .a11y-progress__fill {
      transition: none;
    }
  }
`;

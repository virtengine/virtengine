/**
 * Accessibility Test Utilities
 * VE-UI-002: WCAG 2.1 AA compliance testing
 *
 * These utilities provide automated accessibility testing using axe-core
 */
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import type { AxeResults, Result as AxeViolation } from 'axe-core';

// Conditionally import axe-core to avoid issues in non-browser environments
let axe: typeof import('axe-core') | null = null;
try {
  axe = require('axe-core');
} catch {
  // axe-core not available
}

/**
 * Accessibility test configuration
 */
export interface A11yTestConfig {
  /** Rules to run (default: WCAG 2.1 AA) */
  rules?: string[];
  /** Tags to filter rules by */
  tags?: string[];
  /** Selectors to exclude from testing */
  exclude?: string[];
  /** Whether to include incomplete results */
  includeIncomplete?: boolean;
}

/**
 * Default configuration for WCAG 2.1 AA compliance
 */
export const WCAG_21_AA_CONFIG: A11yTestConfig = {
  tags: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'],
  includeIncomplete: false,
};

/**
 * Format axe violations for readable output
 */
export function formatViolations(violations: AxeViolation[]): string {
  if (violations.length === 0) {
    return 'No accessibility violations found';
  }

  return violations
    .map((violation) => {
      const nodes = violation.nodes
        .map((node) => {
          const target = node.target.join(', ');
          const fix = node.failureSummary || 'No fix suggestion';
          return `    - Element: ${target}\n      Fix: ${fix}`;
        })
        .join('\n');

      return `
  ❌ ${violation.id}: ${violation.description}
     Impact: ${violation.impact}
     Help: ${violation.helpUrl}
     Affected elements:
${nodes}`;
    })
    .join('\n');
}

/**
 * Create an accessibility assertion for use in tests
 */
export function toHaveNoViolations(results: AxeResults): {
  pass: boolean;
  message: () => string;
} {
  const { violations } = results;
  const pass = violations.length === 0;

  return {
    pass,
    message: () =>
      pass
        ? 'Expected accessibility violations but found none'
        : `Found ${violations.length} accessibility violation(s):\n${formatViolations(violations)}`,
  };
}

/**
 * Run axe accessibility tests on an HTML element or string
 */
export async function runA11yTests(
  html: string | Element,
  config: A11yTestConfig = WCAG_21_AA_CONFIG
): Promise<AxeResults> {
  if (!axe) {
    throw new Error('axe-core is not available. Install it with: npm install -D axe-core');
  }

  // Create a container element if html string is provided
  let container: Element;
  let cleanupNeeded = false;

  if (typeof html === 'string') {
    if (typeof document === 'undefined') {
      throw new Error('runA11yTests requires a DOM environment for string inputs');
    }
    container = document.createElement('div');
    container.innerHTML = html;
    document.body.appendChild(container);
    cleanupNeeded = true;
  } else {
    container = html;
  }

  try {
    const options: any = {};

    if (config.rules) {
      options.rules = config.rules.reduce((acc: any, rule: string) => {
        acc[rule] = { enabled: true };
        return acc;
      }, {});
    }

    if (config.tags) {
      options.runOnly = { type: 'tag', values: config.tags };
    }

    if (config.exclude) {
      options.exclude = config.exclude.map((selector) => [selector]);
    }

    const results = await axe.run(container, options);

    if (!config.includeIncomplete) {
      results.incomplete = [];
    }

    return results;
  } finally {
    if (cleanupNeeded && container.parentNode) {
      container.parentNode.removeChild(container);
    }
  }
}

/**
 * Accessibility test helper for component testing
 */
export async function expectNoA11yViolations(
  html: string | Element,
  config: A11yTestConfig = WCAG_21_AA_CONFIG
): Promise<void> {
  const results = await runA11yTests(html, config);
  const assertion = toHaveNoViolations(results);

  if (!assertion.pass) {
    throw new Error(assertion.message());
  }
}

/**
 * Check color contrast ratio
 */
export function checkContrastRatio(
  foreground: string,
  background: string,
  largeText = false
): { ratio: number; passes: boolean; requirement: number } {
  const hexToRgb = (hex: string): [number, number, number] => {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    if (!result) throw new Error(`Invalid hex color: ${hex}`);
    return [parseInt(result[1], 16), parseInt(result[2], 16), parseInt(result[3], 16)];
  };

  const getLuminance = (r: number, g: number, b: number): number => {
    const [rs, gs, bs] = [r, g, b].map((c) => {
      const sRGB = c / 255;
      return sRGB <= 0.03928 ? sRGB / 12.92 : Math.pow((sRGB + 0.055) / 1.055, 2.4);
    });
    return 0.2126 * rs + 0.7152 * gs + 0.0722 * bs;
  };

  const fg = hexToRgb(foreground);
  const bg = hexToRgb(background);
  const l1 = getLuminance(...fg);
  const l2 = getLuminance(...bg);
  const lighter = Math.max(l1, l2);
  const darker = Math.min(l1, l2);
  const ratio = (lighter + 0.05) / (darker + 0.05);
  const requirement = largeText ? 3 : 4.5;

  return {
    ratio: Math.round(ratio * 100) / 100,
    passes: ratio >= requirement,
    requirement,
  };
}

/**
 * Check focus indicator visibility
 */
export function checkFocusIndicator(element: Element): {
  hasFocusStyle: boolean;
  outlineWidth: string;
  outlineColor: string;
  outlineOffset: string;
} {
  if (typeof window === 'undefined') {
    return {
      hasFocusStyle: false,
      outlineWidth: 'N/A',
      outlineColor: 'N/A',
      outlineOffset: 'N/A',
    };
  }

  // Focus the element
  if (element instanceof HTMLElement) {
    element.focus();
  }

  const styles = window.getComputedStyle(element);
  const outlineWidth = styles.outlineWidth;
  const outlineColor = styles.outlineColor;
  const outlineOffset = styles.outlineOffset;

  // Check if focus indicator is visible (not 0 width and not transparent)
  const hasFocusStyle =
    outlineWidth !== '0px' &&
    outlineColor !== 'transparent' &&
    outlineColor !== 'rgba(0, 0, 0, 0)';

  return {
    hasFocusStyle,
    outlineWidth,
    outlineColor,
    outlineOffset,
  };
}

/**
 * Check touch target size (WCAG 2.5.5 Level AAA / 2.5.8 Level AA)
 */
export function checkTouchTargetSize(element: Element): {
  width: number;
  height: number;
  meetsAA: boolean; // 24x24 px minimum
  meetsAAA: boolean; // 44x44 px minimum
} {
  const rect = element.getBoundingClientRect();

  return {
    width: rect.width,
    height: rect.height,
    meetsAA: rect.width >= 24 && rect.height >= 24,
    meetsAAA: rect.width >= 44 && rect.height >= 44,
  };
}

/**
 * Keyboard navigation test helper
 */
export interface KeyboardNavTestResult {
  focusableElements: number;
  tabOrder: string[];
  skipLinks: number;
  focusTrapFound: boolean;
}

export function analyzeKeyboardNav(container: Element): KeyboardNavTestResult {
  const focusableSelectors = [
    'a[href]',
    'button:not([disabled])',
    'input:not([disabled])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    '[tabindex]:not([tabindex="-1"])',
  ].join(', ');

  const focusable = container.querySelectorAll(focusableSelectors);
  const skipLinks = container.querySelectorAll('a.skip-link, [class*="skip"]');
  const focusTraps = container.querySelectorAll('[data-focus-trap], [role="dialog"]');

  const tabOrder = Array.from(focusable).map((el) => {
    const tag = el.tagName.toLowerCase();
    const id = el.id ? `#${el.id}` : '';
    const ariaLabel = el.getAttribute('aria-label');
    return `${tag}${id}${ariaLabel ? ` (${ariaLabel})` : ''}`;
  });

  return {
    focusableElements: focusable.length,
    tabOrder,
    skipLinks: skipLinks.length,
    focusTrapFound: focusTraps.length > 0,
  };
}

/**
 * ARIA attribute validation
 */
export function validateAriaAttributes(element: Element): {
  valid: boolean;
  issues: string[];
} {
  const issues: string[] = [];

  // Check for aria-labelledby pointing to existing element
  const labelledby = element.getAttribute('aria-labelledby');
  if (labelledby) {
    const ids = labelledby.split(' ');
    ids.forEach((id) => {
      if (!document.getElementById(id)) {
        issues.push(`aria-labelledby references non-existent id: ${id}`);
      }
    });
  }

  // Check for aria-describedby pointing to existing element
  const describedby = element.getAttribute('aria-describedby');
  if (describedby) {
    const ids = describedby.split(' ');
    ids.forEach((id) => {
      if (!document.getElementById(id)) {
        issues.push(`aria-describedby references non-existent id: ${id}`);
      }
    });
  }

  // Check for aria-controls pointing to existing element
  const controls = element.getAttribute('aria-controls');
  if (controls && !document.getElementById(controls)) {
    issues.push(`aria-controls references non-existent id: ${controls}`);
  }

  // Check for missing accessible name on interactive elements
  const interactiveRoles = ['button', 'link', 'checkbox', 'radio', 'textbox', 'combobox'];
  const role = element.getAttribute('role');
  if (
    (role && interactiveRoles.includes(role)) ||
    element.tagName.toLowerCase() === 'button' ||
    (element.tagName.toLowerCase() === 'a' && element.hasAttribute('href'))
  ) {
    const hasAccessibleName =
      element.textContent?.trim() ||
      element.getAttribute('aria-label') ||
      element.getAttribute('aria-labelledby') ||
      (element.tagName.toLowerCase() === 'input' &&
        (element as HTMLInputElement).type === 'submit');

    if (!hasAccessibleName) {
      issues.push(`Interactive element missing accessible name: ${element.outerHTML.slice(0, 100)}`);
    }
  }

  return {
    valid: issues.length === 0,
    issues,
  };
}

/**
 * Screen reader content check
 */
export function checkScreenReaderContent(container: Element): {
  hiddenElements: number;
  srOnlyElements: number;
  ariaLabels: number;
  liveRegions: number;
  headingStructure: { level: number; text: string }[];
} {
  const hiddenElements = container.querySelectorAll('[aria-hidden="true"]').length;
  const srOnlyElements = container.querySelectorAll('.sr-only, [class*="visually-hidden"]').length;
  const ariaLabels = container.querySelectorAll('[aria-label], [aria-labelledby]').length;
  const liveRegions = container.querySelectorAll('[aria-live]').length;

  const headings = container.querySelectorAll('h1, h2, h3, h4, h5, h6');
  const headingStructure = Array.from(headings).map((h) => ({
    level: parseInt(h.tagName.charAt(1), 10),
    text: h.textContent?.trim() || '',
  }));

  return {
    hiddenElements,
    srOnlyElements,
    ariaLabels,
    liveRegions,
    headingStructure,
  };
}

/**
 * Generate accessibility test report
 */
export interface A11yReport {
  url?: string;
  timestamp: string;
  violations: AxeViolation[];
  passes: number;
  incomplete: number;
  inapplicable: number;
  contrastIssues: string[];
  keyboardNav: KeyboardNavTestResult | null;
  screenReader: ReturnType<typeof checkScreenReaderContent> | null;
}

export async function generateA11yReport(
  container: Element,
  config: A11yTestConfig = WCAG_21_AA_CONFIG
): Promise<A11yReport> {
  const results = await runA11yTests(container, config);

  return {
    timestamp: new Date().toISOString(),
    violations: results.violations,
    passes: results.passes.length,
    incomplete: results.incomplete.length,
    inapplicable: results.inapplicable.length,
    contrastIssues: results.violations
      .filter((v) => v.id === 'color-contrast')
      .flatMap((v) => v.nodes.map((n) => n.target.join(', '))),
    keyboardNav: analyzeKeyboardNav(container),
    screenReader: checkScreenReaderContent(container),
  };
}

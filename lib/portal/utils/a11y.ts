/**
 * Accessibility Utilities
 * VE-UI-002: WCAG 2.1 AA compliance utilities
 */

/**
 * Unique ID counter for accessibility attributes
 */
let idCounter = 0;

/**
 * Generate a unique ID for accessibility attributes
 */
export function generateA11yId(prefix = 'a11y'): string {
  return `${prefix}-${++idCounter}`;
}

/**
 * Reset ID counter (for testing)
 */
export function resetA11yIdCounter(): void {
  idCounter = 0;
}

/**
 * ARIA live region announcer
 * Announces messages to screen readers
 */
export interface LiveAnnouncer {
  announce: (message: string, priority?: 'polite' | 'assertive') => void;
  clear: () => void;
}

let liveRegionPolite: HTMLElement | null = null;
let liveRegionAssertive: HTMLElement | null = null;

/**
 * Initialize live regions for screen reader announcements
 */
export function initLiveRegions(): void {
  if (typeof document === 'undefined') return;

  if (!liveRegionPolite) {
    liveRegionPolite = document.createElement('div');
    liveRegionPolite.setAttribute('aria-live', 'polite');
    liveRegionPolite.setAttribute('aria-atomic', 'true');
    liveRegionPolite.className = 'sr-only';
    liveRegionPolite.id = 'a11y-live-polite';
    document.body.appendChild(liveRegionPolite);
  }

  if (!liveRegionAssertive) {
    liveRegionAssertive = document.createElement('div');
    liveRegionAssertive.setAttribute('aria-live', 'assertive');
    liveRegionAssertive.setAttribute('aria-atomic', 'true');
    liveRegionAssertive.className = 'sr-only';
    liveRegionAssertive.id = 'a11y-live-assertive';
    document.body.appendChild(liveRegionAssertive);
  }
}

/**
 * Announce a message to screen readers
 */
export function announce(message: string, priority: 'polite' | 'assertive' = 'polite'): void {
  initLiveRegions();
  
  const region = priority === 'assertive' ? liveRegionAssertive : liveRegionPolite;
  if (region) {
    // Clear first to ensure re-announcement of same message
    region.textContent = '';
    // Use setTimeout to ensure the DOM update triggers announcement
    setTimeout(() => {
      region.textContent = message;
    }, 50);
  }
}

/**
 * Clear all live region announcements
 */
export function clearAnnouncements(): void {
  if (liveRegionPolite) liveRegionPolite.textContent = '';
  if (liveRegionAssertive) liveRegionAssertive.textContent = '';
}

/**
 * Focus trap for modal dialogs and other components
 */
export interface FocusTrap {
  activate: () => void;
  deactivate: () => void;
  updateContainer: (container: HTMLElement) => void;
}

/**
 * Get all focusable elements within a container
 */
export function getFocusableElements(container: HTMLElement): HTMLElement[] {
  const focusableSelectors = [
    'a[href]',
    'button:not([disabled])',
    'input:not([disabled])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    '[tabindex]:not([tabindex="-1"])',
    'audio[controls]',
    'video[controls]',
    '[contenteditable]:not([contenteditable="false"])',
  ].join(', ');

  return Array.from(container.querySelectorAll<HTMLElement>(focusableSelectors))
    .filter((el) => {
      // Filter out hidden elements
      const style = window.getComputedStyle(el);
      return style.display !== 'none' && style.visibility !== 'hidden';
    });
}

/**
 * Create a focus trap for a container element
 */
export function createFocusTrap(container: HTMLElement): FocusTrap {
  let active = false;
  let previousActiveElement: HTMLElement | null = null;

  function handleKeyDown(event: KeyboardEvent): void {
    if (event.key !== 'Tab' || !active) return;

    const focusable = getFocusableElements(container);
    if (focusable.length === 0) return;

    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    const activeElement = document.activeElement;

    if (event.shiftKey) {
      // Shift+Tab: go backwards
      if (activeElement === first || !container.contains(activeElement)) {
        event.preventDefault();
        last.focus();
      }
    } else {
      // Tab: go forwards
      if (activeElement === last || !container.contains(activeElement)) {
        event.preventDefault();
        first.focus();
      }
    }
  }

  return {
    activate() {
      if (active) return;
      active = true;
      previousActiveElement = document.activeElement as HTMLElement;
      document.addEventListener('keydown', handleKeyDown);
      
      // Focus first focusable element
      const focusable = getFocusableElements(container);
      if (focusable.length > 0) {
        focusable[0].focus();
      }
    },

    deactivate() {
      if (!active) return;
      active = false;
      document.removeEventListener('keydown', handleKeyDown);
      
      // Restore previous focus
      if (previousActiveElement && previousActiveElement.focus) {
        previousActiveElement.focus();
      }
    },

    updateContainer(newContainer: HTMLElement) {
      container = newContainer;
    },
  };
}

/**
 * Color contrast utilities for WCAG 2.1 AA compliance
 */

/**
 * Calculate relative luminance of a color
 * https://www.w3.org/WAI/WCAG21/Understanding/contrast-minimum.html
 */
export function getLuminance(r: number, g: number, b: number): number {
  const [rs, gs, bs] = [r, g, b].map((c) => {
    const sRGB = c / 255;
    return sRGB <= 0.03928 ? sRGB / 12.92 : Math.pow((sRGB + 0.055) / 1.055, 2.4);
  });
  return 0.2126 * rs + 0.7152 * gs + 0.0722 * bs;
}

/**
 * Calculate contrast ratio between two colors
 */
export function getContrastRatio(rgb1: [number, number, number], rgb2: [number, number, number]): number {
  const l1 = getLuminance(...rgb1);
  const l2 = getLuminance(...rgb2);
  const lighter = Math.max(l1, l2);
  const darker = Math.min(l1, l2);
  return (lighter + 0.05) / (darker + 0.05);
}

/**
 * Parse hex color to RGB
 */
export function hexToRgb(hex: string): [number, number, number] | null {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result
    ? [parseInt(result[1], 16), parseInt(result[2], 16), parseInt(result[3], 16)]
    : null;
}

/**
 * Check if color contrast meets WCAG 2.1 AA requirements
 * Normal text: 4.5:1
 * Large text (18pt+ or 14pt+ bold): 3:1
 */
export function meetsContrastRequirement(
  foreground: string,
  background: string,
  isLargeText = false
): boolean {
  const fg = hexToRgb(foreground);
  const bg = hexToRgb(background);
  
  if (!fg || !bg) return false;
  
  const ratio = getContrastRatio(fg, bg);
  return isLargeText ? ratio >= 3 : ratio >= 4.5;
}

/**
 * WCAG 2.1 AA compliant color pairs
 */
export const A11Y_COLORS = {
  // Status colors with WCAG AA compliant contrast
  success: {
    text: '#166534', // Green 800
    background: '#dcfce7', // Green 100
  },
  warning: {
    text: '#854d0e', // Yellow 800
    background: '#fef9c3', // Yellow 100
  },
  error: {
    text: '#991b1b', // Red 800
    background: '#fee2e2', // Red 100
  },
  info: {
    text: '#1e40af', // Blue 800
    background: '#dbeafe', // Blue 100
  },
  // Neutral colors
  text: {
    primary: '#111827', // Gray 900
    secondary: '#4b5563', // Gray 600
    disabled: '#6b7280', // Gray 500
  },
  border: {
    default: '#d1d5db', // Gray 300
    focus: '#3b82f6', // Blue 500
  },
} as const;

/**
 * Keyboard navigation helpers
 */

/**
 * Handle arrow key navigation in lists/grids
 */
export interface ArrowNavOptions {
  items: HTMLElement[];
  orientation?: 'horizontal' | 'vertical' | 'both';
  loop?: boolean;
  onFocus?: (element: HTMLElement, index: number) => void;
}

export function handleArrowNavigation(
  event: KeyboardEvent,
  options: ArrowNavOptions
): void {
  const { items, orientation = 'vertical', loop = true, onFocus } = options;
  
  if (items.length === 0) return;
  
  const currentIndex = items.findIndex((item) => item === document.activeElement);
  let nextIndex = currentIndex;
  
  const isVertical = orientation === 'vertical' || orientation === 'both';
  const isHorizontal = orientation === 'horizontal' || orientation === 'both';
  
  switch (event.key) {
    case 'ArrowUp':
      if (isVertical) {
        event.preventDefault();
        nextIndex = currentIndex > 0 ? currentIndex - 1 : (loop ? items.length - 1 : 0);
      }
      break;
    case 'ArrowDown':
      if (isVertical) {
        event.preventDefault();
        nextIndex = currentIndex < items.length - 1 ? currentIndex + 1 : (loop ? 0 : items.length - 1);
      }
      break;
    case 'ArrowLeft':
      if (isHorizontal) {
        event.preventDefault();
        nextIndex = currentIndex > 0 ? currentIndex - 1 : (loop ? items.length - 1 : 0);
      }
      break;
    case 'ArrowRight':
      if (isHorizontal) {
        event.preventDefault();
        nextIndex = currentIndex < items.length - 1 ? currentIndex + 1 : (loop ? 0 : items.length - 1);
      }
      break;
    case 'Home':
      event.preventDefault();
      nextIndex = 0;
      break;
    case 'End':
      event.preventDefault();
      nextIndex = items.length - 1;
      break;
    default:
      return;
  }
  
  if (nextIndex !== currentIndex && items[nextIndex]) {
    items[nextIndex].focus();
    onFocus?.(items[nextIndex], nextIndex);
  }
}

/**
 * Roving tabindex management for composite widgets
 */
export function manageRovingTabindex(items: HTMLElement[], activeIndex: number): void {
  items.forEach((item, index) => {
    item.setAttribute('tabindex', index === activeIndex ? '0' : '-1');
  });
}

/**
 * Screen reader only styles (CSS-in-JS)
 */
export const srOnlyStyles = {
  position: 'absolute' as const,
  width: '1px',
  height: '1px',
  padding: 0,
  margin: '-1px',
  overflow: 'hidden',
  clip: 'rect(0, 0, 0, 0)',
  whiteSpace: 'nowrap' as const,
  border: 0,
};

/**
 * Focus visible styles (CSS-in-JS)
 */
export const focusVisibleStyles = {
  outline: '2px solid #3b82f6',
  outlineOffset: '2px',
};

/**
 * Reduced motion check
 */
export function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined') return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * High contrast mode check
 */
export function prefersHighContrast(): boolean {
  if (typeof window === 'undefined') return false;
  return window.matchMedia('(prefers-contrast: more)').matches;
}

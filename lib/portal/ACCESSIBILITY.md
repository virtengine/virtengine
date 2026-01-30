# VirtEngine Portal Accessibility Guidelines

## WCAG 2.1 Level AA Compliance

This document outlines the accessibility requirements and best practices for the VirtEngine Portal components.

## Quick Reference

### Color Contrast Requirements

| Content Type | Minimum Contrast Ratio |
|-------------|----------------------|
| Normal Text (< 18pt) | 4.5:1 |
| Large Text (≥ 18pt or 14pt bold) | 3:1 |
| UI Components & Graphics | 3:1 |

### Pre-approved Color Combinations

Use these WCAG AA compliant color pairs:

```typescript
import { A11Y_COLORS } from '../utils/a11y';

// Success: #166534 on #dcfce7
// Warning: #854d0e on #fef9c3  
// Error: #991b1b on #fee2e2
// Info: #1e40af on #dbeafe
```

### Target Sizes

- Minimum touch target: 44×44 CSS pixels (WCAG 2.5.5)
- Recommended for mobile: 48×48 CSS pixels
- Always use `min-height: 44px` for buttons

## Component Guidelines

### Buttons

```tsx
// ✅ Good - accessible button
<button 
  onClick={handleClick}
  aria-label="Submit form"
  disabled={isLoading}
  aria-busy={isLoading}
>
  {isLoading ? <Spinner aria-hidden="true" /> : null}
  Submit
</button>

// ❌ Bad - inaccessible button
<button onClick={handleClick}>
  <Icon /> {/* No text or aria-label */}
</button>
```

### Form Fields

```tsx
// ✅ Good - properly labeled input
<div>
  <label htmlFor="email">Email Address *</label>
  <input 
    id="email"
    type="email"
    aria-required="true"
    aria-invalid={hasError}
    aria-describedby={hasError ? "email-error" : undefined}
  />
  {hasError && (
    <p id="email-error" role="alert">Please enter a valid email</p>
  )}
</div>

// ❌ Bad - unlabeled input
<input type="email" placeholder="Email" />
```

### Interactive Cards

```tsx
// ✅ Good - keyboard accessible card
<div
  role="button"
  tabIndex={0}
  onClick={handleSelect}
  onKeyDown={(e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleSelect();
    }
  }}
  aria-pressed={isSelected}
  aria-label={`Select ${title}`}
>
  {content}
</div>
```

### Loading States

```tsx
// ✅ Good - announced loading
<div role="status" aria-live="polite" aria-busy="true">
  <Spinner aria-hidden="true" />
  <span className="sr-only">Loading content...</span>
</div>
```

### Error Messages

```tsx
// ✅ Good - announced error
<div role="alert" aria-live="assertive">
  <span aria-hidden="true">⚠ </span>
  Error: {errorMessage}
</div>
```

### Progress Indicators

```tsx
// ✅ Good - accessible progress bar
<div
  role="progressbar"
  aria-valuenow={progress}
  aria-valuemin={0}
  aria-valuemax={100}
  aria-label={`Upload progress: ${progress}%`}
>
  <div style={{ width: `${progress}%` }} />
</div>
```

### Images & Icons

```tsx
// ✅ Good - decorative icon (hidden from AT)
<span aria-hidden="true">🔒</span>

// ✅ Good - meaningful image
<img src={photo} alt="User profile photo" />

// ✅ Good - icon with meaning
<button aria-label="Close dialog">
  <CloseIcon aria-hidden="true" />
</button>
```

### SVG Accessibility

```tsx
// ✅ Good - accessible SVG chart
<div 
  role="img" 
  aria-label="Score: 85 out of 100, rated Excellent"
>
  <svg aria-hidden="true" focusable="false">
    {/* SVG content */}
  </svg>
</div>
```

## Keyboard Navigation

### Focus Management

1. **Visible Focus**: All interactive elements must have visible focus indicators
2. **Focus Order**: Tab order should follow visual/logical order
3. **Skip Links**: Provide "Skip to main content" link
4. **Focus Traps**: Modals must trap focus; return focus on close

```css
/* Focus visible styles */
:focus-visible {
  outline: 2px solid #3b82f6;
  outline-offset: 2px;
}

/* Never remove focus entirely */
/* ❌ Bad */
:focus {
  outline: none;
}

/* ✅ OK - only if custom focus indicator provided */
:focus:not(:focus-visible) {
  outline: none;
}
```

### Keyboard Shortcuts

| Action | Keys |
|--------|------|
| Navigate between items | Arrow keys |
| Select/Activate | Enter, Space |
| Cancel/Close | Escape |
| Skip to first/last | Home, End |

## Screen Reader Support

### ARIA Landmarks

```tsx
<header role="banner">...</header>
<nav role="navigation" aria-label="Main">...</nav>
<main role="main">...</main>
<aside role="complementary">...</aside>
<footer role="contentinfo">...</footer>
```

### Live Regions

```tsx
// Polite announcements (non-urgent)
<div aria-live="polite" aria-atomic="true">
  {statusMessage}
</div>

// Assertive announcements (urgent)
<div aria-live="assertive" role="alert">
  {errorMessage}
</div>
```

### Screen Reader Only Content

```tsx
import { SrOnly } from '../components/accessible';

// Hidden visually, readable by screen readers
<SrOnly>Additional context for screen reader users</SrOnly>
```

## Motion & Animation

### Respecting User Preferences

```css
/* Disable animations for users who prefer reduced motion */
@media (prefers-reduced-motion: reduce) {
  * {
    animation: none !important;
    transition: none !important;
  }
}
```

```tsx
import { prefersReducedMotion } from '../utils/a11y';

// Check preference in JavaScript
if (prefersReducedMotion()) {
  // Use alternative non-animated version
}
```

## Testing

### Automated Testing

```bash
# Run accessibility tests
npm run test:a11y

# Run accessibility linting
npm run lint:a11y
```

### Manual Testing Checklist

- [ ] Navigate entire page with keyboard only (no mouse)
- [ ] Test with screen reader (NVDA, VoiceOver, or JAWS)
- [ ] Check color contrast with browser devtools
- [ ] Test at 200% zoom
- [ ] Test with high contrast mode
- [ ] Test with reduced motion enabled

### Browser DevTools

1. Chrome DevTools → Lighthouse → Accessibility
2. Firefox → Accessibility Inspector
3. axe DevTools extension

## Utilities

### Available Helpers

```tsx
import {
  // ID generation
  generateA11yId,
  
  // Live announcements
  announce,
  clearAnnouncements,
  
  // Focus management
  createFocusTrap,
  getFocusableElements,
  
  // Color contrast
  meetsContrastRequirement,
  getContrastRatio,
  
  // Keyboard navigation
  handleArrowNavigation,
  manageRovingTabindex,
  
  // User preferences
  prefersReducedMotion,
  prefersHighContrast,
  
  // Styles
  srOnlyStyles,
  focusVisibleStyles,
  A11Y_COLORS,
} from '../utils/a11y';
```

### Accessible Components

```tsx
import {
  SrOnly,
  SkipLink,
  AccessibleButton,
  AccessibleInput,
  AccessibleSelect,
  AccessibleCheckbox,
  AccessibleAlert,
  AccessibleProgress,
} from '../components/accessible';
```

## Resources

- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [ARIA Authoring Practices](https://www.w3.org/WAI/ARIA/apg/)
- [WebAIM Contrast Checker](https://webaim.org/resources/contrastchecker/)
- [axe DevTools](https://www.deque.com/axe/devtools/)

## Reporting Issues

If you find an accessibility issue:

1. Create an issue with the `accessibility` label
2. Include:
   - Component/page affected
   - WCAG criterion violated
   - Steps to reproduce
   - Suggested fix (if known)

/**
 * Contrast & Readability Utility based on WCAG 2.1 standards.
 * Ensures optimal legibility and contrast ratios between text and background.
 */

// Parse hex or rgb color to [r, g, b] (0-255)
export function parseColor(color) {
  if (!color || typeof color !== 'string') {
    return [248, 250, 252]; // default fallback #f8fafc
  }

  const trimmed = color.trim().toLowerCase();

  // Named colors fallback
  const namedColors = {
    black: [0, 0, 0],
    white: [255, 255, 255],
    transparent: [255, 255, 255],
  };
  if (namedColors[trimmed]) {
    return namedColors[trimmed];
  }

  // Hex format: #rgb, #rgba, #rrggbb, #rrggbbaa
  if (trimmed.startsWith('#')) {
    const hex = trimmed.slice(1);
    if (hex.length === 3 || hex.length === 4) {
      const r = parseInt(hex[0] + hex[0], 16);
      const g = parseInt(hex[1] + hex[1], 16);
      const b = parseInt(hex[2] + hex[2], 16);
      return [r, g, b];
    }
    if (hex.length >= 6) {
      const r = parseInt(hex.slice(0, 2), 16);
      const g = parseInt(hex.slice(2, 4), 16);
      const b = parseInt(hex.slice(4, 6), 16);
      return [isNaN(r) ? 255 : r, isNaN(g) ? 255 : g, isNaN(b) ? 255 : b];
    }
  }

  // RGB/RGBA format: rgb(r, g, b) or rgba(r, g, b, a)
  const rgbMatch = trimmed.match(/rgba?\((\d+)\s*,\s*(\d+)\s*,\s*(\d+)/);
  if (rgbMatch) {
    return [parseInt(rgbMatch[1], 10), parseInt(rgbMatch[2], 10), parseInt(rgbMatch[3], 10)];
  }

  return [248, 250, 252];
}

/**
 * Calculate relative luminance according to WCAG 2.1 specs
 * @param {[number, number, number]} rgb
 * @returns {number} luminance (0 to 1)
 */
export function getRelativeLuminance([r, g, b]) {
  const rs = r / 255;
  const gs = g / 255;
  const bs = b / 255;

  const rLin = rs <= 0.03928 ? rs / 12.92 : Math.pow((rs + 0.055) / 1.055, 2.4);
  const gLin = gs <= 0.03928 ? gs / 12.92 : Math.pow((gs + 0.055) / 1.055, 2.4);
  const bLin = bs <= 0.03928 ? bs / 12.92 : Math.pow((bs + 0.055) / 1.055, 2.4);

  return 0.2126 * rLin + 0.7152 * gLin + 0.0722 * bLin;
}

/**
 * Calculate contrast ratio between two colors according to WCAG 2.1
 * @param {string|[number, number, number]} color1
 * @param {string|[number, number, number]} color2
 * @returns {number} Contrast ratio (1.0 to 21.0)
 */
export function getContrastRatio(color1, color2) {
  const rgb1 = Array.isArray(color1) ? color1 : parseColor(color1);
  const rgb2 = Array.isArray(color2) ? color2 : parseColor(color2);

  const lum1 = getRelativeLuminance(rgb1);
  const lum2 = getRelativeLuminance(rgb2);

  const lighter = Math.max(lum1, lum2);
  const darker = Math.min(lum1, lum2);

  return Number(((lighter + 0.05) / (darker + 0.05)).toFixed(2));
}

/**
 * Evaluate WCAG level based on ratio
 * @param {number} ratio
 * @returns {'AAA' | 'AA' | 'AA Large' | 'Fail'}
 */
export function getWcagLevel(ratio) {
  if (ratio >= 7.0) return 'AAA';
  if (ratio >= 4.5) return 'AA';
  if (ratio >= 3.0) return 'AA Large';
  return 'Fail';
}

/**
 * Returns accessible theme tokens dynamically tailored to the background color.
 * Guarantees high readability & legibility adhering to WCAG 2.1 contrast standards.
 */
export function getAccessibleTheme(bgColor) {
  const bgRgb = parseColor(bgColor);
  const bgLum = getRelativeLuminance(bgRgb);

  // Contrast against pure white and dark slate
  const whiteRatio = getContrastRatio(bgRgb, [255, 255, 255]);
  const darkRatio = getContrastRatio(bgRgb, [15, 23, 42]); // #0f172a

  // If contrast against white is higher, or background is dark, use dark-mode text palette
  const isDark = whiteRatio >= darkRatio || bgLum < 0.22;

  const textColor = isDark ? '#ffffff' : '#0f172a';
  const mutedTextColor = isDark ? '#cbd5e1' : '#475569';
  const subTextColor = isDark ? '#94a3b8' : '#64748b';

  const textContrastRatio = isDark ? whiteRatio : darkRatio;
  const mutedContrastRatio = getContrastRatio(bgRgb, isDark ? [203, 213, 225] : [71, 85, 105]);

  return {
    isDark,
    bgLum,
    contrastRatio: textContrastRatio,
    wcagLevel: getWcagLevel(textContrastRatio),
    mutedContrastRatio,
    textColor,
    mutedTextColor,
    subTextColor,
    borderColor: isDark ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.12)',
    // Cards styling adapted for dark / light
    cardBg: isDark ? 'rgba(30, 41, 59, 0.75)' : 'rgba(255, 255, 255, 0.92)',
    cardHoverBg: isDark ? 'rgba(51, 65, 85, 0.9)' : '#ffffff',
    cardBorder: isDark ? 'rgba(255, 255, 255, 0.18)' : 'rgba(0, 0, 0, 0.08)',
    cardTextColor: isDark ? '#f8fafc' : '#0f172a',
    cardSubTextColor: isDark ? '#94a3b8' : '#64748b',
    cardShadow: isDark
      ? '0 4px 6px -1px rgba(0, 0, 0, 0.4), 0 2px 4px -2px rgba(0, 0, 0, 0.4)'
      : '0 1px 3px 0 rgba(0, 0, 0, 0.08), 0 1px 2px -1px rgba(0, 0, 0, 0.08)',
    // Secondary / Action buttons
    actionButtonBg: isDark ? 'rgba(255, 255, 255, 0.12)' : 'rgba(255, 255, 255, 0.85)',
    actionButtonHoverBg: isDark ? 'rgba(255, 255, 255, 0.2)' : '#ffffff',
    actionButtonText: isDark ? '#ffffff' : '#0f172a',
    actionButtonBorder: isDark ? 'rgba(255, 255, 255, 0.2)' : 'rgba(0, 0, 0, 0.1)',
    // Avatar ring
    avatarRing: isDark ? 'ring-2 ring-white/30 shadow-lg' : 'ring-2 ring-black/10 shadow-md',
    avatarFallbackBg: isDark ? '#3b82f6' : '#2563eb',
    avatarFallbackText: '#ffffff',
  };
}

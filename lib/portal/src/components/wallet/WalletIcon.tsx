'use client';

import type { WalletType } from '../../wallet/types';

export interface WalletIconProps {
  walletType: WalletType;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const SIZE_MAP = {
  sm: 20,
  md: 32,
  lg: 48,
} as const;

/** Keplr wallet icon SVG */
export const KEPLR_ICON_SVG = `<svg viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
  <rect width="40" height="40" rx="8" fill="#5C6BC0"/>
  <path d="M12 10v20l8-10 8 10V10l-8 6-8-6z" fill="#fff"/>
</svg>`;

/** Leap wallet icon SVG */
export const LEAP_ICON_SVG = `<svg viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
  <rect width="40" height="40" rx="8" fill="#32D583"/>
  <path d="M20 8c-6.6 0-12 5.4-12 12s5.4 12 12 12 12-5.4 12-12S26.6 8 20 8zm0 20c-4.4 0-8-3.6-8-8s3.6-8 8-8 8 3.6 8 8-3.6 8-8 8z" fill="#fff"/>
  <path d="M20 14l4 6-4 6-4-6 4-6z" fill="#fff"/>
</svg>`;

/** Cosmostation wallet icon SVG */
export const COSMOSTATION_ICON_SVG = `<svg viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
  <rect width="40" height="40" rx="8" fill="#9C6CFF"/>
  <circle cx="20" cy="20" r="10" stroke="#fff" stroke-width="2" fill="none"/>
  <circle cx="20" cy="20" r="4" fill="#fff"/>
  <path d="M20 10v4M20 26v4M10 20h4M26 20h4" stroke="#fff" stroke-width="2"/>
</svg>`;

/** WalletConnect icon SVG */
export const WALLETCONNECT_ICON_SVG = `<svg viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
  <rect width="40" height="40" rx="8" fill="#3B99FC"/>
  <path d="M12.5 16c4.1-4 10.9-4 15 0l.5.5c.2.2.2.5 0 .7l-1.7 1.6c-.1.1-.3.1-.4 0l-.7-.6c-2.9-2.8-7.6-2.8-10.5 0l-.7.7c-.1.1-.3.1-.4 0l-1.7-1.6c-.2-.2-.2-.5 0-.7l.6-.6zm18.5 3.4l1.5 1.5c.2.2.2.5 0 .7l-6.8 6.5c-.2.2-.5.2-.7 0l-4.8-4.6c0-.1-.1-.1-.2 0l-4.8 4.6c-.2.2-.5.2-.7 0l-6.8-6.5c-.2-.2-.2-.5 0-.7l1.5-1.5c.2-.2.5-.2.7 0l4.8 4.6c0 .1.1.1.2 0l4.8-4.6c.2-.2.5-.2.7 0l4.8 4.6c0 .1.1.1.2 0l4.8-4.6c.2-.2.5-.2.7 0z" fill="#fff"/>
</svg>`;

const WALLET_ICONS: Record<WalletType, string> = {
  keplr: KEPLR_ICON_SVG,
  leap: LEAP_ICON_SVG,
  cosmostation: COSMOSTATION_ICON_SVG,
  walletconnect: WALLETCONNECT_ICON_SVG,
};

export function WalletIcon({ walletType, size = 'md', className }: WalletIconProps) {
  const dimension = SIZE_MAP[size];
  const svgContent = WALLET_ICONS[walletType];

  return (
    <span
      className={className}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: dimension,
        height: dimension,
        flexShrink: 0,
      }}
      role="img"
      aria-label={`${walletType} wallet icon`}
      dangerouslySetInnerHTML={{ __html: svgContent }}
    />
  );
}

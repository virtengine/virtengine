import type { Metadata, Viewport } from 'next';
import { Inter, JetBrains_Mono } from 'next/font/google';
import { Providers } from './providers';
import { SkipToContent } from '@/components/shared';
import './globals.css';

const inter = Inter({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-sans',
});

const jetbrainsMono = JetBrains_Mono({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-mono',
});

export const metadata: Metadata = {
  title: {
    default: 'VirtEngine Portal',
    template: '%s | VirtEngine Portal',
  },
  description:
    'Decentralized cloud computing marketplace with ML-powered identity verification. Deploy workloads, manage HPC jobs, and access compute resources.',
  keywords: [
    'cloud computing',
    'decentralized',
    'blockchain',
    'HPC',
    'identity verification',
    'marketplace',
  ],
  authors: [{ name: 'VirtEngine Team' }],
  creator: 'VirtEngine',
  publisher: 'VirtEngine',
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  metadataBase: new URL(process.env.NEXT_PUBLIC_APP_URL || 'https://portal.virtengine.io'),
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: 'https://portal.virtengine.io',
    siteName: 'VirtEngine Portal',
    title: 'VirtEngine Portal',
    description: 'Decentralized cloud computing marketplace',
    images: [
      {
        url: '/og-image.png',
        width: 1200,
        height: 630,
        alt: 'VirtEngine Portal',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'VirtEngine Portal',
    description: 'Decentralized cloud computing marketplace',
    images: ['/og-image.png'],
    creator: '@virtengine',
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  icons: {
    icon: '/favicon.ico',
    shortcut: '/favicon-16x16.png',
    apple: '/apple-touch-icon.png',
  },
  manifest: '/manifest.json',
};

export const viewport: Viewport = {
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: '#ffffff' },
    { media: '(prefers-color-scheme: dark)', color: '#0a0a0a' },
  ],
  width: 'device-width',
  initialScale: 1,
  maximumScale: 5,
  viewportFit: 'cover',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" dir="ltr" suppressHydrationWarning>
      <body
        className={`${inter.variable} ${jetbrainsMono.variable} min-h-screen font-sans antialiased`}
      >
        <SkipToContent />
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}

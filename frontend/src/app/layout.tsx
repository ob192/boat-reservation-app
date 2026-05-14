import type { Metadata, Viewport } from 'next';
import { Playfair_Display, DM_Sans } from 'next/font/google';
import './globals.css';
import { QueryProvider } from '@/shared/providers/QueryProvider';
import { SupabaseProvider } from '@/shared/providers/SupabaseProvider';

const playfair = Playfair_Display({
  subsets: ['latin', 'cyrillic'],
  variable: '--font-playfair',
  display: 'swap',
  weight: ['400', '700'],
  style: ['normal', 'italic'],
});

const dmSans = DM_Sans({
  subsets: ['latin'],
  variable: '--font-dm-sans',
  display: 'swap',
  weight: ['300', '400', '500'],
});

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
  viewportFit: 'cover',
};

export const metadata: Metadata = {
  title: 'SUP Chernihiv — Оренда SUP-бордів',
  description: 'Орендуй SUP-борд у Чернігові та насолодись водною прогулянкою у власному темпі.',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="uk" className={`${playfair.variable} ${dmSans.variable}`}>
      <body>
        <div className="bg-ocean" aria-hidden="true">
          <div className="wave-layer" />
          <div className="wave-layer" />
          <div className="wave-layer" />
        </div>
        <QueryProvider>
          <SupabaseProvider>
            {children}
          </SupabaseProvider>
        </QueryProvider>
      </body>
    </html>
  );
}

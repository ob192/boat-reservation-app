import type { Metadata } from 'next';
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

export const metadata: Metadata = {
  title: 'Harbour & Wave — Бронювання човнів',
  description: 'Забронюйте човен і насолоджуйтесь водною прогулянкою у власному ритмі.',
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

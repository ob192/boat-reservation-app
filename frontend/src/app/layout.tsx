import type { Metadata, Viewport } from 'next';
import { Playfair_Display, DM_Sans, Oswald } from 'next/font/google';
import './globals.css';
import { QueryProvider } from '@/shared/providers/QueryProvider';
import { SupabaseProvider } from '@/shared/providers/SupabaseProvider';
import { PromoCapture } from '@/features/booking/components/client/PromoCapture';
import Script from "next/script";
import {FB_PIXEL_ID} from "@/shared/lib/fbq";

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

const oswald = Oswald({
  subsets: ['latin', 'cyrillic'],
  variable: '--font-oswald',
  display: 'swap',
  weight: ['300', '400', '500', '600', '700'],
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
        <html lang="uk" className={`${playfair.variable} ${dmSans.variable} ${oswald.variable}`}>
        <body style={{ fontFamily: 'var(--font-oswald), Oswald, sans-serif' }}>

        {/* ── Meta Pixel ── */}
        <Script id="fb-pixel" strategy="afterInteractive">
            {`
          !function(f,b,e,v,n,t,s)
          {if(f.fbq)return;n=f.fbq=function(){n.callMethod?
          n.callMethod.apply(n,arguments):n.queue.push(arguments)};
          if(!f._fbq)f._fbq=n;n.push=n;n.loaded=!0;n.version='2.0';
          n.queue=[];t=b.createElement(e);t.async=!0;
          t.src=v;s=b.getElementsByTagName(e)[0];
          s.parentNode.insertBefore(t,s)}(window, document,'script',
          'https://connect.facebook.net/en_US/fbevents.js');
          fbq('init', '${FB_PIXEL_ID}');
          fbq('track', 'PageView');
        `}
        </Script>
        <noscript>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
                height="1"
                width="1"
                style={{ display: 'none' }}
                src={`https://www.facebook.com/tr?id=${FB_PIXEL_ID}&ev=PageView&noscript=1`}
                alt=""
            />
        </noscript>

        <div className="bg-ocean" aria-hidden="true">
            <div className="wave-layer" />
            <div className="wave-layer" />
            <div className="wave-layer" />
        </div>
        <QueryProvider>
            <SupabaseProvider>
                <PromoCapture />
                {children}
            </SupabaseProvider>
        </QueryProvider>
        </body>
        </html>
    );
}

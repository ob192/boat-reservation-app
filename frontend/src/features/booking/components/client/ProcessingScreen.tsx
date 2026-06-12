'use client';

import Link from 'next/link';
import Script from 'next/script';
import { useState, useEffect, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { Instrument_Serif, Inter_Tight, JetBrains_Mono } from 'next/font/google';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useBookingStatus } from '@/features/booking/hooks';
import { MESSAGES } from '@/features/booking/messages';
import { routeLabel } from '@/features/booking/routes';
import { formatCurrency } from '@/shared/lib/currency';
import { fbqTrack } from '@/shared/lib/fbq';

const instrument = Instrument_Serif({
    subsets: ['latin'],
    weight: ['400'],
    style: ['normal', 'italic'],
    variable: '--bk-serif',
    display: 'swap',
});
const interTight = Inter_Tight({
    subsets: ['latin', 'cyrillic'],
    weight: ['400', '500', '600'],
    variable: '--bk-sans',
    display: 'swap',
});
const jetbrains = JetBrains_Mono({
    subsets: ['latin'],
    weight: ['400', '500'],
    variable: '--bk-mono',
    display: 'swap',
});

// ─── Marina ────────────────────────────────────────────────────────────────
const MARINA = {
    name: 'SUP Chernihiv',
    address: 'Чернігів, Україна',
    lat: 51.51005849458658,
    lng: 31.356877759864490,
    placeId: '0x46d5490079167cf9:0x9c86ca7f773c93ad',
    mapsEmbedSrc: 'https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d1976.2723499952258!2d31.35687775986449!3d51.51005849458658!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x46d5490079167cf9%3A0x9c86ca7f773c93ad!2ssup_che!5e0!3m2!1sen!2sua!4v1780059190547!5m2!1sen!2sua',
} as const;

const SESSION_DURATION_MINUTES = 120;

// ─── Helpers ───────────────────────────────────────────────────────────────
function buildCalendarUrl({ title, description, location, startISO, endISO }: {
    title: string; description: string; location: string; startISO: string; endISO: string;
}): string {
    const fmt = (iso: string) => iso.replace(/[-:]/g, '').replace(/\.\d{3}/, '');
    return `https://calendar.google.com/calendar/render?${new URLSearchParams({
        action: 'TEMPLATE', text: title, details: description,
        location, dates: `${fmt(startISO)}/${fmt(endISO)}`,
    }).toString()}`;
}

function buildStartDate(date: string, time: string): Date {
    const [h, m] = time.split(':').map(Number);
    const d = new Date(`${date}T00:00:00`);
    d.setHours(h, m, 0, 0);
    return d;
}

function buildMapsOpenUrl(lat: number, lng: number, placeId: string) {
    return `https://www.google.com/maps/search/?api=1&query=${lat},${lng}&query_place_id=${placeId}`;
}

function useIsIOS(): boolean {
    const [isIOS, setIsIOS] = useState(false);
    useEffect(() => {
        const ua = navigator.userAgent || '';
        const iOS =
            /iPad|iPhone|iPod/.test(ua) ||
            // iPadOS 13+ reports as "MacIntel" but has touch points
            (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1);
        setIsIOS(iOS);
    }, []);
    return isIOS;
}

function buildICS({ title, description, location, startISO, endISO }: {
    title: string; description: string; location: string; startISO: string; endISO: string;
}): string {
    // Keep UTC (Z) — startISO/endISO already come from Date.toISOString()
    const fmt = (iso: string) => iso.replace(/[-:]/g, '').replace(/\.\d{3}/, '');
    const esc = (s: string) =>
        s.replace(/\\/g, '\\\\').replace(/;/g, '\\;').replace(/,/g, '\\,').replace(/\n/g, '\\n');

    return [
        'BEGIN:VCALENDAR',
        'VERSION:2.0',
        'PRODID:-//SUP Chernihiv//Booking//UK',
        'CALSCALE:GREGORIAN',
        'METHOD:PUBLISH',
        'BEGIN:VEVENT',
        `UID:${crypto.randomUUID()}@sup-chernihiv`,
        `DTSTAMP:${fmt(new Date().toISOString())}`,
        `DTSTART:${fmt(startISO)}`,
        `DTEND:${fmt(endISO)}`,
        `SUMMARY:${esc(title)}`,
        `DESCRIPTION:${esc(description)}`,
        `LOCATION:${esc(location)}`,
        'END:VEVENT',
        'END:VCALENDAR',
    ].join('\r\n');
}

// ─── Icons ─────────────────────────────────────────────────────────────────
function CalendarIcon() {
    return (
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <rect x="3" y="4" width="18" height="18" rx="2" stroke="currentColor" strokeWidth="1.7" fill="none" />
            <path d="M3 9h18" stroke="currentColor" strokeWidth="1.7" />
            <path d="M8 2v4M16 2v4" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" />
            <rect x="7" y="13" width="3" height="3" rx="0.5" fill="currentColor" />
            <rect x="14" y="13" width="3" height="3" rx="0.5" fill="currentColor" />
        </svg>
    );
}

function MailIcon() {
    return (
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <rect x="2" y="4" width="20" height="16" rx="2" stroke="currentColor" strokeWidth="1.7" fill="none" />
            <path d="M2 7l10 7 10-7" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" />
        </svg>
    );
}

function SunIcon() {
    return (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <circle cx="12" cy="12" r="4" stroke="currentColor" strokeWidth="1.6" />
            <path d="M12 2v2M12 20v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M2 12h2M20 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"
                  stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
        </svg>
    );
}

function DropIcon() {
    return (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M12 2C12 2 5 10 5 15a7 7 0 0 0 14 0c0-5-7-13-7-13z"
                  stroke="currentColor" strokeWidth="1.6" strokeLinejoin="round" />
        </svg>
    );
}

function BagIcon() {
    return (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M6 2l-2 6h16l-2-6" stroke="currentColor" strokeWidth="1.6" strokeLinejoin="round" />
            <rect x="2" y="8" width="20" height="13" rx="2" stroke="currentColor" strokeWidth="1.6" fill="none" />
            <path d="M9 12a3 3 0 0 0 6 0" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
        </svg>
    );
}

function AnchorIcon() {
    return (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <circle cx="12" cy="6" r="3" stroke="currentColor" strokeWidth="1.6" fill="none" />
            <path d="M12 9v12M5 12H2a10 10 0 0 0 20 0h-3"
                  stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
        </svg>
    );
}

// ─── Checklist data ────────────────────────────────────────────────────────
const CHECKLIST = [
    { icon: <SunIcon />, title: 'Сонцезахист', hint: 'Крем SPF 50+ та окуляри' },
    { icon: <DropIcon />, title: 'Вода', hint: 'Мін. 1 л на особу' },
    { icon: <BagIcon />, title: 'Гермомішок', hint: 'Для телефону та речей' },
    { icon: <AnchorIcon />, title: 'Одяг, що сохне', hint: 'Купальник або шорти' },
];

// ─── FAQ data ──────────────────────────────────────────────────────────────
const FAQ = [
    {
        q: 'Потрібен досвід веслування?',
        a: 'Ні. Перед виходом проводимо короткий інструктаж — близько 10 хвилин. Цього достатньо для впевненого старту.',
    },
    {
        q: 'Що буде, якщо зіпсується погода?',
        a: 'За несприятливих умов безкоштовно перенесемо сесію або повернемо повну вартість — на ваш вибір.',
    },
    {
        q: 'Чи можна з дітьми?',
        a: 'Так. Діти від 45 кг — на власному борді. Менших беремо разом із дорослим на великий борд.',
    },
    {
        q: 'Як скасувати або змінити бронювання?',
        a: 'Напишіть нам не пізніше ніж за 24 години до сесії. Пізніше скасування — без повернення коштів.',
    },
];

// ─── ProcessingScreen ──────────────────────────────────────────────────────
export function ProcessingScreen() {
    return (
        <div className="processing-screen">
            <div className="processing-spinner" aria-hidden="true" />
            <h3 style={{ fontFamily: 'var(--font-playfair)', fontSize: '1.4rem', color: 'var(--navy)', marginBottom: '0.5rem' }}>
                {MESSAGES.processing.title}
            </h3>
            <p style={{ color: 'var(--subtle)', fontSize: '0.85rem' }}>{MESSAGES.processing.subtitle}</p>
        </div>
    );
}

// ─── ConfirmationDisplay ───────────────────────────────────────────────────
interface ConfirmationDisplayProps {
    booking: NonNullable<ReturnType<typeof useBookingStatus>['data']>['booking'];
}

function ConfirmationDisplay({ booking }: ConfirmationDisplayProps) {
    const { reset } = useBookingStore();
    const router = useRouter();
    const isIOS = useIsIOS();
    const [openFaq, setOpenFaq] = useState<number | null>(null);

    useEffect(() => {
        if (!booking) return;
        try {
            const key = `sup-fb-purchase-${booking.id}`;
            if (localStorage.getItem(key)) return;
            fbqTrack('Purchase', {
                value: booking.totalAmount,
                currency: 'UAH',
                content_name: routeLabel(booking.routeName),
                content_category: 'sup_booking',
                content_ids: [booking.id],
                contents: [{ id: booking.id, quantity: 1 }],
            });
            localStorage.setItem(key, '1');
        } catch {
            // best-effort
        }
    }, [booking]);

    if (!booking) return null;

    const dateStr = new Date(booking.date).toLocaleDateString('uk-UA', {
        weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
    });

    const total = booking.totalAmount;

    // Build individual board lines for stacked display
    const boardLines: string[] = [];
    if (booking.quantities.big > 0) boardLines.push(`${booking.quantities.big}× великий`);
    if (booking.quantities.medium > 0) boardLines.push(`${booking.quantities.medium}× середній`);
    if (booking.quantities.small > 0) boardLines.push(`${booking.quantities.small}× малий`);
    if (booking.quantities.child > 0) boardLines.push(`${booking.quantities.child}× дитина`);

    // Single comma-separated string for calendar/other uses
    const boardParts = boardLines;

    const startDate = buildStartDate(booking.date, booking.time);
    const endDate = new Date(startDate.getTime() + SESSION_DURATION_MINUTES * 60_000);
    const calendarUrl = buildCalendarUrl({
        title: '⛵ SUP Chernihiv — Прогулянка на SUP-борді',
        description: [
            'Ваше бронювання підтверджено!',
            '',
            `📍 Місце: ${MARINA.name}`,
            `📫 Адреса: ${MARINA.address}`,
            '',
            ...boardParts.map((b) => `• ${b}`),
            '',
            `💶 Сума: ${formatCurrency(total)}`,
            `Бронювання #${booking.id}`,
        ].join('\n'),
        location: `${MARINA.name}, ${MARINA.address}`,
        startISO: startDate.toISOString(),
        endISO: endDate.toISOString(),
    });
    const icsHref = useMemo(() => {
        const ics = buildICS({
            title: '⛵ SUP Chernihiv — Прогулянка на SUP-борді',
            description: [
                'Ваше бронювання підтверджено!',
                '',
                `📍 Місце: ${MARINA.name}`,
                `📫 Адреса: ${MARINA.address}`,
                '',
                ...boardParts.map((b) => `• ${b}`),
                '',
                `💶 Сума: ${formatCurrency(total)}`,
                `Бронювання #${booking.id}`,
            ].join('\n'),
            location: `${MARINA.name}, ${MARINA.address}`,
            startISO: startDate.toISOString(),
            endISO: endDate.toISOString(),
        });
        return `data:text/calendar;charset=utf-8,${encodeURIComponent(ics)}`;
    }, [booking.id, total, boardParts, startDate, endDate]);
    const mapsOpenUrl = buildMapsOpenUrl(MARINA.lat, MARINA.lng, MARINA.placeId);

    const handleNewBooking = () => { reset(); router.replace('/book/route'); };

    return (
        <div className={`${instrument.variable} ${interTight.variable} ${jetbrains.variable} bk-root`}>
            <div className="bk-shell">

                {/* ── Brand bar (no stepper) ── */}
                <header className="bk-topbar">
                    <Link href="/" className="bk-brand">
                        <span className="bk-monogram">S</span>
                        <span className="bk-wordmark">
                            <span className="bk-wordmark-name">SUP Chernihiv</span>
                            <span className="bk-wordmark-sub">оренда SUP-бордів</span>
                        </span>
                    </Link>
                </header>

                {/* ── Hero ── */}
                <section className="bk-confirm-hero">
                    <div className="bk-confirm-circle" aria-hidden="true">
                        <svg width="28" height="28" viewBox="0 0 24 24" fill="none">
                            <path d="M5 12l5 5L19 7" stroke="var(--paper)" strokeWidth="2.2"
                                  strokeLinecap="round" strokeLinejoin="round" />
                        </svg>
                    </div>
                    <p className="bk-confirm-eyebrow">Підтверджено</p>
                    <h1 className="bk-confirm-headline">
                        До зустрічі<br /><em>на воді.</em>
                    </h1>
                    <p className="bk-confirm-sub">
                        Квиток і всі деталі вже на вашій пошті. Збережіть підтвердження —
                        воно знадобиться на місці.
                    </p>
                </section>

                <div className="bk-rule" />

                {/* ── Booking card ── */}
                <div className="bk-order-card" style={{ marginBottom: 20 }}>
                    <div className="bk-order-heading" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <span>Деталі бронювання</span>
                        <span style={{ fontFamily: 'var(--mono)', fontSize: 9, letterSpacing: '0.1em', color: 'var(--ink-muted)' }}>
                            № SC-{booking.id.slice(-4).toUpperCase()}
                        </span>
                    </div>
                    <div className="bk-order-rows">
                        {booking.routeName && (
                            <div className="bk-order-row">
                                <span className="bk-order-key">Маршрут</span>
                                <span className="bk-order-val">{routeLabel(booking.routeName)}</span>
                            </div>
                        )}
                        <div className="bk-order-row">
                            <span className="bk-order-key">Дата</span>
                            <span className="bk-order-val">{dateStr}</span>
                        </div>
                        <div className="bk-order-row">
                            <span className="bk-order-key">Час</span>
                            <span className="bk-order-val">{booking.time}</span>
                        </div>

                        {/* ── Борди row — stacked lines to avoid overflow ── */}
                        <div className="bk-order-row" style={{ alignItems: 'flex-start' }}>
                            <span className="bk-order-key" style={{ paddingTop: 3 }}>Борди</span>
                            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 4 }}>
                                {boardLines.length > 0 ? boardLines.map((line) => (
                                    <span key={line} className="bk-order-val" style={{ lineHeight: 1.3 }}>
                                        {line}
                                    </span>
                                )) : (
                                    <span className="bk-order-val">—</span>
                                )}
                            </div>
                        </div>
                    </div>
                    <div className="bk-order-total-row">
                        <span className="bk-order-total-label">До сплати</span>
                        <span className="bk-order-total-amount" style={{ fontSize: 28 }}>{formatCurrency(total)}</span>
                    </div>
                </div>

                {/* ── CTAs ── */}
                <div className="bk-confirm-ctas">
                    <a href={isIOS ? icsHref : calendarUrl}
                       {...(isIOS
                           ? { download: 'sup-chernihiv.ics' }
                           : { target: '_blank', rel: 'noopener noreferrer' })}
                       className="bk-confirm-cta-dark"
                    >
                        <CalendarIcon />
                        Додати до календаря
                    </a>
                    <button className="bk-confirm-cta-outline" type="button">
                        <MailIcon />
                        Надіслати чек повторно
                    </button>
                </div>

                <div className="bk-rule" style={{ margin: '28px 0' }} />

                {/* ── Checklist ── */}
                <section>
                    <p className="bk-eyebrow bk-eyebrow--lead" style={{ marginBottom: 14 }}>
                        <span className="bk-eyebrow-rule" />
                        Що взяти з собою
                    </p>
                    <h2 className="bk-confirm-section-title">Базовий чек-лист.</h2>

                    <div className="bk-checklist-grid">
                        {CHECKLIST.map((item) => (
                            <div key={item.title} className="bk-checklist-tile">
                                <div className="bk-checklist-icon">{item.icon}</div>
                                <span className="bk-checklist-name">{item.title}</span>
                                <span className="bk-checklist-hint">{item.hint}</span>
                            </div>
                        ))}
                    </div>

                    <div className="bk-checklist-note">
                        ★ Гермомішок, жилет та весло — від нас. Нічого додаткового купувати не потрібно.
                    </div>
                </section>

                <div className="bk-rule" style={{ margin: '28px 0' }} />

                {/* ── FAQ ── */}
                <section>
                    <p className="bk-eyebrow bk-eyebrow--lead" style={{ marginBottom: 14 }}>
                        <span className="bk-eyebrow-rule" />
                        Часті запитання
                    </p>
                    <h2 className="bk-confirm-section-title">Усе про сесію.</h2>

                    <div className="bk-faq-list">
                        {FAQ.map((item, i) => (
                            <div key={item.q} className="bk-faq-item">
                                <button
                                    className="bk-faq-q"
                                    onClick={() => setOpenFaq(openFaq === i ? null : i)}
                                    type="button"
                                    aria-expanded={openFaq === i}
                                >
                                    <span className="bk-faq-qtext">{item.q}</span>
                                    <span className="bk-faq-toggle" aria-hidden="true">
                                        {openFaq === i ? '×' : '+'}
                                    </span>
                                </button>
                                {openFaq === i && (
                                    <div className="bk-faq-a">{item.a}</div>
                                )}
                            </div>
                        ))}
                    </div>
                </section>

                <div className="bk-rule" style={{ margin: '28px 0' }} />

                {/* ── Map — Точка збору ── */}
                <section style={{ marginBottom: 28 }}>
                    <p className="bk-eyebrow bk-eyebrow--lead" style={{ marginBottom: 14 }}>
                        <span className="bk-eyebrow-rule" />
                        Точка збору
                    </p>
                    <h2 className="bk-confirm-section-title">База SUP Chernihiv.</h2>

                    <div className="bk-map-card">
                        {/* Real Google Maps embed */}
                        <div className="bk-map-visual" aria-label="Карта точки збору">
                            <iframe
                                src={MARINA.mapsEmbedSrc}
                                width="100%"
                                height="100%"
                                style={{ border: 0, display: 'block' }}
                                allowFullScreen
                                loading="lazy"
                                referrerPolicy="no-referrer-when-downgrade"
                                title="SUP Chernihiv — точка збору"
                            />
                        </div>
                        <div className="bk-map-info">
                            <div>
                                <div className="bk-map-addr-label">Адреса</div>
                                <div className="bk-map-addr">SUP Chernihiv, Чернігів</div>
                            </div>
                            <a href={mapsOpenUrl} target="_blank" rel="noopener noreferrer" className="bk-map-link">
                                Відкрити →
                            </a>
                        </div>
                    </div>
                </section>

                {/* ── Footer ── */}
                <div className="bk-confirm-footer">
                    <button
                        className="bk-confirm-new-booking"
                        onClick={handleNewBooking}
                        type="button"
                    >
                        + Зробити ще одне бронювання
                    </button>
                    <p className="bk-confirm-footer-copy">
                        SUP Chernihiv · Сезон 2026
                    </p>
                </div>

            </div>
            <Script async src="https://www.instagram.com/embed.js" strategy="lazyOnload" />
        </div>
    );
}

// ─── SuccessPoller ─────────────────────────────────────────────────────────
export function SuccessPoller({ sessionId }: { sessionId: string }) {
    const { data, isLoading } = useBookingStatus(sessionId);
    if (isLoading) return <ProcessingScreen />;
    if (data?.status === 'confirmed') return <ConfirmationDisplay booking={data.booking} />;
    if (data?.status === 'failed' || data?.status === 'expired') {
        return (
            <div className="confirm-screen">
                <div className="confirm-icon" style={{ background: '#fff0ed' }}>❌</div>
                <h3 style={{ fontFamily: 'var(--font-playfair)', fontSize: '1.6rem', color: 'var(--navy)', marginBottom: '0.5rem' }}>
                    {data.status === 'expired' ? MESSAGES.errors.bookingExpired : MESSAGES.errors.paymentFailed}
                </h3>
                <p style={{ color: 'var(--subtle)', fontSize: '0.85rem' }}>
                    Спробуйте ще раз або оберіть інший слот.
                </p>
                <a href="/book/route" className="btn-primary" style={{ marginTop: '1.5rem', textDecoration: 'none' }}>
                    Спробувати знову
                </a>
            </div>
        );
    }
    return <ProcessingScreen />;
}
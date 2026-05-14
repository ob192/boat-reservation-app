'use client';

import { useRouter } from 'next/navigation';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useBookingStatus } from '@/features/booking/hooks';
import { MESSAGES, PRICES } from '@/features/booking/messages';
import { formatCurrency } from '@/shared/lib/currency';

// ─── Mocked marina location ────────────────────────────────────────────────
// Replace coordinates and name with the real departure point.
const MARINA = {
    name: 'Harbour & Wave Marina',
    address: 'Набережна Перемоги, 1, Чернігів, 14000',
    lat: 51.4982,
    lng: 31.2893,
    /** Google Maps place ID — replace with real one */
    placeId: 'ChIJN1t_tDeuEmsRUsoyG83frY4',
} as const;

// Duration of a boat session in minutes (adjust to real session length)
const SESSION_DURATION_MINUTES = 120;

// ─── Helpers ───────────────────────────────────────────────────────────────

/** Build a Google Calendar "add event" URL */
function buildCalendarUrl({
                              title,
                              description,
                              location,
                              startISO,
                              endISO,
                          }: {
    title: string;
    description: string;
    location: string;
    startISO: string;
    endISO: string;
}): string {
    // Google Calendar expects UTC datetimes in the format YYYYMMDDTHHmmssZ
    const fmt = (iso: string) => iso.replace(/[-:]/g, '').replace(/\.\d{3}/, '');
    const params = new URLSearchParams({
        action: 'TEMPLATE',
        text: title,
        details: description,
        location,
        dates: `${fmt(startISO)}/${fmt(endISO)}`,
    });
    return `https://calendar.google.com/calendar/render?${params.toString()}`;
}

/** Parse "HH:MM" time string and return Date set to that time on the given date string */
function buildStartDate(date: string, time: string): Date {
    const [h, m] = time.split(':').map(Number);
    const d = new Date(`${date}T00:00:00`);
    d.setHours(h, m, 0, 0);
    return d;
}

/** Google Static Maps thumbnail URL (no API key needed for embed link, but Static Maps needs a key)
 *  We use the Maps *embed* iframe approach which is free and key-optional for basic usage. */
function buildMapsEmbedUrl(lat: number, lng: number): string {
    // This uses the no-API-key-required embed endpoint.
    return `https://maps.google.com/maps?q=${lat},${lng}&z=16&output=embed`;
}

/** Direct "open in Google Maps" link */
function buildMapsOpenUrl(lat: number, lng: number, placeId: string): string {
    return `https://www.google.com/maps/search/?api=1&query=${lat},${lng}&query_place_id=${placeId}`;
}

// ─── ProcessingScreen ──────────────────────────────────────────────────────

export function ProcessingScreen() {
    return (
        <div className="processing-screen">
            <div className="processing-spinner" aria-hidden="true" />
            <h3
                style={{
                    fontFamily: 'var(--font-playfair)',
                    fontSize: '1.4rem',
                    color: 'var(--navy)',
                    marginBottom: '0.5rem',
                }}
            >
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

    if (!booking) return null;

    const d = new Date(booking.date);
    const dateStr = d.toLocaleDateString('uk-UA', {
        weekday: 'long',
        year: 'numeric',
        month: 'long',
        day: 'numeric',
    });

    const period = MESSAGES.time.periods[booking.time] ?? '';
    const total =
        booking.quantities.big * PRICES.big +
        booking.quantities.medium * PRICES.medium +
        booking.quantities.child * PRICES.child;

    const rows = [
        { label: MESSAGES.success.dateLabel, val: dateStr },
        { label: MESSAGES.success.departureLabel, val: `${booking.time} · ${period}` },
        ...(booking.quantities.big > 0
            ? [
                {
                    label: MESSAGES.success.bigBoatsLabel,
                    val: `${booking.quantities.big} × ${formatCurrency(PRICES.big)} = ${formatCurrency(booking.quantities.big * PRICES.big)}`,
                },
            ]
            : []),
        ...(booking.quantities.medium > 0
            ? [
                {
                    label: MESSAGES.success.mediumBoatsLabel,
                    val: `${booking.quantities.medium} × ${formatCurrency(PRICES.medium)} = ${formatCurrency(booking.quantities.medium * PRICES.medium)}`,
                },
            ]
            : []),
        ...(booking.quantities.child > 0
            ? [
                {
                    label: MESSAGES.success.childrenLabel,
                    val: `${booking.quantities.child} × ${formatCurrency(PRICES.child)} = ${formatCurrency(booking.quantities.child * PRICES.child)}`,
                },
            ]
            : []),
    ];

    // ── Google Calendar invite ──────────────────────────────────────────
    const startDate = buildStartDate(booking.date, booking.time);
    const endDate = new Date(startDate.getTime() + SESSION_DURATION_MINUTES * 60_000);

    const calendarUrl = buildCalendarUrl({
        title: `⛵ Harbour & Wave — Прогулянка на човні`,
        description: [
            `Ваше бронювання підтверджено!`,
            ``,
            `📍 Місце відправлення: ${MARINA.name}`,
            `📫 Адреса: ${MARINA.address}`,
            ``,
            ...(booking.quantities.big > 0
                ? [`🚢 Великих човнів: ${booking.quantities.big}`]
                : []),
            ...(booking.quantities.medium > 0
                ? [`⛵ Середніх човнів: ${booking.quantities.medium}`]
                : []),
            ...(booking.quantities.child > 0
                ? [`👶 Дітей: ${booking.quantities.child}`]
                : []),
            ``,
            `💶 Загальна сума: ${formatCurrency(total)}`,
            ``,
            `Бронювання #${booking.id}`,
        ].join('\n'),
        location: `${MARINA.name}, ${MARINA.address}`,
        startISO: startDate.toISOString(),
        endISO: endDate.toISOString(),
    });

    const mapsEmbedUrl = buildMapsEmbedUrl(MARINA.lat, MARINA.lng);
    const mapsOpenUrl = buildMapsOpenUrl(MARINA.lat, MARINA.lng, MARINA.placeId);

    const handleNewBooking = () => {
        reset();
        router.replace('/book/date');
    };

    return (
        <div className="confirm-screen">
            <div className="confirm-icon">🎉</div>
            <h3>{MESSAGES.success.title}</h3>
            <p>{MESSAGES.success.message}</p>

            {/* Booking details */}
            <div className="confirm-details">
                {rows.map((r) => (
                    <div key={r.label} className="confirm-row">
                        <span className="cr-label">{r.label}</span>
                        <span className="cr-val">{r.val}</span>
                    </div>
                ))}
                <div className="confirm-row total">
                    <span className="cr-label">{MESSAGES.success.totalLabel}</span>
                    <span className="cr-val">{formatCurrency(total)}</span>
                </div>
            </div>

            {/* ── Action buttons ───────────────────────────────────────────── */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.65rem', marginBottom: '1.5rem' }}>
                {/* Add to Google Calendar */}
                <a
                    href={calendarUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="btn-primary"
                    style={{
                        textDecoration: 'none',
                        background: 'var(--navy)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        gap: '0.5rem',
                    }}
                >
                    {/* Google Calendar icon */}
                    <svg width="17" height="17" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                        <rect x="3" y="4" width="18" height="18" rx="2" stroke="currentColor" strokeWidth="2" fill="none"/>
                        <path d="M3 9h18" stroke="currentColor" strokeWidth="2"/>
                        <path d="M8 2v4M16 2v4" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                        <rect x="7" y="13" width="3" height="3" rx="0.5" fill="currentColor"/>
                        <rect x="14" y="13" width="3" height="3" rx="0.5" fill="currentColor"/>
                    </svg>
                    Додати до Google Календаря
                </a>

                {/* New booking */}
                <button
                    className="btn-ghost"
                    onClick={handleNewBooking}
                    type="button"
                    style={{ width: '100%' }}
                >
                    {MESSAGES.buttons.newBooking}
                </button>
            </div>

            {/* ── Departure point map ──────────────────────────────────────── */}
            <div style={{ textAlign: 'left' }}>
                <div
                    style={{
                        fontSize: '0.6rem',
                        letterSpacing: '0.14em',
                        textTransform: 'uppercase',
                        color: 'var(--subtle)',
                        fontWeight: 600,
                        marginBottom: '0.55rem',
                    }}
                >
                    📍 Місце відправлення
                </div>

                {/* Map iframe */}
                <div
                    style={{
                        borderRadius: 12,
                        overflow: 'hidden',
                        border: '1.5px solid var(--mist)',
                        marginBottom: '0.55rem',
                        height: 200,
                        position: 'relative',
                    }}
                >
                    <iframe
                        title="Місце відправлення"
                        src={mapsEmbedUrl}
                        width="100%"
                        height="200"
                        style={{ border: 0, display: 'block' }}
                        allowFullScreen={false}
                        loading="lazy"
                        referrerPolicy="no-referrer-when-downgrade"
                    />
                </div>

                {/* Marina info + "Open in Maps" link */}
                <div
                    style={{
                        background: 'var(--sand)',
                        borderRadius: 10,
                        padding: '0.75rem 0.9rem',
                        display: 'flex',
                        alignItems: 'flex-start',
                        justifyContent: 'space-between',
                        gap: '0.75rem',
                    }}
                >
                    <div style={{ minWidth: 0 }}>
                        <div style={{ fontWeight: 600, fontSize: '0.85rem', color: 'var(--navy)', marginBottom: '0.15rem' }}>
                            {MARINA.name}
                        </div>
                        <div style={{ fontSize: '0.72rem', color: 'var(--subtle)', lineHeight: 1.4 }}>
                            {MARINA.address}
                        </div>
                    </div>
                    <a
                        href={mapsOpenUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        style={{
                            flexShrink: 0,
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '0.3rem',
                            padding: '0.45rem 0.75rem',
                            background: 'var(--teal)',
                            color: 'white',
                            borderRadius: 8,
                            fontSize: '0.72rem',
                            fontWeight: 500,
                            textDecoration: 'none',
                            whiteSpace: 'nowrap',
                        }}
                    >
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                            <path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7z" fill="currentColor"/>
                            <circle cx="12" cy="9" r="2.5" fill="white"/>
                        </svg>
                        Відкрити
                    </a>
                </div>
            </div>
        </div>
    );
}

// ─── SuccessPoller ─────────────────────────────────────────────────────────

export function SuccessPoller({ sessionId }: { sessionId: string }) {
    const { data, isLoading } = useBookingStatus(sessionId);

    if (isLoading) return <ProcessingScreen />;

    if (data?.status === 'confirmed') {
        return <ConfirmationDisplay booking={data.booking} />;
    }

    if (data?.status === 'failed' || data?.status === 'expired') {
        return (
            <div className="confirm-screen">
                <div className="confirm-icon" style={{ background: '#fff0ed' }}>
                    ❌
                </div>
                <h3
                    style={{
                        fontFamily: 'var(--font-playfair)',
                        fontSize: '1.6rem',
                        color: 'var(--navy)',
                        marginBottom: '0.5rem',
                    }}
                >
                    {data.status === 'expired' ? MESSAGES.errors.bookingExpired : MESSAGES.errors.paymentFailed}
                </h3>
                <p style={{ color: 'var(--subtle)', fontSize: '0.85rem' }}>
                    Спробуйте ще раз або оберіть інший слот.
                </p>
                <a href="/book/date" className="btn-primary" style={{ marginTop: '1.5rem', textDecoration: 'none' }}>
                    Спробувати знову
                </a>
            </div>
        );
    }

    return <ProcessingScreen />;
}
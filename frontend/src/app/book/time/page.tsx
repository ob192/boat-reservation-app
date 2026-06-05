'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Instrument_Serif, Inter_Tight, JetBrains_Mono } from 'next/font/google';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useSlots } from '@/features/booking/hooks';
import { useStepGuard } from '@/features/booking/hooks/useStepGuard';
import { UserMenu } from '@/features/auth/components/UserMenu';
import { MESSAGES } from '@/features/booking/messages';

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

const STEPS = [
    { n: '01', label: 'Маршрут' },
    { n: '02', label: 'Дата' },
    { n: '03', label: 'Час' },
    { n: '04', label: 'Човни' },
    { n: '05', label: 'Деталі' },
    { n: '06', label: 'Готово' },
];

function ChevronLeft() {
    return (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M15 5l-7 7 7 7" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
    );
}

function ArrowRight() {
    return (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M5 12h14M13 6l6 6-6 6" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
    );
}

function CheckIcon() {
    return (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M5 12l5 5L19 7" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
    );
}

const TIME_PERIOD: Record<string, string> = {
    '08:00': 'Ранок',
    '10:00': 'Ранок',
    '12:00': 'Опівдні',
    '14:00': 'Опівдні',
    '16:00': 'Вечір',
    '18:00': 'Вечір',
    '20:00': '↘ Захід',
};

function getPeriod(time: string): string {
    return TIME_PERIOD[time] ?? time.slice(0, 5);
}

function getBarColor(available: number, total: number): string {
    if (total === 0) return 'var(--mint)';
    const pct = available / total;
    if (pct <= 0) return 'var(--rust)';
    if (pct <= 0.3) return 'var(--ochre)';
    return 'var(--mint)';
}

function isSlotInPast(date: string, time: string): boolean {
    const [h, m] = time.split(':').map(Number);
    const slotDate = new Date(`${date}T00:00:00`);
    slotDate.setHours(h, m, 0, 0);
    return slotDate <= new Date(Date.now() + 30 * 60 * 1000);
}

export default function TimePage() {
    useStepGuard('time');

    const { selectedDate, selectedRoute, selectedTime, setTime } = useBookingStore();
    const { data: slotsData, isLoading, isError } = useSlots(selectedDate, selectedRoute);
    const router = useRouter();

    const dateLabel = selectedDate
        ? new Date(selectedDate + 'T00:00:00').toLocaleDateString('uk-UA', {
            day: 'numeric',
            month: 'long',
        })
        : '';

    const todayKey = (() => {
        const t = new Date();
        return `${t.getFullYear()}-${String(t.getMonth() + 1).padStart(2, '0')}-${String(t.getDate()).padStart(2, '0')}`;
    })();
    const isToday = selectedDate === todayKey;

    const rawSlots = slotsData?.slots ?? [];
    const dateBlocked = slotsData?.dateBlocked ?? false;
    const fullyBlocked = slotsData?.fullyBlocked ?? false;

    const visibleSlots = isToday
        ? rawSlots.filter((s) => !isSlotInPast(selectedDate!, s.time))
        : rawSlots;

    const ctaLabel = selectedTime ? `Далі — ${selectedTime}` : 'Далі';

    return (
        <div className={`${instrument.variable} ${interTight.variable} ${jetbrains.variable} bk-root`}>
            <div className="bk-shell">

                {/* ── Brand bar ── */}
                <header className="bk-topbar">
                    <Link href="/" className="bk-brand">
                        <span className="bk-monogram">S</span>
                        <span className="bk-wordmark">
                            <span className="bk-wordmark-name">SUP Chernihiv</span>
                            <span className="bk-wordmark-sub">оренда SUP-бордів</span>
                        </span>
                    </Link>
                    <UserMenu />
                </header>

                {/* ── Stepper ── */}
                <nav className="bk-stepper" aria-label="Прогрес бронювання">
                    {STEPS.map((s, i) => (
                        <div
                            key={s.n}
                            className={`bk-seg ${i === 0 ? 'done' : ''} ${i === 1 ? 'active' : ''}`}
                            aria-current={i === 1 ? 'step' : undefined}
                        >
                            <span className="bk-seg-bar" />
                            <span className="bk-seg-label">
                                <span className="bk-seg-num">{s.n}</span> {s.label}
                            </span>
                        </div>
                    ))}
                </nav>

                {/* ── Intro ── */}
                <section className="bk-intro">
                    <p className="bk-eyebrow bk-eyebrow--lead">
                        <span className="bk-eyebrow-rule" />
                        Крок 02 — Час
                    </p>
                    <h1 className="bk-headline">
                        Коли <em>вирушаємо?</em>
                    </h1>
                    <p className="bk-time-sub">
                        Доступні слоти на {dateLabel}. Кожна сесія — 2 години.
                    </p>
                </section>

                {/* ── Slot list ── */}
                {isLoading && (
                    <div className="bk-slots">
                        {[0, 1, 2, 3, 4].map((i) => (
                            <div key={i} className="bk-slot-skeleton" aria-hidden="true" />
                        ))}
                    </div>
                )}

                {isError && (
                    <div className="bk-banner bk-banner--error" role="alert">
                        Не вдалося завантажити слоти. Спробуйте оновити сторінку.
                    </div>
                )}

                {fullyBlocked && !isLoading && (
                    <div className="bk-banner bk-banner--warn" role="alert">
                        <span style={{ fontFamily: 'var(--serif)', fontSize: '15px', color: 'var(--ink)' }}>
                            Тимчасово недоступно
                        </span>
                        <span>Бронювання на цей день призупинено. Оберіть інший день.</span>
                    </div>
                )}

                {!isLoading && !isError && !fullyBlocked && visibleSlots.length === 0 && (
                    <p className="bk-slots-empty">
                        {isToday
                            ? 'На сьогодні більше немає доступних слотів.'
                            : MESSAGES.time.noSlots}
                    </p>
                )}

                {!isLoading && !isError && !fullyBlocked && visibleSlots.length > 0 && (
                    <div className="bk-slots" role="group" aria-label="Доступні часові слоти">
                        {visibleSlots.map((s) => {
                            const isCancelled = s.cancelled;
                            const isBlocked = !isCancelled && (dateBlocked || s.blocked);
                            const isFull = !isCancelled && !isBlocked && s.availableBig <= 0 && s.availableMedium <= 0 && s.availableSmall <= 0;
                            const isUnavailable = isCancelled || isBlocked || isFull;
                            const isSel = selectedTime === s.time;
                            const pct = s.totalBig > 0
                                ? Math.round(((s.totalBig - s.availableBig) / s.totalBig) * 100)
                                : 0;
                            const barColor = isSel
                                ? 'oklch(1 0 0 / 0.5)'
                                : getBarColor(s.availableBig, s.totalBig);

                            return (
                                <div
                                    key={s.time}
                                    className={[
                                        'bk-slot',
                                        isSel ? 'selected' : '',
                                        isUnavailable ? 'unavailable' : '',
                                        isCancelled ? 'cancelled' : '',
                                        s.time === '18:00' && !isUnavailable ? 'popular' : '',
                                    ].filter(Boolean).join(' ')}
                                    onClick={!isUnavailable ? () => setTime(s.time) : undefined}
                                    role="radio"
                                    aria-checked={isSel}
                                    aria-disabled={isUnavailable}
                                    tabIndex={isUnavailable ? -1 : 0}
                                    onKeyDown={(e) => {
                                        if ((e.key === 'Enter' || e.key === ' ') && !isUnavailable) setTime(s.time);
                                    }}
                                >
                                    {/* Left: time + period */}
                                    <div className="bk-slot-left">
                                        <span className="bk-slot-time">{s.time}</span>
                                        <span className="bk-slot-period">{getPeriod(s.time)}</span>
                                    </div>

                                    {/* Middle: availability + bar */}
                                    <div className="bk-slot-mid">
                                        {isCancelled ? (
                                            <span className="bk-slot-avail bk-slot-avail--cancelled">
                                                {MESSAGES.time.cancelledTag}
                                            </span>
                                        ) : isBlocked ? (
                                            <span className="bk-slot-avail">{MESSAGES.time.blockedTag}</span>
                                        ) : isFull ? (
                                            <span className="bk-slot-avail">{MESSAGES.time.fullTag}</span>
                                        ) : (
                                            <span className="bk-slot-avail">
                                                {s.availableBig} великих · {s.availableMedium} середніх · {s.availableSmall} малих
                                            </span>
                                        )}
                                        {!isBlocked && !isCancelled && (
                                            <div className="bk-slot-bar">
                                                <div
                                                    className="bk-slot-bar-fill"
                                                    style={{ width: `${pct}%`, background: barColor }}
                                                />
                                            </div>
                                        )}
                                    </div>

                                    {/* Right: circle button */}
                                    <div className="bk-slot-circ" aria-hidden="true">
                                        {isSel ? <CheckIcon /> : <ArrowRight />}
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                )}

                <div className="bk-dock-spacer" />
            </div>

            {/* ── Sticky dock ── */}
            <div className="bk-dock">
                <div className="bk-dock-inner">
                    <button
                        className="bk-back"
                        onClick={() => router.push('/book/date')}
                        type="button"
                        aria-label="Назад"
                    >
                        <ChevronLeft />
                    </button>
                    <button
                        className="bk-cta"
                        disabled={!selectedTime}
                        onClick={() => router.push('/book/boats')}
                        type="button"
                    >
                        {ctaLabel}
                    </button>
                </div>
            </div>
        </div>
    );
}
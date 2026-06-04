'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Instrument_Serif, Inter_Tight, JetBrains_Mono } from 'next/font/google';
import { Calendar } from '@/features/booking/components/client/Calendar';
import { WeatherPreview } from '@/features/booking/components/client/WeatherPreview';
import { UserMenu } from '@/features/auth/components/UserMenu';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useStepGuard } from '@/features/booking/hooks/useStepGuard';

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

export default function DatePage() {
    useStepGuard('date');
    const { selectedDate } = useBookingStore();
    const router = useRouter();

    const dateLabel = selectedDate
        ? new Date(selectedDate + 'T00:00:00').toLocaleDateString('uk-UA', {
            day: 'numeric',
            month: 'long',
        })
        : null;

    return (
        <div className={`${instrument.variable} ${interTight.variable} ${jetbrains.variable} bk-root`}>
            <div className="bk-shell">
                {/* ── Brand bar ──────────────────────────────────────────── */}
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

                {/* ── Progress stepper ───────────────────────────────────── */}
                <nav className="bk-stepper" aria-label="Прогрес бронювання">
                    {STEPS.map((s, i) => (
                        <div
                            key={s.n}
                            className={`bk-seg ${i < 1 ? 'done' : ''} ${i === 1 ? 'active' : ''}`}
                            aria-current={i === 1 ? 'step' : undefined}
                        >
                            <span className="bk-seg-bar" />
                            <span className="bk-seg-label">
                                <span className="bk-seg-num">{s.n}</span> {s.label}
                            </span>
                        </div>
                    ))}
                </nav>

                {/* ── Intro ──────────────────────────────────────────────── */}
                <section className="bk-intro">
                    <p className="bk-eyebrow bk-eyebrow--lead">
                        <span className="bk-eyebrow-rule" />
                        Крок 01 — Дата
                    </p>
                    <h1 className="bk-headline">
                        Оберіть день
                        <br />
                        <em>на воді.</em>
                    </h1>
                </section>

                {/* ── Calendar ───────────────────────────────────────────── */}
                <Calendar />

                {/*
                  WeatherPreview handles all three states internally:
                  - isLoading  → skeleton
                  - no data for date (out of 16-day window) → null (renders nothing)
                  - data found → weather card
                  So we only render it when a date is actually selected.
                */}
                {selectedDate && dateLabel && (
                    <>
                        <div className="bk-rule" />
                        <WeatherPreview date={selectedDate} dateLabel={dateLabel} />
                    </>
                )}

                <div className="bk-dock-spacer" />
            </div>

            {/* ── Sticky dock ────────────────────────────────────────────── */}
            <div className="bk-dock">
                <div className="bk-dock-inner">
                    <button
                        className="bk-back"
                        onClick={() => router.push('/')}
                        type="button"
                        aria-label="Назад"
                    >
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                            <path d="M15 5l-7 7 7 7" stroke="currentColor" strokeWidth="1.8"
                                  strokeLinecap="round" strokeLinejoin="round" />
                        </svg>
                    </button>
                    <button
                        className="bk-cta"
                        disabled={!selectedDate}
                        onClick={() => router.push('/book/time')}
                        type="button"
                    >
                        {selectedDate ? `Далі — ${dateLabel}` : 'Далі →'}
                    </button>
                </div>
            </div>
        </div>
    );
}
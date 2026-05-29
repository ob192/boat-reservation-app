'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Instrument_Serif, Inter_Tight, JetBrains_Mono } from 'next/font/google';
import { Calendar } from '@/features/booking/components/client/Calendar';
import { UserMenu } from '@/features/auth/components/UserMenu';
import { useBookingStore } from '@/features/booking/store/bookingStore';

// ─── Fonts (scoped via the .bk-root variable classes) ──────────────────────
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
    { n: '01', label: 'Дата' },
    { n: '02', label: 'Час' },
    { n: '03', label: 'Човни' },
    { n: '04', label: 'Деталі' },
    { n: '05', label: 'Готово' },
];

function SunIcon() {
    return (
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <circle cx="12" cy="12" r="4.2" stroke="currentColor" strokeWidth="1.6" />
            <path
                d="M12 2.5v2.2M12 19.3v2.2M4.6 4.6l1.6 1.6M17.8 17.8l1.6 1.6M2.5 12h2.2M19.3 12h2.2M4.6 19.4l1.6-1.6M17.8 6.2l1.6-1.6"
                stroke="currentColor"
                strokeWidth="1.6"
                strokeLinecap="round"
            />
        </svg>
    );
}

/**
 * Weather preview. Values are a static editorial placeholder — wire these to a
 * forecast API (e.g. Open-Meteo for Chernihiv, 51.49 N / 31.29 E) keyed on the
 * selected date when ready.
 */
function WeatherPreview({ dateLabel }: { dateLabel: string }) {
    return (
        <section className="bk-weather" aria-label="Прогноз погоди">
            <div className="bk-weather-head">
                <span className="bk-eyebrow">Прогноз на день</span>
                <span className="bk-weather-date">{dateLabel}</span>
            </div>
            <div className="bk-weather-grid">
                <div className="bk-weather-col">
                    <span className="bk-weather-label">Умови</span>
                    <span className="bk-weather-val">
                        <span className="bk-weather-ico"><SunIcon /></span>
                        Сонячно
                    </span>
                </div>
                <div className="bk-weather-col">
                    <span className="bk-weather-label">Повітря</span>
                    <span className="bk-weather-val">22°C</span>
                    <span className="bk-weather-sub">комфортно</span>
                </div>
                <div className="bk-weather-col">
                    <span className="bk-weather-label">Вода</span>
                    <span className="bk-weather-val">23°C</span>
                    <span className="bk-weather-sub">вітер 3 м/с</span>
                </div>
            </div>
        </section>
    );
}

export default function DatePage() {
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
                {/* ── Brand bar + account pill ──────────────────────────── */}
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

                {/* ── Progress stepper ──────────────────────────────────── */}
                <nav className="bk-stepper" aria-label="Прогрес бронювання">
                    {STEPS.map((s, i) => (
                        <div key={s.n} className={`bk-seg ${i === 0 ? 'active' : ''}`} aria-current={i === 0 ? 'step' : undefined}>
                            <span className="bk-seg-bar" />
                            <span className="bk-seg-label">
                                <span className="bk-seg-num">{s.n}</span> {s.label}
                            </span>
                        </div>
                    ))}
                </nav>

                {/* ── Intro ─────────────────────────────────────────────── */}
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

                {/* ── Calendar ──────────────────────────────────────────── */}
                <Calendar />

                <div className="bk-rule" />

                {/* ── Weather preview / empty state ─────────────────────── */}
                {dateLabel ? (
                    <WeatherPreview dateLabel={dateLabel} />
                ) : (
                    <div className="bk-weather-empty">
                        <div className="bk-spinner" aria-hidden="true" />
                        <p className="bk-weather-empty-text">
                            Оберіть день у календарі, щоб побачити прогноз погоди на воді.
                        </p>
                    </div>
                )}

                <div className="bk-dock-spacer" />
            </div>

            {/* ── Sticky dock ───────────────────────────────────────────── */}
            <div className="bk-dock">
                <div className="bk-dock-inner">
                    <button className="bk-back" onClick={() => router.push('/')} type="button" aria-label="Назад">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                            <path d="M15 5l-7 7 7 7" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
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
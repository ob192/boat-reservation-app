'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Instrument_Serif, Inter_Tight, JetBrains_Mono } from 'next/font/google';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { UserMenu } from '@/features/auth/components/UserMenu';
import { ROUTES } from '@/features/booking/routes';

const instrument = Instrument_Serif({ subsets: ['latin'], weight: ['400'], style: ['normal', 'italic'], variable: '--bk-serif', display: 'swap' });
const interTight = Inter_Tight({ subsets: ['latin', 'cyrillic'], weight: ['400', '500', '600'], variable: '--bk-sans', display: 'swap' });
const jetbrains = JetBrains_Mono({ subsets: ['latin'], weight: ['400', '500'], variable: '--bk-mono', display: 'swap' });

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

export default function RoutePage() {
    
    const { selectedRoute, setRoute } = useBookingStore();
    const router = useRouter();

    return (
        <div className={`${instrument.variable} ${interTight.variable} ${jetbrains.variable} bk-root`}>
            <div className="bk-shell">
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

                <nav className="bk-stepper" aria-label="Прогрес бронювання">
                    {STEPS.map((s, i) => (
                        <div
                            key={s.n}
                            className={`bk-seg ${i === 0 ? 'active' : ''}`}
                            aria-current={i === 0 ? 'step' : undefined}
                        >
                            <span className="bk-seg-bar" />
                            <span className="bk-seg-label">
                                <span className="bk-seg-num">{s.n}</span> {s.label}
                            </span>
                        </div>
                    ))}
                </nav>

                <section className="bk-intro">
                    <p className="bk-eyebrow bk-eyebrow--lead">
                        <span className="bk-eyebrow-rule" />
                        Крок 01 — Маршрут
                    </p>
                    <h1 className="bk-headline">
                        Який <em>маршрут?</em>
                    </h1>
                    <p className="bk-time-sub">Оберіть напрямок сплаву — від нього залежать ціни та доступність</p>
                </section>

                <div className="bk-routes" role="group" aria-label="Доступні маршрути">
                    {ROUTES.map((r) => {
                        const isSel = selectedRoute === r.id;
                        return (
                            <div
                                key={r.id}
                                className={`bk-route-card ${isSel ? 'selected' : ''}`}
                                onClick={() => setRoute(r.id)}
                                role="radio"
                                aria-checked={isSel}
                                tabIndex={0}
                                onKeyDown={(e) => {
                                    if (e.key === 'Enter' || e.key === ' ') {
                                        e.preventDefault();
                                        setRoute(r.id);
                                    }
                                }}
                            >
                                <div className="bk-route-body">
                                    <div className="bk-route-name-row">
                                        <span className="bk-route-name">{r.label}</span>
                                        <span className="bk-route-badge">{r.sub}</span>
                                    </div>
                                    <p className="bk-route-desc">{r.desc}</p>
                                </div>
                                <div className="bk-route-circ" aria-hidden="true">
                                    {isSel ? <CheckIcon /> : <ArrowRight />}
                                </div>
                            </div>
                        );
                    })}
                </div>

                <div className="bk-dock-spacer" />
            </div>

            <div className="bk-dock">
                <div className="bk-dock-inner">
                    <button className="bk-back" onClick={() => router.push('/')} type="button" aria-label="Назад">
                        <ChevronLeft />
                    </button>
                    <button
                        className="bk-cta"
                        disabled={!selectedRoute}
                        onClick={() => router.push('/book/date')}
                        type="button"
                    >
                        Продовжити
                    </button>
                </div>
            </div>
        </div>
    );
}
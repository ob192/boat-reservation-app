'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Instrument_Serif, Inter_Tight, JetBrains_Mono } from 'next/font/google';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useSlots } from '@/features/booking/hooks';
import { useStepGuard } from '@/features/booking/hooks/useStepGuard';
import { UserMenu } from '@/features/auth/components/UserMenu';
import { MESSAGES, PRICES, MAX_BIG, MAX_MEDIUM } from '@/features/booking/messages';
import { formatCurrency } from '@/shared/lib/currency';

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

function ChevronLeft() {
    return (
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M15 5l-7 7 7 7" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
    );
}

function SparkleIcon() {
    return (
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M12 2l2.4 7.4H22l-6.2 4.5 2.4 7.4L12 17l-6.2 4.3 2.4-7.4L2 9.4h7.6z"
                  stroke="currentColor" strokeWidth="1.6" strokeLinejoin="round" />
        </svg>
    );
}

interface QtyCounterProps {
    value: number;
    onDecrement: () => void;
    onIncrement: () => void;
    disableDecrement: boolean;
    disableIncrement: boolean;
}

function QtyCounter({ value, onDecrement, onIncrement, disableDecrement, disableIncrement }: QtyCounterProps) {
    return (
        <div className="bk-qty">
            <button
                className="bk-qty-btn"
                onClick={onDecrement}
                disabled={disableDecrement}
                type="button"
                aria-label={MESSAGES.boats.decrease}
            >
                −
            </button>
            <span className="bk-qty-val">{value}</span>
            <button
                className="bk-qty-btn"
                onClick={onIncrement}
                disabled={disableIncrement}
                type="button"
                aria-label={MESSAGES.boats.increase}
            >
                +
            </button>
        </div>
    );
}

export default function BoatsPage() {
    useStepGuard('boats');

    const { selectedDate, selectedTime, quantities, setQuantity } = useBookingStore();
    const { data: slotsData } = useSlots(selectedDate);
    const router = useRouter();

    const slotInfo = slotsData?.slots.find((s) => s.time === selectedTime);
    const availableBig = slotInfo?.availableBig ?? MAX_BIG;
    const availableMedium = slotInfo?.availableMedium ?? MAX_MEDIUM;

    const change = (type: 'big' | 'medium' | 'child', delta: number) => {
        if (type === 'big') {
            const next = Math.max(0, Math.min(availableBig, quantities.big + delta));
            setQuantity('big', next);
            if (next === 0) setQuantity('child', 0);
            else setQuantity('child', Math.min(quantities.child, next));
        } else if (type === 'medium') {
            setQuantity('medium', Math.max(0, Math.min(availableMedium, quantities.medium + delta)));
        } else {
            setQuantity('child', Math.max(0, Math.min(quantities.big, quantities.child + delta)));
        }
    };

    const total =
        quantities.big * PRICES.big +
        quantities.medium * PRICES.medium +
        quantities.child * PRICES.child;

    const totalPax = quantities.big * 2 + quantities.medium + quantities.child;
    const canProceed = quantities.big + quantities.medium > 0;

    const summaryParts: string[] = [];
    if (quantities.big > 0) summaryParts.push(`${quantities.big}× великий`);
    if (quantities.medium > 0) summaryParts.push(`${quantities.medium}× середній`);
    if (quantities.child > 0) summaryParts.push(`${quantities.child}× дитина`);

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
                            className={`bk-seg ${i < 2 ? 'done' : ''} ${i === 2 ? 'active' : ''}`}
                            aria-current={i === 2 ? 'step' : undefined}
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
                        Крок 03 — Човни
                    </p>
                    <h1 className="bk-headline">
                        Скільки <em>бордів?</em>
                    </h1>
                    <p className="bk-time-sub">
                        {availableBig} великих та {availableMedium} середніх доступно
                    </p>
                </section>

                {/* ── Board cards ── */}
                <div className="bk-boards">

                    {/* Big board card */}
                    <div className={`bk-board-card ${quantities.big > availableBig ? 'bk-board-card--warn' : ''}`}>
                        <div className="bk-board-info">
                            <div className="bk-board-name-row">
                                <span className="bk-board-name">Великий борд</span>
                                <span className="bk-board-badge">1 особа + дитина</span>
                            </div>
                            <p className="bk-board-desc">
                                Просторий та стабільний. 1 доросла особа + дитина до 45 кг.
                            </p>
                        </div>
                        <div className="bk-board-footer">
                            <div className="bk-board-price-col">
                                <span className="bk-board-price">{formatCurrency(PRICES.big)}</span>
                                <span className="bk-board-per">за борд</span>
                            </div>
                            <QtyCounter
                                value={quantities.big}
                                onDecrement={() => change('big', -1)}
                                onIncrement={() => change('big', 1)}
                                disableDecrement={quantities.big === 0}
                                disableIncrement={quantities.big >= availableBig}
                            />
                        </div>
                    </div>

                    {/* Child add-on — revealed when big > 0 */}
                    {quantities.big > 0 && (
                        <div className="bk-child-addon">
                            <div className="bk-child-addon-left">
                                <div className="bk-child-icon">
                                    <SparkleIcon />
                                </div>
                                <div className="bk-child-info">
                                    <div className="bk-child-name-row">
                                        <span className="bk-child-name">Додати дитину</span>
                                        <span className="bk-child-badge">−50%</span>
                                    </div>
                                    <span className="bk-child-hint">до 40 кг · лише на великих бордах</span>
                                </div>
                            </div>
                            <QtyCounter
                                value={quantities.child}
                                onDecrement={() => change('child', -1)}
                                onIncrement={() => change('child', 1)}
                                disableDecrement={quantities.child === 0}
                                disableIncrement={quantities.child >= quantities.big}
                            />
                        </div>
                    )}

                    {/* Medium board card */}
                    <div className={`bk-board-card ${quantities.medium > availableMedium ? 'bk-board-card--warn' : ''}`}>
                        <div className="bk-board-info">
                            <div className="bk-board-name-row">
                                <span className="bk-board-name">Середній борд</span>
                                <span className="bk-board-badge">1 особа</span>
                            </div>
                            <p className="bk-board-desc">
                                Маневрений та легкий. Для досвідчених і тих, хто любить швидкість.
                            </p>
                        </div>
                        <div className="bk-board-footer">
                            <div className="bk-board-price-col">
                                <span className="bk-board-price">{formatCurrency(PRICES.medium)}</span>
                                <span className="bk-board-per">за борд</span>
                            </div>
                            <QtyCounter
                                value={quantities.medium}
                                onDecrement={() => change('medium', -1)}
                                onIncrement={() => change('medium', 1)}
                                disableDecrement={quantities.medium === 0}
                                disableIncrement={quantities.medium >= availableMedium}
                            />
                        </div>
                    </div>
                </div>

                {/* ── Live total strip ── */}
                <div className="bk-total-strip">
                    <div className="bk-total-left">
                        <span className="bk-total-label">Обрано</span>
                        <span className="bk-total-breakdown">
                            {summaryParts.length
                                ? summaryParts.join(', ')
                                : <em>нічого не обрано</em>}
                        </span>
                        {totalPax > 0 && (
                            <span className="bk-total-pax">{totalPax} учасн.</span>
                        )}
                    </div>
                    <div className="bk-total-right">
                        <span className="bk-total-due-label">До сплати</span>
                        <span className="bk-total-amount">{formatCurrency(total)}</span>
                    </div>
                </div>

                <div className="bk-dock-spacer" />
            </div>

            {/* ── Sticky dock ── */}
            <div className="bk-dock">
                <div className="bk-dock-inner">
                    <button
                        className="bk-back"
                        onClick={() => router.push('/book/time')}
                        type="button"
                        aria-label="Назад"
                    >
                        <ChevronLeft />
                    </button>
                    <button
                        className="bk-cta"
                        disabled={!canProceed}
                        onClick={() => router.push('/book/details')}
                        type="button"
                    >
                        Продовжити
                    </button>
                </div>
            </div>
        </div>
    );
}
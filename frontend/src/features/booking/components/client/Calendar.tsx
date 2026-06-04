'use client';

import { useState } from 'react';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useAvailability, useBookingSystemStatus } from '@/features/booking/hooks';
import { MESSAGES, MAX_BIG, PRICES } from '@/features/booking/messages';
import { formatCurrency } from '@/shared/lib/currency';

function dateKey(d: Date): string {
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

function monthKey(y: number, m: number): string {
    return `${y}-${String(m + 1).padStart(2, '0')}`;
}

const LEGEND = [
    { cls: 'available', color: 'var(--mint)', label: MESSAGES.calendar.available },
    { cls: 'limited', color: 'var(--ochre)', label: MESSAGES.calendar.limited },
    { cls: 'booked', color: 'var(--rust)', label: MESSAGES.calendar.booked },
] as const;

function Chevron({ dir }: { dir: 'left' | 'right' }) {
    return (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path
                d={dir === 'left' ? 'M15 5l-7 7 7 7' : 'M9 5l7 7-7 7'}
                stroke="currentColor"
                strokeWidth="1.8"
                strokeLinecap="round"
                strokeLinejoin="round"
            />
        </svg>
    );
}

export function Calendar() {
    const { selectedDate, selectedRoute, setDate } = useBookingStore();
    const [calYear, setCalYear] = useState(() => new Date().getFullYear());
    const [calMonth, setCalMonth] = useState(() => new Date().getMonth());

    const { data: status, isLoading: statusLoading } = useBookingSystemStatus();
    const bookingsEnabled = status?.bookingsEnabled ?? true;

    const month = monthKey(calYear, calMonth);
    const {
        data: availability,
        isLoading: availLoading,
        isError,
    } = useAvailability(month, selectedRoute, !statusLoading && bookingsEnabled);

    const availabilityMap = new Map<string, { slots: number; blocked: boolean }>();
    availability?.days.forEach((d) =>
        availabilityMap.set(d.date, {
            slots: d.availableSlots,
            blocked: d.blocked || d.fullyBlocked,
        }),
    );

    const isLoading = statusLoading || availLoading;

    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const firstDayRaw = new Date(calYear, calMonth, 1).getDay();
    const firstDay = firstDayRaw === 0 ? 6 : firstDayRaw - 1; // Monday-first
    const daysInMonth = new Date(calYear, calMonth + 1, 0).getDate();

    const priceLabel = formatCurrency(PRICES.big); // ₴400

    const prevMonth = () => {
        if (calMonth === 0) { setCalYear((y) => y - 1); setCalMonth(11); }
        else setCalMonth((m) => m - 1);
    };
    const nextMonth = () => {
        if (calMonth === 11) { setCalYear((y) => y + 1); setCalMonth(0); }
        else setCalMonth((m) => m + 1);
    };

    return (
        <div className="bk-cal">
            <div className="bk-cal-nav">
                <button className="bk-cal-chev" onClick={prevMonth} aria-label={MESSAGES.calendar.prevMonth} type="button">
                    <Chevron dir="left" />
                </button>
                <div className="bk-cal-monthwrap">
                    <span className="bk-cal-year">{calYear}</span>
                    <span className="bk-cal-month">{MESSAGES.calendar.months[calMonth]}</span>
                </div>
                <button className="bk-cal-chev" onClick={nextMonth} aria-label={MESSAGES.calendar.nextMonth} type="button">
                    <Chevron dir="right" />
                </button>
            </div>

            {!statusLoading && !bookingsEnabled && (
                <div className="bk-banner bk-banner--warn" role="alert">
                    <span style={{ fontFamily: 'var(--bk-serif), serif', fontSize: '15px', color: 'var(--ink)' }}>
                        Бронювання тимчасово недоступне
                    </span>
                    <span>{status?.reason ?? 'Спробуйте, будь ласка, пізніше.'}</span>
                </div>
            )}

            {bookingsEnabled && isError && !isLoading && (
                <div className="bk-banner bk-banner--error" role="alert">
                    Не вдалося завантажити доступність. Спробуйте оновити сторінку.
                </div>
            )}

            <div className="bk-cal-grid" role="grid" aria-label="Календар">
                {MESSAGES.calendar.days.map((d) => (
                    <div key={d} className="bk-cal-dow" role="columnheader">{d}</div>
                ))}

                {Array.from({ length: firstDay }).map((_, i) => (
                    <div key={`empty-${i}`} className="bk-day empty" aria-hidden="true" />
                ))}

                {Array.from({ length: daysInMonth }).map((_, i) => {
                    const day = i + 1;
                    const thisDate = new Date(calYear, calMonth, day);
                    const dKey = dateKey(thisDate);
                    const isPast = thisDate < today;
                    const info = availabilityMap.get(dKey);
                    const slots = info?.slots ?? MAX_BIG;
                    const isBlocked = info?.blocked ?? false;
                    const isSel = selectedDate === dKey;

                    if (isLoading && !isPast) {
                        return <div key={dKey} className="bk-day-skeleton" aria-hidden="true" />;
                    }

                    let state = 'available';
                    if (isPast) state = 'past';
                    else if (!bookingsEnabled || isBlocked) state = 'blocked';
                    else if (slots <= 0) state = 'booked';
                    else if (slots <= 5) state = 'limited';

                    const clickable = !isPast && bookingsEnabled && !isBlocked && slots > 0;

                    return (
                        <div
                            key={dKey}
                            className={`bk-day ${state} ${isSel ? 'selected' : ''}`}
                            onClick={clickable ? () => setDate(dKey) : undefined}
                            role="gridcell"
                            aria-label={`${day} ${MESSAGES.calendar.months[calMonth]}`}
                            aria-pressed={isSel}
                            aria-disabled={!clickable}
                            tabIndex={clickable ? 0 : -1}
                            onKeyDown={(e) => {
                                if ((e.key === 'Enter' || e.key === ' ') && clickable) setDate(dKey);
                            }}
                        >
                            <span className="bk-day-num">{day}</span>
                            {!isPast && <span className="bk-day-dot" aria-hidden="true" />}
                        </div>
                    );
                })}
            </div>

            <div className="bk-legend" aria-label="Легенда">
                {LEGEND.map((l) => (
                    <div key={l.cls} className="bk-legend-item">
                        <span className="bk-legend-dot" style={{ background: l.color }} />
                        {l.label}
                    </div>
                ))}
            </div>
        </div>
    );
}
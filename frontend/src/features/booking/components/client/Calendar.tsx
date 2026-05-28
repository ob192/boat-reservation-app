'use client';

import { useState } from 'react';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useAvailability, useBookingSystemStatus } from '@/features/booking/hooks';
import { MESSAGES, MAX_BIG } from '@/features/booking/messages';

function dateKey(d: Date): string {
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

function monthKey(y: number, m: number): string {
    return `${y}-${String(m + 1).padStart(2, '0')}`;
}

function BookingsDisabledBanner({ reason }: { reason?: string }) {
    return (
        <div
            style={{
                background: '#fefce8',
                border: '1px solid #fde68a',
                borderRadius: 12,
                padding: '1.25rem 1rem',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: '0.5rem',
                textAlign: 'center',
                marginBottom: '1rem',
            }}
            role="alert"
        >
            <span style={{ fontSize: '1.75rem' }}>🚧</span>
            <div style={{ fontWeight: 600, fontSize: '0.9rem', color: '#92400e' }}>
                Бронювання тимчасово недоступне
            </div>
            <div style={{ fontSize: '0.78rem', color: '#a16207', lineHeight: 1.5, maxWidth: 300 }}>
                {reason ?? 'Спробуйте, будь ласка, пізніше.'}
            </div>
        </div>
    );
}

const LEGEND_ITEMS = [
    { color: 'var(--teal)',                    opacity: 1,   labelKey: 'available'     },
    { color: 'var(--gold)',                    opacity: 1,   labelKey: 'limited'       },
    { color: 'var(--coral)',                   opacity: 1,   labelKey: 'booked'        },
    { color: 'var(--subtle)',                  opacity: 0.6, labelKey: 'blocked'       },
    { color: 'var(--gold)',                    opacity: 0.8, labelKey: 'fullyBlocked'  },
] as const;

export function Calendar() {
    const { selectedDate, setDate } = useBookingStore();
    const [calYear, setCalYear] = useState(() => new Date().getFullYear());
    const [calMonth, setCalMonth] = useState(() => new Date().getMonth());

    const { data: status, isLoading: statusLoading } = useBookingSystemStatus();
    const bookingsEnabled = status?.bookingsEnabled ?? true;

    const month = monthKey(calYear, calMonth);
    const {
        data: availability,
        isLoading: availLoading,
        isError,
    } = useAvailability(month, !statusLoading && bookingsEnabled);

    const availabilityMap = new Map<string, { slots: number; blocked: boolean; fullyBlocked: boolean }>();
    availability?.days.forEach((d) =>
        availabilityMap.set(d.date, {
            slots: d.availableSlots,
            blocked: d.blocked,
            fullyBlocked: d.fullyBlocked,
        }),
    );

    const isLoading = statusLoading || availLoading;

    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const firstDayRaw = new Date(calYear, calMonth, 1).getDay();
    const firstDay = firstDayRaw === 0 ? 6 : firstDayRaw - 1;
    const daysInMonth = new Date(calYear, calMonth + 1, 0).getDate();

    const prevMonth = () => {
        if (calMonth === 0) { setCalYear((y) => y - 1); setCalMonth(11); }
        else setCalMonth((m) => m - 1);
    };

    const nextMonth = () => {
        if (calMonth === 11) { setCalYear((y) => y + 1); setCalMonth(0); }
        else setCalMonth((m) => m + 1);
    };

    return (
        <div>
            <div className="calendar-nav">
                <button className="cal-nav-btn" onClick={prevMonth} aria-label={MESSAGES.calendar.prevMonth} type="button">
                    ‹
                </button>
                <span className="cal-month">
          {MESSAGES.calendar.months[calMonth]} {calYear}
        </span>
                <button className="cal-nav-btn" onClick={nextMonth} aria-label={MESSAGES.calendar.nextMonth} type="button">
                    ›
                </button>
            </div>

            {!statusLoading && !bookingsEnabled && (
                <BookingsDisabledBanner reason={status?.reason} />
            )}

            {bookingsEnabled && isLoading && (
                <div className="availability-loading-banner" role="status" aria-live="polite">
                    <div className="mini-spinner" aria-hidden="true" />
                    Завантажуємо доступність…
                </div>
            )}
            {bookingsEnabled && isError && !isLoading && (
                <div className="availability-error-banner" role="alert">
                    ⚠️ Не вдалося завантажити доступність. Спробуйте оновити сторінку.
                </div>
            )}

            <div
                className={`cal-grid${!bookingsEnabled ? ' cal-grid--disabled' : ''}`}
                role="grid"
                aria-label="Календар"
            >
                {MESSAGES.calendar.days.map((d) => (
                    <div key={d} className="cal-day-label" role="columnheader">{d}</div>
                ))}

                {Array.from({ length: firstDay }).map((_, i) => (
                    <div key={`empty-${i}`} className="cal-day empty" aria-hidden="true" />
                ))}

                {Array.from({ length: daysInMonth }).map((_, i) => {
                    const day = i + 1;
                    const thisDate = new Date(calYear, calMonth, day);
                    const dKey = dateKey(thisDate);
                    const isPast = thisDate < today;
                    const info = availabilityMap.get(dKey);
                    const slotsAvailable = info?.slots ?? MAX_BIG;
                    const isBlocked = info?.blocked ?? false;
                    const isFullyBlocked = info?.fullyBlocked ?? false;
                    const isSel = selectedDate === dKey;

                    if (isLoading && !isPast) {
                        return (
                            <div
                                key={dKey}
                                className="cal-day cal-day-skeleton"
                                aria-hidden="true"
                                tabIndex={-1}
                            >
                                <span className="day-num" style={{ opacity: 0.5 }}>{day}</span>
                            </div>
                        );
                    }

                    if (!bookingsEnabled && !isPast) {
                        return (
                            <div
                                key={dKey}
                                className="cal-day date-blocked"
                                role="gridcell"
                                aria-label={`${day} ${MESSAGES.calendar.months[calMonth]} — недоступно`}
                                aria-disabled="true"
                                tabIndex={-1}
                            >
                                <span className="day-num">{day}</span>
                                <div className="avail-dot" aria-hidden="true" />
                            </div>
                        );
                    }

                    let statusClass = '';
                    if (!isPast) {
                        if (isFullyBlocked) {
                            statusClass = 'date-blocked date-fully-blocked';
                        } else if (isBlocked) {
                            statusClass = 'date-blocked';
                        } else if (slotsAvailable <= 0) {
                            statusClass = 'full';
                        } else if (slotsAvailable <= 5) {
                            statusClass = 'partial';
                        }
                    }

                    const isClickable = !isPast && !isBlocked && !isFullyBlocked && slotsAvailable > 0;

                    return (
                        <div
                            key={dKey}
                            className={`cal-day ${isPast ? 'past' : ''} ${statusClass} ${isSel ? 'selected' : ''}`}
                            onClick={isClickable ? () => setDate(dKey) : undefined}
                            role="gridcell"
                            aria-label={`${day} ${MESSAGES.calendar.months[calMonth]}${isFullyBlocked ? ' — тимчасово недоступно' : isBlocked ? ' — недоступно' : ''}`}
                            aria-pressed={isSel}
                            aria-disabled={!isClickable}
                            tabIndex={isClickable ? 0 : -1}
                            onKeyDown={(e) => {
                                if ((e.key === 'Enter' || e.key === ' ') && isClickable) setDate(dKey);
                            }}
                        >
                            <span className="day-num">{day}</span>
                            <div className="avail-dot" aria-hidden="true" />
                        </div>
                    );
                })}
            </div>

            {/* ── Legend — 2-col grid on mobile, single row on wider screens ── */}
            <div className="cal-legend" aria-label="Легенда">
                {LEGEND_ITEMS.map((item) => (
                    <div key={item.labelKey} className="cal-legend-item">
                        <div
                            className="ld"
                            style={{ background: item.color, opacity: item.opacity }}
                        />
                        {MESSAGES.calendar[item.labelKey as keyof typeof MESSAGES.calendar] as string}
                    </div>
                ))}
            </div>
        </div>
    );
}
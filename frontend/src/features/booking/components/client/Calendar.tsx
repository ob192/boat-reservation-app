'use client';

import { useState } from 'react';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useAvailability } from '@/features/booking/hooks';
import { MESSAGES, MAX_BIG } from '@/features/booking/messages';

function dateKey(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

function monthKey(y: number, m: number): string {
  return `${y}-${String(m + 1).padStart(2, '0')}`;
}

export function Calendar() {
  const { selectedDate, setDate } = useBookingStore();
  const [calYear, setCalYear] = useState(() => new Date().getFullYear());
  const [calMonth, setCalMonth] = useState(() => new Date().getMonth());

  const month = monthKey(calYear, calMonth);
  const { data: availability, isLoading, isError } = useAvailability(month);

  // Map date -> { availableSlots, blocked }
  const availabilityMap = new Map<string, { slots: number; blocked: boolean }>();
  availability?.days.forEach((d) =>
      availabilityMap.set(d.date, { slots: d.availableSlots, blocked: d.blocked }),
  );

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

        {/* Loading / error banner */}
        {isLoading && (
            <div className="availability-loading-banner" role="status" aria-live="polite">
              <div className="mini-spinner" aria-hidden="true" />
              Завантажуємо доступність…
            </div>
        )}
        {isError && !isLoading && (
            <div className="availability-error-banner" role="alert">
              ⚠️ Не вдалося завантажити доступність. Спробуйте оновити сторінку.
            </div>
        )}

        <div className="cal-grid" role="grid" aria-label="Календар">
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
            const isSel = selectedDate === dKey;

            // While loading: show shimmer on non-past days
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

            // Determine status class
            let statusClass = '';
            if (!isPast) {
              if (isBlocked) {
                statusClass = 'date-blocked';
              } else if (slotsAvailable <= 0) {
                statusClass = 'full';
              } else if (slotsAvailable <= 5) {
                statusClass = 'partial';
              }
            }

            const isClickable = !isPast && !isBlocked && slotsAvailable > 0;

            return (
                <div
                    key={dKey}
                    className={`cal-day ${isPast ? 'past' : ''} ${statusClass} ${isSel ? 'selected' : ''}`}
                    onClick={isClickable ? () => setDate(dKey) : undefined}
                    role="gridcell"
                    aria-label={`${day} ${MESSAGES.calendar.months[calMonth]}${isBlocked ? ' — недоступно' : ''}`}
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

        <div className="cal-legend" aria-label="Легенда">
          <div className="cal-legend-item">
            <div className="ld" style={{ background: 'var(--teal)' }} />
            {MESSAGES.calendar.available}
          </div>
          <div className="cal-legend-item">
            <div className="ld" style={{ background: 'var(--gold)' }} />
            {MESSAGES.calendar.limited}
          </div>
          <div className="cal-legend-item">
            <div className="ld" style={{ background: 'var(--coral)' }} />
            {MESSAGES.calendar.booked}
          </div>
          <div className="cal-legend-item">
            <div className="ld" style={{ background: 'var(--subtle)', opacity: 0.6 }} />
            {MESSAGES.calendar.blocked}
          </div>
        </div>
      </div>
  );
}
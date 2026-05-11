'use client';

import { useState } from 'react';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useAvailability } from '@/features/booking/hooks';
import { MESSAGES, MAX_SLOTS } from '@/features/booking/messages';

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
  const { data: availability } = useAvailability(month);

  const availabilityMap = new Map<string, number>();
  availability?.days.forEach((d) => availabilityMap.set(d.date, d.availableSlots));

  const today = new Date();
  today.setHours(0, 0, 0, 0);

  const firstDayRaw = new Date(calYear, calMonth, 1).getDay();
  const firstDay = firstDayRaw === 0 ? 6 : firstDayRaw - 1; // Monday-first
  const daysInMonth = new Date(calYear, calMonth + 1, 0).getDate();

  const prevMonth = () => {
    if (calMonth === 0) { setCalYear((y) => y - 1); setCalMonth(11); }
    else setCalMonth((m) => m - 1);
  };

  const nextMonth = () => {
    if (calMonth === 11) { setCalYear((y) => y + 1); setCalMonth(0); }
    else setCalMonth((m) => m + 1);
  };

  const handleSelect = (d: Date) => {
    setDate(dateKey(d));
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

      <div className="cal-grid" role="grid" aria-label="Календар">
        {/* Day labels */}
        {MESSAGES.calendar.days.map((d) => (
          <div key={d} className="cal-day-label" role="columnheader">
            {d}
          </div>
        ))}

        {/* Empty padding cells */}
        {Array.from({ length: firstDay }).map((_, i) => (
          <div key={`empty-${i}`} className="cal-day empty" aria-hidden="true" />
        ))}

        {/* Day cells */}
        {Array.from({ length: daysInMonth }).map((_, i) => {
          const day = i + 1;
          const thisDate = new Date(calYear, calMonth, day);
          const dKey = dateKey(thisDate);
          const isPast = thisDate < today;
          const slotsAvailable = availabilityMap.get(dKey) ?? MAX_SLOTS;
          const isSel = selectedDate === dKey;

          let statusClass = '';
          if (!isPast) {
            if (slotsAvailable <= 0) statusClass = 'full';
            else if (slotsAvailable <= 5) statusClass = 'partial';
          }

          return (
            <div
              key={dKey}
              className={`cal-day ${isPast ? 'past' : ''} ${statusClass} ${isSel ? 'selected' : ''}`}
              onClick={!isPast && slotsAvailable > 0 ? () => handleSelect(thisDate) : undefined}
              role="gridcell"
              aria-label={`${day} ${MESSAGES.calendar.months[calMonth]}`}
              aria-pressed={isSel}
              aria-disabled={isPast || slotsAvailable <= 0}
              tabIndex={isPast ? -1 : 0}
              onKeyDown={(e) => {
                if ((e.key === 'Enter' || e.key === ' ') && !isPast && slotsAvailable > 0) {
                  handleSelect(thisDate);
                }
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
      </div>
    </div>
  );
}

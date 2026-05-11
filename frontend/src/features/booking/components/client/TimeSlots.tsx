'use client';

import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useSlots } from '@/features/booking/hooks';
import { MESSAGES, MAX_SLOTS, TIME_OPTIONS } from '@/features/booking/messages';

export function TimeSlots() {
  const { selectedDate, selectedTime, setTime } = useBookingStore();
  const { data: slotsData, isLoading } = useSlots(selectedDate);

  if (!selectedDate) return null;

  const slotsMap = new Map<string, { available: number; total: number }>();
  slotsData?.slots.forEach((s) => slotsMap.set(s.time, { available: s.available, total: s.total }));

  return (
    <div className="time-slots" role="group" aria-label="Доступні часові слоти">
      {TIME_OPTIONS.map((t) => {
        const slotInfo = slotsMap.get(t);
        const available = isLoading ? MAX_SLOTS : (slotInfo?.available ?? MAX_SLOTS);
        const total = slotInfo?.total ?? MAX_SLOTS;
        const isFull = available <= 0;
        const pct = ((total - available) / total) * 100;
        const isSel = selectedTime === t;

        const fillColor = isFull
          ? 'var(--coral)'
          : available <= 4
            ? 'var(--gold)'
            : 'var(--teal)';

        const availText = isFull
          ? MESSAGES.time.noSlots
          : MESSAGES.time.slotsAvailable(available);

        return (
          <div
            key={t}
            className={`time-slot ${isFull ? 'full' : ''} ${isSel ? 'selected' : ''}`}
            onClick={!isFull ? () => setTime(t) : undefined}
            role="radio"
            aria-checked={isSel}
            aria-disabled={isFull}
            tabIndex={isFull ? -1 : 0}
            onKeyDown={(e) => {
              if ((e.key === 'Enter' || e.key === ' ') && !isFull) setTime(t);
            }}
          >
            {isFull && <div className="full-tag">{MESSAGES.time.fullTag}</div>}
            <div className="time">{t}</div>
            <div className="period">{MESSAGES.time.periods[t] ?? ''}</div>
            <div className="slots-bar" aria-hidden="true">
              <div
                className="slots-fill"
                style={{
                  width: `${pct}%`,
                  background: isSel ? 'var(--seafoam)' : fillColor,
                }}
              />
            </div>
            <div className="slots-text">🛶 {availText}</div>
          </div>
        );
      })}
    </div>
  );
}

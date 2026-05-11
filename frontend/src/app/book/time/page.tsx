'use client';

import { useRouter } from 'next/navigation';
import { TimeSlots } from '@/features/booking/components/client/TimeSlots';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useStepGuard } from '@/features/booking/hooks/useStepGuard';
import { MESSAGES } from '@/features/booking/messages';

export default function TimePage() {
  useStepGuard('time');

  const { selectedDate, selectedTime } = useBookingStore();
  const router = useRouter();

  const dateLabel = selectedDate
    ? new Date(selectedDate + 'T00:00:00').toLocaleDateString('uk-UA', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      })
    : MESSAGES.time.subtitle;

  return (
    <>
      <div className="card-header">
        <div className="card-header-icon">🕐</div>
        <div>
          <h3>{MESSAGES.time.title}</h3>
          <p>{dateLabel}</p>
        </div>
      </div>
      <div className="card-body">
        <TimeSlots />
      </div>
      <div className="nav-btns">
        <button className="btn-ghost" onClick={() => router.push('/book/date')} type="button">
          {MESSAGES.buttons.back}
        </button>
        <button
          className="btn-primary"
          disabled={!selectedTime}
          onClick={() => router.push('/book/boats')}
          type="button"
        >
          {MESSAGES.buttons.continue}
        </button>
      </div>
    </>
  );
}

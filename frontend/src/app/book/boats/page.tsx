'use client';

import { useRouter } from 'next/navigation';
import { BoatSelector } from '@/features/booking/components/client/BoatSelector';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useSlots } from '@/features/booking/hooks';
import { useStepGuard } from '@/features/booking/hooks/useStepGuard';
import { MESSAGES, MAX_SLOTS } from '@/features/booking/messages';

export default function BoatsPage() {
  useStepGuard('boats');

  const { selectedDate, selectedTime, quantities } = useBookingStore();
  const { data: slotsData } = useSlots(selectedDate);
  const router = useRouter();

  const slotInfo = slotsData?.slots.find((s) => s.time === selectedTime);
  const available = slotInfo?.available ?? MAX_SLOTS;

  const totalBoats = quantities.big + quantities.medium;
  const canProceed = totalBoats > 0 && totalBoats <= available;

  return (
    <>
      <div className="card-header">
        <div className="card-header-icon">⛵</div>
        <div>
          <h3>{MESSAGES.boats.title}</h3>
          <p>{MESSAGES.boats.slotsAvailable(available)}</p>
        </div>
      </div>
      <div className="card-body">
        <BoatSelector />
      </div>
      <div className="nav-btns">
        <button className="btn-ghost" onClick={() => router.push('/book/time')} type="button">
          {MESSAGES.buttons.back}
        </button>
        <button
          className="btn-primary"
          disabled={!canProceed}
          onClick={() => router.push('/book/details')}
          type="button"
        >
          {MESSAGES.buttons.continue}
        </button>
      </div>
    </>
  );
}

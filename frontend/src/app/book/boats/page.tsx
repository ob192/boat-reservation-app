'use client';

import { useRouter } from 'next/navigation';
import { BoatSelector } from '@/features/booking/components/client/BoatSelector';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useSlots } from '@/features/booking/hooks';
import { useStepGuard } from '@/features/booking/hooks/useStepGuard';
import { MESSAGES, MAX_BIG, MAX_MEDIUM } from '@/features/booking/messages';

export default function BoatsPage() {
    useStepGuard('boats');

    const { selectedDate, selectedTime, quantities } = useBookingStore();
    const { data: slotsData } = useSlots(selectedDate);
    const router = useRouter();

    const slotInfo = slotsData?.slots.find((s) => s.time === selectedTime);
    const availableBig = slotInfo?.availableBig ?? MAX_BIG;
    const availableMedium = slotInfo?.availableMedium ?? MAX_MEDIUM;

    const totalBoats = quantities.big + quantities.medium;
    const canProceed = totalBoats > 0 && quantities.big <= availableBig && quantities.medium <= availableMedium;

    const availabilityLabel =
        MESSAGES.boats.slotsAvailable(availableBig) + ' великих \u00b7 ' +
        MESSAGES.boats.slotsAvailable(availableMedium) + ' середніх';

    return (
        <>
            <div className="card-header">
                <div className="card-header-icon">⛵</div>
                <div>
                    <h3>{MESSAGES.boats.title}</h3>
                    <p>{availabilityLabel}</p>
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
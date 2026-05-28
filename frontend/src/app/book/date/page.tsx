'use client';

import { useRouter } from 'next/navigation';
import { Calendar } from '@/features/booking/components/client/Calendar';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { MESSAGES } from '@/features/booking/messages';

export default function DatePage() {
    const { selectedDate } = useBookingStore();
    const router = useRouter();

    return (
        <>
            <div className="card-header">
                <div className="card-header-icon">📅</div>
                <div>
                    <h3>{MESSAGES.calendar.title}</h3>
                    <p>{MESSAGES.calendar.subtitle}</p>
                </div>
            </div>
            <div className="card-body">
                <Calendar />
            </div>
            <div className="nav-btns">
                <div />
                <button
                    className="btn-primary"
                    disabled={!selectedDate}
                    onClick={() => router.push('/book/time')}
                    type="button"
                >
                    {MESSAGES.buttons.continue}
                </button>
            </div>
        </>
    );
}
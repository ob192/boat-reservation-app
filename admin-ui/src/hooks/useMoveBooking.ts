import { useMutation, useQueryClient } from '@tanstack/react-query';
import { adminFetch } from '@/lib/api';
import { MoveBookingResponse } from '@/lib/types';

export interface MoveBookingInput {
    bookingId: string;
    date: string;       // destination YYYY-MM-DD
    time: string;       // destination HH:MM
    routeName: string;  // destination route
}

/**
 * Moves a booking to a different date/time/route slot.
 * Capacity is enforced against the destination by the backend.
 *
 * `fromDate`/`fromTime`/`fromRoute` describe the booking's *current* slot and are
 * only used to invalidate the right caches so both the source and destination
 * views refresh after a successful move.
 */
export function useMoveBooking(fromDate: string, fromTime: string, fromRoute: string) {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ bookingId, date, time, routeName }: MoveBookingInput) =>
            adminFetch<MoveBookingResponse>(`/admin/bookings/${bookingId}/move`, {
                method: 'POST',
                body: JSON.stringify({ date, time, routeName }),
            }),
        onSuccess: (res) => {
            // Source slot loses the booking…
            qc.invalidateQueries({ queryKey: ['slot-bookings', fromDate, fromTime, fromRoute] });
            qc.invalidateQueries({ queryKey: ['slots', fromDate] });
            // …destination slot gains it.
            qc.invalidateQueries({ queryKey: ['slot-bookings', res.date, res.time, res.routeName] });
            qc.invalidateQueries({ queryKey: ['slots', res.date] });
            // Availability dots + history both change.
            qc.invalidateQueries({ queryKey: ['availability'] });
            qc.invalidateQueries({ queryKey: ['booking-history'] });
        },
    });
}
import { useQuery } from '@tanstack/react-query';
import { adminFetch } from '@/lib/api';
import { BookingHistoryResponse } from '@/lib/types';

/** All bookings for a single date, across every route. Used by the day report. */
export function useDayBookings(date: string, enabled = true) {
    return useQuery<BookingHistoryResponse>({
        queryKey: ['day-bookings', date],
        queryFn: () =>
            adminFetch<BookingHistoryResponse>(`/admin/bookings?date=${date}&limit=500&offset=0`),
        enabled: enabled && !!date,
        staleTime: 10_000,
    });
}
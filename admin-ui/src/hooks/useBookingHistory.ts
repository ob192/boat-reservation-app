import { useQuery, keepPreviousData } from '@tanstack/react-query';
import { adminFetch } from '@/lib/api';
import { BookingHistoryResponse, BookingStatus } from '@/lib/types';

export interface BookingHistoryParams {
    date?: string;
    status?: BookingStatus;
    limit?: number;
    offset?: number;
}

export function useBookingHistory(params: BookingHistoryParams) {
    const { date, status, limit = 50, offset = 0 } = params;

    const qs = new URLSearchParams();
    if (date) qs.set('date', date);
    if (status) qs.set('status', status);
    qs.set('limit', String(limit));
    qs.set('offset', String(offset));

    return useQuery<BookingHistoryResponse>({
        queryKey: ['booking-history', date ?? null, status ?? null, limit, offset],
        queryFn: () => adminFetch<BookingHistoryResponse>(`/admin/bookings?${qs.toString()}`),
        placeholderData: keepPreviousData, // smooth paging — no fl. between pages
        staleTime: 10_000,
    });
}
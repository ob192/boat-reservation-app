import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminFetch } from '@/lib/api';
import { SlotBookingsResponse } from '@/lib/types';

export function useSlotBookings(date: string, time: string, route: string, enabled = true) {
  return useQuery<SlotBookingsResponse>({
    queryKey: ['slot-bookings', date, time, route],
    queryFn: () => adminFetch<SlotBookingsResponse>(`/admin/slots/${date}/${time}/${route}/bookings`),
    enabled: enabled && !!date && !!time && !!route,
  });
}

export function useCancelBooking(date: string, time: string, route: string) {
  const qc = useQueryClient();
  return useMutation({
    // Cancel endpoint itself is unchanged; route is only used for cache invalidation.
    mutationFn: ({ bookingId, reason }: { bookingId: string; reason: string }) =>
        adminFetch(`/admin/bookings/${bookingId}/cancel`, {
          method: 'POST',
          body: JSON.stringify({ reason }),
        }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['slot-bookings', date, time, route] });
      qc.invalidateQueries({ queryKey: ['slots', date] });
      qc.invalidateQueries({ queryKey: ['booking-history'] });
    },
  });
}
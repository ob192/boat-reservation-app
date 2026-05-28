import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminFetch } from '@/lib/api';
import { SlotBookingsResponse } from '@/lib/types';

export function useSlotBookings(date: string, time: string, enabled = true) {
  return useQuery<SlotBookingsResponse>({
    queryKey: ['slot-bookings', date, time],
    queryFn: () => adminFetch<SlotBookingsResponse>(`/admin/slots/${date}/${time}/bookings`),
    enabled: enabled && !!date && !!time,
  });
}

export function useCancelBooking(date: string, time: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ bookingId, reason }: { bookingId: string; reason: string }) =>
      adminFetch(`/admin/bookings/${bookingId}/cancel`, {
        method: 'POST',
        body: JSON.stringify({ reason }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['slot-bookings', date, time] });
      qc.invalidateQueries({ queryKey: ['slots', date] });
    },
  });
}

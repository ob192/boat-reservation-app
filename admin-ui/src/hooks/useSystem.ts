import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminFetch, todayStr } from '@/lib/api';
import { SlotsResponse, SystemState } from '@/lib/types';

export function useSystem() {
  return useQuery<SystemState>({
    queryKey: ['system'],
    queryFn: async () => {
      const today = todayStr();
      const data = await adminFetch<SlotsResponse>(`/slots/${today}`);
      return {
        bookingsEnabled: data.bookingsEnabled,
      };
    },
    staleTime: 10_000,
  });
}

export function useToggleBookings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ enabled, reason }: { enabled: boolean; reason?: string }) =>
      adminFetch<{ bookingsEnabled: boolean; reason?: string; updatedAt: string }>(
        '/admin/system/bookings-enabled',
        {
          method: 'PUT',
          body: JSON.stringify({ enabled, reason }),
        }
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['system'] });
      qc.invalidateQueries({ queryKey: ['slots', todayStr()] });
    },
  });
}

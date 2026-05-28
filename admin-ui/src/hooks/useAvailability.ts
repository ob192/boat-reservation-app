import { useQuery } from '@tanstack/react-query';
import { adminFetch } from '@/lib/api';
import { AvailabilityResponse } from '@/lib/types';

export function useAvailability(month: string) {
  return useQuery<AvailabilityResponse>({
    queryKey: ['availability', month],
    queryFn: () => adminFetch<AvailabilityResponse>(`/availability/${month}`),
    enabled: !!month,
    staleTime: 60_000,
  });
}

export function useBlockDate() {
  return {
    blockDate: (date: string, reason?: string) =>
      adminFetch(`/admin/dates/${date}/block`, {
        method: 'PUT',
        body: JSON.stringify({ reason }),
      }),
    unblockDate: (date: string) =>
      adminFetch(`/admin/dates/${date}/block`, { method: 'DELETE' }),
  };
}

import { useQuery } from '@tanstack/react-query';
import { adminFetch } from '@/lib/api';
import { AvailabilityResponse } from '@/lib/types';

export function useAvailability(month: string, route: string) {
  return useQuery<AvailabilityResponse>({
    queryKey: ['availability', month, route],
    queryFn: () => adminFetch<AvailabilityResponse>(`/availability/${month}/${route}`),
    enabled: !!month && !!route,
    staleTime: 60_000,
  });
}
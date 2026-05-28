import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminFetch } from '@/lib/api';
import { SlotsResponse } from '@/lib/types';

export function useSlots(date: string) {
  return useQuery<SlotsResponse>({
    queryKey: ['slots', date],
    queryFn: () => adminFetch<SlotsResponse>(`/slots/${date}`),
    enabled: !!date,
  });
}

export function useUpsertSlot(date: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ time, capacityBig, capacityMedium }: {
      time: string;
      capacityBig: number;
      capacityMedium: number;
    }) =>
      adminFetch(`/admin/slots/${date}/${time}`, {
        method: 'PUT',
        body: JSON.stringify({ capacityBig, capacityMedium }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['slots', date] });
    },
  });
}

export function useBlockSlot(date: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ time, reason }: { time: string; reason?: string }) =>
      adminFetch(`/admin/slots/${date}/${time}/block`, {
        method: 'PUT',
        body: JSON.stringify({ reason }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['slots', date] });
    },
  });
}

export function useUnblockSlot(date: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ time }: { time: string }) =>
      adminFetch(`/admin/slots/${date}/${time}/block`, { method: 'DELETE' }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['slots', date] });
    },
  });
}

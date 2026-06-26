import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminFetch } from '@/lib/api';
import { SlotsResponse, CancelSlotResponse, UncancelSlotResponse } from '@/lib/types';

export function useSlots(date: string, route: string) {
  return useQuery<SlotsResponse>({
    queryKey: ['slots', date, route],
    queryFn: () => adminFetch<SlotsResponse>(`/slots/${date}/${route}`),
    enabled: !!date && !!route,
  });
}

export function useUpsertSlot(date: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ time, route, capacityBig, capacityMedium, capacitySmall }: {
      time: string;
      route: string;
      capacityBig: number;
      capacityMedium: number;
      capacitySmall: number;
    }) =>
        adminFetch(`/admin/slots/${date}/${time}/${route}`, {
          method: 'PUT',
          body: JSON.stringify({ capacityBig, capacityMedium, capacitySmall }),
        }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['slots', date] });
    },
  });
}

export function useBlockSlot(date: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ time, route, reason }: { time: string; route: string; reason?: string }) =>
        adminFetch(`/admin/slots/${date}/${time}/${route}/block`, {
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
    mutationFn: ({ time, route }: { time: string; route: string }) =>
        adminFetch(`/admin/slots/${date}/${time}/${route}/block`, { method: 'DELETE' }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['slots', date] });
    },
  });
}

export function useCancelSlot(date: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ time, route, reason }: { time: string; route: string; reason?: string }) =>
        adminFetch<CancelSlotResponse>(`/admin/slots/${date}/${time}/${route}/cancel`, {
          method: 'PUT',
          body: JSON.stringify({ reason }),
        }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['slots', date] });
      qc.invalidateQueries({ queryKey: ['slot-bookings'] });
      qc.invalidateQueries({ queryKey: ['booking-history'] });
    },
  });
}

export function useUncancelSlot(date: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ time, route }: { time: string; route: string }) =>
        adminFetch<UncancelSlotResponse>(`/admin/slots/${date}/${time}/${route}/cancel`, {
          method: 'DELETE',
        }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['slots', date] });
    },
  });
}

export function useDeleteSlot(date: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ time, route }: { time: string; route: string }) =>
        adminFetch<void>(`/admin/slots/${date}/${time}/${route}`, {
          method: 'DELETE',
          parseJson: false,
        }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['slots', date] });
      qc.invalidateQueries({ queryKey: ['availability'] });
    },
  });
}
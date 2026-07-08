import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminFetch } from '@/lib/api';
import { PromocodeListResponse, Promocode, CreatePromocodeRequest } from '@/lib/types';

/** List all promocodes, newest-created first. Optionally filter to one admin's codes. */
export function usePromocodes(createdBy?: string) {
  const qs = createdBy ? `?createdBy=${encodeURIComponent(createdBy)}` : '';
  return useQuery<PromocodeListResponse>({
    queryKey: ['promocodes', createdBy ?? null],
    queryFn: () => adminFetch<PromocodeListResponse>(`/admin/promocodes${qs}`),
    staleTime: 10_000,
  });
}

export function useCreatePromocode() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreatePromocodeRequest) =>
      adminFetch<Promocode>('/admin/promocodes', {
        method: 'POST',
        body: JSON.stringify(body),
      }),
    // Any createdBy filter variant is affected — invalidate the whole family.
    onSuccess: () => qc.invalidateQueries({ queryKey: ['promocodes'] }),
  });
}
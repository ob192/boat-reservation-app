'use client';

import { useQuery, useMutation } from '@tanstack/react-query';
import { apiFetch } from '@/shared/lib/api/client';
import type {
  StatusResponse,
  AvailabilityResponse,
  SlotsResponse,
  CreateBookingBody,
  CreateBookingResponse,
  CheckoutBody,
  CheckoutResponse,
  BookingStatusResponse,
} from '@/shared/lib/api/types';
import { getStableIdempotencyKey } from '@/shared/lib/idempotency';

export function useBookingSystemStatus() {
  return useQuery({
    queryKey: ['system-status'],
    queryFn: () => apiFetch<StatusResponse>('/status'),
    staleTime: 30_000,
    retry: 1,
  });
}

export function useAvailability(month: string, enabled = true) {
  return useQuery({
    queryKey: ['availability', month],
    queryFn: () => apiFetch<AvailabilityResponse>(`/availability/${month}`),
    staleTime: 30_000,
    enabled: !!month && enabled,
  });
}

export function useSlots(date: string | null) {
  return useQuery({
    queryKey: ['slots', date],
    queryFn: () => apiFetch<SlotsResponse>(`/slots/${date}`),
    staleTime: 15_000,
    enabled: !!date,
  });
}

export function useCreateBooking() {
  return useMutation({
    mutationFn: (body: CreateBookingBody) =>
        apiFetch<CreateBookingResponse>('/bookings', {
          method: 'POST',
          body: JSON.stringify(body),
          headers: { 'X-Idempotency-Key': getStableIdempotencyKey(body) },
        }),
  });
}

export function useCreateCheckout() {
  return useMutation({
    mutationFn: (body: CheckoutBody) =>
        apiFetch<CheckoutResponse>('/checkout', {
          method: 'POST',
          body: JSON.stringify(body),
        }),
  });
}

export function useBookingStatus(sessionId: string | null) {
  return useQuery({
    queryKey: ['booking-status', sessionId],
    queryFn: () => apiFetch<BookingStatusResponse>(`/bookings/${sessionId}`),
    enabled: !!sessionId,
    refetchInterval: (query) =>
        query.state.data?.status === 'confirmed' || query.state.data?.status === 'failed'
            ? false
            : 2000,
    refetchIntervalInBackground: false,
  });
}
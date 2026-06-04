'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useBookingStore } from '../store/bookingStore';

type RequiredStep = 'date' | 'time' | 'boats';

export function useStepGuard(requires: RequiredStep) {
  const router = useRouter();
  const { selectedRoute, selectedDate, selectedTime } = useBookingStore();

  useEffect(() => {
    if (!selectedRoute) {
      router.replace('/book/route');
      return;
    }
    if (requires === 'time' && !selectedDate) {
      router.replace('/book/date');
      return;
    }
    if (requires === 'boats' && (!selectedDate || !selectedTime)) {
      router.replace(!selectedDate ? '/book/date' : '/book/time');
      return;
    }
  }, [requires, selectedRoute, selectedDate, selectedTime, router]);
}
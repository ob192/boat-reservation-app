'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useBookingStore } from '../store/bookingStore';

type RequiredStep = 'date' | 'time' | 'boats';

export function useStepGuard(requires: RequiredStep) {
  const router = useRouter();
  const { selectedDate, selectedTime, quantities } = useBookingStore();

  useEffect(() => {
    if (requires === 'time' && !selectedDate) {
      router.replace('/book/date');
      return;
    }
    if (requires === 'boats' && (!selectedDate || !selectedTime)) {
      router.replace(!selectedDate ? '/book/date' : '/book/time');
      return;
    }
    if (requires === 'boats') {
      const hasBoats = quantities.big + quantities.medium > 0;
      // Only check on details page — boats step itself populates this
      void hasBoats;
    }
  }, [requires, selectedDate, selectedTime, quantities, router]);
}

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { clearIdempotencyKey } from '@/shared/lib/idempotency';

export interface BookingQuantities {
  big: number;
  medium: number;
  child: number;
}

export interface ContactInfo {
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
}

interface BookingState {
  // Step 1 — Date
  selectedDate: string | null; // YYYY-MM-DD

  // Step 2 — Time
  selectedTime: string | null; // HH:MM

  // Step 3 — Boats
  quantities: BookingQuantities;

  // Step 4 — Details
  contact: ContactInfo;

  // After booking is created
  bookingId: string | null;
  sessionId: string | null;

  // Actions
  setDate: (date: string) => void;
  setTime: (time: string) => void;
  setQuantity: (type: keyof BookingQuantities, value: number) => void;
  setContact: (contact: Partial<ContactInfo>) => void;
  setBookingId: (id: string) => void;
  setSessionId: (id: string) => void;
  reset: () => void;
}

const initialState = {
  selectedDate: null,
  selectedTime: null,
  quantities: { big: 0, medium: 0, child: 0 },
  contact: { firstName: '', lastName: '', email: '', phone: '' },
  bookingId: null,
  sessionId: null,
};

export const useBookingStore = create<BookingState>()(
  persist(
    (set) => ({
      ...initialState,

      setDate: (date) =>
        set({ selectedDate: date, selectedTime: null, quantities: { big: 0, medium: 0, child: 0 } }),

      setTime: (time) =>
        set({ selectedTime: time, quantities: { big: 0, medium: 0, child: 0 } }),

      setQuantity: (type, value) =>
        set((state) => ({
          quantities: { ...state.quantities, [type]: value },
        })),

      setContact: (contact) =>
        set((state) => ({
          contact: { ...state.contact, ...contact },
        })),

      setBookingId: (id) => set({ bookingId: id }),
      setSessionId: (id) => set({ sessionId: id }),

        reset: () => {
            clearIdempotencyKey();
            set(initialState);
        },
    }),
    {
      name: 'harbour-wave-booking',
      partialize: (state) => ({
        selectedDate: state.selectedDate,
        selectedTime: state.selectedTime,
        quantities: state.quantities,
      }),
    },
  ),
);


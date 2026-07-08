import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { clearIdempotencyKey } from '@/shared/lib/idempotency';
import type { RouteId } from '@/features/booking/routes';

export interface BookingQuantities {
    big: number;
    medium: number;
    small: number;
    child: number;
}

export interface ContactInfo {
    firstName: string;
    lastName: string;
    email: string;
    phone: string;
}

interface BookingState {
    selectedRoute: RouteId | null; // Step 1
    selectedDate: string | null;   // Step 2
    selectedTime: string | null;   // Step 3
    quantities: BookingQuantities; // Step 4
    contact: ContactInfo;          // Step 5
    promoCode: string | null;      // captured from ?promo= URL, applied at booking
    bookingId: string | null;
    sessionId: string | null;

    setRoute: (route: RouteId) => void;
    setDate: (date: string) => void;
    setTime: (time: string) => void;
    setQuantity: (type: keyof BookingQuantities, value: number) => void;
    setContact: (contact: Partial<ContactInfo>) => void;
    setPromoCode: (code: string | null) => void;
    setBookingId: (id: string) => void;
    setSessionId: (id: string) => void;
    reset: () => void;
}

const initialState = {
    selectedRoute: null,
    selectedDate: null,
    selectedTime: null,
    quantities: { big: 0, medium: 0, small: 0, child: 0 },
    contact: { firstName: '', lastName: '', email: '', phone: '' },
    promoCode: null,
    bookingId: null,
    sessionId: null,
};

export const useBookingStore = create<BookingState>()(
    persist(
        (set) => ({
            ...initialState,

            // Changing route invalidates date/time/quantities (route-scoped availability & prices)
            setRoute: (route) =>
                set({
                    selectedRoute: route,
                    selectedDate: null,
                    selectedTime: null,
                    quantities: { big: 0, medium: 0, small: 0, child: 0 },
                }),

            setDate: (date) =>
                set({ selectedDate: date, selectedTime: null, quantities: { big: 0, medium: 0, small: 0, child: 0 }}),

            setTime: (time) =>
                set({ selectedTime: time, quantities: { big: 0, medium: 0, small: 0, child: 0 }}),

            setQuantity: (type, value) =>
                set((state) => ({ quantities: { ...state.quantities, [type]: value } })),

            setContact: (contact) =>
                set((state) => ({ contact: { ...state.contact, ...contact } })),

            // Promo is independent of the route/date/time cascade — it must NOT
            // be cleared when those change, only on an explicit reset.
            setPromoCode: (code) => set({ promoCode: code }),

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
                selectedRoute: state.selectedRoute,
                selectedDate: state.selectedDate,
                selectedTime: state.selectedTime,
                quantities: state.quantities,
                promoCode: state.promoCode,
            }),
        },
    ),
);
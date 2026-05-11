// ─── Availability ────────────────────────────────────────────────────────────

export interface DayAvailability {
  date: string; // YYYY-MM-DD
  availableSlots: number;
}

export interface AvailabilityResponse {
  month: string; // YYYY-MM
  days: DayAvailability[];
}

// ─── Slots ───────────────────────────────────────────────────────────────────

export interface TimeSlot {
  time: string; // HH:MM
  available: number;
  total: number;
}

export interface SlotsResponse {
  date: string;
  slots: TimeSlot[];
}

// ─── Booking ─────────────────────────────────────────────────────────────────

export interface BookingQuantities {
  big: number;
  medium: number;
  child: number;
}

export interface BookingContact {
  firstName: string;
  lastName: string;
  email: string;
  phone?: string;
}

export interface CreateBookingBody {
  date: string; // YYYY-MM-DD
  time: string; // HH:MM
  quantities: BookingQuantities;
  contact: BookingContact;
}

export interface CreateBookingResponse {
  bookingId: string;
  totalAmount: number;
  expiresAt: string;
}

// ─── Checkout ────────────────────────────────────────────────────────────────

export interface CheckoutBody {
  bookingId: string;
}

export interface CheckoutResponse {
  checkoutUrl: string;
  sessionId: string;
}

// ─── Booking Status ──────────────────────────────────────────────────────────

export type BookingStatus = 'pending' | 'confirmed' | 'failed' | 'expired';

export interface BookingDetail {
  id: string;
  date: string;
  time: string;
  quantities: BookingQuantities;
  contact: BookingContact;
  totalAmount: number;
  status: BookingStatus;
}

export interface BookingStatusResponse {
  status: BookingStatus;
  booking?: BookingDetail;
}

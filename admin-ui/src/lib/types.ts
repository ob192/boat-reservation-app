export type BookingStatus = 'pending' | 'confirmed' | 'failed' | 'expired' | 'cancelled';

export interface SlotInfo {
  time: string;
  routeName: string;
  availableBig: number;
  totalBig: number;
  availableMedium: number;
  totalMedium: number;
  availableSmall: number;
  totalSmall: number;
  blocked: boolean;
  blockReason?: string;
  cancelled: boolean;          // NEW
  cancelReason?: string;       // NEW
  cancelledAt?: string;        // NEW
}

export interface SlotsResponse {
  date: string;
  dateBlocked: boolean;
  bookingsEnabled: boolean;
  fullyBlocked: boolean;
  slots: SlotInfo[];
}

export interface DayAvailability {
  date: string;
  availableSlots: number;
  blocked: boolean;
}

export interface AvailabilityResponse {
  month: string;
  days: DayAvailability[];
}

export interface Booking {
  id: string;
  date?: string;
  time?: string;
  routeName?: string;
  userEmail: string;
  firstName: string;
  lastName: string;
  phone?: string;
  quantities: {
    big: number;
    medium: number;
    small: number;
    child: number;
  };
  totalAmount: number;
  effectiveAmount: number;
  status: BookingStatus;
  createdAt: string;
  posterIncomingOrderId?: number | null;
  posterIncomingTransactionId?: number | null;
  // Frozen promo snapshot taken at booking time — null/absent if no promo was used.
  promoCode?: string | null;
  discountPercent?: number | null;
  discountAmount?: number | null;
}

// ── Promocodes ────────────────────────────────────────────────────
export interface Promocode {
  code: string;
  discountPercent: number;
  maxUses: number;
  timesUsed: number;           // only increments once a booking is confirmed (paid)
  active: boolean;
  createdBy: string;
  createdAt: string;
}

export interface PromocodeListResponse {
  promocodes: Promocode[];
}

export interface CreatePromocodeRequest {
  code: string;
  discountPercent: number;
  maxUses: number;
}

export interface SlotBookingsResponse {
  date: string;
  time: string;
  routeName: string;
  bookings: Booking[];
}

export interface BookingHistoryResponse {
  bookings: Booking[];
}

export interface SystemState {
  bookingsEnabled: boolean;
  reason?: string;
  updatedAt?: string;
}

// Cancel slot responses
export interface CancelSlotResponse {
  date: string;
  time: string;
  routeName: string;
  cancelled: true;
  reason?: string;
  cancelledAt: string;
  cancelledBookings: number;
}

export interface UncancelSlotResponse {
  date: string;
  time: string;
  routeName: string;
  cancelled: false;
}

// Move booking response (POST /admin/bookings/:bookingId/move)
export interface MoveBookingResponse {
  bookingId: string;
  date: string;
  time: string;
  routeName: string;
  status: BookingStatus;
}
// Status
export interface StatusResponse {
  bookingsEnabled: boolean;
  reason?: string;
}

// Availability
export interface DayAvailability {
  date: string;
  availableSlots: number;
  blocked: boolean;
  fullyBlocked: boolean;
}

export interface AvailabilityResponse {
  month: string;
  days: DayAvailability[];
}

// Slots
export interface TimeSlot {
  time: string;
  routeName: string;
  availableBig: number;
  availableMedium: number;
  availableSmall: number;
  totalBig: number;
  totalMedium: number;
  totalSmall: number;
  blocked: boolean;
  cancelled: boolean; // NEW — slot cancelled by admin, cascades to bookings
}

export interface SlotsResponse {
  date: string;
  dateBlocked: boolean;
  bookingsEnabled: boolean;
  fullyBlocked: boolean;
  slots: TimeSlot[];
}

// Booking
export interface BookingQuantities {
  big: number;
  medium: number;
  small: number;
  child: number;
}

export interface BookingContact {
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
}

export interface CreateBookingBody {
  routeName: string;
  date: string;
  time: string;
  quantities: BookingQuantities;
  contact: BookingContact;
  consent?: ConsentRecord;
  promoCode?: string; // optional — omit for no promo
}

export interface CreateBookingResponse {
  bookingId: string;
  totalAmount: number; // already net of any discount
  expiresAt: string;
  // Present only when a promo was applied — for display only.
  promoCode?: string | null;
  discountPercent?: number | null;
  discountAmount?: number | null;
}

// GET /promocodes/:code — deliberately minimal preview
export interface PromoPreviewResponse {
  discountPercent: number;
}

export interface CheckoutBody {
  bookingId: string;
  resultUrl: string;
}

export interface CheckoutResponse {
  checkoutUrl: string;
  sessionId: string;
}

// Booking Status
export type BookingStatus = 'pending' | 'confirmed' | 'failed' | 'expired';

export interface BookingDetail {
  id: string;
  date: string;
  time: string;
  routeName: string;
  quantities: BookingQuantities;
  contact: BookingContact;
  totalAmount: number; // already net of any discount
  status: BookingStatus;
  // Present only when a promo was applied — for display only.
  promoCode?: string | null;
  discountPercent?: number | null;
  discountAmount?: number | null;
}

export interface BookingStatusResponse {
  status: BookingStatus;
  booking?: BookingDetail;
}

export interface ConsentRecord {
  event: 'consent_agreed';
  consentId: string;
  agreementId: string;
  agreementVersion: string;
  agreementHash: string;
  user: { id?: string; email: string; name: string };
  device: {
    fingerprint: string;
    userAgent: string;
    platform: string;
    timezone: string;
    language: string;
    screen: string;
  };
  clientTimestamp: string;
}

// Admin — Slot cancellation
export interface SlotCancelBody {
  reason?: string;
}

export interface SlotCancelResponse {
  date: string;
  time: string;
  routeName: string;
  cancelled: true;
  reason?: string;
  cancelledAt: string;
  cancelledBookings: number;
}

export interface SlotUncancelResponse {
  date: string;
  time: string;
  routeName: string;
  cancelled: false;
}
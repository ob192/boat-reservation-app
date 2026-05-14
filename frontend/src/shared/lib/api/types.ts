// Availability
export interface DayAvailability {
  date: string;
  availableSlots: number;
  blocked: boolean;
}

export interface AvailabilityResponse {
  month: string;
  days: DayAvailability[];
}

// Slots
export interface TimeSlot {
  time: string;
  availableBig: number;
  availableMedium: number;
  totalBig: number;
  totalMedium: number;
  blocked: boolean;
}

export interface SlotsResponse {
  date: string;
  dateBlocked: boolean;
  bookingsEnabled: boolean;
  slots: TimeSlot[];
}

// Booking
export interface BookingQuantities {
  big: number;
  medium: number;
  child: number;
}

export interface BookingContact {
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
}

export interface CreateBookingBody {
  date: string;
  time: string;
  quantities: BookingQuantities;
  contact: BookingContact;
}

export interface CreateBookingResponse {
  bookingId: string;
  totalAmount: number;
  expiresAt: string;
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
  quantities: BookingQuantities;
  contact: BookingContact;
  totalAmount: number;
  status: BookingStatus;
}

export interface BookingStatusResponse {
  status: BookingStatus;
  booking?: BookingDetail;
}
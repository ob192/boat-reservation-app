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
  routeName: string;        // NEW
  availableBig: number;
  availableMedium: number;
  availableSmall: number;   // NEW
  totalBig: number;
  totalMedium: number;
  totalSmall: number;       // NEW
  blocked: boolean;
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
  routeName: string;        // NEW — required
  date: string;
  time: string;
  quantities: BookingQuantities;
  contact: BookingContact;
  consent?: ConsentRecord;
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
  routeName: string;        // NEW
  quantities: BookingQuantities;
  contact: BookingContact;
  totalAmount: number;
  status: BookingStatus;
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
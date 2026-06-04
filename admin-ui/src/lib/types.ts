export type BookingStatus = 'pending' | 'confirmed' | 'failed' | 'expired' | 'cancelled';

export interface SlotInfo {
  time: string;
  routeName: string;
  availableBig: number;
  totalBig: number;
  availableMedium: number;
  totalMedium: number;
  availableSmall: number;   // now confirmed always present
  totalSmall: number;       // now confirmed always present
  blocked: boolean;
  blockReason?: string;
}

export interface SlotsResponse {
  date: string;
  dateBlocked: boolean;
  bookingsEnabled: boolean;
  fullyBlocked: boolean;    // NEW
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
  date?: string;                // NEW (present on slot-bookings + history)
  time?: string;                // NEW
  routeName?: string;           // NEW
  userEmail: string;
  firstName: string;
  lastName: string;
  phone?: string;
  quantities: {
    big: number;
    medium: number;
    small: number;              // NEW
    child: number;
  };
  totalAmount: number;
  effectiveAmount: number;
  status: BookingStatus;
  createdAt: string;
  posterIncomingOrderId?: number | null;
  posterIncomingTransactionId?: number | null;
}

export interface SlotBookingsResponse {
  date: string;
  time: string;
  routeName: string;            // NEW
  bookings: Booking[];
}

export interface BookingHistoryResponse {   // NEW
  bookings: Booking[];
}

export interface SystemState {
  bookingsEnabled: boolean;
  reason?: string;
  updatedAt?: string;
}
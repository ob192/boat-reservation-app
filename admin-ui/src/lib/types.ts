export type BookingStatus = 'pending' | 'confirmed' | 'failed' | 'expired' | 'cancelled';

export interface SlotInfo {
  time: string;
  availableBig: number;
  totalBig: number;
  availableMedium: number;
  totalMedium: number;
  blocked: boolean;
  blockReason?: string;
}

export interface SlotsResponse {
  date: string;
  dateBlocked: boolean;
  bookingsEnabled: boolean;
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
  userEmail: string;
  firstName: string;
  lastName: string;
  phone?: string;
  quantities: {
    big: number;
    medium: number;
    child: number;
  };
  totalAmount: number;
  effectiveAmount: number;
  status: BookingStatus;
  createdAt: string;
}

export interface SlotBookingsResponse {
  date: string;
  time: string;
  bookings: Booking[];
}

export interface SystemState {
  bookingsEnabled: boolean;
  reason?: string;
  updatedAt?: string;
}

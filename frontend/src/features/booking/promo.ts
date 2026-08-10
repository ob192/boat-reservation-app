'use client';

// Client-side promo-code bookkeeping. The backend is authoritative for whether
// a code is valid / how much it discounts and for counting redemptions — this
// module only keeps a small "receipt" captured at booking-creation time, so the
// success page can show the discount even though `GET /bookings/by-session/:id`
// isn't guaranteed to echo promo fields back.
//
// There is deliberately NO local "already redeemed" list: a code may be applied
// as many times as the backend allows (codes created with `maxUses = 0` are
// uncapped), so the client never blocks a repeat redemption on its own.

const RECEIPT_PREFIX = 'harbour-wave-promo-receipt:';

/** Codes are matched case-insensitively/trimmed everywhere (matches backend). */
export function normalizePromo(code: string): string {
    return code.trim().toUpperCase();
}

export interface PromoReceipt {
    promoCode: string;
    discountPercent: number;
    discountAmount: number;
}

/** Persist the applied-promo snapshot from the booking-creation response. */
export function savePromoReceipt(bookingId: string, receipt: PromoReceipt): void {
    try {
        localStorage.setItem(RECEIPT_PREFIX + bookingId, JSON.stringify(receipt));
    } catch {
        // best effort
    }
}

export function getPromoReceipt(bookingId: string): PromoReceipt | null {
    try {
        const raw = localStorage.getItem(RECEIPT_PREFIX + bookingId);
        return raw ? (JSON.parse(raw) as PromoReceipt) : null;
    } catch {
        return null;
    }
}

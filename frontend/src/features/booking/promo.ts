'use client';

// Client-side promo-code bookkeeping. The backend is authoritative for whether
// a code is valid / how much it discounts and for counting redemptions — this
// module only tracks two things locally:
//   1. which codes this browser has already redeemed (so a `?promo=` link isn't
//      re-applied on a second booking), and
//   2. a small "receipt" captured at booking-creation time, so the success page
//      can show the discount even though `GET /bookings/by-session/:id` isn't
//      guaranteed to echo promo fields back.

const USED_KEY = 'harbour-wave-promo-used';
const RECEIPT_PREFIX = 'harbour-wave-promo-receipt:';

/** Codes are matched case-insensitively/trimmed everywhere (matches backend). */
export function normalizePromo(code: string): string {
    return code.trim().toUpperCase();
}

function getUsedPromoCodes(): string[] {
    try {
        const raw = localStorage.getItem(USED_KEY);
        const arr = raw ? JSON.parse(raw) : [];
        return Array.isArray(arr) ? (arr as string[]) : [];
    } catch {
        return [];
    }
}

/** True once this browser has completed a booking that redeemed `code`. */
export function isPromoUsed(code: string | null | undefined): boolean {
    if (!code) return false;
    return getUsedPromoCodes().includes(normalizePromo(code));
}

/** Remember that `code` was redeemed so we never auto-apply it again here. */
export function markPromoUsed(code: string | null | undefined): void {
    if (!code) return;
    const norm = normalizePromo(code);
    const used = getUsedPromoCodes();
    if (used.includes(norm)) return;
    try {
        localStorage.setItem(USED_KEY, JSON.stringify([...used, norm]));
    } catch {
        // quota / SSR — best effort
    }
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

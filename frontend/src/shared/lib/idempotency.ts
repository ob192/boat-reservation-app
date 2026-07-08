'use client';

import type { CreateBookingBody } from '@/shared/lib/api/types';

const STORAGE_KEY = 'harbour-wave-idempotency';
const TTL_MS = 5 * 60 * 1000; // 5 minutes — hard cap, anchored to creation

interface IdemRecord {
    key: string;
    createdAt: number;
    fingerprint: string;
}

// Only the price/capacity-affecting fields define booking identity.
// Contact edits (typo fix in phone) must NOT mint a new key, or you'd
// double-hold the slot. Slot + quantities + promo is the dedupe boundary —
// promo is included so applying/changing a code re-mints the key rather than
// reusing an earlier discount-free hold for the same slot.
function fingerprint(body: CreateBookingBody): string {
    const { routeName, date, time, quantities, promoCode } = body;
    return `${routeName}|${date}|${time}|${quantities.big}|${quantities.medium}|${quantities.child}|${promoCode ?? ''}`;
}

/**
 * Returns a stable idempotency key for this booking attempt.
 * Reused across retries/refreshes for ≤5 min if the payload is unchanged;
 * otherwise a fresh key is generated and persisted.
 */
export function getStableIdempotencyKey(body: CreateBookingBody): string {
    const fp = fingerprint(body);
    const now = Date.now();

    try {
        const raw = localStorage.getItem(STORAGE_KEY);
        if (raw) {
            const rec = JSON.parse(raw) as IdemRecord;
            const fresh = now - rec.createdAt < TTL_MS;
            if (fresh && rec.fingerprint === fp && rec.key) {
                return rec.key;
            }
        }
    } catch {
        // corrupt / unavailable storage — fall through to a new key
    }

    const key = crypto.randomUUID();
    try {
        localStorage.setItem(
            STORAGE_KEY,
            JSON.stringify({ key, createdAt: now, fingerprint: fp } satisfies IdemRecord),
        );
    } catch {
        // quota / SSR — still return a usable key
    }
    return key;
}

export function clearIdempotencyKey(): void {
    try {
        localStorage.removeItem(STORAGE_KEY);
    } catch {
        // ignore
    }
}
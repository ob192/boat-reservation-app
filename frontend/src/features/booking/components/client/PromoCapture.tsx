'use client';

import { Suspense, useEffect } from 'react';
import { useSearchParams } from 'next/navigation';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { normalizePromo } from '@/features/booking/promo';

/**
 * Captures a `?promo=CODE` query param anywhere in the app and stashes it in the
 * booking store so it can be applied at checkout. A previously redeemed code is
 * captured again — re-visiting a promo link re-applies it, and the backend
 * decides whether the code still has redemptions left.
 */
function PromoCaptureInner() {
    const params = useSearchParams();
    const setPromoCode = useBookingStore((s) => s.setPromoCode);

    useEffect(() => {
        const raw = params.get('promo');
        if (!raw) return;
        const code = normalizePromo(raw);
        if (!code) return;
        setPromoCode(code);
    }, [params, setPromoCode]);

    return null;
}

export function PromoCapture() {
    // useSearchParams requires a Suspense boundary to avoid a CSR bailout.
    return (
        <Suspense fallback={null}>
            <PromoCaptureInner />
        </Suspense>
    );
}
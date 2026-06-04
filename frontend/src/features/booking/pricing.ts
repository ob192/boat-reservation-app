import type { RouteId } from '@/features/booking/routes';

export interface RoutePrices {
    big: number;
    medium: number;
    small: number;
    child: number;
}

// Per-unit prices, per route. These are CLIENT-SIDE ESTIMATES for live display
// only — the backend's `totalAmount` is always authoritative for the actual
// charge. Keep this table in sync with the backend route/price config.
export const ROUTE_PRICES: Record<RouteId, RoutePrices> = {
    Desna:    { big: 450, medium: 350, small: 250, child: 225 },
    Klochkov: { big: 600, medium: 480, small: 340, child: 300 },
};

// Used before/if a route is somehow missing (the step guard prevents this in
// practice, but we never want pricing to crash the boats/details UI).
const FALLBACK_PRICES: RoutePrices = ROUTE_PRICES.Desna;

export function getRoutePrices(route: RouteId | string | null | undefined): RoutePrices {
    if (route && route in ROUTE_PRICES) return ROUTE_PRICES[route as RouteId];
    return FALLBACK_PRICES;
}

export interface PriceableQuantities {
    big?: number;
    medium?: number;
    small?: number;
    child?: number;
}

/** Live estimate of the booking total for a given route + quantities. */
export function calculateBookingTotal(
    route: RouteId | string | null | undefined,
    q: PriceableQuantities,
): number {
    const p = getRoutePrices(route);
    return (
        (q.big ?? 0) * p.big +
        (q.medium ?? 0) * p.medium +
        (q.small ?? 0) * p.small +
        (q.child ?? 0) * p.child
    );
}
// Route identifiers are backend-owned and NOT localized (see API doc).
// The list can grow, so routeLabel() falls back to the raw name.
export const ROUTES = ['Desna', 'Klochkov'] as const;
export type RouteName = typeof ROUTES[number];

export const ROUTE_LABELS: Record<string, string> = {
    Desna: 'Сплав на Десні',
    Klochkov: 'Сплав Клочков',
};

export function routeLabel(route: string): string {
    return ROUTE_LABELS[route] ?? route;
}
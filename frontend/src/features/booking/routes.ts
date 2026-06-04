export const ROUTE_IDS = ['Desna', 'Klochkov'] as const;
export type RouteId = (typeof ROUTE_IDS)[number];

export interface RouteOption {
    id: RouteId;   // sent to backend verbatim — case-sensitive
    label: string; // Ukrainian display label (backend does NOT localise)
    sub: string;
    desc: string;
}

export const ROUTES: RouteOption[] = [
    {
        id: 'Desna',
        label: 'Сплав на Десні',
        sub: 'Класичний маршрут',
        desc: 'Спокійна вода вздовж Десни — мальовничі краєвиди й неспішний темп. Гарний вибір для першого виходу.',
    },
    {
        id: 'Klochkov',
        label: 'Сплав на Снові',
        sub: 'Маршрут Клочків -> Сновянка',
        desc: 'Прогулянка в районі Клочкова — трохи динамічніша вода та інші ракурси міста.',
    },
];

export const ROUTE_LABELS: Record<string, string> = Object.fromEntries(
    ROUTES.map((r) => [r.id, r.label]),
);

export function routeLabel(id: string | null | undefined): string {
    if (!id) return '';
    return ROUTE_LABELS[id] ?? id;
}
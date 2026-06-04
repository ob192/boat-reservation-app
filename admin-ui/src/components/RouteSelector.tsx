'use client';

import { ROUTES, RouteName, routeLabel } from '@/lib/routes';

interface RouteSelectorProps {
    value: RouteName;
    onChange: (route: RouteName) => void;
}

export function RouteSelector({ value, onChange }: RouteSelectorProps) {
    return (
        <div className="route-selector" role="tablist" aria-label="Маршрут">
            {ROUTES.map(r => (
                <button
                    key={r}
                    type="button"
                    role="tab"
                    aria-selected={value === r}
                    className={`route-selector-btn${value === r ? ' active' : ''}`}
                    onClick={() => onChange(r)}
                >
                    {routeLabel(r)}
                </button>
            ))}
        </div>
    );
}
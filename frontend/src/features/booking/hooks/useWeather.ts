'use client';

import { useQuery } from '@tanstack/react-query';
import type { WeatherDay } from '@/app/api/weather/cache';

export type { WeatherDay };

export interface WeatherAllResponse {
    days: WeatherDay[];
    fetchedAt: string;
}

export interface WeatherDayResponse {
    day: WeatherDay;
    fetchedAt: string;
}

// Fetch all 60 days at once on mount — single request, no per-date calls
async function fetchAllWeather(): Promise<WeatherAllResponse> {
    const res = await fetch('/api/weather');
    if (!res.ok) throw new Error('WEATHER_UNAVAILABLE');
    return res.json() as Promise<WeatherAllResponse>;
}

/**
 * Loads all 60 days into React Query cache once.
 * staleTime matches the server-side 4-hour TTL.
 */
export function useWeather() {
    return useQuery<WeatherAllResponse>({
        queryKey: ['weather'],
        queryFn: fetchAllWeather,
        staleTime: 4 * 60 * 60 * 1000, // 4 hours
        placeholderData: (prev) => prev,
        retry: 2,
    });
}

/**
 * Returns the WeatherDay for a specific YYYY-MM-DD date.
 * Returns null while loading or if date is out of the 60-day window.
 */
export function useWeatherForDate(date: string | null): WeatherDay | null {
    const { data } = useWeather();
    if (!date || !data) return null;
    return data.days.find((d) => d.date === date) ?? null;
}

// ─── WMO helpers (shared with WeatherPreview) ─────────────────────────────

export function wmoLabel(code: number): string {
    if (code === 0)  return 'Ясно';
    if (code <= 3)   return 'Хмарно';
    if (code <= 48)  return 'Туман';
    if (code <= 57)  return 'Мряка';
    if (code <= 67)  return 'Дощ';
    if (code <= 77)  return 'Сніг';
    if (code <= 82)  return 'Зливи';
    if (code <= 86)  return 'Снігопад';
    if (code <= 99)  return 'Гроза';
    return 'Невідомо';
}

export function wmoIcon(code: number): string {
    if (code === 0)  return '☀️';
    if (code <= 3)   return '⛅';
    if (code <= 48)  return '🌫️';
    if (code <= 57)  return '🌦️';
    if (code <= 67)  return '🌧️';
    if (code <= 77)  return '❄️';
    if (code <= 82)  return '🌧️';
    if (code <= 86)  return '🌨️';
    return '⛈️';
}
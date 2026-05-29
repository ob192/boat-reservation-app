import { getWeatherForDate, getAllWeatherDays, getCacheFetchedAt } from './cache';
import type { WeatherDay } from './cache';

export type { WeatherDay };

import('./cache').then(({ warmWeatherCache }) => {
    warmWeatherCache().catch((err) =>
        console.error('[weather] initial cache warm failed:', err),
    );
});

/**
 * GET /api/weather?date=YYYY-MM-DD  → single day
 * GET /api/weather                  → all 7 days
 */
export async function GET(req: Request) {
    const { searchParams } = new URL(req.url);
    const date = searchParams.get('date');

    try {
        if (date) {
            if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) {
                return Response.json(
                    { error: 'INVALID_DATE', message: 'Use YYYY-MM-DD format' },
                    { status: 400 },
                );
            }

            const day = await getWeatherForDate(date);

            if (!day) {
                return Response.json(
                    { error: 'DATE_NOT_IN_RANGE', message: `No forecast available for ${date}` },
                    { status: 404 },
                );
            }

            const fetchedAt = await getCacheFetchedAt();
            return Response.json(
                { day, fetchedAt: new Date(fetchedAt).toISOString() },
                { headers: { 'Cache-Control': 'public, s-maxage=14400, stale-while-revalidate=3600' } },
            );
        }

        const days = await getAllWeatherDays();
        const fetchedAt = await getCacheFetchedAt();
        return Response.json(
            { days, fetchedAt: new Date(fetchedAt).toISOString() },
            { headers: { 'Cache-Control': 'public, s-maxage=14400, stale-while-revalidate=3600' } },
        );
    } catch (err) {
        console.error('[weather] route error:', err);
        return Response.json({ error: 'WEATHER_UNAVAILABLE' }, { status: 503 });
    }
}
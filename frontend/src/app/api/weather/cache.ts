const LAT = 51.4982;
const LNG = 31.2893;
const TIMEZONE = 'Europe%2FKyiv';

/**
 * Open-Meteo free tier hard limit: 16 days ahead.
 * Chunks are fetched sequentially to avoid 429 Too Many Concurrent Requests.
 */
const FORECAST_WINDOW_DAYS = 16;
const CHUNK_DAYS = 7;
const TTL_MS = 4 * 60 * 60 * 1000; // 4 hours

export interface WeatherDay {
    date: string;            // YYYY-MM-DD
    sunrise: string;         // HH:MM
    sunset: string;          // HH:MM
    airTempMax: number;
    airTempMin: number;
    rainSum: number;
    rainProbability: number;
    waterTemp: number | null;
    weatherCode: number;
}

interface CacheEntry {
    byDate: Map<string, WeatherDay>;
    fetchedAt: number;
}

let cache: CacheEntry | null = null;
let inflight: Promise<CacheEntry> | null = null;

// ─── Date helpers ──────────────────────────────────────────────────────────

function toDateString(d: Date): string {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
}

function addDays(base: Date, n: number): Date {
    const d = new Date(base);
    d.setDate(d.getDate() + n);
    return d;
}

/** "2026-06-01T04:43" → "04:43" */
function toHHMM(iso: string): string {
    const t = iso.split('T')[1];
    return t ? t.slice(0, 5) : iso;
}

function buildChunks(): Array<{ startDate: string; endDate: string }> {
    const start = new Date();
    start.setHours(0, 0, 0, 0);

    const windowEnd = addDays(start, FORECAST_WINDOW_DAYS);
    const chunks: Array<{ startDate: string; endDate: string }> = [];
    let cursor = new Date(start);

    while (cursor < windowEnd) {
        const chunkStart = new Date(cursor);
        const chunkEnd = addDays(cursor, CHUNK_DAYS - 1);
        chunks.push({
            startDate: toDateString(chunkStart),
            endDate:   toDateString(chunkEnd < windowEnd ? chunkEnd : addDays(windowEnd, -1)),
        });
        cursor = addDays(cursor, CHUNK_DAYS);
    }

    return chunks;
}

// ─── Single-chunk fetcher ──────────────────────────────────────────────────

async function fetchChunk(
    startDate: string,
    endDate: string,
): Promise<WeatherDay[] | null> {
    const forecastUrl =
        `https://api.open-meteo.com/v1/forecast` +
        `?latitude=${LAT}&longitude=${LNG}` +
        `&daily=temperature_2m_max,temperature_2m_min,precipitation_sum` +
        `,precipitation_probability_max,sunrise,sunset,weather_code` +
        `&timezone=${TIMEZONE}` +
        `&start_date=${startDate}&end_date=${endDate}`;

    const marineUrl =
        `https://marine-api.open-meteo.com/v1/marine` +
        `?latitude=${LAT}&longitude=${LNG}` +
        `&hourly=sea_surface_temperature` +
        `&timezone=${TIMEZONE}` +
        `&start_date=${startDate}&end_date=${endDate}`;

    const [forecastRes, marineRes] = await Promise.allSettled([
        fetch(forecastUrl),
        fetch(marineUrl),
    ]);

    if (forecastRes.status === 'rejected') {
        throw new Error(`Network error for chunk ${startDate}→${endDate}: ${forecastRes.reason}`);
    }

    if (!forecastRes.value.ok) {
        const body = await forecastRes.value.text();
        if (forecastRes.value.status === 400) {
            console.warn(`[weather] chunk ${startDate}→${endDate} out of range, skipping.`);
            return null;
        }
        throw new Error(
            `Open-Meteo chunk ${startDate}→${endDate} failed HTTP ${forecastRes.value.status}: ${body}`,
        );
    }

    const forecast = await forecastRes.value.json();

    const sstByDate: Record<string, number> = {};
    if (marineRes.status === 'fulfilled' && marineRes.value.ok) {
        const marine = await marineRes.value.json();
        const times: string[] = marine?.hourly?.time ?? [];
        const ssts: (number | null)[] = marine?.hourly?.sea_surface_temperature ?? [];
        times.forEach((t, i) => {
            const [datePart, hourPart] = t.split('T');
            if (hourPart?.startsWith('12') && ssts[i] != null) {
                sstByDate[datePart] = ssts[i] as number;
            }
        });
    }

    const {
        daily: {
            time,
            temperature_2m_max,
            temperature_2m_min,
            precipitation_sum,
            precipitation_probability_max,
            sunrise,
            sunset,
            weather_code,
        },
    } = forecast;

    return (time as string[]).map((date: string, i: number) => ({
        date,
        sunrise:         toHHMM(sunrise[i] as string),
        sunset:          toHHMM(sunset[i]  as string),
        airTempMax:      Math.round(temperature_2m_max[i] as number),
        airTempMin:      Math.round(temperature_2m_min[i] as number),
        rainSum:         Math.round(((precipitation_sum[i] as number) ?? 0) * 10) / 10,
        rainProbability: (precipitation_probability_max[i] as number) ?? 0,
        waterTemp:       sstByDate[date] != null ? Math.round(sstByDate[date]) : null,
        weatherCode:     (weather_code[i] as number) ?? 0,
    }));
}

// ─── Full fetch — sequential to respect rate limits ────────────────────────

async function fetchOpenMeteo(): Promise<WeatherDay[]> {
    const chunks = buildChunks();

    console.log(
        `[weather] fetching ${FORECAST_WINDOW_DAYS} days in ${chunks.length} chunks (sequential): ` +
        chunks.map((c) => `${c.startDate}→${c.endDate}`).join(', '),
    );

    const byDate = new Map<string, WeatherDay>();

    for (const chunk of chunks) {
        const days = await fetchChunk(chunk.startDate, chunk.endDate);
        if (!days) continue;
        for (const day of days) {
            byDate.set(day.date, day);
        }
    }

    const sorted = Array.from(byDate.values()).sort((a, b) => a.date.localeCompare(b.date));
    console.log(`[weather] collected ${sorted.length} days (${sorted[0]?.date} → ${sorted.at(-1)?.date})`);
    return sorted;
}

// ─── Cache management ──────────────────────────────────────────────────────

async function buildCache(): Promise<CacheEntry> {
    const days = await fetchOpenMeteo();
    const byDate = new Map<string, WeatherDay>(days.map((d) => [d.date, d]));
    return { byDate, fetchedAt: Date.now() };
}

async function getCache(): Promise<CacheEntry> {
    const now = Date.now();

    if (cache && now - cache.fetchedAt < TTL_MS) return cache;

    if (inflight) return inflight;

    inflight = buildCache()
        .then((entry) => {
            cache = entry;
            inflight = null;
            console.log(
                `[weather] cache refreshed at ${new Date().toISOString()} — ${entry.byDate.size} days cached`,
            );
            return entry;
        })
        .catch((err) => {
            inflight = null;
            if (cache) {
                console.error('[weather] refresh failed, serving stale cache:', err);
                return cache;
            }
            throw err;
        });

    return inflight;
}

// ─── Public API ────────────────────────────────────────────────────────────

export async function getWeatherForDate(date: string): Promise<WeatherDay | null> {
    const entry = await getCache();
    return entry.byDate.get(date) ?? null;
}

export async function getAllWeatherDays(): Promise<WeatherDay[]> {
    const entry = await getCache();
    return Array.from(entry.byDate.values()).sort((a, b) => a.date.localeCompare(b.date));
}

export async function getCacheFetchedAt(): Promise<number> {
    const entry = await getCache();
    return entry.fetchedAt;
}

export async function warmWeatherCache(): Promise<void> {
    await getCache();
}
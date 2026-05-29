'use client';

import { useWeatherForDate, useWeather, wmoLabel, wmoIcon } from '@/features/booking/hooks/useWeather';

// ─── Skeleton ──────────────────────────────────────────────────────────────

function WeatherSkeleton({ dateLabel }: { dateLabel: string }) {
    const shimmer: React.CSSProperties = {
        display: 'block',
        width: '65%',
        height: 20,
        borderRadius: 6,
        background:
            'linear-gradient(90deg, oklch(0.93 0.012 80) 25%, oklch(0.965 0.012 80) 50%, oklch(0.93 0.012 80) 75%)',
        backgroundSize: '200% 100%',
        animation: 'shimmer 1.4s ease-in-out infinite',
    };

    return (
        <section className="bk-weather" aria-label="Прогноз погоди" aria-busy="true">
            <div className="bk-weather-head">
                <span className="bk-eyebrow">Прогноз на день</span>
                <span className="bk-weather-date">{dateLabel}</span>
            </div>
            <div className="bk-weather-grid">
                {(['Умови', 'Повітря', 'Сонце'] as const).map((label) => (
                    <div className="bk-weather-col" key={label}>
                        <span className="bk-weather-label">{label}</span>
                        <span aria-hidden="true" style={shimmer} />
                    </div>
                ))}
            </div>
        </section>
    );
}

// ─── Unavailable note ──────────────────────────────────────────────────────

function WeatherUnavailable({ dateLabel }: { dateLabel: string }) {
    return (
        <section className="bk-weather" aria-label="Прогноз погоди">
            <div className="bk-weather-head">
                <span className="bk-eyebrow">Прогноз на день</span>
                <span className="bk-weather-date">{dateLabel}</span>
            </div>
            <p
                style={{
                    fontFamily: 'var(--mono)',
                    fontSize: 10,
                    letterSpacing: '0.1em',
                    textTransform: 'uppercase',
                    color: 'var(--ink-muted)',
                    marginTop: 10,
                }}
            >
                Прогноз доступний лише на найближчі 16 днів
            </p>
        </section>
    );
}

// ─── Main ──────────────────────────────────────────────────────────────────

interface WeatherPreviewProps {
    dateLabel: string;
    date: string; // YYYY-MM-DD
}

export function WeatherPreview({ dateLabel, date }: WeatherPreviewProps) {
    const { isLoading } = useWeather();
    const weather = useWeatherForDate(date);

    if (isLoading) {
        return <WeatherSkeleton dateLabel={dateLabel} />;
    }

    if (!weather) {
        return <WeatherUnavailable dateLabel={dateLabel} />;
    }

    return (
        <section className="bk-weather" aria-label="Прогноз погоди">
            <div className="bk-weather-head">
                <span className="bk-eyebrow">Прогноз на день</span>
                <span className="bk-weather-date">{dateLabel}</span>
            </div>

            <div className="bk-weather-grid">
                {/* Col 1 — Conditions */}
                <div className="bk-weather-col">
                    <span className="bk-weather-label">Умови</span>
                    <span className="bk-weather-val">
            <span className="bk-weather-ico" aria-hidden="true">
              {wmoIcon(weather.weatherCode)}
            </span>
                        {wmoLabel(weather.weatherCode)}
          </span>
                    {weather.rainProbability > 0 && (
                        <span className="bk-weather-sub">
              дощ {weather.rainProbability}%
                            {weather.rainSum > 0 ? ` · ${weather.rainSum} мм` : ''}
            </span>
                    )}
                </div>

                {/* Col 2 — Air temperature */}
                <div className="bk-weather-col">
                    <span className="bk-weather-label">Повітря</span>
                    <span className="bk-weather-val">{weather.airTempMax}°C</span>
                    <span className="bk-weather-sub">мін {weather.airTempMin}°C</span>
                </div>

                {/* Col 3 — Water temp or sunrise/sunset */}
                <div className="bk-weather-col">
          <span className="bk-weather-label">
            {weather.waterTemp != null ? 'Вода' : 'Сонце'}
          </span>
                    <span className="bk-weather-val">
            {weather.waterTemp != null
                ? `${weather.waterTemp}°C`
                : `↑ ${weather.sunrise}`}
          </span>
                    <span className="bk-weather-sub">
            {weather.waterTemp != null
                ? `↑ ${weather.sunrise} ↓ ${weather.sunset}`
                : `↓ ${weather.sunset}`}
          </span>
                </div>
            </div>
        </section>
    );
}
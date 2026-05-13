'use client';

import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useSlots } from '@/features/booking/hooks';
import { MESSAGES, MAX_BIG, MAX_MEDIUM, TIME_OPTIONS } from '@/features/booking/messages';

function TimeSlotSkeleton({ time }: { time: string }) {
    return (
        <div className="time-slot time-slot-skeleton" aria-hidden="true">
            <div className="skel-line" style={{ width: '55%', height: '1.5rem' }} />
            <div className="skel-line" style={{ width: '38%', height: '0.65rem' }} />
            <div className="skel-line" style={{ width: '100%', height: '3px', marginTop: '0.6rem' }} />
            <div className="skel-line" style={{ width: '72%', height: '0.65rem', marginTop: '0.35rem', marginBottom: 0 }} />
        </div>
    );
}

export function TimeSlots() {
    const { selectedDate, selectedTime, setTime } = useBookingStore();
    const { data: slotsData, isLoading, isError } = useSlots(selectedDate);

    if (!selectedDate) return null;

    // ── Loading state ─────────────────────────────────────────────────
    if (isLoading) {
        return (
            <div>
                <div className="availability-loading-banner" role="status" aria-live="polite">
                    <div className="mini-spinner" aria-hidden="true" />
                    Перевіряємо доступні часові слоти…
                </div>
                <div className="time-slots" aria-busy="true">
                    {TIME_OPTIONS.map((t) => (
                        <TimeSlotSkeleton key={t} time={t} />
                    ))}
                </div>
            </div>
        );
    }

    // ── Error state ───────────────────────────────────────────────────
    if (isError) {
        return (
            <div className="availability-error-banner" role="alert">
                ⚠️ Не вдалося завантажити слоти. Спробуйте оновити сторінку.
            </div>
        );
    }

    // ── Loaded ────────────────────────────────────────────────────────
    const slotsMap = new Map<
        string,
        { availableBig: number; availableMedium: number; totalBig: number; totalMedium: number; blocked: boolean }
    >();
    slotsData?.slots.forEach((s) =>
        slotsMap.set(s.time, {
            availableBig: s.availableBig,
            availableMedium: s.availableMedium,
            totalBig: s.totalBig,
            totalMedium: s.totalMedium,
            blocked: s.blocked,
        }),
    );

    const dateBlocked = slotsData?.dateBlocked ?? false;

    return (
        <div className="time-slots" role="group" aria-label="Доступні часові слоти">
            {TIME_OPTIONS.map((t) => {
                const slotInfo = slotsMap.get(t);
                const isBlocked = dateBlocked || (slotInfo?.blocked ?? false);

                const availableBig = slotInfo?.availableBig ?? MAX_BIG;
                const availableMedium = slotInfo?.availableMedium ?? MAX_MEDIUM;
                const totalBig = slotInfo?.totalBig ?? MAX_BIG;

                const isFull = !isBlocked && availableBig <= 0 && availableMedium <= 0;
                const isUnavailable = isBlocked || isFull;
                const isSel = selectedTime === t;

                const pctBig = totalBig > 0 ? ((totalBig - availableBig) / totalBig) * 100 : 0;

                const bigFillColor =
                    availableBig <= 0
                        ? 'var(--coral)'
                        : availableBig <= 4
                            ? 'var(--gold)'
                            : 'var(--teal)';

                return (
                    <div
                        key={t}
                        className={`time-slot ${isFull ? 'full' : ''} ${isBlocked ? 'slot-blocked' : ''} ${isSel ? 'selected' : ''}`}
                        onClick={!isUnavailable ? () => setTime(t) : undefined}
                        role="radio"
                        aria-checked={isSel}
                        aria-disabled={isUnavailable}
                        tabIndex={isUnavailable ? -1 : 0}
                        onKeyDown={(e) => {
                            if ((e.key === 'Enter' || e.key === ' ') && !isUnavailable) setTime(t);
                        }}
                    >
                        {isBlocked && <div className="full-tag blocked-tag">{MESSAGES.time.blockedTag}</div>}
                        {!isBlocked && isFull && <div className="full-tag">{MESSAGES.time.fullTag}</div>}

                        <div className="time">{t}</div>
                        <div className="period">{MESSAGES.time.periods[t] ?? ''}</div>

                        {!isBlocked && (
                            <div className="slots-bar" aria-hidden="true">
                                <div
                                    className="slots-fill"
                                    style={{
                                        width: `${pctBig}%`,
                                        background: isSel ? 'var(--seafoam)' : bigFillColor,
                                    }}
                                />
                            </div>
                        )}

                        <div className="slots-text">
                            {isBlocked ? (
                                <span>{MESSAGES.time.blockedTag}</span>
                            ) : (
                                <>
                                    <span>⛵ {MESSAGES.time.bigAvailable(availableBig)}</span>
                                    <span className="slots-sep">·</span>
                                    <span>🚤 {MESSAGES.time.mediumAvailable(availableMedium)}</span>
                                </>
                            )}
                        </div>
                    </div>
                );
            })}
        </div>
    );
}
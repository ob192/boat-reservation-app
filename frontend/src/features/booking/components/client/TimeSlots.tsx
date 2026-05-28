'use client';

import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useSlots } from '@/features/booking/hooks';
import { MESSAGES } from '@/features/booking/messages';

function TimeSlotSkeleton() {
    return (
        <div className="time-slot time-slot-skeleton" aria-hidden="true">
            <div className="skel-line" style={{ width: '3.5rem', height: '1.4rem' }} />
            <div style={{ flex: 1 }}>
                <div className="skel-line" style={{ width: '100%', height: '3px' }} />
                <div className="skel-line" style={{ width: '60%', height: '0.65rem', marginTop: '0.3rem', marginBottom: 0 }} />
            </div>
        </div>
    );
}

/** Returns true if the slot time on the given date is in the past or within 30 min from now */
function isSlotInPast(date: string, time: string): boolean {
    const [h, m] = time.split(':').map(Number);
    const slotDate = new Date(`${date}T00:00:00`);
    slotDate.setHours(h, m, 0, 0);
    const cutoff = new Date(Date.now() + 30 * 60 * 1000); // 30 min buffer
    return slotDate <= cutoff;
}

export function TimeSlots() {
    const { selectedDate, selectedTime, setTime } = useBookingStore();
    const { data: slotsData, isLoading, isError } = useSlots(selectedDate);

    if (!selectedDate) return null;

    if (isLoading) {
        return (
            <div>
                <div className="availability-loading-banner" role="status" aria-live="polite">
                    <div className="mini-spinner" aria-hidden="true" />
                    Перевіряємо доступні часові слоти…
                </div>
                <div className="time-slots" aria-busy="true">
                    {[0, 1, 2, 3].map((i) => <TimeSlotSkeleton key={i} />)}
                </div>
            </div>
        );
    }

    if (isError) {
        return (
            <div className="availability-error-banner" role="alert">
                ⚠️ Не вдалося завантажити слоти. Спробуйте оновити сторінку.
            </div>
        );
    }

    const slots = slotsData?.slots ?? [];
    const dateBlocked = slotsData?.dateBlocked ?? false;
    const fullyBlocked = slotsData?.fullyBlocked ?? false;

    // ── Fully blocked ────────────────────────────────────────────────
    if (fullyBlocked) {
        return (
            <div
                style={{
                    background: '#fefce8',
                    border: '1px solid #fde68a',
                    borderRadius: 12,
                    padding: '1.25rem 1rem',
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    gap: '0.5rem',
                    textAlign: 'center',
                }}
                role="alert"
            >
                <span style={{ fontSize: '1.75rem' }}>🚧</span>
                <div style={{ fontWeight: 600, fontSize: '0.9rem', color: '#92400e' }}>
                    Тимчасово недоступно
                </div>
                <div style={{ fontSize: '0.78rem', color: '#a16207', lineHeight: 1.5, maxWidth: 280 }}>
                    Бронювання на цей день тимчасово призупинено. Спробуйте вибрати інший день або зайдіть пізніше.
                </div>
            </div>
        );
    }

    if (slots.length === 0) {
        return (
            <p style={{ color: 'var(--subtle)', fontSize: '0.85rem', textAlign: 'center', padding: '1rem 0' }}>
                {MESSAGES.time.noSlots}
            </p>
        );
    }

    // Check if today — only then apply past-slot filtering
    const todayKey = (() => {
        const t = new Date();
        return `${t.getFullYear()}-${String(t.getMonth() + 1).padStart(2, '0')}-${String(t.getDate()).padStart(2, '0')}`;
    })();
    const isToday = selectedDate === todayKey;

    // Filter out slots that are in the past when viewing today
    const visibleSlots = isToday
        ? slots.filter((s) => !isSlotInPast(selectedDate, s.time))
        : slots;

    if (visibleSlots.length === 0) {
        return (
            <p style={{ color: 'var(--subtle)', fontSize: '0.85rem', textAlign: 'center', padding: '1rem 0' }}>
                На сьогодні більше немає доступних слотів.
            </p>
        );
    }

    return (
        <div className="time-slots" role="group" aria-label="Доступні часові слоти">
            {visibleSlots.map((s) => {
                const isBlocked = dateBlocked || s.blocked;
                const isFull = !isBlocked && s.availableBig <= 0 && s.availableMedium <= 0;
                const isUnavailable = isBlocked || isFull;
                const isSel = selectedTime === s.time;

                const pctBig = s.totalBig > 0 ? ((s.totalBig - s.availableBig) / s.totalBig) * 100 : 0;
                const bigFillColor =
                    s.availableBig <= 0 ? 'var(--coral)'
                        : s.availableBig <= 4 ? 'var(--gold)'
                            : 'var(--teal)';

                return (
                    <div
                        key={s.time}
                        className={`time-slot ${isFull ? 'full' : ''} ${isBlocked ? 'slot-blocked' : ''} ${isSel ? 'selected' : ''}`}
                        onClick={!isUnavailable ? () => setTime(s.time) : undefined}
                        role="radio"
                        aria-checked={isSel}
                        aria-disabled={isUnavailable}
                        tabIndex={isUnavailable ? -1 : 0}
                        onKeyDown={(e) => {
                            if ((e.key === 'Enter' || e.key === ' ') && !isUnavailable) setTime(s.time);
                        }}
                    >
                        {isBlocked && <div className="full-tag blocked-tag">{MESSAGES.time.blockedTag}</div>}
                        {!isBlocked && isFull && <div className="full-tag">{MESSAGES.time.fullTag}</div>}

                        <div className="time">{s.time}</div>

                        <div className="slot-right">
                            {!isBlocked && (
                                <div className="slots-bar" aria-hidden="true">
                                    <div
                                        className="slots-fill"
                                        style={{
                                            width: `${pctBig}%`,
                                            background: isSel ? 'rgba(255,255,255,0.4)' : bigFillColor,
                                        }}
                                    />
                                </div>
                            )}
                            <div className="slots-text">
                                {isBlocked ? (
                                    <span>{MESSAGES.time.blockedTag}</span>
                                ) : (
                                    <>
                                        <span>⛵ {MESSAGES.time.bigAvailable(s.availableBig)}</span>
                                        <span className="slots-sep">·</span>
                                        <span>🚤 {MESSAGES.time.mediumAvailable(s.availableMedium)}</span>
                                    </>
                                )}
                            </div>
                        </div>
                    </div>
                );
            })}
        </div>
    );
}
'use client';

import { useState, useMemo } from 'react';
import { Modal } from '@/components/Modal';
import { Calendar } from '@/components/Calendar';
import { RouteSelector } from '@/components/RouteSelector';
import { CapacityBar } from '@/components/CapacityBar';
import { useSlots } from '@/hooks/useSlots';
import { useMoveBooking } from '@/hooks/useMoveBooking';
import { toast } from '@/hooks/useToast';
import { ApiError, formatDate } from '@/lib/api';
import { Booking, SlotInfo } from '@/lib/types';
import { RouteName, ROUTES, routeLabel } from '@/lib/routes';

interface BookingMoveModalProps {
    open: boolean;
    onClose: () => void;
    booking: Booking;
    /** The slot the booking currently sits on (the source). */
    fromDate: string;
    fromTime: string;
    fromRoute: string;
}

// Maps a backend error `message` code to admin-facing Ukrainian copy.
function moveErrorText(e: unknown): string {
    if (!(e instanceof ApiError)) return 'Помилка сервера';
    switch (e.message) {
        case 'SLOT_TAKEN':          return 'Недостатньо місця в обраному слоті';
        case 'SLOT_NOT_FOUND':      return 'Обраний слот не існує';
        case 'SLOT_BLOCKED':        return 'Обраний слот заблоковано';
        case 'SLOT_CANCELLED':      return 'Обраний слот скасовано';
        case 'BOOKING_NOT_PENDING': return 'Це бронювання не можна перемістити';
        case 'BOOKING_NOT_FOUND':   return 'Бронювання не знайдено';
        case 'INVALID_DATE':        return 'Невірна дата';
        case 'INVALID_TIME':        return 'Невірний час';
        case 'INVALID_ROUTE':       return 'Невірний маршрут';
        case 'INVALID_INPUT':       return 'Невірні дані';
        case 'FORBIDDEN':           return 'Недостатньо прав';
        default:                    return 'Помилка сервера';
    }
}

function fmtNiceDate(date: string) {
    return new Date(date + 'T00:00:00').toLocaleDateString('uk-UA', {
        day: 'numeric', month: 'long', year: 'numeric',
    });
}

// ── Capacity preview: does the booking fit the chosen destination slot? ──
interface FitRow {
    label: string;
    need: number;
    available: number;
    total: number;
}

function buildFitRows(b: Booking, slot: SlotInfo, isSameSlot: boolean): FitRow[] {
    // Same-slot moves exclude the booking's own boats from the check (backend
    // does this too), so the booking always "fits" its current slot.
    const offset = isSameSlot
        ? { big: b.quantities.big, medium: b.quantities.medium, small: b.quantities.small ?? 0 }
        : { big: 0, medium: 0, small: 0 };

    return [
        { label: 'Великі',  need: b.quantities.big,           available: slot.availableBig + offset.big,       total: slot.totalBig },
        { label: 'Середні', need: b.quantities.medium,        available: slot.availableMedium + offset.medium, total: slot.totalMedium },
        { label: 'Малі',    need: b.quantities.small ?? 0,    available: slot.availableSmall + offset.small,   total: slot.totalSmall },
    ];
}

export function BookingMoveModal({ open, onClose, booking, fromDate, fromTime, fromRoute }: BookingMoveModalProps) {
    const today = formatDate(new Date());

    const [destRoute, setDestRoute] = useState<RouteName>(
        (ROUTES as readonly string[]).includes(fromRoute) ? (fromRoute as RouteName) : ROUTES[0]
    );
    const [destDate, setDestDate] = useState(fromDate || today);
    const [destTime, setDestTime] = useState<string>('');

    const { mutateAsync, isPending } = useMoveBooking(fromDate, fromTime, fromRoute);
    const { data: slotsData, isLoading: slotsLoading } = useSlots(destDate, destRoute);

    // Reset the chosen time whenever the date or route changes — it may not exist
    // on the new day, and we never want a stale selection submitted.
    const slots = slotsData?.slots ?? [];
    const selectedSlot = useMemo(
        () => slots.find(s => s.time === destTime) ?? null,
        [slots, destTime]
    );

    const isSameSlot =
        destDate === fromDate && destTime === fromTime && destRoute === fromRoute;

    const fitRows = selectedSlot ? buildFitRows(booking, selectedSlot, isSameSlot) : [];
    const fits = fitRows.every(r => r.need <= r.available);
    const slotUnavailable = !!selectedSlot && (selectedSlot.blocked || selectedSlot.cancelled);

    const canSubmit = !!selectedSlot && !slotUnavailable && (isSameSlot || fits) && !isPending;

    const onPickDate = (d: string) => {
        setDestDate(d);
        setDestTime('');
    };
    const onPickRoute = (r: RouteName) => {
        setDestRoute(r);
        setDestTime('');
    };

    const handleClose = () => {
        if (isPending) return;
        onClose();
    };

    const handleMove = async () => {
        if (!selectedSlot) return;
        try {
            const res = await mutateAsync({
                bookingId: booking.id,
                date: destDate,
                time: destTime,
                routeName: destRoute,
            });
            toast(
                isSameSlot ? 'Бронювання залишилось у тому ж слоті' : `Переміщено на ${res.time}`,
                'success'
            );
            onClose();
        } catch (e) {
            // Keep the picker open and tell the admin what to change.
            toast(moveErrorText(e), 'error');
        }
    };

    const boatLine =
        `В:${booking.quantities.big} · С:${booking.quantities.medium} · ` +
        `М:${booking.quantities.small ?? 0} · Д:${booking.quantities.child}`;

    return (
        <Modal
            open={open}
            onClose={handleClose}
            title="Перемістити бронювання"
            wide
            footer={
                <>
                    <button className="btn btn-secondary" onClick={handleClose} disabled={isPending}>
                        Відміна
                    </button>
                    <button className="btn btn-primary" onClick={handleMove} disabled={!canSubmit}>
                        {isPending ? 'Переміщення…' : isSameSlot ? 'Залишити тут' : 'Перемістити'}
                    </button>
                </>
            }
        >
            {/* ── From → To transfer header (the signature element) ───────── */}
            <div className="move-transfer">
                <div className="move-node">
                    <span className="move-node-label">Звідки</span>
                    <span className="move-node-route">{routeLabel(fromRoute)}</span>
                    <span className="move-node-when">{fmtNiceDate(fromDate)}</span>
                    <span className="move-node-time">{fromTime}</span>
                </div>

                <div className="move-arrow" aria-hidden>→</div>

                <div className={`move-node move-node-dest${selectedSlot && !slotUnavailable ? ' is-set' : ''}`}>
                    <span className="move-node-label">Куди</span>
                    {selectedSlot ? (
                        <>
                            <span className="move-node-route">{routeLabel(destRoute)}</span>
                            <span className="move-node-when">{fmtNiceDate(destDate)}</span>
                            <span className="move-node-time">{destTime}</span>
                        </>
                    ) : (
                        <span className="move-node-empty">Оберіть слот нижче</span>
                    )}
                </div>
            </div>

            <div className="move-booking-chip" title="Човни цього бронювання">
                <span className="booking-name">{booking.firstName} {booking.lastName}</span>
                <span className="text-subtle" style={{ fontSize: '0.78rem' }}>{boatLine}</span>
            </div>

            {/* ── Destination picker ──────────────────────────────────────── */}
            <div className="form-group">
                <label className="form-label">Маршрут призначення</label>
                <RouteSelector value={destRoute} onChange={onPickRoute} />
            </div>

            <div className="move-picker">
                <Calendar selected={destDate} onSelect={onPickDate} route={destRoute} allowPast={false} />

                <div className="move-slots">
                    <div className="form-label" style={{ marginBottom: 8 }}>
                        Слоти на {fmtNiceDate(destDate)}
                    </div>

                    {slotsLoading && (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                            {[1, 2, 3].map(i => <div key={i} className="skeleton" style={{ height: 38 }} />)}
                        </div>
                    )}

                    {!slotsLoading && slots.length === 0 && (
                        <div className="empty-state" style={{ padding: '24px 12px' }}>
                            Немає слотів на цю дату
                        </div>
                    )}

                    {!slotsLoading && slots.length > 0 && (
                        <div className="move-slot-list">
                            {slots.map(s => {
                                const unavailable = s.blocked || s.cancelled;
                                const isCurrent = s.time === fromTime && destDate === fromDate && destRoute === fromRoute;
                                const active = s.time === destTime;
                                return (
                                    <button
                                        key={s.time}
                                        type="button"
                                        className={`move-slot-pill${active ? ' active' : ''}${unavailable ? ' disabled' : ''}`}
                                        onClick={() => !unavailable && setDestTime(s.time)}
                                        disabled={unavailable}
                                        title={
                                            s.cancelled ? 'Слот скасовано'
                                                : s.blocked ? 'Слот заблоковано'
                                                    : undefined
                                        }
                                    >
                                        <span className="move-slot-time">{s.time}</span>
                                        {isCurrent && <span className="move-slot-tag">поточний</span>}
                                        {s.cancelled && <span className="move-slot-tag danger">скасовано</span>}
                                        {s.blocked && !s.cancelled && <span className="move-slot-tag danger">заблок.</span>}
                                    </button>
                                );
                            })}
                        </div>
                    )}
                </div>
            </div>

            {/* ── Capacity fit preview ────────────────────────────────────── */}
            {selectedSlot && !slotUnavailable && (
                <div className={`move-fit${isSameSlot ? '' : fits ? ' ok' : ' over'}`}>
                    <div className="move-fit-head">
                        {isSameSlot
                            ? 'Це поточний слот бронювання'
                            : fits
                                ? 'Поміститься в обраний слот'
                                : 'Не вистачає місця в обраному слоті'}
                    </div>
                    <div className="move-fit-rows">
                        {fitRows.map(r => {
                            const rowFits = r.need <= r.available;
                            return (
                                <div key={r.label} className="move-fit-row">
                                    <span className="move-fit-label">{r.label}</span>
                                    <span className={`move-fit-need${!rowFits ? ' bad' : ''}`}>
                    треба {r.need}
                  </span>
                                    <CapacityBar available={r.available} total={r.total} />
                                </div>
                            );
                        })}
                    </div>
                </div>
            )}

            {selectedSlot && slotUnavailable && (
                <div className="move-fit over">
                    <div className="move-fit-head">
                        {selectedSlot.cancelled ? 'Цей слот скасовано — переміщення неможливе'
                            : 'Цей слот заблоковано — переміщення неможливе'}
                    </div>
                </div>
            )}

            <p className="text-subtle" style={{ fontSize: '0.78rem', marginTop: 4 }}>
                Сума бронювання не перераховується при переміщенні.
            </p>
        </Modal>
    );
}
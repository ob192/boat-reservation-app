'use client';

import { useState } from 'react';
import { CapacityBar } from '@/components/CapacityBar';
import { ConfirmInline } from '@/components/ConfirmInline';
import { SlotEditModal } from './SlotEditModal';
import { BookingsDrawer } from './BookingsDrawer';
import { useCancelSlot, useUncancelSlot, useDeleteSlot } from '@/hooks/useSlots';
import { toast } from '@/hooks/useToast';
import { ApiError } from '@/lib/api';
import { SlotInfo } from '@/lib/types';

interface SlotCardProps {
    date: string;
    slot: SlotInfo;
}

type CardAction = 'cancel' | 'uncancel' | 'delete' | null;

export function SlotCard({ date, slot }: SlotCardProps) {
    const [editOpen, setEditOpen] = useState(false);
    const [bookingsOpen, setBookingsOpen] = useState(false);
    const [cancelAction, setCancelAction] = useState<CardAction>(null);
    const [cancelReason, setCancelReason] = useState('');

    const { mutateAsync: cancelSlot, isPending: cancelling } = useCancelSlot(date);
    const { mutateAsync: uncancelSlot, isPending: uncancelling } = useUncancelSlot(date);
    const { mutateAsync: deleteSlot, isPending: deleting } = useDeleteSlot(date);
    const isActionPending = cancelling || uncancelling || deleting;

    const handleCancel = async () => {
        try {
            const result = await cancelSlot({ time: slot.time, route: slot.routeName, reason: cancelReason || undefined });
            const bookingsMsg = result.cancelledBookings > 0
                ? `, ${result.cancelledBookings} бронювань скасовано`
                : '';
            toast(`Слот скасовано${bookingsMsg}`, 'info');
            setCancelAction(null);
            setCancelReason('');
        } catch (e) {
            if (e instanceof ApiError && e.status === 409) toast('Вже скасовано', 'error');
            else toast(e instanceof ApiError ? e.message : 'Помилка сервера', 'error');
        }
    };

    const handleUncancel = async () => {
        try {
            await uncancelSlot({ time: slot.time, route: slot.routeName });
            toast('Слот відновлено', 'success');
            setCancelAction(null);
        } catch (e) {
            toast(e instanceof ApiError ? e.message : 'Помилка сервера', 'error');
        }
    };

    const handleDelete = async () => {
        try {
            await deleteSlot({ time: slot.time, route: slot.routeName });
            toast('Слот видалено', 'success');
            setCancelAction(null);
        } catch (e) {
            // Slot has active bookings — route the admin to Cancel instead.
            if (e instanceof ApiError && e.status === 409) {
                if (slot.cancelled) {
                    // A cancelled slot can't be deleted because of preserved active
                    // records; nothing to cancel further, so just inform.
                    toast('Слот має активні бронювання — видалення неможливе', 'error');
                    setCancelAction(null);
                } else {
                    toast('Слот має активні бронювання. Скасуйте слот замість видалення.', 'error');
                    setCancelAction('cancel');
                }
            } else if (e instanceof ApiError && e.status === 404) {
                // Already gone — refresh will drop it from the list.
                toast('Слот не знайдено', 'error');
                setCancelAction(null);
            } else {
                toast(e instanceof ApiError ? e.message : 'Помилка сервера', 'error');
            }
        }
    };

    const isCancelled = slot.cancelled;
    const isBlocked = slot.blocked;

    return (
        <>
            <div className={`slot-card${isBlocked ? ' slot-blocked' : ''}${isCancelled ? ' slot-cancelled' : ''}`}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 12 }}>
                    <div className="slot-time" style={{ color: isCancelled ? 'var(--coral)' : undefined }}>
                        {slot.time}
                    </div>
                    <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', justifyContent: 'flex-end' }}>
                        {isCancelled && (
                            <span className="badge badge-cancelled-slot">🚫 Скасовано</span>
                        )}
                        {isBlocked && !isCancelled && (
                            <span className="badge badge-blocked">🔒 Заблоковано</span>
                        )}
                    </div>
                </div>

                {isCancelled && slot.cancelReason && (
                    <div className="text-subtle mt-4" style={{ fontSize: '0.8rem', marginBottom: 10 }}>
                        Причина: {slot.cancelReason}
                    </div>
                )}

                <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 14, opacity: isCancelled ? 0.45 : 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                        <span className="text-subtle" style={{ width: 80, flexShrink: 0, fontSize: '0.8rem' }}>Великі:</span>
                        <CapacityBar available={slot.availableBig} total={slot.totalBig} />
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                        <span className="text-subtle" style={{ width: 80, flexShrink: 0, fontSize: '0.8rem' }}>Середні:</span>
                        <CapacityBar available={slot.availableMedium} total={slot.totalMedium} />
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                        <span className="text-subtle" style={{ width: 80, flexShrink: 0, fontSize: '0.8rem' }}>Малі:</span>
                        <CapacityBar available={slot.availableSmall} total={slot.totalSmall} />
                    </div>
                </div>

                <div className="divider" style={{ margin: '0 0 12px' }} />

                {/* Actions row */}
                {cancelAction === null && (
                    isCancelled ? (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                            <button
                                className="btn btn-sm w-full"
                                style={{ borderColor: 'var(--teal)', color: 'var(--teal)', justifyContent: 'center' }}
                                onClick={() => setCancelAction('uncancel')}
                            >
                                Відновити слот
                            </button>
                            <div style={{ display: 'flex', gap: 8 }}>
                                <button
                                    className="btn btn-ghost btn-sm"
                                    style={{ flex: 1, justifyContent: 'center' }}
                                    onClick={() => setBookingsOpen(true)}
                                >
                                    📋 Бронювання
                                </button>
                                <button
                                    className="btn btn-ghost btn-sm"
                                    style={{ flex: 1, justifyContent: 'center', color: 'var(--coral)' }}
                                    onClick={() => setCancelAction('delete')}
                                >
                                    🗑 Видалити
                                </button>
                            </div>
                        </div>
                    ) : (
                        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
                            <button className="btn btn-secondary btn-sm" onClick={() => setEditOpen(true)}>
                                ✏️ Редагувати
                            </button>
                            <button className="btn btn-ghost btn-sm" onClick={() => setBookingsOpen(true)}>
                                📋 Бронювання
                            </button>
                            <button
                                className="btn btn-ghost btn-sm"
                                style={{ color: 'var(--coral)', marginLeft: 'auto' }}
                                onClick={() => setCancelAction('delete')}
                            >
                                🗑 Видалити
                            </button>
                            <button
                                className="btn btn-sm btn-danger"
                                onClick={() => setCancelAction('cancel')}
                            >
                                Скасувати слот
                            </button>
                        </div>
                    )
                )}

                {/* Delete confirmation */}
                {cancelAction === 'delete' && (
                    <ConfirmInline
                        onConfirm={handleDelete}
                        onCancel={() => setCancelAction(null)}
                        loading={isActionPending}
                        confirmLabel="Видалити слот"
                    >
                        <div style={{ padding: '2px 0 6px' }}>
                            <p className="text-subtle" style={{ fontSize: '0.85rem', lineHeight: 1.5 }}>
                                Слот буде видалено повністю. Цю дію можна виконати лише для слотів без активних бронювань.
                            </p>
                            <p style={{ fontSize: '0.8rem', color: 'var(--subtle)', marginTop: 6 }}>
                                Якщо в слоті є активні бронювання, скасуйте слот замість видалення.
                            </p>
                        </div>
                    </ConfirmInline>
                )}

                {/* Cancel confirmation */}
                {cancelAction === 'cancel' && (
                    <ConfirmInline
                        onConfirm={handleCancel}
                        onCancel={() => { setCancelAction(null); setCancelReason(''); }}
                        loading={isActionPending}
                        confirmLabel="Скасувати слот"
                    >
                        <div style={{ marginBottom: 10 }}>
                            <p className="text-subtle" style={{ fontSize: '0.85rem', lineHeight: 1.5 }}>
                                Усі активні бронювання цього слоту буде скасовано автоматично.
                            </p>
                        </div>
                        <div className="form-group">
                            <label className="form-label">Причина (необов&apos;язково)</label>
                            <input
                                type="text"
                                className="form-input"
                                placeholder="Технічне обслуговування…"
                                value={cancelReason}
                                onChange={e => setCancelReason(e.target.value)}
                                autoFocus
                            />
                        </div>
                    </ConfirmInline>
                )}

                {/* Uncancel confirmation */}
                {cancelAction === 'uncancel' && (
                    <ConfirmInline
                        onConfirm={handleUncancel}
                        onCancel={() => setCancelAction(null)}
                        loading={isActionPending}
                        confirmLabel="Відновити слот"
                        danger={false}
                    >
                        <div style={{ padding: '2px 0 6px' }}>
                            <p className="text-subtle" style={{ fontSize: '0.85rem', lineHeight: 1.5 }}>
                                Слот буде відкрито для нових бронювань.
                            </p>
                            <p style={{
                                fontSize: '0.8rem',
                                color: 'var(--coral)',
                                marginTop: 6,
                                fontWeight: 500,
                            }}>
                                ⚠️ Раніше скасовані бронювання не відновлюються.
                            </p>
                        </div>
                    </ConfirmInline>
                )}
            </div>

            {!isCancelled && (
                <SlotEditModal open={editOpen} onClose={() => setEditOpen(false)} date={date} slot={slot} />
            )}
            <BookingsDrawer
                open={bookingsOpen}
                onClose={() => setBookingsOpen(false)}
                date={date}
                time={slot.time}
                route={slot.routeName}
                slotCancelled={slot.cancelled}
            />
        </>
    );
}
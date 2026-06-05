'use client';

import { useState } from 'react';
import { CapacityBar } from '@/components/CapacityBar';
import { ConfirmInline } from '@/components/ConfirmInline';
import { SlotEditModal } from './SlotEditModal';
import { BookingsDrawer } from './BookingsDrawer';
import { useCancelSlot, useUncancelSlot } from '@/hooks/useSlots';
import { toast } from '@/hooks/useToast';
import { ApiError } from '@/lib/api';
import { SlotInfo } from '@/lib/types';

interface SlotCardProps {
    date: string;
    slot: SlotInfo;
}

export function SlotCard({ date, slot }: SlotCardProps) {
    const [editOpen, setEditOpen] = useState(false);
    const [bookingsOpen, setBookingsOpen] = useState(false);
    const [cancelAction, setCancelAction] = useState<'cancel' | 'uncancel' | null>(null);
    const [cancelReason, setCancelReason] = useState('');

    const { mutateAsync: cancelSlot, isPending: cancelling } = useCancelSlot(date);
    const { mutateAsync: uncancelSlot, isPending: uncancelling } = useUncancelSlot(date);
    const isActionPending = cancelling || uncancelling;

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

    const isCancelled = slot.cancelled;
    const isBlocked = slot.blocked;
    const isUnavailable = isCancelled || isBlocked;

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
                    <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                        {!isCancelled && (
                            <button className="btn btn-secondary btn-sm" onClick={() => setEditOpen(true)}>
                                ✏️ Редагувати
                            </button>
                        )}
                        <button className="btn btn-ghost btn-sm" onClick={() => setBookingsOpen(true)}>
                            📋 Бронювання
                        </button>
                        {isCancelled ? (
                            <button
                                className="btn btn-sm"
                                style={{ borderColor: 'var(--teal)', color: 'var(--teal)', marginLeft: 'auto' }}
                                onClick={() => setCancelAction('uncancel')}
                            >
                                Відновити слот
                            </button>
                        ) : (
                            <button
                                className="btn btn-sm btn-danger"
                                style={{ marginLeft: 'auto' }}
                                onClick={() => setCancelAction('cancel')}
                            >
                                Скасувати слот
                            </button>
                        )}
                    </div>
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
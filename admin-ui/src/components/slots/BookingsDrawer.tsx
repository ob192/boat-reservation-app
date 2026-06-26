'use client';

import { useState } from 'react';
import { Drawer } from '@/components/Drawer';
import { StatusBadge } from '@/components/StatusBadge';
import { ConfirmInline } from '@/components/ConfirmInline';
import { BookingMoveModal } from './BookingMoveModal';
import { useSlotBookings, useCancelBooking } from '@/hooks/useSlotBookings';
import { toast } from '@/hooks/useToast';
import { Booking } from '@/lib/types';
import { ApiError } from '@/lib/api';
import { routeLabel } from '@/lib/routes';
import { PosterCopy, CopyText } from '@/components/PosterCopy';

interface BookingsDrawerProps {
  open: boolean;
  onClose: () => void;
  date: string;
  time: string;
  route: string;
  slotCancelled?: boolean;
}

// ── Copy-all button ───────────────────────────────────────────────
function CopyAllButton({ bookings, date, time, route }: {
  bookings: Booking[];
  date: string;
  time: string;
  route: string;
}) {
  const [copied, setCopied] = useState(false);

  const active = bookings.filter(b => b.status !== 'cancelled' && b.status !== 'expired' && b.status !== 'failed');

  const buildText = () => {
    const header = `${date} · ${time} · ${routeLabel(route)}`;
    const divider = '─'.repeat(40);
    const lines = active.map((b, i) => {
      const name   = `${b.firstName} ${b.lastName}`;
      const email  = b.userEmail;
      const phone  = b.phone ?? '—';
      const boats  = `В:${b.quantities.big}  С:${b.quantities.medium}  М:${b.quantities.small ?? 0}  Д:${b.quantities.child}`;
      const amount = `${b.effectiveAmount.toFixed(2)} ₴`;
      return [
        `${i + 1}. ${name}`,
        `   📧 ${email}`,
        `   📞 ${phone}`,
        `   🚣 ${boats}`,
        `   💰 ${amount}`,
      ].join('\n');
    });

    return [header, divider, ...lines, divider, `Всього: ${active.length}`].join('\n');
  };

  const handleCopy = async () => {
    const text = buildText();
    try {
      await navigator.clipboard.writeText(text);
    } catch {
      // clipboard.js–style fallback for restricted environments
      const el = document.createElement('textarea');
      el.value = text;
      el.style.cssText = 'position:fixed;top:-9999px;left:-9999px;opacity:0';
      document.body.appendChild(el);
      el.focus();
      el.select();
      document.execCommand('copy');
      document.body.removeChild(el);
    }
    setCopied(true);
    toast('Список клієнтів скопійовано', 'success');
    setTimeout(() => setCopied(false), 2000);
  };

  if (active.length === 0) return null;

  return (
      <button
          onClick={handleCopy}
          title="Скопіювати список клієнтів"
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 6,
            padding: '5px 12px',
            borderRadius: 99,
            border: `1.5px solid ${copied ? 'rgba(14,124,123,0.45)' : 'rgba(168,213,209,0.5)'}`,
            background: copied ? 'rgba(14,124,123,0.10)' : 'rgba(168,213,209,0.10)',
            color: copied ? 'var(--teal)' : 'var(--subtle)',
            fontSize: '0.75rem',
            fontWeight: 600,
            cursor: 'pointer',
            transition: 'all 180ms ease',
            fontFamily: 'DM Sans, sans-serif',
            whiteSpace: 'nowrap',
            flexShrink: 0,
          }}
      >
        {copied ? '✓' : '⎘'} {copied ? 'Скопійовано' : `Копіювати (${active.length})`}
      </button>
  );
}

// ── Booking row ───────────────────────────────────────────────────
function BookingRow({ booking, date, time, route, slotCancelled }: {
  booking: Booking;
  date: string;
  time: string;
  route: string;
  slotCancelled?: boolean;
}) {
  const [confirming, setConfirming] = useState(false);
  const [moveOpen, setMoveOpen] = useState(false);
  const [reason, setReason] = useState('');
  const { mutateAsync, isPending } = useCancelBooking(date, time, route);

  // Only pending/confirmed bookings can be cancelled or moved.
  const isMovable = booking.status === 'pending' || booking.status === 'confirmed';
  const canCancel = !slotCancelled && isMovable;
  const canMove = isMovable;

  const handleCancel = async () => {
    if (!reason.trim()) return;
    try {
      await mutateAsync({ bookingId: booking.id, reason });
      toast('Бронювання скасовано', 'success');
      setConfirming(false);
    } catch (e) {
      toast(e instanceof ApiError ? e.message : 'Помилка сервера', 'error');
    }
  };

  return (
      <div className="booking-row">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 8 }}>
          <div>
            <div className="booking-name">{booking.firstName} {booking.lastName}</div>
            <div className="booking-contact">
              {booking.userEmail}{booking.phone && <><span style={{ color: 'var(--mist)', margin: '0 4px' }}>·</span><CopyText value={booking.phone} /></>}
            </div>
            <div className="booking-quantities mt-4">
              Великих: {booking.quantities.big} · Середніх: {booking.quantities.medium} · Малих: {booking.quantities.small ?? 0} · Дітей: {booking.quantities.child}
            </div>
            <PosterCopy
                orderId={booking.posterIncomingOrderId}
                transactionId={booking.posterIncomingTransactionId}
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 6 }}>
            <StatusBadge status={booking.status} />
            <span className="booking-amount">{booking.effectiveAmount.toFixed(2)} ₴</span>
          </div>
        </div>

        {!confirming && (canMove || canCancel) && (
            <div style={{ display: 'flex', gap: 8, marginTop: 8, flexWrap: 'wrap' }}>
              {canMove && (
                  <button className="btn btn-secondary btn-sm" onClick={() => setMoveOpen(true)}>
                    ⇄ Перемістити
                  </button>
              )}
              {canCancel && (
                  <button className="btn btn-danger btn-sm" onClick={() => setConfirming(true)}>
                    Скасувати
                  </button>
              )}
            </div>
        )}

        {confirming && (
            <ConfirmInline
                onConfirm={handleCancel}
                onCancel={() => { setConfirming(false); setReason(''); }}
                loading={isPending}
                confirmLabel="Підтвердити скасування"
            >
              <div className="form-group">
                <label className="form-label">Причина скасування</label>
                <input
                    type="text"
                    className="form-input"
                    placeholder="Вкажіть причину…"
                    value={reason}
                    onChange={e => setReason(e.target.value)}
                    autoFocus
                />
              </div>
            </ConfirmInline>
        )}

        <BookingMoveModal
            open={moveOpen}
            onClose={() => setMoveOpen(false)}
            booking={booking}
            fromDate={date}
            fromTime={time}
            fromRoute={route}
        />
      </div>
  );
}

// ── Drawer ────────────────────────────────────────────────────────
export function BookingsDrawer({ open, onClose, date, time, route, slotCancelled }: BookingsDrawerProps) {
  const { data, isLoading, isError } = useSlotBookings(date, time, route, open);

  const formattedDate = date
      ? new Date(date + 'T00:00:00').toLocaleDateString('uk-UA', { day: 'numeric', month: 'long', year: 'numeric' })
      : '';

  const totalCount  = data?.bookings.length ?? 0;
  const cancelledCount = data?.bookings.filter(b => b.status === 'cancelled').length ?? 0;

  const subtitleParts: string[] = [];
  if (totalCount > 0)      subtitleParts.push(`${totalCount} бронювань`);
  if (cancelledCount > 0)  subtitleParts.push(`${cancelledCount} скасовано`);

  return (
      <Drawer
          open={open}
          onClose={onClose}
          title={`${formattedDate} · ${time} · ${routeLabel(route)}`}
          subtitle={subtitleParts.join(' · ') || undefined}
      >
        {/* Cancelled-slot banner */}
        {slotCancelled && (
            <div style={{
              background: 'rgba(224,90,78,0.08)',
              border: '1px solid rgba(224,90,78,0.25)',
              borderRadius: 'var(--radius-sm)',
              padding: '10px 14px',
              marginBottom: 16,
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              fontSize: '0.82rem',
              color: 'var(--coral)',
              fontWeight: 500,
            }}>
              <span>🚫</span>
              <span>Цей слот скасовано — нові бронювання неможливі</span>
            </div>
        )}

        {/* Copy-all toolbar — shown only when data is ready and non-empty */}
        {data && data.bookings.length > 0 && (
            <div style={{
              display: 'flex',
              justifyContent: 'flex-end',
              marginBottom: 12,
            }}>
              <CopyAllButton bookings={data.bookings} date={date} time={time} route={route} />
            </div>
        )}

        {isLoading && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              {[1, 2, 3].map(i => (
                  <div key={i} style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                    <div className="skeleton" style={{ height: 18, width: '60%' }} />
                    <div className="skeleton" style={{ height: 14, width: '80%' }} />
                    <div className="skeleton" style={{ height: 14, width: '40%' }} />
                  </div>
              ))}
            </div>
        )}

        {isError && (
            <div className="empty-state">
              <div className="empty-state-icon">⚠</div>
              <div>Не вдалося завантажити бронювання</div>
            </div>
        )}

        {data && data.bookings.length === 0 && (
            <div className="empty-state">
              <div className="empty-state-icon">📋</div>
              <div>Немає бронювань для цього слоту</div>
            </div>
        )}

        {data?.bookings.map(b => (
            <BookingRow
                key={b.id}
                booking={b}
                date={date}
                time={time}
                route={route}
                slotCancelled={slotCancelled}
            />
        ))}
      </Drawer>
  );
}
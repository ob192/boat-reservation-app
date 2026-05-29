'use client';

import { useState } from 'react';
import { Drawer } from '@/components/Drawer';
import { StatusBadge } from '@/components/StatusBadge';
import { ConfirmInline } from '@/components/ConfirmInline';
import { useSlotBookings, useCancelBooking } from '@/hooks/useSlotBookings';
import { toast } from '@/hooks/useToast';
import { Booking } from '@/lib/types';
import { ApiError } from '@/lib/api';

interface BookingsDrawerProps {
  open: boolean;
  onClose: () => void;
  date: string;
  time: string;
}

function BookingRow({ booking, date, time }: { booking: Booking; date: string; time: string }) {
  const [confirming, setConfirming] = useState(false);
  const [reason, setReason] = useState('');
  const { mutateAsync, isPending } = useCancelBooking(date, time);

  const canCancel = booking.status === 'pending' || booking.status === 'confirmed';

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
            {booking.userEmail}{booking.phone ? ` · ${booking.phone}` : ''}
          </div>
          <div className="booking-quantities mt-4">
            Великих: {booking.quantities.big} · Середніх: {booking.quantities.medium} · Дітей: {booking.quantities.child}
          </div>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 6 }}>
          <StatusBadge status={booking.status} />
          <span className="booking-amount">{booking.effectiveAmount.toFixed(2)} ₴</span>
        </div>
      </div>

      {canCancel && !confirming && (
        <button
          className="btn btn-danger btn-sm mt-8"
          onClick={() => setConfirming(true)}
        >
          Скасувати
        </button>
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
    </div>
  );
}

export function BookingsDrawer({ open, onClose, date, time }: BookingsDrawerProps) {
  const { data, isLoading, isError } = useSlotBookings(date, time, open);

  const formattedDate = date
    ? new Date(date + 'T00:00:00').toLocaleDateString('uk-UA', { day: 'numeric', month: 'long', year: 'numeric' })
    : '';

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={`${formattedDate} · ${time}`}
      subtitle={data ? `${data.bookings.length} бронювань` : undefined}
    >
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
        <BookingRow key={b.id} booking={b} date={date} time={time} />
      ))}
    </Drawer>
  );
}

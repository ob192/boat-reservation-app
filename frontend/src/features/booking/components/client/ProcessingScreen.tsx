'use client';

import { useRouter } from 'next/navigation';
import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useBookingStatus } from '@/features/booking/hooks';
import { MESSAGES, PRICES } from '@/features/booking/messages';
import { formatCurrency } from '@/shared/lib/currency';

export function ProcessingScreen() {
  return (
    <div className="processing-screen">
      <div className="processing-spinner" aria-hidden="true" />
      <h3
        style={{
          fontFamily: 'var(--font-playfair)',
          fontSize: '1.4rem',
          color: 'var(--navy)',
          marginBottom: '0.5rem',
        }}
      >
        {MESSAGES.processing.title}
      </h3>
      <p style={{ color: 'var(--subtle)', fontSize: '0.85rem' }}>{MESSAGES.processing.subtitle}</p>
    </div>
  );
}

interface ConfirmationDisplayProps {
  booking: NonNullable<ReturnType<typeof useBookingStatus>['data']>['booking'];
}

function ConfirmationDisplay({ booking }: ConfirmationDisplayProps) {
  const { reset } = useBookingStore();
  const router = useRouter();

  if (!booking) return null;

  const d = new Date(booking.date);
  const dateStr = d.toLocaleDateString('uk-UA', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });

  const period = MESSAGES.time.periods[booking.time] ?? '';
  const total =
    booking.quantities.big * PRICES.big +
    booking.quantities.medium * PRICES.medium +
    booking.quantities.child * PRICES.child;

  const rows = [
    { label: MESSAGES.success.dateLabel, val: dateStr },
    { label: MESSAGES.success.departureLabel, val: `${booking.time} · ${period}` },
    ...(booking.quantities.big > 0
      ? [
          {
            label: MESSAGES.success.bigBoatsLabel,
            val: `${booking.quantities.big} × ${formatCurrency(PRICES.big)} = ${formatCurrency(booking.quantities.big * PRICES.big)}`,
          },
        ]
      : []),
    ...(booking.quantities.medium > 0
      ? [
          {
            label: MESSAGES.success.mediumBoatsLabel,
            val: `${booking.quantities.medium} × ${formatCurrency(PRICES.medium)} = ${formatCurrency(booking.quantities.medium * PRICES.medium)}`,
          },
        ]
      : []),
    ...(booking.quantities.child > 0
      ? [
          {
            label: MESSAGES.success.childrenLabel,
            val: `${booking.quantities.child} × ${formatCurrency(PRICES.child)} = ${formatCurrency(booking.quantities.child * PRICES.child)}`,
          },
        ]
      : []),
  ];

  const handleNewBooking = () => {
    reset();
    router.replace('/book/date');
  };

  return (
    <div className="confirm-screen">
      <div className="confirm-icon">🎉</div>
      <h3>{MESSAGES.success.title}</h3>
      <p>{MESSAGES.success.message}</p>

      <div className="confirm-details">
        {rows.map((r) => (
          <div key={r.label} className="confirm-row">
            <span className="cr-label">{r.label}</span>
            <span className="cr-val">{r.val}</span>
          </div>
        ))}
        <div className="confirm-row total">
          <span className="cr-label">{MESSAGES.success.totalLabel}</span>
          <span className="cr-val">{formatCurrency(total)}</span>
        </div>
      </div>

      <button
        className="btn-primary"
        onClick={handleNewBooking}
        type="button"
        style={{ margin: '0 auto', flex: '0 0 auto' }}
      >
        {MESSAGES.buttons.newBooking}
      </button>
    </div>
  );
}

export function SuccessPoller({ sessionId }: { sessionId: string }) {
  const { data, isLoading } = useBookingStatus(sessionId);

  if (isLoading) return <ProcessingScreen />;

  if (data?.status === 'confirmed') {
    return <ConfirmationDisplay booking={data.booking} />;
  }

  if (data?.status === 'failed' || data?.status === 'expired') {
    return (
      <div className="confirm-screen">
        <div className="confirm-icon" style={{ background: '#fff0ed' }}>
          ❌
        </div>
        <h3 style={{ fontFamily: 'var(--font-playfair)', fontSize: '1.6rem', color: 'var(--navy)', marginBottom: '0.5rem' }}>
          {data.status === 'expired' ? MESSAGES.errors.bookingExpired : MESSAGES.errors.paymentFailed}
        </h3>
        <p style={{ color: 'var(--subtle)', fontSize: '0.85rem' }}>
          Спробуйте ще раз або оберіть інший слот.
        </p>
        <a href="/book/date" className="btn-primary" style={{ marginTop: '1.5rem', textDecoration: 'none' }}>
          Спробувати знову
        </a>
      </div>
    );
  }

  return <ProcessingScreen />;
}

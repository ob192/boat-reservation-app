'use client';

import { useRouter } from 'next/navigation';

export default function CancelledPage() {
  const router = useRouter();

  return (
    <div className="confirm-screen">
      <div className="confirm-icon" style={{ background: '#fff0ed' }}>
        ↩️
      </div>
      <h3 style={{ fontFamily: 'var(--font-playfair)', fontSize: '1.6rem', color: 'var(--navy)', marginBottom: '0.5rem' }}>
        Оплата скасована
      </h3>
      <p style={{ color: 'var(--subtle)', fontSize: '0.85rem', maxWidth: 360, margin: '0 auto 1.75rem' }}>
        Ви скасували оплату. Ваше бронювання не завершено. Ви можете спробувати ще раз.
      </p>
      <button
        className="btn-primary"
        onClick={() => router.push('/book/route')}
        type="button"
        style={{ margin: '0 auto', flex: '0 0 auto' }}
      >
        Спробувати знову
      </button>
    </div>
  );
}

'use client';

import { useState } from 'react';
import { AdminGuard } from '@/components/AdminGuard';
import { Header } from '@/components/Header';
import { Modal } from '@/components/Modal';
import { useSystem, useToggleBookings } from '@/hooks/useSystem';
import { toast } from '@/hooks/useToast';
import { ApiError } from '@/lib/api';

function KillSwitchCard() {
  const { data, isLoading, isError } = useSystem();
  const { mutateAsync, isPending } = useToggleBookings();
  const [modalOpen, setModalOpen] = useState(false);
  const [reason, setReason] = useState('');

  const isEnabled = data?.bookingsEnabled ?? true;

  const handleConfirm = async () => {
    try {
      await mutateAsync({ enabled: !isEnabled, reason: reason || undefined });
      toast(
        isEnabled ? 'Бронювання вимкнено' : 'Бронювання увімкнено',
        isEnabled ? 'info' : 'success'
      );
      setModalOpen(false);
      setReason('');
    } catch (e) {
      toast(e instanceof ApiError ? e.message : 'Помилка сервера', 'error');
    }
  };

  if (isLoading) {
    return (
      <div className="card" style={{ maxWidth: 600, padding: 36 }}>
        <div className="skeleton" style={{ height: 32, width: '50%', marginBottom: 20 }} />
        <div className="skeleton" style={{ height: 52, width: '100%', marginBottom: 16 }} />
        <div className="skeleton" style={{ height: 44 }} />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="card" style={{ maxWidth: 600 }}>
        <div className="empty-state">
          <div className="empty-state-icon">⚠</div>
          <div>Не вдалося завантажити стан системи</div>
        </div>
      </div>
    );
  }

  return (
    <>
      <div className={`card system-card ${isEnabled ? 'enabled' : 'disabled'}`}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 28 }}>
          <div>
            <h2 style={{ fontFamily: 'Playfair Display, serif', fontSize: '1.3rem', marginBottom: 6 }}>
              Приймати бронювання
            </h2>
            <p className="text-subtle">
              Глобальне перемикання. При вимкненні жодне нове бронювання неможливе.
            </p>
          </div>
          <label className="toggle-wrap" onClick={() => setModalOpen(true)} style={{ cursor: 'pointer' }}>
            <div className="toggle">
              <input type="checkbox" checked={isEnabled} readOnly />
              <div className="toggle-track" />
              <div className="toggle-thumb" />
            </div>
          </label>
        </div>

        <div style={{
          background: isEnabled ? 'rgba(14,124,123,0.08)' : 'rgba(224,90,78,0.08)',
          borderRadius: 'var(--radius-sm)',
          padding: '16px 20px',
          display: 'flex',
          alignItems: 'center',
          gap: 14,
          marginBottom: 24,
        }}>
          <span style={{ fontSize: '1.5rem' }}>{isEnabled ? '✅' : '⛔'}</span>
          <div>
            <div className={`system-status-label ${isEnabled ? 'enabled' : 'disabled'}`}>
              {isEnabled ? 'УВІМКНЕНО' : 'ВИМКНЕНО'}
            </div>
            <div className="text-subtle mt-4">
              {isEnabled
                ? 'Клієнти можуть створювати бронювання'
                : 'Бронювання заборонені для всіх користувачів'}
            </div>
          </div>
        </div>

        <button
          className={`btn w-full ${isEnabled ? 'btn-danger' : 'btn-primary'}`}
          style={{ justifyContent: 'center', fontSize: '0.95rem' }}
          onClick={() => setModalOpen(true)}
          disabled={isPending}
        >
          {isEnabled ? '⛔ Вимкнути бронювання' : '✅ Увімкнути бронювання'}
        </button>
      </div>

      <Modal
        open={modalOpen}
        onClose={() => { setModalOpen(false); setReason(''); }}
        title={isEnabled ? 'Вимкнути бронювання?' : 'Увімкнути бронювання?'}
        footer={
          <>
            <button
              className="btn btn-secondary"
              onClick={() => { setModalOpen(false); setReason(''); }}
              disabled={isPending}
            >
              Відміна
            </button>
            <button
              className={`btn ${isEnabled ? 'btn-danger-solid' : 'btn-primary'}`}
              onClick={handleConfirm}
              disabled={isPending}
            >
              {isPending ? '…' : 'Підтвердити'}
            </button>
          </>
        }
      >
        <p style={{ color: 'var(--navy)', lineHeight: 1.6 }}>
          {isEnabled
            ? 'Всі нові бронювання будуть заборонені одразу після підтвердження. Існуючі бронювання залишаться незмінними.'
            : 'Клієнти зможуть знову створювати бронювання одразу після підтвердження.'}
        </p>
        <div className="form-group">
          <label className="form-label">Причина (необов&apos;язково)</label>
          <input
            type="text"
            className="form-input"
            placeholder="Технічне обслуговування, свято…"
            value={reason}
            onChange={e => setReason(e.target.value)}
            autoFocus
          />
        </div>
      </Modal>
    </>
  );
}

export default function SystemPage() {
  return (
    <AdminGuard>
      <Header />
      <main className="admin-main">
        <div className="page-container">
          <h1 className="page-title">Система</h1>
          <p className="page-subtitle">Глобальні налаштування платформи</p>

          <KillSwitchCard />

          <div className="card" style={{ maxWidth: 600, marginTop: 20, padding: 24 }}>
            <h3 style={{ fontFamily: 'Playfair Display, serif', fontSize: '1rem', marginBottom: 12 }}>
              ℹ Про kill-switch
            </h3>
            <ul style={{ color: 'var(--subtle)', fontSize: '0.875rem', lineHeight: 1.8, paddingLeft: 18 }}>
              <li>Зміна набуває чинності негайно — без кешування.</li>
              <li>При вимкненні <code>POST /bookings</code> повертає <code>503 BOOKINGS_DISABLED</code>.</li>
              <li>Існуючі та вже підтверджені бронювання залишаються незмінними.</li>
              <li>Стан синхронізується з полем <code>bookingsEnabled</code> у відповіді <code>GET /slots/:date</code>.</li>
            </ul>
          </div>
        </div>
      </main>
    </AdminGuard>
  );
}

'use client';

import { useState, useEffect, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { AdminGuard } from '@/components/AdminGuard';
import { Header } from '@/components/Header';
import { Calendar } from '@/components/Calendar';
import { ConfirmInline } from '@/components/ConfirmInline';
import { Modal } from '@/components/Modal';
import { useSlots, useBlockSlot, useUnblockSlot } from '@/hooks/useSlots';
import { toast } from '@/hooks/useToast';
import { adminFetch, formatDate, formatMonth } from '@/lib/api';
import { ApiError } from '@/lib/api';
import { DayAvailability } from '@/lib/types';
import { routeLabel } from '@/lib/routes';
import { RouteSelector } from '@/components/RouteSelector';
import { RouteName, ROUTES } from '@/lib/routes';

// ── Tab A: Slot Blocking ─────────────────────────────────────────
interface SlotBlockRowProps {
  date: string;
  time: string;
  route: string;            // NEW
  blocked: boolean;
  blockReason?: string;
}

function SlotBlockRow({ date, time, route, blocked, blockReason }: SlotBlockRowProps) {
  const [action, setAction] = useState<'block' | 'unblock' | null>(null);
  const [reason, setReason] = useState('');
  const { mutateAsync: blockSlot, isPending: blocking } = useBlockSlot(date);
  const { mutateAsync: unblockSlot, isPending: unblocking } = useUnblockSlot(date);
  const isPending = blocking || unblocking;

  const handleBlock = async () => {
    try {
      await blockSlot({ time, route, reason: reason || undefined });   // route added
      toast('Слот заблоковано', 'success');
      setAction(null);
      setReason('');
    } catch (e) {
      if (e instanceof ApiError && e.status === 409) toast('Вже заблоковано', 'error');
      else toast('Помилка сервера', 'error');
    }
  };

  const handleUnblock = async () => {
    try {
      await unblockSlot({ time, route });                              // route added
      toast('Слот розблоковано', 'success');
      setAction(null);
    } catch (e) {
      toast(e instanceof ApiError ? e.message : 'Помилка', 'error');
    }
  };

  return (
      <div style={{ padding: '14px 0', borderBottom: '1px solid var(--mist)' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12 }}>
          <div>
            <span style={{ fontFamily: 'Playfair Display, serif', fontWeight: 700, fontSize: '1.05rem' }}>{time}</span>
            {blocked && blockReason && (
                <div className="text-subtle mt-4">{blockReason}</div>
            )}
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <span className={`badge ${blocked ? 'badge-blocked' : 'badge-available'}`}>
            {blocked ? 'Заблоковано' : 'Доступний'}
          </span>
            {!action && (
                blocked
                    ? <button className="btn btn-sm" style={{ borderColor: 'var(--teal)', color: 'var(--teal)' }}
                              onClick={() => setAction('unblock')}>Розблокувати</button>
                    : <button className="btn btn-sm btn-danger" onClick={() => setAction('block')}>Заблокувати</button>
            )}
          </div>
        </div>

        {action === 'block' && (
            <ConfirmInline
                onConfirm={handleBlock}
                onCancel={() => { setAction(null); setReason(''); }}
                loading={isPending}
                confirmLabel="Заблокувати"
            >
              <div className="form-group">
                <label className="form-label">Причина (необов&apos;язково)</label>
                <input
                    type="text"
                    className="form-input"
                    placeholder="Технічне обслуговування…"
                    value={reason}
                    onChange={e => setReason(e.target.value)}
                    autoFocus
                />
              </div>
            </ConfirmInline>
        )}

        {action === 'unblock' && (
            <ConfirmInline
                onConfirm={handleUnblock}
                onCancel={() => setAction(null)}
                loading={isPending}
                confirmLabel="Розблокувати"
                danger={false}
            >
              <p className="text-subtle">Розблокувати слот {time} для бронювань?</p>
            </ConfirmInline>
        )}
      </div>
  );
}

function SlotBlockingTab() {
  const today = formatDate(new Date());
  const [date, setDate] = useState(today);
  const [route, setRoute] = useState<RouteName>('Desna');
  const { data, isLoading } = useSlots(date, route);

  return (
      <>
        <RouteSelector value={route} onChange={setRoute} />
        <div className="two-panel">
          <div>
            <Calendar selected={date} onSelect={setDate} route={route} />
          </div>
          <div className="card card-body">
            <h3 style={{ fontFamily: 'Playfair Display, serif', marginBottom: 16 }}>Слоти на {date}</h3>
            {isLoading && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                  {[1,2,3,4].map(i => <div key={i} className="skeleton" style={{ height: 48 }} />)}
                </div>
            )}
            {data?.slots.map(slot => (
                <SlotBlockRow
                    key={slot.time}
                    date={date}
                    time={slot.time}
                    route={route}
                    blocked={slot.blocked}
                    blockReason={slot.blockReason}
                />
            ))}
            {data?.slots.length === 0 && (
                <div className="empty-state">Немає слотів для цієї дати</div>
            )}
          </div>
        </div>
      </>
  );
}

// ── Tab B: Date Blocking ─────────────────────────────────────────
function DateBlockingTab() {
  const qc = useQueryClient();
  const [blockedDates, setBlockedDates] = useState<DayAvailability[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [newDate, setNewDate] = useState('');
  const [newReason, setNewReason] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [unblockingDate, setUnblockingDate] = useState<string | null>(null);
  const [confirmUnblock, setConfirmUnblock] = useState<string | null>(null);

  const fetchBlockedDates = useCallback(async () => {
    setLoading(true);
    const now = new Date();
    const months = [
      formatMonth(now),
      formatMonth(new Date(now.getFullYear(), now.getMonth() + 1, 1)),
    ];

    const allDays: DayAvailability[] = [];
    await Promise.all(months.map(async m => {
      try {
        const data = await adminFetch<{ days: DayAvailability[] }>(`/availability/${m}/${ROUTES[0]}`);
        allDays.push(...data.days.filter(d => d.blocked));
      } catch { /* ignore */ }
    }));

    allDays.sort((a, b) => a.date.localeCompare(b.date));
    setBlockedDates(allDays);
    setLoading(false);
  }, []);

  useEffect(() => { fetchBlockedDates(); }, [fetchBlockedDates]);

  const handleBlockDate = async () => {
    if (!newDate) return;
    setSubmitting(true);
    try {
      await adminFetch(`/admin/dates/${newDate}/block`, {
        method: 'PUT',
        body: JSON.stringify({ reason: newReason || undefined }),
      });
      toast('Дату заблоковано', 'success');
      qc.invalidateQueries({ queryKey: ['availability'] });
      setCreateOpen(false);
      setNewDate('');
      setNewReason('');
      fetchBlockedDates();
    } catch (e) {
      if (e instanceof ApiError && e.status === 409) toast('Вже заблоковано', 'error');
      else toast('Помилка сервера', 'error');
    } finally {
      setSubmitting(false);
    }
  };

  const handleUnblock = async (date: string) => {
    setUnblockingDate(date);
    try {
      await adminFetch(`/admin/dates/${date}/block`, { method: 'DELETE' });
      toast('Дату розблоковано', 'success');
      qc.invalidateQueries({ queryKey: ['availability'] });
      setConfirmUnblock(null);
      fetchBlockedDates();
    } catch (e) {
      toast(e instanceof ApiError ? e.message : 'Помилка', 'error');
    } finally {
      setUnblockingDate(null);
    }
  };

  const fmtDate = (d: string) => new Date(d + 'T00:00:00').toLocaleDateString('uk-UA', {
    day: 'numeric', month: 'long', year: 'numeric',
  });

  return (
      <div>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
          <h3 style={{ fontFamily: 'Playfair Display, serif' }}>Заблоковані дати</h3>
          <button className="btn btn-danger" onClick={() => setCreateOpen(true)}>
            + Заблокувати дату
          </button>
        </div>

        <div className="card">
          {loading && (
              <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 14 }}>
                {[1,2,3].map(i => <div key={i} className="skeleton" style={{ height: 48 }} />)}
              </div>
          )}

          {!loading && blockedDates.length === 0 && (
              <div className="empty-state">
                <div className="empty-state-icon">✅</div>
                <div>Немає заблокованих дат</div>
              </div>
          )}

          {!loading && blockedDates.map(d => (
              <div key={d.date} style={{ padding: '16px 20px', borderBottom: '1px solid var(--mist)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12 }}>
                  <div>
                    <div style={{ fontWeight: 600 }}>{fmtDate(d.date)}</div>
                  </div>
                  {confirmUnblock !== d.date ? (
                      <button
                          className="btn btn-sm"
                          style={{ borderColor: 'var(--teal)', color: 'var(--teal)' }}
                          onClick={() => setConfirmUnblock(d.date)}
                      >
                        Розблокувати
                      </button>
                  ) : (
                      <ConfirmInline
                          onConfirm={() => handleUnblock(d.date)}
                          onCancel={() => setConfirmUnblock(null)}
                          loading={unblockingDate === d.date}
                          confirmLabel="Розблокувати цю дату?"
                          danger={false}
                      />
                  )}
                </div>
              </div>
          ))}
        </div>

        <Modal
            open={createOpen}
            onClose={() => setCreateOpen(false)}
            title="Заблокувати дату"
            footer={
              <>
                <button className="btn btn-secondary" onClick={() => setCreateOpen(false)} disabled={submitting}>
                  Відміна
                </button>
                <button className="btn btn-danger-solid" onClick={handleBlockDate} disabled={submitting || !newDate}>
                  {submitting ? 'Блокування…' : 'Заблокувати'}
                </button>
              </>
            }
        >
          <div className="form-group">
            <label className="form-label">Дата</label>
            <input
                type="date"
                className="form-input"
                value={newDate}
                min={formatDate(new Date())}
                onChange={e => setNewDate(e.target.value)}
            />
          </div>
          <div className="form-group">
            <label className="form-label">Причина (необов&apos;язково)</label>
            <input
                type="text"
                className="form-input"
                placeholder="Свято, технічне обслуговування…"
                value={newReason}
                onChange={e => setNewReason(e.target.value)}
            />
          </div>
        </Modal>
      </div>
  );
}

// ── Page ─────────────────────────────────────────────────────────
export default function BlocksPage() {
  const [tab, setTab] = useState<'slots' | 'dates'>('slots');

  return (
      <AdminGuard>
        <Header />
        <main className="admin-main">
          <div className="page-container">
            <h1 className="page-title">Блокування</h1>
            <p className="page-subtitle">Керування блокуванням слотів та дат</p>

            <div className="page-tabs">
              <button className={`page-tab${tab === 'slots' ? ' active' : ''}`} onClick={() => setTab('slots')}>
                Слоти
              </button>
              <button className={`page-tab${tab === 'dates' ? ' active' : ''}`} onClick={() => setTab('dates')}>
                Дати
              </button>
            </div>

            {tab === 'slots' ? <SlotBlockingTab /> : <DateBlockingTab />}
          </div>
        </main>
      </AdminGuard>
  );
}
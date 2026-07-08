'use client';

import { useState, useEffect, useMemo } from 'react';
import { AdminGuard } from '@/components/AdminGuard';
import { Header } from '@/components/Header';
import { Modal } from '@/components/Modal';
import { usePromocodes, useCreatePromocode } from '@/hooks/usePromocodes';
import { toast } from '@/hooks/useToast';
import { ApiError } from '@/lib/api';
import { supabase } from '@/lib/supabase';
import { Promocode } from '@/lib/types';

// Backend puts the error code in the `message` field (see httpx.Err) — map the
// promo-specific ones to friendly Ukrainian copy, fall back to a generic message.
function createErrorMessage(e: unknown): string {
  if (e instanceof ApiError) {
    switch (e.message) {
      case 'PROMO_ALREADY_EXISTS': return 'Такий код вже існує.';
      case 'INVALID_INPUT':        return 'Некоректні дані. Перевірте поля.';
      case 'SERVICE_UNAVAILABLE':  return 'Сервіс тимчасово недоступний. Спробуйте ще раз.';
    }
  }
  return 'Не вдалося створити промокод';
}

function fmtCreated(iso?: string) {
  if (!iso) return '—';
  return new Date(iso).toLocaleDateString('uk-UA', { day: 'numeric', month: 'short', year: 'numeric' });
}

// ── Usage bar (reuses the slot capacity bar styles) ──────────────────
function UsageBar({ used, max }: { used: number; max: number }) {
  const pct = max > 0 ? Math.min(100, Math.round((used / max) * 100)) : 0;
  const cls = pct < 50 ? 'fill-good' : pct < 85 ? 'fill-warn' : 'fill-full';
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <div className="slots-bar" style={{ flex: 1, minWidth: 60 }}>
        <div className={`slots-fill ${cls}`} style={{ width: `${pct}%` }} />
      </div>
      <span className="text-subtle" style={{ fontSize: '0.75rem', minWidth: 52, whiteSpace: 'nowrap' }}>
        {used}/{max}
      </span>
    </div>
  );
}

function StatusPill({ active, exhausted }: { active: boolean; exhausted: boolean }) {
  if (!active) return <span className="badge badge-cancelled">Неактивний</span>;
  if (exhausted) return <span className="badge badge-blocked">Вичерпано</span>;
  return <span className="badge badge-available">Активний</span>;
}

// ── Stat tiles ───────────────────────────────────────────────────────
function StatTiles({ codes }: { codes: Promocode[] }) {
  const total = codes.length;
  const active = codes.filter(c => c.active).length;
  const redemptions = codes.reduce((s, c) => s + c.timesUsed, 0);
  const remaining = codes.reduce((s, c) => s + Math.max(0, c.maxUses - c.timesUsed), 0);

  const tiles = [
    { label: 'Усього кодів', value: total, hint: `${active} активних` },
    { label: 'Використань', value: redemptions, hint: 'підтверджених бронювань' },
    { label: 'Залишок', value: remaining, hint: 'доступно застосувань' },
  ];

  return (
    <div className="promo-stats">
      {tiles.map(t => (
        <div key={t.label} className="promo-stat">
          <div className="promo-stat-value">{t.value}</div>
          <div className="promo-stat-label">{t.label}</div>
          <div className="promo-stat-hint">{t.hint}</div>
        </div>
      ))}
    </div>
  );
}

// ── Create modal ─────────────────────────────────────────────────────
function CreatePromoModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { mutateAsync, isPending } = useCreatePromocode();
  const [code, setCode] = useState('');
  const [percent, setPercent] = useState('10');
  const [maxUses, setMaxUses] = useState('100');

  const reset = () => { setCode(''); setPercent('10'); setMaxUses('100'); };
  const close = () => { reset(); onClose(); };

  const percentNum = Number(percent);
  const maxUsesNum = Number(maxUses);
  const valid =
    code.trim().length >= 1 && code.trim().length <= 64 &&
    Number.isFinite(percentNum) && percentNum >= 0 && percentNum <= 100 &&
    Number.isInteger(maxUsesNum) && maxUsesNum >= 1;

  const handleSubmit = async () => {
    if (!valid) return;
    try {
      const created = await mutateAsync({
        code: code.trim(),
        discountPercent: percentNum,
        maxUses: maxUsesNum,
      });
      toast(`Промокод ${created.code} створено`, 'success');
      close();
    } catch (e) {
      toast(createErrorMessage(e), 'error');
    }
  };

  return (
    <Modal
      open={open}
      onClose={close}
      title="Новий промокод"
      footer={
        <>
          <button className="btn btn-secondary" onClick={close} disabled={isPending}>Відміна</button>
          <button className="btn btn-primary" onClick={handleSubmit} disabled={isPending || !valid}>
            {isPending ? '…' : 'Створити'}
          </button>
        </>
      }
    >
      <div className="form-group">
        <label className="form-label">Код</label>
        <input
          type="text"
          className="form-input"
          placeholder="SUMMER10"
          value={code}
          onChange={e => setCode(e.target.value)}
          maxLength={64}
          autoFocus
          style={{ textTransform: 'uppercase' }}
        />
        <span className="text-subtle" style={{ fontSize: '0.75rem', marginTop: 4, display: 'block' }}>
          1–64 символи. Регістр не має значення (зберігається у верхньому).
        </span>
      </div>
      <div style={{ display: 'flex', gap: 12 }}>
        <div className="form-group" style={{ flex: 1 }}>
          <label className="form-label">Знижка, %</label>
          <input
            type="number" min={0} max={100} className="form-input"
            value={percent} onChange={e => setPercent(e.target.value)}
          />
        </div>
        <div className="form-group" style={{ flex: 1 }}>
          <label className="form-label">Ліміт застосувань</label>
          <input
            type="number" min={1} step={1} className="form-input"
            value={maxUses} onChange={e => setMaxUses(e.target.value)}
          />
        </div>
      </div>
      <p className="text-subtle" style={{ fontSize: '0.78rem', lineHeight: 1.6 }}>
        Ліміт — глобальний, спільний для всіх клієнтів. Лічильник зростає лише після
        <b> підтвердженого (оплаченого)</b> бронювання.
      </p>
    </Modal>
  );
}

// ── Table row / mobile card ──────────────────────────────────────────
function isExhausted(c: Promocode) { return c.timesUsed >= c.maxUses; }

function PromoRow({ c }: { c: Promocode }) {
  return (
    <tr style={!c.active ? { opacity: 0.7 } : undefined}>
      <td><span className="promo-code">{c.code}</span></td>
      <td style={{ whiteSpace: 'nowrap' }} className="booking-amount">−{c.discountPercent}%</td>
      <td style={{ minWidth: 160 }}><UsageBar used={c.timesUsed} max={c.maxUses} /></td>
      <td><StatusPill active={c.active} exhausted={isExhausted(c)} /></td>
      <td style={{ whiteSpace: 'nowrap', color: 'var(--subtle)' }}>{fmtCreated(c.createdAt)}</td>
    </tr>
  );
}

function PromoCard({ c }: { c: Promocode }) {
  return (
    <div className="history-card" style={!c.active ? { opacity: 0.75 } : undefined}>
      <div className="history-card-top">
        <span className="promo-code">{c.code}</span>
        <StatusPill active={c.active} exhausted={isExhausted(c)} />
      </div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, marginBottom: 10 }}>
        <span className="booking-amount">−{c.discountPercent}%</span>
        <span className="text-subtle" style={{ fontSize: '0.78rem' }}>Створено: {fmtCreated(c.createdAt)}</span>
      </div>
      <UsageBar used={c.timesUsed} max={c.maxUses} />
    </div>
  );
}

// ── Page ─────────────────────────────────────────────────────────────
export default function PromocodesPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const [mineOnly, setMineOnly] = useState(false);
  const [myId, setMyId] = useState<string>('');

  useEffect(() => {
    supabase.auth.getSession().then(({ data: { session } }) => {
      setMyId(session?.user?.id ?? '');
    });
  }, []);

  const { data, isLoading, isError } = usePromocodes(mineOnly && myId ? myId : undefined);
  const codes = useMemo(() => data?.promocodes ?? [], [data]);

  return (
    <AdminGuard>
      <Header />
      <main className="admin-main">
        <div className="page-container">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 12, flexWrap: 'wrap' }}>
            <div>
              <h1 className="page-title">Промокоди</h1>
              <p className="page-subtitle">Створення та статистика знижкових кодів</p>
            </div>
            <button className="btn btn-primary" onClick={() => setCreateOpen(true)}>+ Новий код</button>
          </div>

          {!isLoading && !isError && <StatTiles codes={codes} />}

          <div className="history-filters">
            <label style={{ display: 'inline-flex', alignItems: 'center', gap: 8, cursor: myId ? 'pointer' : 'not-allowed', fontSize: '0.875rem', color: 'var(--navy)' }}>
              <input
                type="checkbox"
                checked={mineOnly}
                disabled={!myId}
                onChange={e => setMineOnly(e.target.checked)}
              />
              Лише створені мною
            </label>
          </div>

          <div className="card">
            {isLoading && (
              <div style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}>
                {[1, 2, 3, 4].map(i => <div key={i} className="skeleton" style={{ height: 44 }} />)}
              </div>
            )}

            {isError && (
              <div className="empty-state">
                <div className="empty-state-icon">⚠</div>
                <div>Не вдалося завантажити промокоди</div>
              </div>
            )}

            {!isLoading && !isError && codes.length === 0 && (
              <div className="empty-state">
                <div className="empty-state-icon">🎟</div>
                <div>{mineOnly ? 'Ви ще не створили жодного коду' : 'Промокодів ще немає'}</div>
              </div>
            )}

            {!isLoading && !isError && codes.length > 0 && (
              <>
                <div className="history-table-wrap">
                  <table className="history-table">
                    <thead>
                      <tr>
                        <th>Код</th>
                        <th>Знижка</th>
                        <th>Використання</th>
                        <th>Статус</th>
                        <th>Створено</th>
                      </tr>
                    </thead>
                    <tbody>
                      {codes.map(c => <PromoRow key={c.code} c={c} />)}
                    </tbody>
                  </table>
                </div>
                <div className="history-cards">
                  {codes.map(c => <PromoCard key={c.code} c={c} />)}
                </div>
              </>
            )}
          </div>

          <p className="text-subtle" style={{ fontSize: '0.78rem', marginTop: 12, lineHeight: 1.6 }}>
            ℹ Коди наразі можна лише створювати та переглядати — редагування чи деактивація через API
            поки не підтримуються.
          </p>

          <CreatePromoModal open={createOpen} onClose={() => setCreateOpen(false)} />
        </div>
      </main>
    </AdminGuard>
  );
}
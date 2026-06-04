'use client';

import { useState } from 'react';
import { AdminGuard } from '@/components/AdminGuard';
import { Header } from '@/components/Header';
import { StatusBadge } from '@/components/StatusBadge';
import { useBookingHistory } from '@/hooks/useBookingHistory';
import { routeLabel } from '@/lib/routes';
import { Booking, BookingStatus } from '@/lib/types';

const STATUS_OPTIONS: { value: '' | BookingStatus; label: string }[] = [
    { value: '', label: 'Усі статуси' },
    { value: 'pending', label: 'Очікує' },
    { value: 'confirmed', label: 'Підтверджено' },
    { value: 'failed', label: 'Помилка' },
    { value: 'expired', label: 'Прострочено' },
    { value: 'cancelled', label: 'Скасовано' },
];

const LIMIT_OPTIONS = [50, 100, 200];

function fmtSlot(date?: string, time?: string) {
    if (!date) return '—';
    const d = new Date(date + 'T00:00:00').toLocaleDateString('uk-UA', { day: 'numeric', month: 'short', year: 'numeric' });
    return time ? `${d} · ${time}` : d;
}

function fmtCreated(iso?: string) {
    if (!iso) return '—';
    return new Date(iso).toLocaleString('uk-UA', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' });
}

function boatsLine(q: Booking['quantities']) {
    return `В:${q.big} · С:${q.medium} · М:${q.small ?? 0} · Д:${q.child}`;
}

function PosterTags({ b, compact = false }: { b: Booking; compact?: boolean }) {
    const hasOrder = b.posterIncomingOrderId != null;
    const hasTxn = b.posterIncomingTransactionId != null;

    if (!hasOrder && !hasTxn) {
        return compact ? null : <span className="text-subtle">—</span>;
    }

    return (
        <div className="text-subtle" style={{ fontSize: '0.78rem', lineHeight: 1.5 }}>
            {hasOrder && <div>зам. <strong>#{b.posterIncomingOrderId}</strong></div>}
            {hasTxn && <div>трз. <strong>#{b.posterIncomingTransactionId}</strong></div>}
        </div>
    );
}

// ── Desktop table row ─────────────────────────────────────────────
function HistoryRow({ b }: { b: Booking }) {
    return (
        <tr>
            <td style={{ whiteSpace: 'nowrap' }}>{fmtSlot(b.date, b.time)}</td>
            <td><span className="badge badge-route">{b.routeName ? routeLabel(b.routeName) : '—'}</span></td>
            <td>
                <div className="booking-name">{b.firstName} {b.lastName}</div>
                <div className="booking-contact">{b.userEmail}{b.phone ? ` · ${b.phone}` : ''}</div>
            </td>
            <td style={{ whiteSpace: 'nowrap' }} title="Великі · Середні · Малі · Діти">{boatsLine(b.quantities)}</td>
            <td className="booking-amount" style={{ whiteSpace: 'nowrap' }}>{b.effectiveAmount.toFixed(2)} ₴</td>
            <td><StatusBadge status={b.status} /></td>
            <td style={{ whiteSpace: 'nowrap' }}><PosterTags b={b} /></td>
            <td style={{ whiteSpace: 'nowrap', color: 'var(--subtle)' }}>{fmtCreated(b.createdAt)}</td>
        </tr>
    );
}

// ── Mobile card ───────────────────────────────────────────────────
function HistoryCard({ b }: { b: Booking }) {
    return (
        <div className="history-card">
            <div className="history-card-top">
                <div>
                    <div className="history-card-slot">{fmtSlot(b.date, b.time)}</div>
                    <span className="badge badge-route mt-4" style={{ display: 'inline-flex' }}>
            {b.routeName ? routeLabel(b.routeName) : '—'}
          </span>
                </div>
                <StatusBadge status={b.status} />
            </div>

            <div className="booking-name">{b.firstName} {b.lastName}</div>
            <div className="booking-contact">{b.userEmail}{b.phone ? ` · ${b.phone}` : ''}</div>

            <div className="history-card-footer">
                <span className="text-subtle" title="Великі · Середні · Малі · Діти">{boatsLine(b.quantities)}</span>
                <span className="booking-amount">{b.effectiveAmount.toFixed(2)} ₴</span>
            </div>

            {(b.posterIncomingOrderId != null || b.posterIncomingTransactionId != null) && (
                <div className="mt-4" style={{ fontSize: '0.78rem' }}>
                    <span className="text-subtle">Poster: </span>
                    {b.posterIncomingOrderId != null && <>зам. <strong>#{b.posterIncomingOrderId}</strong></>}
                    {b.posterIncomingOrderId != null && b.posterIncomingTransactionId != null && <span> · </span>}
                    {b.posterIncomingTransactionId != null && <>трз. <strong>#{b.posterIncomingTransactionId}</strong></>}
                </div>
            )}

            <div className="text-subtle mt-4" style={{ fontSize: '0.78rem' }}>
                Створено: {fmtCreated(b.createdAt)}
            </div>
        </div>
    );
}

export default function HistoryPage() {
    const [date, setDate] = useState('');
    const [status, setStatus] = useState<'' | BookingStatus>('');
    const [limit, setLimit] = useState(50);
    const [offset, setOffset] = useState(0);

    const { data, isLoading, isError, isFetching } = useBookingHistory({
        date: date || undefined,
        status: status || undefined,
        limit,
        offset,
    });

    const bookings = data?.bookings ?? [];
    const hasNext = bookings.length === limit;  // no total count from backend (per API doc)
    const hasPrev = offset > 0;

    const onFilter = (fn: () => void) => { fn(); setOffset(0); };

    return (
        <AdminGuard>
            <Header />
            <main className="admin-main">
                <div className="page-container">
                    <h1 className="page-title">Історія бронювань</h1>
                    <p className="page-subtitle">Найновіші зверху · Човни: В — великі, С — середні, М — малі, Д — діти</p>

                    <div className="history-filters">
                        <div className="form-group" style={{ flex: '1 1 160px' }}>
                            <label className="form-label">Дата слоту</label>
                            <input type="date" className="form-input" value={date}
                                   onChange={e => onFilter(() => setDate(e.target.value))} />
                        </div>
                        <div className="form-group" style={{ flex: '1 1 160px' }}>
                            <label className="form-label">Статус</label>
                            <select className="form-input" value={status}
                                    onChange={e => onFilter(() => setStatus(e.target.value as '' | BookingStatus))}>
                                {STATUS_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
                            </select>
                        </div>
                        <div className="form-group" style={{ flex: '1 1 110px' }}>
                            <label className="form-label">На сторінці</label>
                            <select className="form-input" value={limit}
                                    onChange={e => onFilter(() => setLimit(Number(e.target.value)))}>
                                {LIMIT_OPTIONS.map(n => <option key={n} value={n}>{n}</option>)}
                            </select>
                        </div>
                        {(date || status) && (
                            <button className="btn btn-ghost btn-sm history-reset"
                                    onClick={() => onFilter(() => { setDate(''); setStatus(''); })}>
                                Скинути фільтри
                            </button>
                        )}
                    </div>

                    <div className="card" style={{ opacity: isFetching && !isLoading ? 0.6 : 1, transition: 'opacity 150ms' }}>
                        {isLoading && (
                            <div style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}>
                                {[1, 2, 3, 4, 5].map(i => <div key={i} className="skeleton" style={{ height: 44 }} />)}
                            </div>
                        )}

                        {isError && (
                            <div className="empty-state">
                                <div className="empty-state-icon">⚠</div>
                                <div>Не вдалося завантажити історію</div>
                            </div>
                        )}

                        {!isLoading && !isError && bookings.length === 0 && (
                            <div className="empty-state">
                                <div className="empty-state-icon">📋</div>
                                <div>Бронювань не знайдено</div>
                            </div>
                        )}

                        {!isLoading && !isError && bookings.length > 0 && (
                            <>
                                {/* Desktop / tablet */}
                                <div className="history-table-wrap">
                                    <table className="history-table">
                                        <thead>
                                        <tr>
                                            <th>Слот</th>
                                            <th>Маршрут</th>
                                            <th>Клієнт</th>
                                            <th title="Великі · Середні · Малі · Діти">Човни</th>
                                            <th>Сума</th>
                                            <th>Статус</th>
                                            <th>Poster</th>
                                            <th>Створено</th>
                                        </tr>
                                        </thead>
                                        <tbody>
                                        {bookings.map(b => <HistoryRow key={b.id} b={b} />)}
                                        </tbody>
                                    </table>
                                </div>

                                {/* Mobile */}
                                <div className="history-cards">
                                    {bookings.map(b => <HistoryCard key={b.id} b={b} />)}
                                </div>
                            </>
                        )}
                    </div>

                    {!isError && (
                        <div className="history-pager">
              <span className="text-subtle">
                {bookings.length > 0 ? `Показано ${offset + 1}–${offset + bookings.length}` : '\u00A0'}
              </span>
                            <div style={{ display: 'flex', gap: 8 }}>
                                <button className="btn btn-secondary btn-sm" disabled={!hasPrev || isFetching}
                                        onClick={() => setOffset(o => Math.max(0, o - limit))}>‹ Назад</button>
                                <button className="btn btn-secondary btn-sm" disabled={!hasNext || isFetching}
                                        onClick={() => setOffset(o => o + limit)}>Далі ›</button>
                            </div>
                        </div>
                    )}
                </div>
            </main>
        </AdminGuard>
    );
}
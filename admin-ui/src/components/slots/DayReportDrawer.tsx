'use client';

import { useState, useMemo } from 'react';
import { Drawer } from '@/components/Drawer';
import { StatusBadge } from '@/components/StatusBadge';
import { PosterCopy, CopyText } from '@/components/PosterCopy';
import { PromoBadge } from '@/components/PromoBadge';
import { useDayBookings } from '@/hooks/useDayBookings';
import { toast } from '@/hooks/useToast';
import { Booking } from '@/lib/types';
import { ROUTES, routeLabel } from '@/lib/routes';

interface DayReportDrawerProps {
    open: boolean;
    onClose: () => void;
    date: string;
}

const isActive = (b: Booking) =>
    b.status !== 'cancelled' && b.status !== 'expired' && b.status !== 'failed';

const fmtLong = (d: string) =>
    d ? new Date(d + 'T00:00:00').toLocaleDateString('uk-UA', {
        day: 'numeric', month: 'long', year: 'numeric',
    }) : '';

function boatsLine(q: Booking['quantities']) {
    return `В:${q.big} · С:${q.medium} · М:${q.small ?? 0} · Д:${q.child}`;
}

function sumBoats(list: Booking[]) {
    return list.reduce(
        (a, b) => ({
            big: a.big + b.quantities.big,
            medium: a.medium + b.quantities.medium,
            small: a.small + (b.quantities.small ?? 0),
            child: a.child + b.quantities.child,
        }),
        { big: 0, medium: 0, small: 0, child: 0 },
    );
}

async function copyToClipboard(text: string): Promise<boolean> {
    try {
        await navigator.clipboard.writeText(text);
        return true;
    } catch {
        try {
            const el = document.createElement('textarea');
            el.value = text;
            el.style.cssText = 'position:fixed;top:-9999px;left:-9999px;opacity:0';
            document.body.appendChild(el);
            el.focus();
            el.select();
            document.execCommand('copy');
            document.body.removeChild(el);
            return true;
        } catch { return false; }
    }
}

// ── Grouping: route → time ────────────────────────────────────────
interface SlotGroup {
    route: string;
    time: string;
    bookings: Booking[];
}

function groupBySlot(bookings: Booking[]): SlotGroup[] {
    const map = new Map<string, Booking[]>();
    for (const b of bookings) {
        const key = `${b.routeName ?? ''}||${b.time ?? ''}`;
        const arr = map.get(key);
        if (arr) arr.push(b);
        else map.set(key, [b]);
    }

    const rank = (r: string) => {
        const i = (ROUTES as readonly string[]).indexOf(r);
        return i === -1 ? 999 : i;
    };

    const groups: SlotGroup[] = [];
    map.forEach((list, key) => {
        const [route, time] = key.split('||');
        groups.push({ route, time, bookings: list });
    });
    groups.sort((a, b) => rank(a.route) - rank(b.route) || a.time.localeCompare(b.time));
    return groups;
}

// ── Copy text builders ────────────────────────────────────────────
function buildGroupText(g: SlotGroup): string {
    const active = g.bookings.filter(isActive);
    const lines = [`${routeLabel(g.route)} · ${g.time}`, '─'.repeat(34)];
    active.forEach((b, i) => {
        lines.push(`${i + 1}. ${b.firstName} ${b.lastName}`);
        lines.push(`   📞 ${b.phone ?? '—'}  🚣 ${boatsLine(b.quantities)}  💰 ${b.effectiveAmount.toFixed(2)} ₴`);
    });
    lines.push('─'.repeat(34), `Всього: ${active.length}`);
    return lines.join('\n');
}

function buildDayText(date: string, groups: SlotGroup[]): string {
    const lines: string[] = [`📋 Звіт за ${fmtLong(date)}`, '═'.repeat(40)];
    const all: Booking[] = [];

    for (const g of groups) {
        const active = g.bookings.filter(isActive);
        if (active.length === 0) continue;
        all.push(...active);
        lines.push('', `🚣 ${routeLabel(g.route)} · ⏰ ${g.time}  (${active.length})`);
        active.forEach((b, i) => {
            lines.push(`  ${i + 1}. ${b.firstName} ${b.lastName}${b.phone ? ` · ${b.phone}` : ''}`);
            lines.push(`     ${boatsLine(b.quantities)} · ${b.effectiveAmount.toFixed(2)} ₴`);
        });
    }

    const tot = sumBoats(all);
    const amount = all.reduce((s, b) => s + b.effectiveAmount, 0);
    const discount = all.reduce((s, b) => s + (b.discountAmount ?? 0), 0);
    const promoCount = all.filter(b => b.promoCode).length;
    lines.push(
        '', '═'.repeat(40),
        `Всього бронювань: ${all.length}`,
        `Човни — В:${tot.big} · С:${tot.medium} · М:${tot.small} · Д:${tot.child}`,
        `Сума: ${amount.toFixed(2)} ₴`,
    );
    if (discount > 0) {
        lines.push(`Знижки (${promoCount} з промокодом): −${discount.toFixed(2)} ₴`);
    }
    return lines.join('\n');
}

// ── Small copy button ─────────────────────────────────────────────
function CopyBtn({ text, label, onDone }: { text: string; label: string; onDone?: () => void }) {
    const [copied, setCopied] = useState(false);
    return (
        <button
            onClick={async () => {
                const ok = await copyToClipboard(text);
                setCopied(ok);
                onDone?.();
                setTimeout(() => setCopied(false), 1800);
            }}
            className="btn btn-sm"
            style={{
                borderColor: copied ? 'rgba(14,124,123,0.45)' : 'var(--mist)',
                color: copied ? 'var(--teal)' : 'var(--subtle)',
                background: copied ? 'rgba(14,124,123,0.08)' : 'transparent',
                whiteSpace: 'nowrap',
            }}
        >
            {copied ? '✓ ' : '⎘ '}{label}
        </button>
    );
}

// ── Read-only booking row ─────────────────────────────────────────
function ReportRow({ b, idx }: { b: Booking; idx: number }) {
    return (
        <div className="booking-row" style={b.status === 'cancelled' ? { opacity: 0.6 } : undefined}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 8 }}>
                <div>
                    <div className="booking-name">{idx}. {b.firstName} {b.lastName}</div>
                    <div className="booking-contact">
                        {b.userEmail}
                        {b.phone && <><span style={{ color: 'var(--mist)', margin: '0 4px' }}>·</span><CopyText value={b.phone} /></>}
                    </div>
                    <div className="booking-quantities mt-4">{boatsLine(b.quantities)}</div>
                    <PosterCopy
                        orderId={b.posterIncomingOrderId}
                        transactionId={b.posterIncomingTransactionId}
                    />
                    <PromoBadge
                        code={b.promoCode}
                        discountPercent={b.discountPercent}
                        discountAmount={b.discountAmount}
                    />
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 6 }}>
                    <StatusBadge status={b.status} />
                    <span className="booking-amount">{b.effectiveAmount.toFixed(2)} ₴</span>
                </div>
            </div>
        </div>
    );
}

// ── Drawer ────────────────────────────────────────────────────────
export function DayReportDrawer({ open, onClose, date }: DayReportDrawerProps) {
    const { data, isLoading, isError } = useDayBookings(date, open);

    const bookings = data?.bookings ?? [];
    const groups = useMemo(() => groupBySlot(bookings), [bookings]);

    const active = bookings.filter(isActive);
    const cancelledCount = bookings.filter(b => b.status === 'cancelled').length;
    const totals = sumBoats(active);
    const totalAmount = active.reduce((s, b) => s + b.effectiveAmount, 0);
    const totalDiscount = active.reduce((s, b) => s + (b.discountAmount ?? 0), 0);
    const promoCount = active.filter(b => b.promoCode).length;

    const subtitleParts: string[] = [];
    if (active.length) subtitleParts.push(`${active.length} активних`);
    if (cancelledCount) subtitleParts.push(`${cancelledCount} скасовано`);

    return (
        <Drawer
            open={open}
            onClose={onClose}
            title={`Звіт · ${fmtLong(date)}`}
            subtitle={subtitleParts.join(' · ') || undefined}
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
                    <div>Не вдалося завантажити звіт</div>
                </div>
            )}

            {!isLoading && !isError && active.length === 0 && (
                <div className="empty-state">
                    <div className="empty-state-icon">📋</div>
                    <div>Немає активних бронювань на цей день</div>
                </div>
            )}

            {!isLoading && !isError && active.length > 0 && (
                <>
                    {/* Summary + copy-all-day */}
                    <div style={{
                        background: 'rgba(14,124,123,0.06)',
                        border: '1px solid rgba(14,124,123,0.2)',
                        borderRadius: 'var(--radius-sm)',
                        padding: '14px 16px',
                        marginBottom: 16,
                    }}>
                        <div style={{
                            display: 'flex', justifyContent: 'space-between',
                            alignItems: 'center', gap: 12, flexWrap: 'wrap', marginBottom: 10,
                        }}>
                            <div>
                                <div style={{ fontWeight: 700, color: 'var(--navy)' }}>
                                    {active.length} бронювань
                                </div>
                                <div className="text-subtle" style={{ fontSize: '0.8rem', marginTop: 2 }}>
                                    Човни — {`В:${totals.big} · С:${totals.medium} · М:${totals.small} · Д:${totals.child}`}
                                </div>
                            </div>
                            <div style={{ textAlign: 'right' }}>
                                <div className="booking-amount">{totalAmount.toFixed(2)} ₴</div>
                                {totalDiscount > 0 && (
                                    <div style={{ fontSize: '0.78rem', color: 'var(--coral)', fontWeight: 600, marginTop: 2 }}>
                                        🎟 Знижки: −{totalDiscount.toFixed(2)} ₴
                                    </div>
                                )}
                            </div>
                        </div>
                        {totalDiscount > 0 && (
                            <div className="text-subtle" style={{ fontSize: '0.78rem', marginBottom: 10 }}>
                                Промокоди застосовано у {promoCount} з {active.length} бронювань
                            </div>
                        )}
                        <CopyBtn
                            text={buildDayText(date, groups)}
                            label="Копіювати весь день"
                            onDone={() => toast('Звіт за день скопійовано', 'success')}
                        />
                    </div>

                    {cancelledCount > 0 && (
                        <div style={{
                            background: 'rgba(224,90,78,0.06)',
                            border: '1px solid rgba(224,90,78,0.2)',
                            borderRadius: 'var(--radius-sm)',
                            padding: '8px 14px',
                            marginBottom: 16,
                            fontSize: '0.82rem',
                            color: 'var(--coral)',
                        }}>
                            🚫 {cancelledCount} скасованих бронювань на цей день (нижче не показані)
                        </div>
                    )}

                    {/* Groups by slot */}
                    {groups.map(g => {
                        const groupActive = g.bookings.filter(isActive);
                        if (groupActive.length === 0) return null;
                        return (
                            <div key={`${g.route}-${g.time}`} style={{ marginBottom: 22 }}>
                                <div style={{
                                    display: 'flex', justifyContent: 'space-between',
                                    alignItems: 'center', gap: 10, marginBottom: 6,
                                    paddingBottom: 6, borderBottom: '2px solid var(--mist)',
                                }}>
                                    <div>
                                        <span style={{
                                            fontFamily: 'Playfair Display, serif',
                                            fontWeight: 700, fontSize: '1.05rem',
                                        }}>
                                            {g.time}
                                        </span>
                                        <span className="badge badge-route" style={{ marginLeft: 8 }}>
                                            {routeLabel(g.route)}
                                        </span>
                                        <span className="text-subtle" style={{ marginLeft: 8, fontSize: '0.8rem' }}>
                                            {groupActive.length} чол.
                                        </span>
                                    </div>
                                    <CopyBtn text={buildGroupText(g)} label="Копіювати" />
                                </div>

                                {groupActive.map((b, i) => (
                                    <ReportRow key={b.id} b={b} idx={i + 1} />
                                ))}
                            </div>
                        );
                    })}
                </>
            )}
        </Drawer>
    );
}
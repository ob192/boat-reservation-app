'use client';

import { useState } from 'react';
import { AdminGuard } from '@/components/AdminGuard';
import { Header } from '@/components/Header';
import { Calendar } from '@/components/Calendar';
import { RouteSelector } from '@/components/RouteSelector';
import { SlotCard } from '@/components/slots/SlotCard';
import { SlotCreateModal } from '@/components/slots/SlotCreateModal';
import { useSlots } from '@/hooks/useSlots';
import { formatDate } from '@/lib/api';
import { RouteName } from '@/lib/routes';
import { DayReportDrawer } from '@/components/slots/DayReportDrawer';

function SlotsList({ date, route }: { date: string; route: RouteName }) {
    const { data, isLoading, isError } = useSlots(date, route);
    const [createOpen, setCreateOpen] = useState(false);
    const [reportOpen, setReportOpen] = useState(false);

    if (isLoading) {
        return (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                {[1, 2, 3, 4].map(i => (
                    <div key={i} className="slot-card">
                        <div className="skeleton" style={{ height: 28, width: '30%', marginBottom: 8 }} />
                        <div className="skeleton" style={{ height: 14, width: '20%', marginBottom: 16 }} />
                        <div className="skeleton" style={{ height: 8, width: '100%', marginBottom: 8 }} />
                        <div className="skeleton" style={{ height: 8, width: '100%' }} />
                    </div>
                ))}
            </div>
        );
    }

    if (isError) {
        return (
            <div className="empty-state">
                <div className="empty-state-icon">⚠</div>
                <div>Не вдалося завантажити слоти</div>
            </div>
        );
    }

    const formattedDate = new Date(date + 'T00:00:00').toLocaleDateString('uk-UA', {
        weekday: 'long', day: 'numeric', month: 'long', year: 'numeric',
    });

    return (
        <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20, gap: 12, flexWrap: 'wrap' }}>
                <div>
                    <h2 style={{ fontFamily: 'Playfair Display, serif', fontSize: '1.1rem', color: 'var(--navy)', textTransform: 'capitalize' }}>
                        {formattedDate}
                    </h2>
                    {data?.dateBlocked && (
                        <span className="badge badge-blocked mt-4" style={{ display: 'inline-flex' }}>
                            🗓 Дата заблокована
                        </span>
                    )}
                    {data && !data.bookingsEnabled && (
                        <span className="badge badge-cancelled mt-4" style={{ display: 'inline-flex', marginLeft: 8 }}>
                            ⛔ Бронювання вимкнено
                        </span>
                    )}
                </div>

                {/* NEW: per-day report */}
                <button className="btn btn-secondary btn-sm" onClick={() => setReportOpen(true)}>
                    📋 Звіт за день
                </button>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                {data?.slots.map(slot => (
                    <SlotCard key={slot.time} date={date} slot={slot} />
                ))}
                {data?.slots.length === 0 && (
                    <div className="empty-state">Немає слотів для цієї дати</div>
                )}
            </div>

            <button
                className="btn btn-secondary w-full mt-16"
                onClick={() => setCreateOpen(true)}
                style={{ justifyContent: 'center', marginTop: 20 }}
            >
                + Додати слот
            </button>

            <SlotCreateModal
                open={createOpen}
                onClose={() => setCreateOpen(false)}
                date={date}
                route={route}
            />

            <DayReportDrawer
                open={reportOpen}
                onClose={() => setReportOpen(false)}
                date={date}
            />
        </div>
    );
}

export default function SlotsPage() {
    const today = formatDate(new Date());
    const [selectedDate, setSelectedDate] = useState(today);
    const [route, setRoute] = useState<RouteName>('Desna');

    return (
        <AdminGuard>
            <Header />
            <main className="admin-main">
                <div className="page-container">
                    <h1 className="page-title">Слоти</h1>
                    <p className="page-subtitle">Перегляд та управління часовими слотами</p>

                    <RouteSelector value={route} onChange={setRoute} />

                    <div className="two-panel">
                        <div>
                            <Calendar selected={selectedDate} onSelect={setSelectedDate} route={route} />
                        </div>
                        <div>
                            <SlotsList date={selectedDate} route={route} />
                        </div>
                    </div>
                </div>
            </main>
        </AdminGuard>
    );
}
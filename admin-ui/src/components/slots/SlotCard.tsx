'use client';

import { useState } from 'react';
import { CapacityBar } from '@/components/CapacityBar';
import { SlotEditModal } from './SlotEditModal';
import { BookingsDrawer } from './BookingsDrawer';
import { SlotInfo } from '@/lib/types';

interface SlotCardProps {
  date: string;
  slot: SlotInfo;
}

export function SlotCard({ date, slot }: SlotCardProps) {
  const [editOpen, setEditOpen] = useState(false);
  const [bookingsOpen, setBookingsOpen] = useState(false);

  return (
      <>
        <div className={`slot-card${slot.blocked ? ' slot-blocked' : ''}`}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 12 }}>
            <div className="slot-time">{slot.time}</div>
            {slot.blocked && (
                <span className="badge badge-blocked">🔒 Заблоковано</span>
            )}
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 14 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <span className="text-subtle" style={{ width: 80, flexShrink: 0, fontSize: '0.8rem' }}>Великі:</span>
              <CapacityBar available={slot.availableBig} total={slot.totalBig} />
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <span className="text-subtle" style={{ width: 80, flexShrink: 0, fontSize: '0.8rem' }}>Середні:</span>
              <CapacityBar available={slot.availableMedium} total={slot.totalMedium} />
            </div>
          </div>

          <div className="divider" style={{ margin: '0 0 12px' }} />

          <div style={{ display: 'flex', gap: 8 }}>
            <button className="btn btn-secondary btn-sm" onClick={() => setEditOpen(true)}>
              ✏️ Редагувати
            </button>
            <button className="btn btn-ghost btn-sm" onClick={() => setBookingsOpen(true)}>
              📋 Бронювання
            </button>
          </div>
        </div>

        <SlotEditModal
            open={editOpen}
            onClose={() => setEditOpen(false)}
            date={date}
            slot={slot}
        />

        <BookingsDrawer
            open={bookingsOpen}
            onClose={() => setBookingsOpen(false)}
            date={date}
            time={slot.time}
        />
      </>
  );
}
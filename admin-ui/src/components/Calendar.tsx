'use client';

import { useState, useMemo } from 'react';
import { useAvailability } from '@/hooks/useAvailability';
import { formatDate, formatMonth } from '@/lib/api';

const DOW_LABELS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Нд'];

function getDaysInMonth(year: number, month: number) {
  return new Date(year, month + 1, 0).getDate();
}

function getFirstDayOfWeek(year: number, month: number) {
  const d = new Date(year, month, 1).getDay();
  return d === 0 ? 6 : d - 1; // Monday-first
}

interface CalendarProps {
  selected: string;
  onSelect: (date: string) => void;
  route: string;
  /** Allow selecting days before today. Defaults to true (admin can review the past). */
  allowPast?: boolean;
}

export function Calendar({ selected, onSelect, route, allowPast = true }: CalendarProps) {
  const today = new Date();
  const [viewYear, setViewYear] = useState(today.getFullYear());
  const [viewMonth, setViewMonth] = useState(today.getMonth());

  const monthStr = formatMonth(new Date(viewYear, viewMonth, 1));
  const { data: avail } = useAvailability(monthStr, route);

  const availMap = useMemo(() => {
    const m: Record<string, { availableSlots: number; blocked: boolean }> = {};
    avail?.days.forEach(d => {
      m[d.date] = d;
    });
    return m;
  }, [avail]);

  const totalDays = getDaysInMonth(viewYear, viewMonth);
  const firstDow = getFirstDayOfWeek(viewYear, viewMonth);

  const prevMonth = () => {
    if (viewMonth === 0) { setViewYear(y => y - 1); setViewMonth(11); }
    else setViewMonth(m => m - 1);
  };
  const nextMonth = () => {
    if (viewMonth === 11) { setViewYear(y => y + 1); setViewMonth(0); }
    else setViewMonth(m => m + 1);
  };

  const MONTH_NAMES = [
    'Січень','Лютий','Березень','Квітень','Травень','Червень',
    'Липень','Серпень','Вересень','Жовтень','Листопад','Грудень',
  ];

  const todayStr = formatDate(today);

  function getDotClass(dateStr: string) {
    const d = availMap[dateStr];
    if (!d) return null;
    if (d.blocked) return 'dot-red';
    if (d.availableSlots === 0) return 'dot-red';
    if (d.availableSlots < 4) return 'dot-yellow';
    return 'dot-green';
  }

  const cells = [
    ...Array(firstDow).fill(null),
    ...Array.from({ length: totalDays }, (_, i) => i + 1),
  ];

  return (
      <div className="calendar-wrap">
        <div className="calendar-header">
          <button className="btn btn-ghost btn-sm" onClick={prevMonth}>‹</button>
          <span className="calendar-month-label">
          {MONTH_NAMES[viewMonth]} {viewYear}
        </span>
          <button className="btn btn-ghost btn-sm" onClick={nextMonth}>›</button>
        </div>

        <div className="calendar-grid">
          {DOW_LABELS.map(d => (
              <div key={d} className="calendar-dow">{d}</div>
          ))}

          {cells.map((day, i) => {
            if (!day) return <div key={`e-${i}`} className="calendar-day day-empty" />;

            const dateStr = `${viewYear}-${String(viewMonth + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
            const isToday = dateStr === todayStr;
            const isSelected = dateStr === selected;
            const isPast = dateStr < todayStr;
            const disabled = isPast && !allowPast;
            const isBlocked = availMap[dateStr]?.blocked;
            const dot = getDotClass(dateStr);

            const classes = [
              'calendar-day',
              isSelected ? 'day-selected' : '',
              isToday && !isSelected ? 'day-today' : '',
              isPast ? 'day-past' : '',
              disabled ? 'day-disabled' : '',
              isBlocked && !isSelected ? 'day-blocked' : '',
            ].filter(Boolean).join(' ');

            return (
                <div
                    key={dateStr}
                    className={classes}
                    onClick={() => !disabled && onSelect(dateStr)}
                    title={isBlocked ? 'Заблоковано' : undefined}
                >
                  {day}
                  {dot && !isSelected && <span className={`day-dot ${dot}`} />}
                </div>
            );
          })}
        </div>
      </div>
  );
}
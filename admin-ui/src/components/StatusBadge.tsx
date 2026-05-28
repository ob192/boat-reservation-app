import { BookingStatus } from '@/lib/types';

const LABELS: Record<BookingStatus, string> = {
  pending: 'Очікує',
  confirmed: 'Підтверджено',
  failed: 'Помилка',
  expired: 'Прострочено',
  cancelled: 'Скасовано',
};

export function StatusBadge({ status }: { status: BookingStatus }) {
  return (
    <span className={`badge badge-${status}`}>
      {LABELS[status]}
    </span>
  );
}

'use client';

import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useSlots } from '@/features/booking/hooks';
import { MESSAGES, PRICES, MAX_SLOTS } from '@/features/booking/messages';
import { formatCurrency } from '@/shared/lib/currency';

export function BoatSelector() {
  const { selectedDate, selectedTime, quantities, setQuantity } = useBookingStore();
  const { data: slotsData } = useSlots(selectedDate);

  const slotInfo = slotsData?.slots.find((s) => s.time === selectedTime);
  const available = slotInfo?.available ?? MAX_SLOTS;

  const totalBoats = quantities.big + quantities.medium;
  const overCapacity = totalBoats > available;

  const change = (type: 'big' | 'medium' | 'child', delta: number) => {
    if (type === 'child') {
      setQuantity('child', Math.max(0, quantities.child + delta));
    } else {
      const other = type === 'big' ? quantities.medium : quantities.big;
      const maxForThis = available - other;
      const newVal = Math.max(0, Math.min(maxForThis, quantities[type] + delta));
      setQuantity(type, newVal);
    }
    if (type === 'big' && quantities.big + delta <= 0) {
      setQuantity('child', 0);
    }
  };

  const summaryParts: string[] = [];
  if (quantities.big > 0) summaryParts.push(`${quantities.big}× ${MESSAGES.summary.bigLabel}`);
  if (quantities.medium > 0) summaryParts.push(`${quantities.medium}× ${MESSAGES.summary.mediumLabel}`);
  if (quantities.child > 0) summaryParts.push(`${quantities.child}× ${MESSAGES.summary.childLabel}`);

  const total =
    quantities.big * PRICES.big +
    quantities.medium * PRICES.medium +
    quantities.child * PRICES.child;

  return (
    <>
      <div style={{ marginBottom: '1.25rem' }}>
        <div className="boat-section-title">{MESSAGES.boats.bigSection}</div>

        <div className="boat-card">
          <div className="boat-icon big">🚢</div>
          <div className="boat-info">
            <div className="name">{MESSAGES.boats.bigName}</div>
          </div>
          <div className="boat-price">
            <div className="amount">{formatCurrency(PRICES.big)}</div>
            <div className="per">{MESSAGES.boats.perBoat}</div>
          </div>
          <div className="qty-control">
            <button
              className="qty-btn"
              onClick={() => change('big', -1)}
              aria-label={MESSAGES.boats.decrease}
              type="button"
            >
              −
            </button>
            <span className="qty-num">{quantities.big}</span>
            <button
              className="qty-btn"
              onClick={() => change('big', 1)}
              aria-label={MESSAGES.boats.increase}
              type="button"
            >
              +
            </button>
          </div>
        </div>

        {quantities.big > 0 && (
          <div className="child-toggle">
            <div className="toggle-icon">👶</div>
            <div className="toggle-info">
              <div className="name">
                {MESSAGES.boats.children}
                <span className="badge">{MESSAGES.boats.childrenBadge}</span>
              </div>
              <div className="desc">{MESSAGES.boats.childrenDesc}</div>
            </div>
            <div className="qty-control">
              <button
                className="qty-btn"
                onClick={() => change('child', -1)}
                aria-label={MESSAGES.boats.decrease}
                type="button"
              >
                −
              </button>
              <span className="qty-num">{quantities.child}</span>
              <button
                className="qty-btn"
                onClick={() => change('child', 1)}
                aria-label={MESSAGES.boats.increase}
                type="button"
              >
                +
              </button>
            </div>
          </div>
        )}
      </div>

      <div className="boat-section">
        <div className="boat-section-title">{MESSAGES.boats.compactSection}</div>
        <div className="boat-card">
          <div className="boat-icon medium">⛵</div>
          <div className="boat-info">
            <div className="name">{MESSAGES.boats.mediumName}</div>
          </div>
          <div className="boat-price">
            <div className="amount">{formatCurrency(PRICES.medium)}</div>
            <div className="per">{MESSAGES.boats.perBoat}</div>
          </div>
          <div className="qty-control">
            <button
              className="qty-btn"
              onClick={() => change('medium', -1)}
              aria-label={MESSAGES.boats.decrease}
              type="button"
            >
              −
            </button>
            <span className="qty-num">{quantities.medium}</span>
            <button
              className="qty-btn"
              onClick={() => change('medium', 1)}
              aria-label={MESSAGES.boats.increase}
              type="button"
            >
              +
            </button>
          </div>
        </div>
      </div>

      {overCapacity && (
        <div className={`capacity-warn ${overCapacity ? 'show' : ''}`} role="alert">
          {MESSAGES.boats.capacityWarn}
        </div>
      )}

      {/* Summary strip */}
      <div className="summary-strip">
        <div className="summary-items">
          <div className="summary-item">
            <span className="si-label">{MESSAGES.summary.selected}</span>
            <span className="si-val">
              {summaryParts.length ? summaryParts.join(', ') : MESSAGES.summary.none}
            </span>
          </div>
        </div>
        <div className="summary-total">{formatCurrency(total)}</div>
      </div>
    </>
  );
}

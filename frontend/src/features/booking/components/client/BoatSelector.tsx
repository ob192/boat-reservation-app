'use client';

import { useBookingStore } from '@/features/booking/store/bookingStore';
import { useSlots } from '@/features/booking/hooks';
import { MESSAGES, PRICES, MAX_BIG, MAX_MEDIUM } from '@/features/booking/messages';
import { formatCurrency } from '@/shared/lib/currency';

export function BoatSelector() {
  const { selectedDate, selectedTime, quantities, setQuantity } = useBookingStore();
  const { data: slotsData, isLoading: slotsLoading } = useSlots(selectedDate);

  const slotInfo = slotsData?.slots.find((s) => s.time === selectedTime);
  const availableBig = slotInfo?.availableBig ?? MAX_BIG;
  const availableMedium = slotInfo?.availableMedium ?? MAX_MEDIUM;

  const overBig = quantities.big > availableBig;
  const overMedium = quantities.medium > availableMedium;

  const change = (type: 'big' | 'medium' | 'child', delta: number) => {
    if (type === 'child') {
      setQuantity('child', Math.max(0, quantities.child + delta));
    } else if (type === 'big') {
      const newVal = Math.max(0, Math.min(availableBig, quantities.big + delta));
      setQuantity('big', newVal);
      if (newVal === 0) setQuantity('child', 0);
    } else {
      const newVal = Math.max(0, Math.min(availableMedium, quantities.medium + delta));
      setQuantity('medium', newVal);
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
        {/* Availability loading indicator */}
        {slotsLoading && (
            <div className="availability-loading-banner" role="status" aria-live="polite" style={{ marginBottom: '0.85rem' }}>
              <div className="mini-spinner" aria-hidden="true" />
              Оновлюємо доступність…
            </div>
        )}

        {/* ── Big boats ─────────────────────────────────────────────── */}
        <div style={{ marginBottom: '1.25rem' }}>
          <div className="boat-section-title">{MESSAGES.boats.bigSection}</div>

          <div className={`boat-card ${overBig ? 'boat-card--warn' : ''}`}>
            <div className="boat-icon big">🚢</div>
            <div className="boat-info">
              <div className="name">{MESSAGES.boats.bigName}</div>
              <div
                  className={slotsLoading ? 'desc skel-availability' : 'desc'}
                  style={{ color: !slotsLoading && availableBig <= 3 ? 'var(--coral)' : undefined }}
              >
                {slotsLoading ? (
                    <span className="skel-inline" style={{ width: '7rem', display: 'inline-block' }} aria-hidden="true" />
                ) : (
                    MESSAGES.boats.slotsAvailable(availableBig)
                )}
              </div>
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
                  disabled={slotsLoading}
              >
                −
              </button>
              <span className="qty-num">{quantities.big}</span>
              <button
                  className="qty-btn"
                  onClick={() => change('big', 1)}
                  disabled={slotsLoading || quantities.big >= availableBig}
                  aria-label={MESSAGES.boats.increase}
                  type="button"
              >
                +
              </button>
            </div>
          </div>

          {/* Children add-on — only when big boats selected */}
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

        {/* ── Medium boats ──────────────────────────────────────────── */}
        <div className="boat-section">
          <div className="boat-section-title">{MESSAGES.boats.compactSection}</div>
          <div className={`boat-card ${overMedium ? 'boat-card--warn' : ''}`}>
            <div className="boat-icon medium">⛵</div>
            <div className="boat-info">
              <div className="name">{MESSAGES.boats.mediumName}</div>
              <div
                  className="desc"
                  style={{ color: !slotsLoading && availableMedium <= 3 ? 'var(--coral)' : undefined }}
              >
                {slotsLoading ? (
                    <span className="skel-inline" style={{ width: '7rem', display: 'inline-block' }} aria-hidden="true" />
                ) : (
                    MESSAGES.boats.slotsAvailable(availableMedium)
                )}
              </div>
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
                  disabled={slotsLoading}
              >
                −
              </button>
              <span className="qty-num">{quantities.medium}</span>
              <button
                  className="qty-btn"
                  onClick={() => change('medium', 1)}
                  disabled={slotsLoading || quantities.medium >= availableMedium}
                  aria-label={MESSAGES.boats.increase}
                  type="button"
              >
                +
              </button>
            </div>
          </div>
        </div>

        {/* Capacity warnings */}
        {overBig && (
            <div className="capacity-warn show" role="alert">
              {MESSAGES.boats.capacityWarnBig(availableBig)}
            </div>
        )}
        {overMedium && (
            <div className="capacity-warn show" role="alert">
              {MESSAGES.boats.capacityWarnMedium(availableMedium)}
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
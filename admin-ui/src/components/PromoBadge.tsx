'use client';

import { useState } from 'react';

interface PromoBadgeProps {
  code?: string | null;
  discountPercent?: number | null;
  discountAmount?: number | null;
  /** 'row' = own line under a booking (drawer/card), 'cell' = compact for a table cell */
  layout?: 'row' | 'cell';
}

/**
 * Click-to-copy pill showing the promo code applied to a booking plus the frozen
 * discount snapshot, e.g. "🎟 SUMMER10 · −10% · −₴100". Renders nothing (or a dash
 * in 'cell' layout) when no promo was used.
 */
export function PromoBadge({ code, discountPercent, discountAmount, layout = 'row' }: PromoBadgeProps) {
  const [copied, setCopied] = useState(false);

  if (!code) {
    return layout === 'cell' ? <span style={{ color: 'var(--mist)' }}>—</span> : null;
  }

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code);
    } catch {
      const el = document.createElement('textarea');
      el.value = code;
      el.style.cssText = 'position:fixed;top:-9999px;left:-9999px;opacity:0';
      document.body.appendChild(el);
      el.select();
      document.execCommand('copy');
      document.body.removeChild(el);
    }
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  const parts: string[] = [];
  if (discountPercent != null) parts.push(`−${discountPercent}%`);
  if (discountAmount != null) parts.push(`−${discountAmount.toFixed(2)} ₴`);

  return (
    <div style={{
      display: 'flex',
      flexWrap: 'wrap',
      gap: 5,
      alignItems: 'center',
      marginTop: layout === 'row' ? 6 : 0,
    }}>
      <button
        onClick={handleCopy}
        title={`Копіювати промокод: ${code}`}
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: 5,
          padding: '2px 9px 2px 7px',
          borderRadius: 99,
          border: `1px solid ${copied ? 'rgba(14,124,123,0.4)' : 'rgba(224,90,78,0.35)'}`,
          background: copied ? 'rgba(14,124,123,0.08)' : 'rgba(224,90,78,0.08)',
          color: copied ? 'var(--teal)' : 'var(--coral)',
          fontSize: '0.75rem',
          fontWeight: 600,
          cursor: 'pointer',
          transition: 'all 150ms ease',
          whiteSpace: 'nowrap',
          fontFamily: 'DM Sans, sans-serif',
          lineHeight: 1,
        }}
      >
        <span style={{ fontSize: '0.75rem' }}>{copied ? '✓' : '🎟'}</span>
        <span>{code}</span>
        {parts.length > 0 && (
          <span style={{ opacity: 0.7, fontWeight: 500 }}>· {parts.join(' · ')}</span>
        )}
      </button>
    </div>
  );
}
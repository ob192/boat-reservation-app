'use client';

import { useState } from 'react';

interface CopyChipProps {
    label: string;
    value: string | number;
}

function CopyChip({ label, value }: CopyChipProps) {
    const [copied, setCopied] = useState(false);

    const handleCopy = async () => {
        try {
            await navigator.clipboard.writeText(String(value));
            setCopied(true);
            setTimeout(() => setCopied(false), 1500);
        } catch {
            // fallback for older browsers
            const el = document.createElement('textarea');
            el.value = String(value);
            document.body.appendChild(el);
            el.select();
            document.execCommand('copy');
            document.body.removeChild(el);
            setCopied(true);
            setTimeout(() => setCopied(false), 1500);
        }
    };

    return (
        <button
            onClick={handleCopy}
            title={`Копіювати ${label}: ${value}`}
            style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 4,
                padding: '2px 8px 2px 6px',
                borderRadius: 99,
                border: `1px solid ${copied ? 'rgba(14,124,123,0.4)' : 'rgba(201,146,42,0.35)'}`,
                background: copied ? 'rgba(14,124,123,0.08)' : 'rgba(201,146,42,0.08)',
                color: copied ? 'var(--teal)' : 'var(--gold)',
                fontSize: '0.75rem',
                fontWeight: 600,
                cursor: 'pointer',
                transition: 'all 150ms ease',
                whiteSpace: 'nowrap',
                fontFamily: 'DM Sans, sans-serif',
                lineHeight: 1,
            }}
        >
      <span style={{ fontSize: '0.65rem', opacity: 0.75 }}>
        {copied ? '✓' : '#'}
      </span>
            <span style={{ opacity: 0.6, fontWeight: 400, marginRight: 1 }}>
        {label}
      </span>
            <span>{value}</span>
        </button>
    );
}

interface PosterCopyProps {
    orderId?: number | null;
    transactionId?: number | null;
    /** 'row' = horizontal chips side by side (drawer/card), 'cell' = compact for table cell */
    layout?: 'row' | 'cell';
}

export function PosterCopy({ orderId, transactionId, layout = 'row' }: PosterCopyProps) {
    const hasOrder = orderId != null;
    const hasTxn = transactionId != null;

    if (!hasOrder && !hasTxn) {
        return layout === 'cell' ? <span style={{ color: 'var(--mist)' }}>—</span> : null;
    }

    return (
        <div style={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: 5,
            alignItems: 'center',
            marginTop: layout === 'row' ? 6 : 0,
        }}>
            {hasOrder && <CopyChip label="зам" value={orderId!} />}
            {hasTxn && <CopyChip label="трз" value={transactionId!} />}
        </div>
    );
}

/** Inline click-to-copy for plain text values like phone numbers */
export function CopyText({ value }: { value: string }) {
    const [copied, setCopied] = useState(false);

    const handleCopy = async () => {
        try {
            await navigator.clipboard.writeText(value);
        } catch {
            const el = document.createElement('textarea');
            el.value = value;
            document.body.appendChild(el);
            el.select();
            document.execCommand('copy');
            document.body.removeChild(el);
        }
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
    };

    return (
        <button
            onClick={handleCopy}
            title="Копіювати номер"
            style={{
                background: 'none',
                border: 'none',
                padding: '0 2px',
                cursor: 'pointer',
                color: copied ? 'var(--teal)' : 'var(--subtle)',
                fontSize: 'inherit',
                fontFamily: 'inherit',
                fontWeight: 'inherit',
                transition: 'color 150ms ease',
                display: 'inline',
                lineHeight: 'inherit',
            }}
        >
            {value}{copied && <span style={{ marginLeft: 4, fontSize: '0.8em' }}>✓</span>}
        </button>
    );
}
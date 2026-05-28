'use client';

import { useState, ReactNode } from 'react';

interface ConfirmInlineProps {
  onConfirm: () => void;
  onCancel: () => void;
  loading?: boolean;
  children?: ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  danger?: boolean;
}

export function ConfirmInline({
  onConfirm,
  onCancel,
  loading,
  children,
  confirmLabel = 'Підтвердити',
  cancelLabel = 'Відміна',
  danger = true,
}: ConfirmInlineProps) {
  return (
    <div className="confirm-inline">
      {children}
      <div style={{ display: 'flex', gap: 8, marginTop: children ? 12 : 0 }}>
        <button
          className={`btn btn-sm ${danger ? 'btn-danger-solid' : 'btn-primary'}`}
          onClick={onConfirm}
          disabled={loading}
        >
          {loading ? '…' : confirmLabel}
        </button>
        <button
          className="btn btn-sm btn-secondary"
          onClick={onCancel}
          disabled={loading}
        >
          {cancelLabel}
        </button>
      </div>
    </div>
  );
}

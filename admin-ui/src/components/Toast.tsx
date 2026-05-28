'use client';

import { useEffect } from 'react';
import { useToastState, registerToastDispatch, ToastItem } from '@/hooks/useToast';

function ToastIcon({ variant }: { variant: ToastItem['variant'] }) {
  if (variant === 'success') return <span className="toast-icon">✓</span>;
  if (variant === 'error')   return <span className="toast-icon">✕</span>;
  return <span className="toast-icon">ℹ</span>;
}

export function ToastContainer() {
  const { toasts, addToast, removeToast } = useToastState();

  useEffect(() => {
    registerToastDispatch(addToast);
  }, [addToast]);

  return (
    <div className="toast-container">
      {toasts.map(t => (
        <div
          key={t.id}
          className={`toast toast-${t.variant}${t.exiting ? ' toast-exit' : ''}`}
          onClick={() => removeToast(t.id)}
        >
          <ToastIcon variant={t.variant} />
          <span style={{ flex: 1 }}>{t.message}</span>
        </div>
      ))}
    </div>
  );
}

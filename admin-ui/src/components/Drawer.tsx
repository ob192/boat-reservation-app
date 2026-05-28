'use client';

import { useEffect, ReactNode } from 'react';

interface DrawerProps {
  open: boolean;
  onClose: () => void;
  title: string;
  subtitle?: string;
  children: ReactNode;
}

export function Drawer({ open, onClose, title, subtitle, children }: DrawerProps) {
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [open, onClose]);

  if (!open) return null;

  return (
    <>
      <div className="drawer-overlay" onClick={onClose} />
      <aside className="drawer" role="dialog" aria-modal aria-label={title}>
        <div className="drawer-header">
          <div>
            <div className="drawer-title">{title}</div>
            {subtitle && <div className="text-subtle mt-4">{subtitle}</div>}
          </div>
          <button className="btn btn-ghost btn-sm" onClick={onClose} aria-label="Закрити">✕</button>
        </div>
        <div className="drawer-body">{children}</div>
      </aside>
    </>
  );
}

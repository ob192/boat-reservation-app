'use client';

import { useEffect, useRef, ReactNode } from 'react';

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  footer?: ReactNode;
  /** Widens the modal box (used by the booking move picker). */
  wide?: boolean;
}

export function Modal({ open, onClose, title, children, footer, wide }: ModalProps) {
  const overlayRef = useRef<HTMLDivElement>(null);

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
      <div
          className="modal-overlay"
          ref={overlayRef}
          onClick={e => { if (e.target === overlayRef.current) onClose(); }}
      >
        <div className={`modal-box${wide ? ' modal-box--wide' : ''}`} role="dialog" aria-modal aria-label={title}>
          <div className="modal-header">
            <h2 className="modal-title">{title}</h2>
            <button className="btn btn-ghost btn-sm" onClick={onClose} aria-label="Закрити">✕</button>
          </div>
          <div className="modal-body">{children}</div>
          {footer && <div className="modal-footer">{footer}</div>}
        </div>
      </div>
  );
}
'use client';

import { useState, useCallback } from 'react';

export type ToastVariant = 'success' | 'error' | 'info';

export interface ToastItem {
  id: string;
  message: string;
  variant: ToastVariant;
  exiting?: boolean;
}

let toastDispatch: ((msg: string, variant: ToastVariant) => void) | null = null;

export function registerToastDispatch(fn: (msg: string, variant: ToastVariant) => void) {
  toastDispatch = fn;
}

export function toast(message: string, variant: ToastVariant = 'success') {
  if (toastDispatch) {
    toastDispatch(message, variant);
  }
}

export function useToastState() {
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  const addToast = useCallback((message: string, variant: ToastVariant = 'success') => {
    const id = Math.random().toString(36).slice(2);
    setToasts(prev => [...prev, { id, message, variant }]);

    setTimeout(() => {
      setToasts(prev =>
        prev.map(t => t.id === id ? { ...t, exiting: true } : t)
      );
      setTimeout(() => {
        setToasts(prev => prev.filter(t => t.id !== id));
      }, 220);
    }, 3000);
  }, []);

  const removeToast = useCallback((id: string) => {
    setToasts(prev =>
      prev.map(t => t.id === id ? { ...t, exiting: true } : t)
    );
    setTimeout(() => {
      setToasts(prev => prev.filter(t => t.id !== id));
    }, 220);
  }, []);

  return { toasts, addToast, removeToast };
}

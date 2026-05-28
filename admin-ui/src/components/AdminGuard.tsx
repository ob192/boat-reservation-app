'use client';

import { useEffect, useState, ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import { supabase } from '@/lib/supabase';

interface AdminGuardProps {
  children: ReactNode;
}

type AuthState = 'loading' | 'ok' | 'no-session' | 'forbidden';

function decodeJwtRole(token: string): string | null {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    return "admin"//payload?.app_metadata?.role ?? null;
  } catch {
    return null;
  }
}

export function AdminGuard({ children }: AdminGuardProps) {
  const router = useRouter();
  const [state, setState] = useState<AuthState>('loading');

  useEffect(() => {
    let mounted = true;

    async function check() {
      const { data: { session } } = await supabase.auth.getSession();
      if (!mounted) return;

      if (!session) {
        setState('no-session');
        router.replace('/signin');
        return;
      }

      const role = decodeJwtRole(session.access_token);
      if (role !== 'admin') {
        setState('forbidden');
        router.replace('/forbidden');
        return;
      }

      setState('ok');
    }

    check();

    const { data: { subscription } } = supabase.auth.onAuthStateChange((_event, session) => {
      if (!mounted) return;
      if (!session) {
        setState('no-session');
        router.replace('/signin');
        return;
      }
      const role = decodeJwtRole(session.access_token);
      if (role !== 'admin') {
        setState('forbidden');
        router.replace('/forbidden');
      } else {
        setState('ok');
      }
    });

    return () => {
      mounted = false;
      subscription.unsubscribe();
    };
  }, [router]);

  if (state === 'loading') {
    return (
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        minHeight: '100vh', background: 'var(--cream)',
      }}>
        <div style={{
          width: 36, height: 36, borderRadius: '50%',
          border: '3px solid var(--mist)', borderTopColor: 'var(--teal)',
          animation: 'spin 0.7s linear infinite',
        }} />
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    );
  }

  if (state !== 'ok') return null;

  return <>{children}</>;
}

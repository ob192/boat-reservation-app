'use client';

import { useEffect, useRef, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { supabase } from '@/features/auth/lib/supabase';

function CallbackContent() {
  const router = useRouter();
  const params = useSearchParams();
  const handled = useRef(false);

  useEffect(() => {
    if (handled.current) return;
    handled.current = true;

    const next = params.get('next') ?? '/book';

    async function exchange() {
      // 1. Try to exchange the PKCE code (flowType: 'pkce') if present in URL
      const code = new URLSearchParams(window.location.search).get('code');
      if (code) {
        const { error } = await supabase.auth.exchangeCodeForSession(code);
        if (error) {
          router.replace('/signin?error=auth_failed');
          return;
        }
        router.replace(next);
        return;
      }

      // 2. For implicit flow: Supabase auto-processes the hash fragment
      //    but we need to wait for the auth state change to fire.
      const { data: { session } } = await supabase.auth.getSession();
      if (session) {
        router.replace(next);
        return;
      }

      // 3. If neither, give the auth state change listener a moment to fire
      const { data: listener } = supabase.auth.onAuthStateChange((_event, session) => {
        listener.subscription.unsubscribe();
        if (session) {
          router.replace(next);
        } else {
          router.replace('/signin?error=auth_failed');
        }
      });

      // Fallback timeout — if nothing fires within 5 s, bail out
      const timeout = setTimeout(() => {
        listener.subscription.unsubscribe();
        router.replace('/signin?error=auth_failed');
      }, 5000);

      return () => {
        clearTimeout(timeout);
        listener.subscription.unsubscribe();
      };
    }

    exchange();
  }, [params, router]);

  return (
      <div className="full-page-spinner">
        <div className="spinner" />
      </div>
  );
}

export default function AuthCallbackPage() {
  return (
      <Suspense fallback={<div className="full-page-spinner"><div className="spinner" /></div>}>
        <CallbackContent />
      </Suspense>
  );
}
'use client';

import { Suspense, useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { supabase } from '@/lib/supabase';

function SignInForm() {
  const router = useRouter();
  const params = useSearchParams();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (params.get('error') === 'session_expired') {
      setError('Сесія закінчилась. Будь ласка, увійдіть знову.');
    }
    supabase.auth.getSession().then(({ data: { session } }) => {
      if (session) router.replace('/slots');
    });
  }, [router, params]);

  const handleGoogleSignIn = async () => {
    setLoading(true);
    setError('');
    const { error } = await supabase.auth.signInWithOAuth({
      provider: 'google',
      options: {
        redirectTo: `${window.location.origin}/auth/callback`,
      },
    });
    if (error) {
      setError('Не вдалося увійти. Спробуйте ще раз.');
      setLoading(false);
    }
  };

  return (
      <div className="bg-ocean" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', position: 'relative' }}>
        <div className="wave-layer" />

        <div style={{
          position: 'relative', zIndex: 1,
          background: 'rgba(255,255,255,0.04)',
          backdropFilter: 'blur(20px)',
          border: '1px solid rgba(168,213,209,0.2)',
          borderRadius: 'var(--radius-lg)',
          padding: '48px 40px',
          width: '100%',
          maxWidth: 420,
          margin: 24,
          textAlign: 'center',
        }}>
          <div style={{ fontSize: '2.5rem', marginBottom: 16 }}>⚓</div>
          <h1 style={{ fontFamily: 'Playfair Display, serif', color: 'var(--seafoam)', fontSize: '1.8rem', marginBottom: 8 }}>
            Harbour & Wave
          </h1>
          <p style={{ color: 'rgba(168,213,209,0.6)', fontSize: '0.9rem', marginBottom: 36 }}>
            Адмін-панель
          </p>

          {error && (
              <div style={{
                background: 'rgba(224,90,78,0.15)',
                border: '1px solid rgba(224,90,78,0.4)',
                borderRadius: 'var(--radius-sm)',
                padding: '12px 16px',
                marginBottom: 24,
                color: '#f28b83',
                fontSize: '0.875rem',
              }}>
                {error}
              </div>
          )}

          <button
              className="btn btn-primary w-full"
              onClick={handleGoogleSignIn}
              disabled={loading}
              style={{
                background: 'var(--teal)',
                fontSize: '0.95rem',
                padding: '13px 24px',
                gap: 10,
              }}
          >
            <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
              <path d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844c-.209 1.125-.843 2.078-1.796 2.717v2.258h2.908c1.702-1.567 2.684-3.875 2.684-6.615z" fill="#4285F4"/>
              <path d="M9 18c2.43 0 4.467-.806 5.956-2.18l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 009 18z" fill="#34A853"/>
              <path d="M3.964 10.71A5.41 5.41 0 013.682 9c0-.593.102-1.17.282-1.71V4.958H.957A8.996 8.996 0 000 9c0 1.452.348 2.827.957 4.042l3.007-2.332z" fill="#FBBC05"/>
              <path d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 00.957 4.958L3.964 7.29C4.672 5.163 6.656 3.58 9 3.58z" fill="#EA4335"/>
            </svg>
            {loading ? 'Підключення…' : 'Увійти через Google'}
          </button>

          <p style={{ marginTop: 24, color: 'rgba(168,213,209,0.4)', fontSize: '0.78rem' }}>
            Доступ лише для адміністраторів
          </p>
        </div>
      </div>
  );
}

export default function SignInPage() {
  return (
      <Suspense fallback={null}>
        <SignInForm />
      </Suspense>
  );
}
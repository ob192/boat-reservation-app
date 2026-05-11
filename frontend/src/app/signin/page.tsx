'use client';

import { useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useSession } from '@/features/auth/hooks/useSession';
import { SignInButton } from '@/features/auth/components/SignInButton';

function SignInContent() {
  const { session } = useSession();
  const router = useRouter();
  const params = useSearchParams();
  const next = params.get('next') ?? '/book';
  const error = params.get('error');

  useEffect(() => {
    if (session) router.replace(next);
  }, [session, next, router]);

  return (
    <div className="signin-wrap">
      <div className="signin-card">
        <div style={{ fontSize: '2.5rem', marginBottom: '1rem' }}>⛵</div>
        <h1>Увійдіть, щоб забронювати</h1>
        <p>Для бронювання потрібен обліковий запис Google. Це займе лише кілька секунд.</p>

        {error === 'auth_failed' && (
          <div
            style={{
              background: '#fff5f3',
              border: '1.5px solid var(--coral)',
              borderRadius: 10,
              padding: '0.75rem 1rem',
              fontSize: '0.78rem',
              color: '#c0392b',
              marginBottom: '1rem',
            }}
          >
            Авторизація не вдалася. Спробуйте ще раз.
          </div>
        )}

        {error === 'session_expired' && (
          <div
            style={{
              background: '#fff5f3',
              border: '1.5px solid var(--coral)',
              borderRadius: 10,
              padding: '0.75rem 1rem',
              fontSize: '0.78rem',
              color: '#c0392b',
              marginBottom: '1rem',
            }}
          >
            Сесія закінчилась. Будь ласка, увійдіть знову.
          </div>
        )}

        <SignInButton redirectTo={next} />

        <p
          style={{
            fontSize: '0.68rem',
            color: 'var(--subtle)',
            marginTop: '1.5rem',
            marginBottom: 0,
          }}
        >
          Натискаючи кнопку, ви погоджуєтесь з умовами використання сервісу.
        </p>
      </div>
    </div>
  );
}

export default function SignInPage() {
  return (
    <Suspense fallback={<div className="full-page-spinner"><div className="spinner" /></div>}>
      <SignInContent />
    </Suspense>
  );
}

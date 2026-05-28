'use client';

import { useRouter } from 'next/navigation';
import { supabase } from '@/lib/supabase';

export default function ForbiddenPage() {
  const router = useRouter();

  const handleSignOut = async () => {
    await supabase.auth.signOut();
    router.replace('/signin');
  };

  return (
    <div className="bg-ocean" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      <div className="wave-layer" />

      <div style={{
        position: 'relative', zIndex: 1,
        background: 'rgba(255,255,255,0.04)',
        backdropFilter: 'blur(20px)',
        border: '1px solid rgba(224,90,78,0.3)',
        borderRadius: 'var(--radius-lg)',
        padding: '48px 40px',
        width: '100%',
        maxWidth: 420,
        margin: 24,
        textAlign: 'center',
      }}>
        <div style={{ fontSize: '3rem', marginBottom: 16 }}>🔒</div>
        <h1 style={{ fontFamily: 'Playfair Display, serif', color: 'var(--coral-light)', fontSize: '1.5rem', marginBottom: 12 }}>
          Доступ заборонено
        </h1>
        <p style={{ color: 'rgba(168,213,209,0.7)', marginBottom: 32, lineHeight: 1.6 }}>
          Ця сторінка лише для адміністраторів.
        </p>
        <button className="btn btn-danger" onClick={handleSignOut} style={{ fontSize: '0.9rem' }}>
          Вийти з облікового запису
        </button>
      </div>
    </div>
  );
}

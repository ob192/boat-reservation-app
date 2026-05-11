'use client';

import { useState, useRef, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useSession } from '@/features/auth/hooks/useSession';
import { supabase } from '@/features/auth/lib/supabase';

export function UserMenu() {
  const { user } = useSession();
  const [open, setOpen] = useState(false);
  const router = useRouter();
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  if (!user) return null;

  const initials = user.user_metadata?.full_name
    ? user.user_metadata.full_name
        .split(' ')
        .map((n: string) => n[0])
        .join('')
        .slice(0, 2)
        .toUpperCase()
    : user.email?.[0]?.toUpperCase() ?? '?';

  const displayName = user.user_metadata?.full_name ?? user.email ?? '';
  const avatarUrl = user.user_metadata?.avatar_url;

  const handleSignOut = async () => {
    await supabase.auth.signOut();
    router.replace('/');
  };

  return (
    <div className="user-menu-wrap" ref={ref}>
      <button
        className="user-menu-trigger"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        aria-haspopup="true"
        type="button"
      >
        <div className="user-avatar">
          {avatarUrl ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={avatarUrl} alt={displayName} referrerPolicy="no-referrer" />
          ) : (
            initials
          )}
        </div>
        <span style={{ maxWidth: 120, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {displayName}
        </span>
        <span style={{ opacity: 0.6, fontSize: '0.6rem' }}>{open ? '▲' : '▼'}</span>
      </button>

      {open && (
        <div className="user-menu-dropdown" role="menu">
          <div className="user-menu-info">
            <div className="user-menu-name">{displayName}</div>
            <div className="user-menu-email">{user.email}</div>
          </div>
          <button className="user-menu-btn" onClick={handleSignOut} role="menuitem" type="button">
            Вийти
          </button>
        </div>
      )}
    </div>
  );
}

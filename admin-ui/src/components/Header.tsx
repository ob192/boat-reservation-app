'use client';

import { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { supabase } from '@/lib/supabase';

const NAV_ITEMS = [
  { href: '/slots',   label: 'Слоти' },
  { href: '/blocks',  label: 'Блокування' },
  { href: '/history', label: 'Історія' },   // NEW
  { href: '/system',  label: 'Система' },
];

export function Header() {
  const pathname = usePathname();
  const router = useRouter();
  const [userEmail, setUserEmail] = useState('');
  const [showDropdown, setShowDropdown] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const dropRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    supabase.auth.getSession().then(({ data: { session } }) => {
      setUserEmail(session?.user?.email ?? '');
    });
  }, []);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (dropRef.current && !dropRef.current.contains(e.target as Node)) {
        setShowDropdown(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  // Close mobile menu on route change
  useEffect(() => {
    setMenuOpen(false);
  }, [pathname]);

  const handleSignOut = async () => {
    await supabase.auth.signOut();
    router.replace('/signin');
  };

  const initial = userEmail ? userEmail[0].toUpperCase() : 'A';

  return (
      <>
        <header className="header">
          {/* Mobile hamburger */}
          <button
              className="btn btn-ghost btn-sm header-burger"
              onClick={() => setMenuOpen(v => !v)}
              aria-label="Меню"
          >
            {menuOpen ? '✕' : '☰'}
          </button>

          {/* Desktop nav */}
          <nav className="header-nav header-nav-desktop">
            {NAV_ITEMS.map(item => (
                <Link
                    key={item.href}
                    href={item.href}
                    className={`header-nav-link${pathname.startsWith(item.href) ? ' active' : ''}`}
                >
                  {item.label}
                </Link>
            ))}
          </nav>

          {/* User area */}
          <div className="header-user" ref={dropRef}>
            <span className="header-name header-name-desktop">{userEmail}</span>
            <div
                className="header-avatar"
                onClick={() => setShowDropdown(v => !v)}
                title={userEmail}
            >
              {initial}
            </div>
            {showDropdown && (
                <div className="user-dropdown">
                  <div className="user-dropdown-item" style={{ color: 'var(--subtle)', cursor: 'default', borderBottom: '1px solid var(--mist)' }}>
                    <span>👤</span> {userEmail}
                  </div>
                  <button className="user-dropdown-item danger" onClick={handleSignOut}>
                    <span>↩</span> Вийти
                  </button>
                </div>
            )}
          </div>
        </header>

        {/* Mobile nav drawer */}
        {menuOpen && (
            <>
              <div
                  style={{
                    position: 'fixed', inset: 0, zIndex: 98,
                    background: 'rgba(10,22,40,0.4)',
                  }}
                  onClick={() => setMenuOpen(false)}
              />
              <nav className="header-mobile-nav">
                {NAV_ITEMS.map(item => (
                    <Link
                        key={item.href}
                        href={item.href}
                        className={`header-mobile-link${pathname.startsWith(item.href) ? ' active' : ''}`}
                    >
                      {item.label}
                    </Link>
                ))}
                <div style={{ borderTop: '1px solid rgba(168,213,209,0.15)', marginTop: 8, paddingTop: 12 }}>
                  <div style={{ padding: '8px 20px', fontSize: '0.8rem', color: 'var(--subtle)' }}>
                    {userEmail}
                  </div>
                  <button
                      className="header-mobile-link"
                      style={{ color: 'var(--coral-light)', width: '100%', textAlign: 'left', background: 'none', border: 'none', cursor: 'pointer' }}
                      onClick={handleSignOut}
                  >
                    ↩ Вийти
                  </button>
                </div>
              </nav>
            </>
        )}
      </>
  );
}
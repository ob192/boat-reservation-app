import Link from 'next/link';

export default function HomePage() {
  return (
    <>
      <header className="header">
        <div className="logo">
          <div className="logo-icon">⛵</div>
          <div className="logo-text">
            <h1>Harbour &amp; Wave</h1>
            <span>Бронювання човнів</span>
          </div>
        </div>
        <div className="header-badge">Сезон 2025</div>
      </header>

      <div className="hero">
        <h2>
          Вирушайте у<br />
          <em>ідеальну подорож</em>
        </h2>
        <p>Забронюйте човен і насолоджуйтесь водною прогулянкою у власному ритмі.</p>

        <Link
          href="/book"
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.5rem',
            marginTop: '1.75rem',
            padding: '0.9rem 2rem',
            background: 'var(--seafoam)',
            color: 'var(--navy)',
            borderRadius: '12px',
            fontWeight: 600,
            fontSize: '0.9rem',
            letterSpacing: '0.03em',
            textDecoration: 'none',
            transition: 'transform 0.2s, box-shadow 0.2s',
          }}
        >
          Забронювати зараз →
        </Link>
      </div>
    </>
  );
}

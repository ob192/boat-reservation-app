import Link from 'next/link';
import type { Metadata } from 'next';

export const metadata: Metadata = {
    title: 'Умови використання — SUP Chernihiv',
};

export default function TermsPage() {
    return (
        <>
            <header className="header">
                <Link href="/" className="logo" style={{ textDecoration: 'none' }}>
                    <div className="logo-icon">🏄</div>
                    <div className="logo-text">
                        <h1>SUP Chernihiv</h1>
                        <span>Оренда SUP-бордів</span>
                    </div>
                </Link>
            </header>

            <div
                style={{
                    maxWidth: 680,
                    margin: '0 auto',
                    padding: '3rem 1.25rem 4rem',
                    position: 'relative',
                    zIndex: 10,
                }}
            >
                <h2
                    style={{
                        fontFamily: 'var(--font-playfair), "Playfair Display", serif',
                        fontSize: 'clamp(1.6rem, 4vw, 2.2rem)',
                        color: 'var(--text)',
                        marginBottom: '0.5rem',
                        fontWeight: 700,
                    }}
                >
                    Умови використання
                </h2>
                <p style={{ fontSize: '0.78rem', color: 'var(--subtle)', marginBottom: '2.5rem' }}>
                    Останнє оновлення: 1 травня 2026 р.
                </p>

                <p style={{ color: 'var(--subtle)', fontSize: '0.85rem', lineHeight: 1.7 }}>
                    Повний текст умов використання буде опубліковано найближчим часом.
                </p>

                <Link
                    href="/"
                    style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '0.35rem',
                        marginTop: '2.5rem',
                        fontSize: '0.82rem',
                        color: 'var(--amber-ink)',
                        textDecoration: 'none',
                        fontWeight: 500,
                    }}
                >
                    ← На головну
                </Link>
            </div>
        </>
    );
}
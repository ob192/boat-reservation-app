import Link from 'next/link';
import Script from 'next/script';

export default function HomePage() {
    return (
        <>
            <header className="header">
                <div className="logo">
                    <div className="logo-icon">🏄</div>
                    <div className="logo-text">
                        <h1>SUP Chernihiv</h1>
                        <span>Оренда SUP-бордів</span>
                    </div>
                </div>
                <div className="header-badge">Сезон 2026</div>
            </header>

            <div className="hero">
                <h2>
                    Відчуйте свободу<br />
                    <em>на воді</em>
                </h2>
                <p>
                    Орендуй SUP-борд, вийди на воду та насолодись Черніговом з нового
                    ракурсу — спокійно, у власному темпі.
                </p>

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
                    Забронювати зараз
                </Link>
            </div>

            <div
                style={{
                    maxWidth: 680,
                    margin: '0 auto',
                    padding: '0 0.5rem 3rem',
                    position: 'relative',
                    zIndex: 10,
                }}
            >
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        marginBottom: '1rem',
                        padding: '0 0.25rem',
                    }}
                >
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
                        <div
                            style={{
                                width: 36,
                                height: 36,
                                borderRadius: 10,
                                background:
                                    'linear-gradient(135deg, #f09433 0%, #e6683c 25%, #dc2743 50%, #cc2366 75%, #bc1888 100%)',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                flexShrink: 0,
                            }}
                        >
                            <svg
                                width="20"
                                height="20"
                                viewBox="0 0 24 24"
                                fill="none"
                                aria-hidden="true"
                            >
                                <rect
                                    x="2"
                                    y="2"
                                    width="20"
                                    height="20"
                                    rx="5"
                                    stroke="white"
                                    strokeWidth="2"
                                />
                                <circle cx="12" cy="12" r="4" stroke="white" strokeWidth="2" />
                                <circle cx="17.5" cy="6.5" r="1" fill="white" />
                            </svg>
                        </div>
                        <div>
                            <div
                                style={{
                                    color: 'var(--sand)',
                                    fontWeight: 600,
                                    fontSize: '0.9rem',
                                    lineHeight: 1.2,
                                }}
                            >
                                Ми в Instagram
                            </div>
                            <div
                                style={{
                                    color: 'var(--seafoam)',
                                    fontSize: '0.65rem',
                                    letterSpacing: '0.1em',
                                    textTransform: 'uppercase',
                                    fontWeight: 300,
                                }}
                            >
                                @supboard_che
                            </div>
                        </div>
                    </div>

                    <a
                        href="https://www.instagram.com/supboard_che/"
                        target="_blank"
                        rel="noopener noreferrer"
                        style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '0.35rem',
                            padding: '0.45rem 0.9rem',
                            background: 'rgba(255,255,255,0.08)',
                            border: '1px solid rgba(255,255,255,0.18)',
                            borderRadius: 20,
                            color: 'var(--sand)',
                            fontSize: '0.72rem',
                            fontWeight: 500,
                            textDecoration: 'none',
                            letterSpacing: '0.03em',
                            whiteSpace: 'nowrap',
                        }}
                    >
                        Підписатись
                    </a>
                </div>

                <div
                    style={{
                        borderRadius: 18,
                        overflow: 'hidden',
                        boxShadow:
                            '0 20px 50px rgba(0,0,0,0.35), 0 0 0 1px rgba(255,255,255,0.08)',
                    }}
                >
                    <blockquote
                        className="instagram-media"
                        data-instgrm-captioned=""
                        data-instgrm-permalink="https://www.instagram.com/reel/DYPbuVwyOZX/?utm_source=ig_embed&utm_campaign=loading"
                        data-instgrm-version="14"
                        style={{
                            background: '#FFF',
                            border: 0,
                            borderRadius: 0,
                            boxShadow: 'none',
                            margin: 0,
                            maxWidth: '100%',
                            minWidth: 0,
                            padding: 0,
                            width: '100%',
                        }}
                    />
                </div>

                <p
                    style={{
                        textAlign: 'center',
                        color: 'rgba(255,255,255,0.35)',
                        fontSize: '0.65rem',
                        marginTop: '0.85rem',
                        letterSpacing: '0.05em',
                    }}
                >
                    Більше відео та фото — в нашому Instagram
                </p>
            </div>

            <Script
                async
                src="https://www.instagram.com/embed.js"
                strategy="lazyOnload"
            />
        </>
    );
}
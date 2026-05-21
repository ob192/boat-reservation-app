import Link from 'next/link';
import Script from 'next/script';

// ─── Contact & location constants ─────────────────────────────────────────
const PHONE_DISPLAY = '+38 (067) 123-45-67';
const PHONE_HREF = 'tel:+380671234567';

/** Replace with the real coordinates and place ID */
const MARINA_LAT = 51.4982;
const MARINA_LNG = 31.2893;
const MARINA_PLACE_ID = 'ChIJN1t_tDeuEmsRUsoyG83frY4';

function buildMapsEmbedUrl(lat: number, lng: number): string {
    return `https://maps.google.com/maps?q=${lat},${lng}&z=16&output=embed`;
}

function buildMapsOpenUrl(lat: number, lng: number, placeId: string): string {
    return `https://www.google.com/maps/search/?api=1&query=${lat},${lng}&query_place_id=${placeId}`;
}

// ─── Shared inline styles ──────────────────────────────────────────────────

const subscribeStyle: React.CSSProperties = {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '0.35rem',
    padding: '0.45rem 0.9rem',
    background: 'var(--surface)',
    border: '1px solid var(--mist)',
    borderRadius: 20,
    color: 'var(--amber-ink)',
    fontSize: '0.72rem',
    fontWeight: 500,
    textDecoration: 'none',
    letterSpacing: '0.03em',
    whiteSpace: 'nowrap',
};

// ─── Component ────────────────────────────────────────────────────────────

export default function HomePage() {
    const mapsEmbedUrl = buildMapsEmbedUrl(MARINA_LAT, MARINA_LNG);
    const mapsOpenUrl = buildMapsOpenUrl(MARINA_LAT, MARINA_LNG, MARINA_PLACE_ID);

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
                        color: 'var(--on-amber)',
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

                {/* ── Phone number ──────────────────────────────────────────── */}
                {/*<div*/}
                {/*    style={{*/}
                {/*        marginTop: '2rem',*/}
                {/*        background: 'var(--sand)',*/}
                {/*        border: '1px solid var(--mist)',*/}
                {/*        borderRadius: 14,*/}
                {/*        padding: '1rem 1.25rem',*/}
                {/*        display: 'flex',*/}
                {/*        alignItems: 'center',*/}
                {/*        justifyContent: 'space-between',*/}
                {/*        gap: '0.75rem',*/}
                {/*    }}*/}
                {/*>*/}
                {/*    <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem' }}>*/}
                {/*        <div*/}
                {/*            style={{*/}
                {/*                width: 36,*/}
                {/*                height: 36,*/}
                {/*                borderRadius: 10,*/}
                {/*                background: 'var(--cream)',*/}
                {/*                border: '1px solid var(--mist)',*/}
                {/*                display: 'flex',*/}
                {/*                alignItems: 'center',*/}
                {/*                justifyContent: 'center',*/}
                {/*                fontSize: '1.1rem',*/}
                {/*                flexShrink: 0,*/}
                {/*            }}*/}
                {/*        >*/}
                {/*            📞*/}
                {/*        </div>*/}
                {/*        <div>*/}
                {/*            <div*/}
                {/*                style={{*/}
                {/*                    fontSize: '0.6rem',*/}
                {/*                    textTransform: 'uppercase',*/}
                {/*                    letterSpacing: '0.12em',*/}
                {/*                    color: 'var(--subtle)',*/}
                {/*                    fontWeight: 600,*/}
                {/*                    marginBottom: '0.15rem',*/}
                {/*                }}*/}
                {/*            >*/}
                {/*                Зв'язатися з нами*/}
                {/*            </div>*/}
                {/*            <a*/}
                {/*                href={PHONE_HREF}*/}
                {/*                style={{*/}
                {/*                    fontFamily: 'var(--font-playfair), "Playfair Display", serif',*/}
                {/*                    fontSize: '1.05rem',*/}
                {/*                    fontWeight: 700,*/}
                {/*                    color: 'var(--text)',*/}
                {/*                    textDecoration: 'none',*/}
                {/*                    letterSpacing: '0.02em',*/}
                {/*                }}*/}
                {/*            >*/}
                {/*                {PHONE_DISPLAY}*/}
                {/*            </a>*/}
                {/*        </div>*/}
                {/*    </div>*/}

                {/*    <a*/}
                {/*        href={PHONE_HREF}*/}
                {/*        style={{*/}
                {/*            flexShrink: 0,*/}
                {/*            display: 'inline-flex',*/}
                {/*            alignItems: 'center',*/}
                {/*            gap: '0.3rem',*/}
                {/*            padding: '0.5rem 0.9rem',*/}
                {/*            background: 'var(--seafoam)',*/}
                {/*            color: 'white',*/}
                {/*            borderRadius: 10,*/}
                {/*            fontSize: '0.78rem',*/}
                {/*            fontWeight: 600,*/}
                {/*            textDecoration: 'none',*/}
                {/*            whiteSpace: 'nowrap',*/}
                {/*        }}*/}
                {/*    >*/}
                {/*        Подзвонити*/}
                {/*    </a>*/}
                {/*</div>*/}

                {/* ── Phone number ──────────────────────────────────────────── */}
                <div
                    style={{
                        marginTop: '2rem',
                        background: 'var(--sand)',
                        border: '1px solid var(--mist)',
                        borderRadius: 14,
                        padding: '0.85rem 1rem',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        gap: '0.75rem',
                        flexWrap: 'wrap',
                    }}
                >
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem', minWidth: 0, flex: '1 1 auto' }}>
                        <div
                            style={{
                                width: 36,
                                height: 36,
                                borderRadius: 10,
                                background: 'var(--cream)',
                                border: '1px solid var(--mist)',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                fontSize: '1.1rem',
                                flexShrink: 0,
                            }}
                        >
                            📞
                        </div>
                        <div style={{ minWidth: 0 }}>
                            <div
                                style={{
                                    fontSize: '0.6rem',
                                    textTransform: 'uppercase',
                                    letterSpacing: '0.12em',
                                    color: 'var(--subtle)',
                                    fontWeight: 600,
                                    marginBottom: '0.15rem',
                                }}
                            >
                                Зв'язатися з нами
                            </div>
                            <a
                            href={PHONE_HREF}
                            style={{
                            fontFamily: 'var(--font-playfair), "Playfair Display", serif',
                            fontSize: 'clamp(0.9rem, 4vw, 1.05rem)',
                            fontWeight: 700,
                            color: 'var(--text)',
                            textDecoration: 'none',
                            letterSpacing: '0.02em',
                            whiteSpace: 'nowrap',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            display: 'block',
                        }}
                            >
                            {PHONE_DISPLAY}
                        </a>
                    </div>
                </div>

                <a
                href={PHONE_HREF}
                style={{
                flexShrink: 0,
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.3rem',
                padding: '0.55rem 1rem',
                background: 'var(--seafoam)',
                color: 'white',
                borderRadius: 10,
                fontSize: '0.82rem',
                fontWeight: 600,
                textDecoration: 'none',
                whiteSpace: 'nowrap',
                minHeight: 44,
            }}
                >
                Подзвонити
            </a>
        </div>


                {/* ── Instagram embed ───────────────────────────────────────── */}
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        marginBottom: '1rem',
                        padding: '0 0.25rem',
                    }}
                >
                    {/*<div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>*/}
                    {/*    <div*/}
                    {/*        style={{*/}
                    {/*            width: 36,*/}
                    {/*            height: 36,*/}
                    {/*            borderRadius: 10,*/}
                    {/*            background:*/}
                    {/*                'linear-gradient(135deg, #f09433 0%, #e6683c 25%, #dc2743 50%, #cc2366 75%, #bc1888 100%)',*/}
                    {/*            display: 'flex',*/}
                    {/*            alignItems: 'center',*/}
                    {/*            justifyContent: 'center',*/}
                    {/*            flexShrink: 0,*/}
                    {/*        }}*/}
                    {/*    >*/}
                    {/*        <svg*/}
                    {/*            width="20"*/}
                    {/*            height="20"*/}
                    {/*            viewBox="0 0 24 24"*/}
                    {/*            fill="none"*/}
                    {/*            aria-hidden="true"*/}
                    {/*        >*/}
                    {/*            <rect*/}
                    {/*                x="2"*/}
                    {/*                y="2"*/}
                    {/*                width="20"*/}
                    {/*                height="20"*/}
                    {/*                rx="5"*/}
                    {/*                stroke="white"*/}
                    {/*                strokeWidth="2"*/}
                    {/*            />*/}
                    {/*            <circle cx="12" cy="12" r="4" stroke="white" strokeWidth="2" />*/}
                    {/*            <circle cx="17.5" cy="6.5" r="1" fill="white" />*/}
                    {/*        </svg>*/}
                    {/*    </div>*/}
                    {/*    <div>*/}
                    {/*        <div*/}
                    {/*            style={{*/}
                    {/*                color: 'var(--text)',*/}
                    {/*                fontWeight: 600,*/}
                    {/*                fontSize: '0.9rem',*/}
                    {/*                lineHeight: 1.2,*/}
                    {/*            }}*/}
                    {/*        >*/}
                    {/*            Ми в Instagram*/}
                    {/*        </div>*/}
                    {/*        <div*/}
                    {/*            style={{*/}
                    {/*                color: 'var(--amber-ink)',*/}
                    {/*                fontSize: '0.65rem',*/}
                    {/*                letterSpacing: '0.1em',*/}
                    {/*                textTransform: 'uppercase',*/}
                    {/*                fontWeight: 500,*/}
                    {/*            }}*/}
                    {/*        >*/}
                    {/*            @supboard_che*/}
                    {/*        </div>*/}
                    {/*    </div>*/}
                    {/*</div>*/}

                    {/*<a*/}
                    {/*    href="https://www.instagram.com/supboard_che/"*/}
                    {/*    target="_blank"*/}
                    {/*    rel="noopener noreferrer"*/}
                    {/*    style={subscribeStyle}*/}
                    {/*>*/}
                    {/*    Підписатись*/}
                    {/*</a>*/}
                </div>

                <div
                    style={{
                        borderRadius: 18,
                        overflow: 'hidden',
                        boxShadow:
                            '0 20px 50px rgba(0,0,0,0.18), 0 0 0 1px rgba(0,0,0,0.06)',
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
                        color: 'var(--subtle)',
                        fontSize: '0.65rem',
                        marginTop: '0.85rem',
                        letterSpacing: '0.05em',
                    }}
                >
                    Більше відео та фото — в нашому Instagram
                </p>

                {/* ── Location / Google Maps ─────────────────────────────────── */}
                <div style={{ marginTop: '2rem' }}>
                    <div
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'space-between',
                            marginBottom: '0.75rem',
                            padding: '0 0.25rem',
                        }}
                    >
                        <div
                            style={{
                                fontSize: '0.6rem',
                                textTransform: 'uppercase',
                                letterSpacing: '0.14em',
                                color: 'var(--subtle)',
                                fontWeight: 600,
                            }}
                        >
                            📍 Місце відправлення
                        </div>
                        <a
                            href={mapsOpenUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                            style={subscribeStyle}
                        >
                            Відкрити в Maps
                        </a>
                    </div>

                    {/* Map iframe */}
                    <div
                        style={{
                            borderRadius: 16,
                            overflow: 'hidden',
                            border: '1px solid var(--mist)',
                            boxShadow: '0 4px 20px rgba(0,0,0,0.08)',
                            height: 240,
                        }}
                    >
                        <iframe
                            title="Місце відправлення SUP Chernihiv"
                            src={mapsEmbedUrl}
                            width="100%"
                            height="240"
                            style={{ border: 0, display: 'block' }}
                            allowFullScreen={false}
                            loading="lazy"
                            referrerPolicy="no-referrer-when-downgrade"
                        />
                    </div>

                    {/* Address strip */}
                    <div
                        style={{
                            marginTop: '0.6rem',
                            background: 'var(--sand)',
                            borderRadius: 10,
                            padding: '0.7rem 0.9rem',
                        }}
                    >
                        <div style={{ fontWeight: 600, fontSize: '0.85rem', color: 'var(--navy)', marginBottom: '0.1rem' }}>
                            Harbour &amp; Wave Marina
                        </div>
                        <div style={{ fontSize: '0.72rem', color: 'var(--subtle)' }}>
                            Набережна Перемоги, 1, Чернігів, 14000
                        </div>
                    </div>
                </div>

                {/* ── Footer legal links ─────────────────────────────────────── */}
                <div
                    style={{
                        marginTop: '2.5rem',
                        paddingTop: '1.25rem',
                        borderTop: '1px solid var(--mist)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        gap: '0.5rem',
                        flexWrap: 'wrap',
                    }}
                >
                    <Link
                        href="/privacy"
                        style={{
                            fontSize: '0.68rem',
                            color: 'var(--subtle)',
                            textDecoration: 'none',
                            letterSpacing: '0.03em',
                        }}
                    >
                        Політика конфіденційності
                    </Link>
                    <span style={{ fontSize: '0.6rem', color: 'var(--mist)' }}>·</span>
                    <Link
                        href="/terms"
                        style={{
                            fontSize: '0.68rem',
                            color: 'var(--subtle)',
                            textDecoration: 'none',
                            letterSpacing: '0.03em',
                        }}
                    >
                        Умови використання
                    </Link>
                    <span style={{ fontSize: '0.6rem', color: 'var(--mist)' }}>·</span>
                    <span style={{ fontSize: '0.68rem', color: 'var(--subtle)', letterSpacing: '0.03em' }}>
                    © 2026 SUP Chernihiv
                </span>
                </div>
            </div>

            <Script
                async
                src="https://www.instagram.com/embed.js"
                strategy="lazyOnload"
            />
        </>
    );
}
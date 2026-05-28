import Link from 'next/link';
import Script from 'next/script';

// ─── Contact & location constants ─────────────────────────────────────────
const PHONE_DISPLAY = '+38 (050) 367-66-70';
const PHONE_HREF = 'tel:+380503676670';

const MARINA_LAT = 51.51083547181443;
const MARINA_LNG = 31.35220277765311;
const MARINA_PLACE_ID = '0x46d5484aca07d961:0xdc3b564198209bee';

const MAPS_EMBED_SRC =
    'https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d2483.129752207899!2d31.354777700000003!3d51.5108355!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x46d5484aca07d961%3A0xdc3b564198209bee!2sKyryla%20Rozumovskoho%20St%2C%205%2C%20Chernihiv%2C%20Chernihivs%27ka%20oblast%2C%2014000!5e0!3m2!1sen!2sua!4v1779964836720!5m2!1sen!2sua';

function buildMapsEmbedUrl(_lat: number, _lng: number): string {
    return MAPS_EMBED_SRC;
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
                    Відчуйте свободу<br/>
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
                    <div style={{display: 'flex', alignItems: 'center', gap: '0.65rem', minWidth: 0, flex: '1 1 auto'}}>
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
                        <div style={{minWidth: 0}}>
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
                <div style={{marginTop: '2rem'}}>
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
                </div>

                {/* ── Location / Google Maps ─────────────────────────────────── */}
                <div style={{marginTop: '2rem'}}>
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
                            style={{border: 0, display: 'block'}}
                            allowFullScreen
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
                        <div style={{
                            fontWeight: 600,
                            fontSize: '0.85rem',
                            color: 'var(--navy)',
                            marginBottom: '0.1rem'
                        }}>
                            Kyryla Rozumovskoho St, 5
                        </div>
                        <div style={{fontSize: '0.72rem', color: 'var(--subtle)'}}>
                            Chernihiv, Chernihivs'ka oblast, 14000
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
                    <span style={{fontSize: '0.6rem', color: 'var(--mist)'}}>·</span>
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
                    <span style={{fontSize: '0.6rem', color: 'var(--mist)'}}>·</span>
                    <span style={{fontSize: '0.68rem', color: 'var(--subtle)', letterSpacing: '0.03em'}}>
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
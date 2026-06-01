import Link from 'next/link';
import { Instrument_Serif, Inter_Tight, JetBrains_Mono } from 'next/font/google';
import Script from 'next/script';

// ─── Fonts (scoped to this page via the .ed-root variable classes) ──────────
const instrument = Instrument_Serif({
    subsets: ['latin'],
    weight: ['400'],
    style: ['normal', 'italic'],
    variable: '--ed-serif',
    display: 'swap',
});

const interTight = Inter_Tight({
    subsets: ['latin', 'cyrillic'],
    weight: ['400', '500', '600'],
    variable: '--ed-sans',
    display: 'swap',
});

const jetbrains = JetBrains_Mono({
    subsets: ['latin'],
    weight: ['400', '500'],
    variable: '--ed-mono',
    display: 'swap',
});

// ─── Contact constants ──────────────────────────────────────────────────────
const PHONE_DISPLAY = '+38 (050) 367-66-70';
const PHONE_HREF = 'tel:+380503676670';

// ─── Departure point (Google Maps) ──────────────────────────────────────────
const MAPS_EMBED_SRC =
    'https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d2483.129752207899!2d31.354777700000003!3d51.5108355!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x46d5484aca07d961%3A0xdc3b564198209bee!2sKyryla%20Rozumovskoho%20St%2C%205%2C%20Chernihiv%2C%20Chernihivs%27ka%20oblast%2C%2014000!5e0!3m2!1sen!2sua!4v1779964836720!5m2!1sen!2sua';
const MAPS_OPEN_URL =
    'https://www.google.com/maps/search/?api=1&query=51.51083547181443,31.35220277765311&query_place_id=0x46d5484aca07d961:0xdc3b564198209bee';

// ─── Instagram ───────────────────────────────────────────────────────────────
const IG_URL = 'https://www.instagram.com/supboard_che/';

// ─── FAQ items 2–9 (the first, "Де старт і фініш?", embeds a map) ────────────
const FAQ_REST: { q: string; a: string }[] = [
    {
        q: 'Як дістатися автомобілем?',
        a: 'Біля бази є місце для безкоштовного паркування. Заїзд зручний із боку набережної — орієнтуйтеся на наш банер біля води.',
    },
    {
        q: 'Який мінімум для групи та чи робите корпоративи?',
        a: 'Для корпоративного сплаву мінімальна кількість — 8 дорослих осіб. Для груп та корпоративних заходів готуємо окремі умови й маршрут — просто зателефонуйте нам.',
    },
    {
        q: 'Що взяти з собою?',
        a: 'Рушник, змінний одяг, головний убір, сонцезахисний крем і воду. За потреби візьміть з собою засіб від комарів. Решту спорядження видаємо на місці.',
    },
    {
        q: 'Яке спорядження ви даєте і чи треба вміти плавати?',
        a: 'Видаємо борд, весло, ліш, гермомішок для ваших речей та рятувальний жилет на кожного учасника. Базові навички плавання бажані, але жилет обовʼязковий для всіх без винятку.',
    },
    {
        q: 'Чи можна з дітьми?',
        a: 'Так. Діти вагою від 45 кг можуть веслувати на власному борді. Молодших беремо разом із дорослим на один борд.',
    },
    {
        q: 'Як відбувається оплата?',
        a: 'Оплата онлайн карткою під час бронювання. Підтвердження з усіма деталями надходить на вашу пошту одразу після успішної оплати.',
    },
    {
        q: 'А якщо погода зіпсується?',
        a: 'За несприятливих умов ми безкоштовно перенесемо сеанс на інший зручний день або повернемо повну вартість — на ваш вибір.',
    },
    {
        q: 'Чи буде інструктаж перед виходом?',
        a: 'Так. Перед стартом проводимо короткий інструктаж: техніка веслування, рівновага й правила безпеки. Загалом близько десяти хвилин.',
    },
];

export default function HomePage() {
    return (
        <div className={`${instrument.variable} ${interTight.variable} ${jetbrains.variable} ed-root`}>
            <div className="ed-shell">
                {/* ── Top bar ──────────────────────────────────────────────── */}
                <header className="ed-topbar">
                    <div className="ed-brand">
                        <span className="ed-monogram">S</span>
                        <span className="ed-wordmark">
              <span className="ed-wordmark-name">SUP Chernihiv</span>
              <span className="ed-wordmark-sub">оренда SUP-бордів</span>
            </span>
                    </div>
                    <span className="ed-season">Сезон 2026</span>
                </header>

                <div className="ed-rule" />

                {/* ── Hero / Intro ─────────────────────────────────────────── */}
                <section className="ed-hero">
                    <p className="ed-eyebrow ed-eyebrow--lead">
                        <span className="ed-eyebrow-rule" />
                        № 01 — Десна, Чернігів
                    </p>

                    <h1 className="ed-headline">
                        Відчуйте
                        <br />
                        свободу
                        <br />
                        <em>на воді.</em>
                    </h1>

                    <p className="ed-lede">
                        Орендуйте SUP-борд, вийдіть на спокійну воду Десни та подивіться на Чернігів
                        із нового ракурсу — у власному темпі, без поспіху.
                    </p>
                </section>

                {/* ── Stats strip ──────────────────────────────────────────── */}
                <section className="ed-stats">
                    <div className="ed-stat">
                        <span className="ed-stat-num">15</span>
                        <span className="ed-eyebrow">бордів</span>
                    </div>
                    <div className="ed-stat">
                        <span className="ed-stat-num">2 год</span>
                        <span className="ed-eyebrow">мінімум</span>
                    </div>
                    <div className="ed-stat">
                        <span className="ed-stat-num">₴400</span>
                        <span className="ed-eyebrow">за борд</span>
                    </div>
                </section>

                <div className="ed-rule" />

                {/* ── Contact row ──────────────────────────────────────────── */}
                <section className="ed-contact">
                    <a href={PHONE_HREF} className="ed-phone-btn" aria-label="Подзвонити">
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                            <path
                                d="M6.5 4h3l1.5 4-2 1.5a11 11 0 0 0 5 5l1.5-2 4 1.5v3a1.5 1.5 0 0 1-1.6 1.5A15.5 15.5 0 0 1 5 6.6 1.5 1.5 0 0 1 6.5 4Z"
                                stroke="currentColor"
                                strokeWidth="1.6"
                                strokeLinejoin="round"
                            />
                        </svg>
                    </a>

                    <div className="ed-contact-body">
                        <span className="ed-eyebrow">Звʼязатися</span>
                        <a href={PHONE_HREF} className="ed-phone-num">{PHONE_DISPLAY}</a>
                    </div>

                    <a href={PHONE_HREF} className="ed-call-link">Дзвінок →</a>
                </section>

                <div className="ed-rule" />

                {/* ── FAQ ──────────────────────────────────────────────────── */}
                <section className="ed-faq">
                    <p className="ed-eyebrow ed-eyebrow--lead">
                        <span className="ed-eyebrow-rule" />
                        Усе, що варто знати
                    </p>

                    <h2 className="ed-faq-title">
                        Питання про <em>сплав.</em>
                    </h2>
                    <p className="ed-faq-sub">
                        Найчастіше запитують про логістику, спорядження та оплату. Якщо чогось бракує — телефонуйте.
                    </p>

                    <div className="ed-faq-list">
                        {/* Item 01 — start/finish, with embedded map */}
                        <details className="ed-faq-item">
                            <summary className="ed-faq-q">
                                <span className="ed-faq-qtext">Де старт і фініш?</span>
                                <span className="ed-faq-toggle" aria-hidden="true">+</span>
                            </summary>
                            <div className="ed-faq-a">
                                <p>
                                    Старт і фініш — на нашій базі на вулиці Кирила Розумовського, 5. Маршрут
                                    пролягає вздовж Десни й повертається до тієї самої точки, тож загубитися
                                    неможливо.
                                </p>
                                <div className="ed-faq-map">
                                    <iframe
                                        title="Місце старту та фінішу"
                                        src={MAPS_EMBED_SRC}
                                        loading="lazy"
                                        allowFullScreen
                                        referrerPolicy="no-referrer-when-downgrade"
                                    />
                                </div>
                                <a
                                    href={MAPS_OPEN_URL}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="ed-faq-maplink"
                                >
                                    Відкрити маршрут ↗
                                </a>
                            </div>
                        </details>

                        {/* Items 02–09 */}
                        {FAQ_REST.map((item) => (
                            <details className="ed-faq-item" key={item.q}>
                                <summary className="ed-faq-q">
                                    <span className="ed-faq-qtext">{item.q}</span>
                                    <span className="ed-faq-toggle" aria-hidden="true">+</span>
                                </summary>
                                <div className="ed-faq-a">
                                    <p>{item.a}</p>
                                </div>
                            </details>
                        ))}
                    </div>
                </section>

                <div className="ed-rule" />

                {/* ── Instagram card ───────────────────────────────────────── */}
                <section className="ed-ig-sec">
                    <p className="ed-eyebrow ed-eyebrow--lead">
                        <span className="ed-eyebrow-rule" />
                        Свіже з води
                    </p>

                    <div style={{ marginTop: 16 }}>
                        <blockquote
                            className="instagram-media"
                            data-instgrm-captioned
                            data-instgrm-permalink="https://www.instagram.com/reel/DYPbuVwyOZX/?utm_source=ig_embed&utm_campaign=loading"
                            data-instgrm-version="14"
                            style={{
                                background: '#FFF',
                                border: 0,
                                borderRadius: 3,
                                boxShadow: '0 0 1px 0 rgba(0,0,0,0.5),0 1px 10px 0 rgba(0,0,0,0.15)',
                                margin: '1px',
                                maxWidth: 540,
                                minWidth: 326,
                                padding: 0,
                                width: '100%',
                            }}
                        />
                    </div>
                </section>

                {/* ── Footer legal ─────────────────────────────────────────── */}
                <footer className="ed-footer">
                    <Link href="/privacy" className="ed-foot-link">Політика конфіденційності</Link>
                    <span className="ed-foot-dot">·</span>
                    <Link href="/terms" className="ed-foot-link">Умови використання</Link>
                    <span className="ed-foot-dot">·</span>
                    <span className="ed-foot-link">© 2026 SUP Chernihiv</span>
                </footer>

                <div className="ed-dock-spacer" />
            </div>

            {/* ── Sticky bottom dock ─────────────────────────────────────── */}
            <div className="ed-dock">
                <Link href="/book" className="ed-cta">Забронювати зараз</Link>
            </div>

            <Script async src="https://www.instagram.com/embed.js" strategy="lazyOnload" />
        </div>
    );
}
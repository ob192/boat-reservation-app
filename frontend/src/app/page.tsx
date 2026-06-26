import Link from 'next/link';
import {Instrument_Serif, Inter_Tight, JetBrains_Mono} from 'next/font/google';
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
const PHONE_DISPLAY = '+38 (073) 169-69-09';
const PHONE_HREF = 'tel:+380731696909';

// ─── Departure point (Google Maps) ──────────────────────────────────────────
// ─── Desna route — start (base) and finish ──────────────────────────────────
const DESNA_START_EMBED =
    'https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d1220.806632518116!2d31.35466982996041!3d51.50937928244668!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x46d5490079167cf9%3A0x9c86ca7f773c93ad!2ssup_che!5e0!3m2!1sen!2sua!4v1780663713931!5m2!1sen!2sua';
const DESNA_END_URL = 'https://maps.app.goo.gl/j5qfTenHVyyuTfwh8';

// ─── Snov route — MOCK points, replace before launch ────────────────────────
const SNOV_START_EMBED =
    'https://www.google.com/maps?q=Klochkiv,+Chernihiv,+Ukraine&output=embed';
const SNOV_END_URL =
    'https://www.google.com/maps/search/?api=1&query=Snovyanka%2C+Chernihiv+Oblast';

// ─── Instagram ───────────────────────────────────────────────────────────────
const IG_URL = 'https://www.instagram.com/supboard_che/';

// ─── FAQ items 2–9 (the first, "Де старт і фініш?", embeds a map) ────────────
const FAQ_REST: { q: string; a: string }[] = [
    {
        q: 'Як дістатися автомобілем?',
        a: 'Біля бази є місце для безкоштовного паркування. Заїзд зручний із боку набережної — орієнтуйтеся на наш банер біля води.',
    },
    {
        q: 'Який мінімум для групи на сплав та чи робите корпоративи?',
        a: 'Щоб сплав відбувся, має зібратися група не менше 6 осіб — саме стільки потрібно для виходу на воду. Для корпоративних заходів мінімум — 8 учасників. Готуємо окремі умови й маршрут — просто зателефонуйте нам.',
    },
    {
        q: 'Що буде, якщо я запізнюся?',
        a: 'Просимо приходити на місце за 15 хвилин до початку — для інструктажу та підготовки спорядження. Група відправляється строго у призначений час: якщо сплав о 10:00, старт о 10:15 незалежно від кількості присутніх. Тих, хто запізнився більш ніж на 15 хвилин, група не чекає, а вартість сесії не повертається.',
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

                <div className="ed-rule"/>

                {/* ── Hero / Intro ─────────────────────────────────────────── */}
                <section className="ed-hero">
                    <p className="ed-eyebrow ed-eyebrow--lead">
                        <span className="ed-eyebrow-rule"/>
                        № 01 — Десна, Чернігів
                    </p>

                    <h1 className="ed-headline">
                        Відчуйте
                        <br/>
                        свободу
                        <br/>
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
                        <span className="ed-stat-num">18</span>
                        <span className="ed-eyebrow">бордів</span>
                    </div>
                    <div className="ed-stat">
                        <span className="ed-stat-num">2 год</span>
                        <span className="ed-eyebrow">мінімум</span>
                    </div>
                    <div className="ed-stat">
                        <span className="ed-stat-num">₴450</span>
                        <span className="ed-eyebrow">за борд</span>
                    </div>
                </section>

                <div className="ed-rule"/>

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

                <div className="ed-rule"/>

                {/* ── FAQ ──────────────────────────────────────────────────── */}
                <section className="ed-faq">
                    <p className="ed-eyebrow ed-eyebrow--lead">
                        <span className="ed-eyebrow-rule"/>
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
                        {/* Item 01 — start/finish for both routes */}
                        <details className="ed-faq-item">
                            <summary className="ed-faq-q">
                                <span className="ed-faq-qtext">Де старт і фініш?</span>
                                <span className="ed-faq-toggle" aria-hidden="true">+</span>
                            </summary>
                            <div className="ed-faq-a">
                                <p>
                                    Точки старту й фінішу залежать від обраного маршруту. Точну адресу
                                    збору ви отримаєте у листі-підтвердженні разом із квитком.
                                </p>

                                {/* Route — Desna */}
                                <div className="ed-faq-route">
                                    <h4 className="ed-faq-route-title">Сплав по Десні</h4>
                                    <p className="ed-faq-route-label">
                                        Старт — наша база, вул. Кирила Розумовського, 5
                                    </p>
                                    <div className="ed-faq-map">
                                        <iframe
                                            title="Старт — Сплав по Десні"
                                            src={DESNA_START_EMBED}
                                            loading="lazy"
                                            allowFullScreen
                                            referrerPolicy="no-referrer-when-downgrade"
                                        />
                                    </div>
                                    <a
                                        href={DESNA_END_URL}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="ed-faq-maplink"
                                    >
                                        Відкрити точку фінішу ↗
                                    </a>
                                </div>

                                {/* Route — Snov (mock, replace before launch) */}
                                <div className="ed-faq-route">
                                    <h4 className="ed-faq-route-title">Сплав по Снову</h4>
                                    <p className="ed-faq-route-label">Старт — район Клочків</p>
                                    <div className="ed-faq-map">
                                        <iframe
                                            title="Старт — Сплав по Снову"
                                            src={SNOV_START_EMBED}
                                            loading="lazy"
                                            allowFullScreen
                                            referrerPolicy="no-referrer-when-downgrade"
                                        />
                                    </div>
                                    <a
                                        href={SNOV_END_URL}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="ed-faq-maplink"
                                    >
                                        Відкрити точку фінішу — Сновянка ↗
                                    </a>
                                </div>
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

                <div className="ed-rule"/>

                {/* ── Instagram card ───────────────────────────────────────── */}
                <section className="ed-ig-sec">
                    <p className="ed-eyebrow ed-eyebrow--lead">
                        <span className="ed-eyebrow-rule"/>
                        Свіже з води
                    </p>

                    <div style={{marginTop: 16}}>
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

                <div className="ed-dock-spacer"/>
            </div>

            {/* ── Sticky bottom dock ─────────────────────────────────────── */}
            <div className="ed-dock">
                <Link href="/book" className="ed-cta">Забронювати зараз</Link>
            </div>

            <Script async src="https://www.instagram.com/embed.js" strategy="lazyOnload"/>
        </div>
    );
}
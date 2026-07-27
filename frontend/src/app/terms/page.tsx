import Link from 'next/link';
import type { Metadata } from 'next';
import { Instrument_Serif, Inter_Tight, JetBrains_Mono } from 'next/font/google';
import { COMPANY, CONTACTS } from '@/features/legal/company';
import { PUBLIC_OFFER, OFFER_REQUISITES } from '@/features/legal/offer-text';

// ─── Fonts (scoped to this page via the .lg-root variable classes) ──────────
// Deliberately NOT the Oswald ramp used by .ed-root / .bk-root: a condensed
// display face is unreadable across a document this long.
const instrument = Instrument_Serif({
    subsets: ['latin'],
    weight: ['400'],
    style: ['normal', 'italic'],
    variable: '--lg-serif',
    display: 'swap',
});

const interTight = Inter_Tight({
    subsets: ['latin', 'cyrillic'],
    weight: ['400', '500', '600'],
    variable: '--lg-sans',
    display: 'swap',
});

const jetbrains = JetBrains_Mono({
    subsets: ['latin'],
    weight: ['400', '500'],
    variable: '--lg-mono',
    display: 'swap',
});

export const metadata: Metadata = {
    title: 'Публічна оферта — SUP Chernihiv',
    description:
        'Договір про надання послуг оренди SUP-бордів та проведення водних прогулянок. ' +
        'Умови бронювання, оплати, скасування та повернення коштів.',
    alternates: { canonical: '/terms' },
    openGraph: {
        title: 'Публічна оферта — SUP Chernihiv',
        description:
            'Договір про надання послуг оренди SUP-бордів та проведення водних прогулянок.',
        url: '/terms',
        type: 'article',
    },
};

export default function TermsPage() {
    const offer = PUBLIC_OFFER;

    return (
        <div className={`${instrument.variable} ${interTight.variable} ${jetbrains.variable} lg-root`}>
            <div className="lg-shell">
                {/* ── Top bar ── */}
                <header className="lg-topbar">
                    <Link href="/" className="lg-brand">
                        <span className="lg-monogram">S</span>
                        <span className="lg-wordmark">
                            <span className="lg-wordmark-name">{COMPANY.brand}</span>
                            <span className="lg-wordmark-sub">оренда SUP-бордів</span>
                        </span>
                    </Link>
                    <Link href="/" className="lg-back">← На головну</Link>
                </header>

                <div className="lg-rule" style={{ margin: '0 0 4px' }} />

                {/* ── Document head ── */}
                <div className="lg-doc-head">
                    <p className="lg-eyebrow">Юридичний документ</p>
                    <h1 className="lg-title">{offer.title}</h1>
                    <p className="lg-subtitle">{offer.subtitle}</p>

                    <div className="lg-stamp">
                        <span>
                            Редакція <span className="lg-stamp-val">{offer.version}</span>
                        </span>
                        <span>
                            Чинна з <span className="lg-stamp-val">{offer.effectiveDateLabel}</span>
                        </span>
                        <span>
                            Виконавець <span className="lg-stamp-val">{COMPANY.shortLegalName}</span>
                        </span>
                    </div>

                    <p className="lg-preamble">{offer.preamble}</p>
                </div>

                {/* ── Table of contents (pure anchors, no client JS) ── */}
                <nav className="lg-toc" aria-label="Зміст документа">
                    <p className="lg-toc-title">Зміст</p>
                    <ol className="lg-toc-list">
                        {offer.sections.map((section) => (
                            <li key={section.id}>
                                <a href={`#${section.id}`} className="lg-toc-link">
                                    <span className="lg-toc-num">{section.n}.</span>
                                    <span>{section.title}</span>
                                </a>
                            </li>
                        ))}
                    </ol>
                </nav>

                <div className="lg-rule" />

                {/* ── Sections ── */}
                {offer.sections.map((section) => (
                    <section key={section.id} id={section.id} className="lg-section">
                        <div className="lg-section-head">
                            <span className="lg-section-num">{section.n}</span>
                            <h2 className="lg-section-title">{section.title}</h2>
                        </div>

                        {section.lead && <p className="lg-section-lead">{section.lead}</p>}

                        <div className="lg-clauses">
                            {section.clauses.map((clause, i) => (
                                <div className="lg-clause" key={`${section.id}-${i}`}>
                                    {/* Clause numbering is derived, never authored in the text. */}
                                    <span className="lg-clause-num">{`${section.n}.${i + 1}.`}</span>
                                    <div className="lg-clause-body">
                                        <p>{clause.text}</p>
                                        {clause.items && (
                                            <ul className="lg-list">
                                                {clause.items.map((item) => (
                                                    <li key={item}>{item}</li>
                                                ))}
                                            </ul>
                                        )}
                                        {clause.tail && <p className="lg-clause-tail">{clause.tail}</p>}
                                    </div>
                                </div>
                            ))}
                        </div>

                        {section.requisites && (
                            <div className="lg-req">
                                {OFFER_REQUISITES.map((row) => (
                                    <div className="lg-req-row" key={row.label}>
                                        <span className="lg-req-key">{row.label}</span>
                                        <span className="lg-req-val">
                                            {row.href ? (
                                                <a
                                                    href={row.href}
                                                    {...(row.href.startsWith('http')
                                                        ? { target: '_blank', rel: 'noopener noreferrer' }
                                                        : {})}
                                                >
                                                    {row.value}
                                                </a>
                                            ) : (
                                                row.value
                                            )}
                                        </span>
                                    </div>
                                ))}
                            </div>
                        )}
                    </section>
                ))}

                {/* ── Closing ── */}
                <p className="lg-closing">{offer.closing}</p>

                <div className="lg-rule" />

                {/* ── Footer ── */}
                <footer className="lg-footer">
                    <Link href="/privacy" className="lg-foot-link">Політика конфіденційності</Link>
                    <span className="lg-foot-dot">·</span>
                    <a href={CONTACTS.phoneHref} className="lg-foot-link">{CONTACTS.phoneDisplay}</a>
                    <span className="lg-foot-dot">·</span>
                    <span>
                        {offer.id} · {offer.version} · © 2026 {COMPANY.brand}
                    </span>
                </footer>
            </div>
        </div>
    );
}

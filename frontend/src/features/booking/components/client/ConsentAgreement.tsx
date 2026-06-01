'use client';

import { useUser } from '@/features/auth/hooks/useUser';
import { CONSENT_AGREEMENT } from '@/features/booking/consent-text';

interface ConsentAgreementProps {
    agreed: boolean;
    onAgreedChange: (v: boolean) => void;
    docHash: string;
}

export function ConsentAgreement({ agreed, onAgreedChange, docHash }: ConsentAgreementProps) {
    const user = useUser();
    const a = CONSENT_AGREEMENT;

    const displayName = user?.user_metadata?.full_name ?? user?.email ?? '';
    const initials = displayName
        ? displayName.trim().split(/\s+/).map((w: string) => w[0]).slice(0, 2).join('').toUpperCase()
        : '?';

    return (
        <section className="bk-consent" aria-label="Угода учасника">
            <p className="bk-eyebrow bk-eyebrow--lead" style={{ marginBottom: 12 }}>
                <span className="bk-eyebrow-rule" />
                Угода учасника
            </p>

            <div className="bk-consent-doc" tabIndex={0} role="region" aria-label="Текст угоди">
                <div className="bk-consent-org">{a.org}</div>
                <h3 className="bk-consent-title">{a.title}</h3>
                <p className="bk-consent-sub">{a.subtitle}</p>

                <p className="bk-consent-intro">{a.intro}</p>

                <ol className="bk-consent-list">
                    {a.clauses.map((c, i) => (
                        <li key={i}>
                            <span>{c.text}</span>
                            {c.items && (
                                <ul className="bk-consent-risks">
                                    {c.items.map((it) => <li key={it}>{it}</li>)}
                                </ul>
                            )}
                            {c.tail && <span className="bk-consent-tail">{c.tail}</span>}
                        </li>
                    ))}
                </ol>

                <p className="bk-consent-closing">{a.closing}</p>
            </div>

            <div className="bk-consent-meta">
                <span>Хеш документа</span>
                <code>{docHash ? docHash.slice(0, 16) + '…' : 'обчислення…'}</code>
            </div>

            <div className="bk-consent-who">
                <div className="bk-consent-avatar">{initials}</div>
                <div className="bk-consent-who-body">
                    <div className="bk-consent-who-name">{displayName || 'Користувач'}</div>
                    <div className="bk-consent-who-sub">{user?.email ?? ''}</div>
                </div>
            </div>

            <label className="bk-consent-check">
                <input
                    type="checkbox"
                    checked={agreed}
                    onChange={(e) => onAgreedChange(e.target.checked)}
                />
                <span>
                    Я прочитав(-ла) угоду, мені виповнилося 18 років, і я приймаю її умови
                    добровільно.
                </span>
            </label>

            <p className="bk-consent-note">
                Ваша згода, відбиток пристрою та час підписання будуть збережені для обліку.
            </p>
        </section>
    );
}
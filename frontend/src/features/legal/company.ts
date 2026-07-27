/**
 * Single source of truth for the operator’s requisites, contacts and the policy
 * constants the public offer (`offer-text.ts`) and the marketing/booking copy
 * both quote. Anything that appears in a legal document AND in the UI must live
 * here, so the two can never drift apart.
 *
 * ⚠ РНОКПП 3662305123 was supplied by the client as an «ЄДРПОУ». It is 10
 *   digits, so it is a РНОКПП (an ЄДРПОУ is 8 digits and belongs to legal
 *   entities, not to a ФОП). Labelled as РНОКПП here — pending verification
 *   against the витяг з реєстру платників єдиного податку.
 */

export const COMPANY = {
    /** Short trading name used in headings and running copy. */
    brand: 'SUP Chernihiv',
    /** Short legal form + name, as used inside the waiver. */
    shortLegalName: 'ФОП Васюк Юлія Андріївна',
    /** Full legal name for the requisites block. */
    legalName: 'Фізична особа — підприємець Васюк Юлія Андріївна',
    taxIdLabel: 'РНОКПП',
    taxId: '3662305123',
    taxStatus: 'Єдиний податок, 2 група. Не є платником ПДВ.',
    /** Same fact, phrased for use inside a sentence. */
    taxGroupInline: 'єдиний податок, 2 група',
    /** Registered address of the ФОП. */
    address: 'м. Чернігів, Новозаводський р-н, вул. Івана Мазепи, буд. 72-А',
    /** Base / departure point for the sessions. */
    baseAddress: 'м. Чернігів, вул. Кирила Розумовського, 5',
    /** Payment service provider shown to the customer at checkout. */
    paymentProvider: 'LiqPay',
} as const;

export const CONTACTS = {
    phoneDisplay: '+38 (073) 169-69-09',
    phoneHref: 'tel:+380731696909',
    /** The phone number doubles as Telegram and Viber — there is no support e-mail. */
    messengers: 'Telegram, Viber',
    telegramUrl: 'https://t.me/sup_che',
    telegramHandle: '@sup_che',
    instagramUrl: 'https://www.instagram.com/supboard_che/',
    instagramHandle: '@supboard_che',
    siteUrl: 'https://sup-chernihiv.restreto-labs.com',
    siteDisplay: 'sup-chernihiv.restreto-labs.com',
} as const;

export const POLICY = {
    /** Length of one session, in minutes. */
    sessionMinutes: 120,
    /** Human-readable session length for copy. */
    sessionLabel: '120 хвилин',
    /** How early the participant must arrive at the start point, in minutes. */
    arrivalLeadMinutes: 15,
    /** Grace period after the scheduled start; past it the group leaves without the participant. */
    latenessGraceMinutes: 15,
    /** Minimum group size for a session to run. */
    minGroupSize: 6,
    /** Minimum group size for a corporate booking. */
    minCorporateGroupSize: 8,
    /** Below this weight a child shares an adult’s big board at the child rate. */
    childMaxWeightKg: 40,
    /** Minimum age for an independent participant (per the waiver). */
    minAdultAge: 18,
    /** Flat adult tariff per board, ₴. Backend `totalAmount` is authoritative. */
    adultPrice: 450,
    /** Flat child tariff (shared big board), ₴. */
    childPrice: 225,
    /** Currency of all tariffs. */
    currency: 'UAH',
    /** Typical refund turnaround when the operator cancels. */
    refundTerm: 'зазвичай до 10 банківських днів',
} as const;

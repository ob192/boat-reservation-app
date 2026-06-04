import {z} from 'zod';

export const quantitiesSchema = z.object({
    big: z.number().int().min(0),
    medium: z.number().int().min(0),
    small: z.number().int().min(0),
    child: z.number().int().min(0),
});

export const contactSchema = z.object({
    firstName: z.string().min(1, "Введіть ім'я").max(50),
    lastName: z.string().min(1, 'Введіть прізвище').max(50),
    email: z.string().email('Невірна адреса електронної пошти'),
    phone: z.string().min(1), // required by backend; no custom message (soft)
});

export const consentSchema = z.object({
    event: z.literal('consent_agreed'),
    consentId: z.string(),
    agreementId: z.string(),
    agreementVersion: z.string(),
    agreementHash: z.string(),
    user: z.object({
        id: z.string().optional(),
        email: z.string().email(),
        name: z.string(),
    }),
    device: z.object({
        fingerprint: z.string(),
        userAgent: z.string(),
        platform: z.string(),
        timezone: z.string(),
        language: z.string(),
        screen: z.string(),
    }),
    clientTimestamp: z.string(),
});

export const bookingSchema = z.object({
    routeName: z.string().min(1, 'Оберіть маршрут'),
    date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Невірний формат дати'),
    time: z.string().regex(/^\d{2}:\d{2}$/, 'Невірний формат часу'),
    quantities: quantitiesSchema
        .refine((q) => q.big + q.medium + q.small > 0, {
            message: 'Оберіть хоча б один човен',
        })
        .refine((q) => q.child === 0 || q.big > 0, {
            message: 'Дитячі місця доступні лише з великим бордом',
        }),
    contact: contactSchema,
});

export type BookingFormValues = z.infer<typeof bookingSchema>;
export type ContactFormValues = z.infer<typeof contactSchema>;
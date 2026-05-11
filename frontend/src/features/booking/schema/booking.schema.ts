import { z } from 'zod';

export const quantitiesSchema = z.object({
  big: z.number().int().min(0),
  medium: z.number().int().min(0),
  child: z.number().int().min(0),
});

export const contactSchema = z.object({
  firstName: z.string().min(1, "Введіть ім'я").max(50),
  lastName: z.string().min(1, 'Введіть прізвище').max(50),
  email: z.string().email('Невірна адреса електронної пошти'),
  phone: z.string().optional(),
});

export const bookingSchema = z.object({
  date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Невірний формат дати'),
  time: z.string().regex(/^\d{2}:\d{2}$/, 'Невірний формат часу'),
  quantities: quantitiesSchema.refine((q) => q.big + q.medium > 0, {
    message: 'Оберіть хоча б один човен',
  }),
  contact: contactSchema,
});

export type BookingFormValues = z.infer<typeof bookingSchema>;
export type ContactFormValues = z.infer<typeof contactSchema>;

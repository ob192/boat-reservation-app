import { z } from 'zod';

const clientEnvSchema = z.object({
  NEXT_PUBLIC_SUPABASE_URL: z.string().url(),
  NEXT_PUBLIC_SUPABASE_ANON_KEY: z.string().min(1),
  NEXT_PUBLIC_APP_URL: z.string().url().optional(),
});

const serverEnvSchema = z.object({
  BACKEND_URL: z.string().url(),
});

function validateClientEnv() {
  const result = clientEnvSchema.safeParse({
    NEXT_PUBLIC_SUPABASE_URL: process.env.NEXT_PUBLIC_SUPABASE_URL,
    NEXT_PUBLIC_SUPABASE_ANON_KEY: process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY,
    NEXT_PUBLIC_APP_URL: process.env.NEXT_PUBLIC_APP_URL,
  });

  if (!result.success) {
    console.error('❌ Invalid client environment variables:', result.error.flatten().fieldErrors);
    throw new Error('Invalid client environment variables');
  }

  return result.data;
}

function validateServerEnv() {
  // Only run on the server
  if (typeof window !== 'undefined') return {} as z.infer<typeof serverEnvSchema>;

  const result = serverEnvSchema.safeParse({
    BACKEND_URL: process.env.BACKEND_URL,
  });

  if (!result.success) {
    console.error('❌ Invalid server environment variables:', result.error.flatten().fieldErrors);
    throw new Error('Invalid server environment variables');
  }

  return result.data;
}

export const clientEnv = validateClientEnv();
export const env = {
  ...clientEnv,
  ...validateServerEnv(),
};

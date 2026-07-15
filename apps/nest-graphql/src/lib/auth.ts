import { betterAuth } from 'better-auth'
import { Pool } from 'pg'

export const auth = betterAuth({
  database: new Pool({
    connectionString: process.env.DATABASE_URL,
  }),
  emailAndPassword: {
    enabled: true,
  },
  trustedOrigins: [
    process.env.BETTER_AUTH_URL ?? 'http://localhost:3000',
    'http://localhost:5174',
    'http://localhost:80',
  ],
  baseURL: process.env.BETTER_AUTH_URL ?? 'http://localhost:3000',
})

export async function ensureMigrations() {
  const ctx = await (auth as any).$context
  await ctx.runMigrations()
}

export type SessionUser = typeof auth.$Infer.Session.user
export type AuthSession = typeof auth.$Infer.Session.session

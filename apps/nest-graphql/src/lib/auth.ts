import { betterAuth } from 'better-auth'
import { magicLink } from 'better-auth/plugins/magic-link'
import { Pool } from 'pg'
import { sendEmail } from '../email/email-sender'

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
    'http://localhost:9000',
  ],
  baseURL: process.env.BETTER_AUTH_URL ?? 'http://localhost:3000',
  plugins: [
    magicLink({
      sendMagicLink: async ({ email, url }) => {
        console.log('\n═══════════════════════════════════════════════════')
        console.log('  MAGIC LINK for', email)
        console.log('  ───────────────────────────────────────────────')
        console.log('  ', url)
        console.log('═══════════════════════════════════════════════════\n')

        try {
          await sendEmail({
            to: email,
            subject: 'Sign in to Flatty Budget',
            html: `<a href="${url}">Click here to sign in to Flatty Budget</a>`,
          })
          console.log(`[MagicLink] Email also sent to ${email}`)
        } catch (err) {
          console.error(`[MagicLink] Email send failed (ignored in dev): ${err}`)
        }
      },
    }),
  ],
})

export async function ensureMigrations() {
  const ctx = await (auth as any).$context
  await ctx.runMigrations()
}

export type SessionUser = typeof auth.$Infer.Session.user
export type AuthSession = typeof auth.$Infer.Session.session

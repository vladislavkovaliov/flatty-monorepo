import nodemailer from 'nodemailer'

let transporter: nodemailer.Transporter | null = null

function getTransport(): nodemailer.Transporter {
  if (!transporter) {
    transporter = nodemailer.createTransport({
      host: process.env.EMAIL_HOST,
      port: parseInt(process.env.EMAIL_PORT ?? '587', 10),
      secure: parseInt(process.env.EMAIL_PORT ?? '587', 10) === 465,
      auth: {
        user: process.env.EMAIL_USER ?? '',
        pass: process.env.EMAIL_PASS ?? '',
      },
    })
  }
  return transporter
}

export interface SendEmailOptions {
  to: string
  subject: string
  html: string
}

export async function sendEmail(options: SendEmailOptions): Promise<void> {
  const from = process.env.EMAIL_FROM ?? 'noreply@flattybudget.com'

  await getTransport().sendMail({
    from,
    to: options.to,
    subject: options.subject,
    html: options.html,
  })
}

/** For testing: reset the cached transport */
export function resetTransport(): void {
  transporter = null
}

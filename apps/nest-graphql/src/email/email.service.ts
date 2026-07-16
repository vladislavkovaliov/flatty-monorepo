import { Injectable } from '@nestjs/common'
import { sendEmail, SendEmailOptions } from './email-sender'

@Injectable()
export class EmailService {
  async send(options: SendEmailOptions): Promise<void> {
    await sendEmail(options)
  }
}

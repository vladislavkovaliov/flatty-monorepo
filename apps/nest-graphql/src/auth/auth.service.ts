import { Injectable } from '@nestjs/common'
import { auth } from '../lib/auth'

@Injectable()
export class AuthService {
  async validateSession(token: string): Promise<string | null> {
    const headers = new Headers()
    headers.set('cookie', `better-auth.session_token=${token}`)

    const session = await auth.api.getSession({ headers })
    return session?.user.id ?? null
  }
}

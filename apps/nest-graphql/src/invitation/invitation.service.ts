import {
  Injectable,
  ConflictException,
  NotFoundException,
  Logger,
} from '@nestjs/common'
import { InvitationRepository } from './invitation.repository'
import { Invitation, InvitationStatus } from './entities/invitation.entity'
import { InvitationListResponse } from './dto/invitation-list-response'
import { UsersRepository } from '../controllers/users/users.repository'
import { auth } from '../lib/auth'
import type { Request } from 'express'

function headersFromRequest(req: Request): Headers {
  const headers = new Headers()
  for (const [key, value] of Object.entries(req.headers)) {
    if (value) {
      headers.set(key, Array.isArray(value) ? value.join(', ') : value)
    }
  }
  return headers
}

@Injectable()
export class InvitationService {
  private readonly logger = new Logger(InvitationService.name)

  constructor(
    private readonly invitationRepository: InvitationRepository,
    private readonly usersRepository: UsersRepository,
  ) {}

  async sendInvitation(email: string, invitedBy: string, req?: Request): Promise<Invitation> {
    const existingUser = await this.usersRepository.findByEmail(email)
    if (existingUser) {
      throw new ConflictException(
        `User with email ${email} already exists`,
      )
    }

    const existingInvitation =
      await this.invitationRepository.findByEmail(email)
    if (
      existingInvitation &&
      existingInvitation.status === InvitationStatus.PENDING
    ) {
      throw new ConflictException(
        `Pending invitation already sent to ${email}`,
      )
    }

    try {
      await (auth.api as any).signInMagicLink({
        body: {
          email,
          callbackURL:
            process.env.MAGIC_LINK_CALLBACK_URL ??
            'http://localhost:9000/accept-invite',
        },
        headers: req ? headersFromRequest(req) : new Headers({ 'Content-Type': 'application/json' }),
      })
    } catch (error) {
      this.logger.error(`Failed to send magic link to ${email}`, error)
      throw error
    }

    const invitation = await this.invitationRepository.create({
      email,
      invitedBy,
    })

    this.logger.log(`Magic link sent to ${email} (invitation ${invitation.id})`)

    return invitation
  }

  async acceptInvitation(email: string): Promise<Invitation> {
    const invitation = await this.invitationRepository.findByEmail(email)

    if (!invitation) {
      throw new NotFoundException(
        `No invitation found for email ${email}`,
      )
    }

    if (invitation.status !== InvitationStatus.PENDING) {
      throw new ConflictException(
        `Invitation for ${email} is already ${invitation.status}`,
      )
    }

    await this.invitationRepository.updateStatus(
      invitation.id,
      InvitationStatus.ACCEPTED,
    )

    return this.invitationRepository.findById(invitation.id) as Promise<Invitation>
  }

  async list(limit = 10, offset = 0): Promise<InvitationListResponse> {
    const [data, total] = await this.invitationRepository.list(limit, offset)
    return { data, total }
  }

  async count(): Promise<number> {
    return this.invitationRepository.count()
  }

  async findByEmail(email: string): Promise<Invitation | null> {
    return this.invitationRepository.findByEmail(email)
  }
}

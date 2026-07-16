import { Injectable } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { Invitation, InvitationStatus } from './entities/invitation.entity'

@Injectable()
export class InvitationRepository {
  constructor(
    @InjectRepository(Invitation)
    private readonly repository: Repository<Invitation>,
  ) {}

  async findByEmail(email: string): Promise<Invitation | null> {
    return this.repository.findOneBy({ email })
  }

  async create(data: { email: string; invitedBy: string }): Promise<Invitation> {
    const invitation = this.repository.create(data)
    return this.repository.save(invitation)
  }

  async updateStatus(id: string, status: InvitationStatus): Promise<void> {
    await this.repository.update(id, { status, acceptedAt: status === InvitationStatus.ACCEPTED ? new Date().toISOString() : undefined })
  }

  async findById(id: string): Promise<Invitation | null> {
    return this.repository.findOneBy({ id })
  }

  async list(limit: number, offset: number): Promise<[Invitation[], number]> {
    return this.repository.findAndCount({ skip: offset, take: limit, order: { createdAt: 'DESC' } })
  }

  async count(): Promise<number> {
    return this.repository.count()
  }
}

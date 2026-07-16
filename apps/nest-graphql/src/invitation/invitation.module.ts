import { Module } from '@nestjs/common'
import { TypeOrmModule } from '@nestjs/typeorm'
import { Invitation } from './entities/invitation.entity'
import { InvitationRepository } from './invitation.repository'
import { InvitationService } from './invitation.service'
import { InvitationResolver } from './invitation.resolver'
import { UsersRepository } from '../controllers/users/users.repository'
import { User } from '../controllers/users/entities/user.entity'

@Module({
  imports: [TypeOrmModule.forFeature([Invitation, User])],
  providers: [
    InvitationRepository,
    InvitationService,
    InvitationResolver,
    UsersRepository,
  ],
  exports: [InvitationService],
})
export class InvitationModule {}

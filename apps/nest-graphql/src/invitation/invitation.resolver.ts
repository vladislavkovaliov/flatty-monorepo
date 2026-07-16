import { Args, Context, Int, Mutation, Query, Resolver } from '@nestjs/graphql'
import type { Request } from 'express'
import { Invitation } from './entities/invitation.entity'
import { InvitationListResponse } from './dto/invitation-list-response'
import { SendInvitationInput } from './dto/send-invitation-input'
import { InvitationService } from './invitation.service'
import { CurrentUser } from '../auth/current-user.decorator'

@Resolver(() => Invitation)
export class InvitationResolver {
  constructor(private readonly invitationService: InvitationService) {}

  @Query(() => InvitationListResponse, { name: 'invitationList' })
  async list(
    @Args('limit', { type: () => Int, defaultValue: 10 }) limit: number,
    @Args('offset', { type: () => Int, defaultValue: 0 }) offset: number,
  ): Promise<InvitationListResponse> {
    return this.invitationService.list(limit, offset)
  }

  @Query(() => Int, { name: 'invitationCount' })
  async count(): Promise<number> {
    return this.invitationService.count()
  }

  @Mutation(() => Invitation, { name: 'sendInvitation' })
  async sendInvitation(
    @Args('input', { type: () => SendInvitationInput })
    input: SendInvitationInput,
    @CurrentUser() userID: string,
    @Context('req') req: Request,
  ): Promise<Invitation> {
    return this.invitationService.sendInvitation(input.email, userID, req)
  }

  @Mutation(() => Boolean, { name: 'acceptInvitation' })
  async acceptInvitation(
    @Args('email', { type: () => String }) email: string,
  ): Promise<boolean> {
    await this.invitationService.acceptInvitation(email)
    return true
  }
}

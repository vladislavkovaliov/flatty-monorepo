import { Field, Int, ObjectType } from '@nestjs/graphql'
import { Invitation } from '../entities/invitation.entity'

@ObjectType()
export class InvitationListResponse {
  @Field(() => [Invitation])
  data!: Invitation[]

  @Field(() => Int)
  total!: number
}

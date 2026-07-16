import { Field, InputType } from '@nestjs/graphql'

@InputType()
export class SendInvitationInput {
  @Field()
  email!: string
}

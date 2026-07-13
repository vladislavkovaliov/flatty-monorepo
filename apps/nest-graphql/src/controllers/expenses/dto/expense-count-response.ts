import { Field, Int, ObjectType } from '@nestjs/graphql';

@ObjectType()
export class ExpenseCountResponse {
  @Field(() => Int)
  total!: number;
}

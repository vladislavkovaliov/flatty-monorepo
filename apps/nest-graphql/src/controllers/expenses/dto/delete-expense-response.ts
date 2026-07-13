import { Field, Int, ObjectType } from '@nestjs/graphql';

@ObjectType()
export class DeleteExpenseResponse {
  @Field(() => Int)
  data!: number;
}

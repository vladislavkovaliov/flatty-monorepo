import { Field, Float, InputType, Int } from '@nestjs/graphql';

@InputType()
export class ExpenseInput {
  @Field(() => Int)
  residentLocationId!: number;

  @Field(() => Int)
  categoryId!: number;

  @Field(() => Float)
  amount!: number;

  @Field(() => Int)
  month!: number;

  @Field(() => Int)
  year!: number;
}

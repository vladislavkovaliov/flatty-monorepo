import { Field, Int, ObjectType } from '@nestjs/graphql';
import { Expense } from '../entities/expense.entity';

@ObjectType()
export class ListExpenseResponse {
  @Field(() => Int)
  total!: number;

  @Field(() => [Expense])
  data!: Expense[];
}

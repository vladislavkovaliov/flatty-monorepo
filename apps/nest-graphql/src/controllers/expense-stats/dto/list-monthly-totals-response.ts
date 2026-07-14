import { Field, ObjectType } from '@nestjs/graphql';
import { ExpenseMonthlyTotal } from '../entities/expense-monthly-total.entity';

@ObjectType()
export class ListMonthlyTotalsResponse {
  @Field(() => [ExpenseMonthlyTotal])
  data!: ExpenseMonthlyTotal[];
}

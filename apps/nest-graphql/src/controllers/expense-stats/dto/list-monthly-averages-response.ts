import { Field, ObjectType } from '@nestjs/graphql';
import { ExpenseMonthlyAverage } from '../entities/expense-monthly-average.entity';

@ObjectType()
export class ListMonthlyAveragesResponse {
  @Field(() => [ExpenseMonthlyAverage])
  data!: ExpenseMonthlyAverage[];
}

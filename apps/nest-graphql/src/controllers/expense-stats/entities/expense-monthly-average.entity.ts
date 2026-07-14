import { Field, Float, Int, ObjectType } from '@nestjs/graphql';
import { Column, Entity, PrimaryColumn, UpdateDateColumn } from 'typeorm';

@ObjectType()
@Entity('expense_monthly_averages')
export class ExpenseMonthlyAverage {
  @Field(() => Int)
  @PrimaryColumn()
  month!: number;

  @Field(() => Int)
  @PrimaryColumn()
  year!: number;

  @Field(() => Float)
  @Column({ name: 'average_amount', type: 'numeric', precision: 12, scale: 2 })
  averageAmount!: number;

  @Field(() => Int)
  @Column({ name: 'expense_count' })
  expenseCount!: number;

  @Field()
  @UpdateDateColumn({ name: 'updated_at', type: 'timestamptz' })
  updatedAt!: Date;
}

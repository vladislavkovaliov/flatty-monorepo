import { Field, Float, Int, ObjectType } from '@nestjs/graphql';
import { Column, Entity, PrimaryColumn, UpdateDateColumn } from 'typeorm';

@ObjectType()
@Entity('expense_monthly_totals')
export class ExpenseMonthlyTotal {
  @Field(() => Int)
  @PrimaryColumn()
  month!: number;

  @Field(() => Int)
  @PrimaryColumn()
  year!: number;

  @Field(() => Float)
  @Column({ name: 'total_spent', type: 'numeric', precision: 12, scale: 2 })
  totalSpent!: number;

  @Field()
  @UpdateDateColumn({ name: 'updated_at', type: 'timestamptz' })
  updatedAt!: Date;
}

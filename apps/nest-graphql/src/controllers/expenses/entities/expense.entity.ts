import { Field, Float, Int, ObjectType } from '@nestjs/graphql';
import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm';

@ObjectType()
@Entity('expenses')
export class Expense {
  @Field(() => Int)
  @PrimaryGeneratedColumn()
  id!: number;

  @Field(() => Int)
  @Column({ name: 'resident_location_id' })
  residentLocationId!: number;

  @Field(() => Int)
  @Column({ name: 'category_id' })
  categoryId!: number;

  @Field(() => Float)
  @Column({ type: 'numeric', precision: 12, scale: 2 })
  amount!: number;

  @Field(() => Int)
  @Column()
  month!: number;

  @Field(() => Int)
  @Column()
  year!: number;

  @Field()
  @CreateDateColumn({ name: 'created_at', type: 'timestamptz' })
  createdAt!: Date;

  @Field()
  @UpdateDateColumn({ name: 'updated_at', type: 'timestamptz' })
  updatedAt!: Date;
}

import { Field, ObjectType, Int } from "@nestjs/graphql";
import { Column, CreateDateColumn, UpdateDateColumn, Entity, PrimaryGeneratedColumn } from 'typeorm'

@ObjectType()
@Entity('resident_locations')
export class ResidentLocation {
    @Field(() => Int)
    @PrimaryGeneratedColumn()
    id!: number;

    @Column({ name: 'user_id', type: 'text' })
    userId!: string;  // NOTE: no @Field() — intentionally NOT exposed via GraphQL

    @Field()
    @Column()
    country!: string;

    @Field()
    @Column()
    city!: string;

    @Field()
    @Column({ name: 'postal_code' })
    postalCode!: string;

    @Field()
    @Column()
    street!: string;

    @Field()
    @Column()
    house!: string;

    @Field()
    @Column()
    apartment!: string;

    @Field()
    @CreateDateColumn({ name: 'created_at', type: 'timestamptz' })
    createdAt!: Date;

    @Field()
    @UpdateDateColumn({ name: 'updated_at', type: 'timestamptz' })
    updatedAt!: Date;    
}
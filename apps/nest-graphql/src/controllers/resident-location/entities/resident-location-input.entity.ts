import { Field, ObjectType, Int, InputType } from "@nestjs/graphql";
import { Column, CreateDateColumn, UpdateDateColumn, Entity, PrimaryGeneratedColumn } from 'typeorm'

@InputType()
export class ResidentLocationInput {
    @Field()
    country!: string;

    @Field()
    city!: string;

    @Field()
    postalCode!: string;

    @Field()
    street!: string;

    @Field()
    house!: string;

    @Field()
    apartment!: string;
}
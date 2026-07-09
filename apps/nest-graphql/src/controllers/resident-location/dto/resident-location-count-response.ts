import { ObjectType } from "@nestjs/graphql";
import { Field, Int } from "@nestjs/graphql";


@ObjectType()
export class ResidentLocationCountResponse {
    @Field(() => Int)
    total!: number; 
}
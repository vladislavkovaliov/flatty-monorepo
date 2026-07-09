import { ObjectType } from "@nestjs/graphql";
import { Field, Int } from "@nestjs/graphql";
import { ResidentLocation } from "../entities/resident-location.entity";


@ObjectType()
export class ResidentLocationCountResponse {
    @Field(() => Int)
    total!: number; 
}


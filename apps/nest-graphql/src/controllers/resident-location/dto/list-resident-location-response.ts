import { ObjectType } from "@nestjs/graphql";
import { Field, Int } from "@nestjs/graphql";
import { ResidentLocation } from "../entities/resident-location.entity";


@ObjectType()
export class ListResidentLocationResponse {
    @Field(() => Int)
    total!: number; 

    @Field(() => [ResidentLocation])
    data!: ResidentLocation[]
}


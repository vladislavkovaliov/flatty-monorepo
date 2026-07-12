import { ObjectType } from "@nestjs/graphql";
import { Field, Int } from "@nestjs/graphql";

@ObjectType()
export class CategoryCountResponse {
    @Field(() => Int)
    total!: number;
}

import { ObjectType } from "@nestjs/graphql";
import { Field, Int } from "@nestjs/graphql";
import { Category } from "../entities/category.entity";

@ObjectType()
export class ListCategoryResponse {
    @Field(() => Int)
    total!: number;

    @Field(() => [Category])
    data!: Category[]
}

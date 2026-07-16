import { ObjectType } from "@nestjs/graphql";
import { Field, Int } from "@nestjs/graphql";
import { User } from "../entities/user.entity";

@ObjectType()
export class ListUserResponse {
    @Field(() => Int)
    total!: number;

    @Field(() => [User])
    data!: User[];
}

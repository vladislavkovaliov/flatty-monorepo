import { ObjectType, Field, Int } from "@nestjs/graphql";

@ObjectType()
export class DeleteResidentLocationResponse {
    @Field(() => Int)
    data!: number;
}
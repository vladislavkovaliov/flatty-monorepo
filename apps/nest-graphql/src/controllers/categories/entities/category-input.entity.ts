import { Field, InputType } from "@nestjs/graphql";

@InputType()
export class CategoryInput {
    @Field()
    name!: string;

    @Field()
    description!: string;
}

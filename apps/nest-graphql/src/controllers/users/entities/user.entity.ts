import { Field, ObjectType } from "@nestjs/graphql";
import { Column, CreateDateColumn, UpdateDateColumn, Entity } from "typeorm";

@ObjectType()
@Entity('user')
export class User {
    @Field()
    @Column({ primary: true })
    id!: string;

    @Field()
    @Column()
    name!: string;

    @Field()
    @Column({ unique: true })
    email!: string;

    @Field()
    @Column({ name: 'emailVerified', default: false })
    emailVerified!: boolean;

    @Field(() => String, { nullable: true })
    @Column({ name: 'image', type: 'text', nullable: true })
    image!: string | null;

    @Field()
    @CreateDateColumn({ name: 'createdAt', type: 'timestamptz' })
    createdAt!: Date;

    @Field()
    @UpdateDateColumn({ name: 'updatedAt', type: 'timestamptz' })
    updatedAt!: Date;
}

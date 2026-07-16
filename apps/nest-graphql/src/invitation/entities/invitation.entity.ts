import { Field, ObjectType, registerEnumType } from '@nestjs/graphql'
import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn } from 'typeorm'

export enum InvitationStatus {
  PENDING = 'pending',
  ACCEPTED = 'accepted',
  EXPIRED = 'expired',
}

registerEnumType(InvitationStatus, {
  name: 'InvitationStatus',
})

@ObjectType()
@Entity('invitation')
export class Invitation {
  @Field()
  @PrimaryGeneratedColumn('uuid')
  id!: string

  @Field()
  @Column()
  email!: string

  @Field()
  @Column({ name: 'invitedBy' })
  invitedBy!: string

  @Field(() => InvitationStatus)
  @Column({
    type: 'enum',
    enum: InvitationStatus,
    default: InvitationStatus.PENDING,
  })
  status!: InvitationStatus

  @Field()
  @CreateDateColumn({ name: 'createdAt', type: 'timestamptz' })
  createdAt!: Date

  @Field(() => String, { nullable: true })
  @Column({ name: 'acceptedAt', type: 'timestamptz', nullable: true })
  acceptedAt!: string | null
}

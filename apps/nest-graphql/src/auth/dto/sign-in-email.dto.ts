import { ApiProperty } from '@nestjs/swagger'

export class SignInEmailDto {
  @ApiProperty({ example: 'user@example.com' })
  email!: string

  @ApiProperty({ example: 'password123' })
  password!: string

  @ApiProperty({ required: false })
  callbackURL?: string

  @ApiProperty({ required: false, default: true })
  rememberMe?: boolean
}

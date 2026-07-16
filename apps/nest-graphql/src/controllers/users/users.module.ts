import { Module } from '@nestjs/common';
import { UsersService } from './users.service';
import { UsersResolver } from './users.resolver';
import { TypeOrmModule } from '@nestjs/typeorm';
import { User } from './entities/user.entity';
import { UsersRepository } from './users.repository';

@Module({
    imports: [TypeOrmModule.forFeature([User])],
    providers: [UsersRepository, UsersService, UsersResolver],
})
export class UsersModule {}

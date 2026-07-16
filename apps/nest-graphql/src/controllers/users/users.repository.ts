import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { User } from './entities/user.entity';

@Injectable()
export class UsersRepository {
    constructor(
        @InjectRepository(User)
        private readonly usersRepository: Repository<User>,
    ) {}

    async count(): Promise<number> {
        return this.usersRepository.count();
    }

    async list(limit = 10, offset = 0): Promise<[User[], number]> {
        return this.usersRepository.findAndCount({
            skip: offset,
            take: limit,
        });
    }

    async findById(id: string): Promise<User | null> {
        return this.usersRepository.findOneBy({ id });
    }
}

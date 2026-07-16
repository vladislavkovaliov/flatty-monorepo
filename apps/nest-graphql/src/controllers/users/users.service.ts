import { Injectable, NotFoundException } from '@nestjs/common';
import { UsersRepository } from './users.repository';
import { User } from './entities/user.entity';
import { ListUserResponse } from './dto/list-user-response';

@Injectable()
export class UsersService {
    constructor(private readonly usersRepository: UsersRepository) {}

    async count(): Promise<number> {
        return this.usersRepository.count();
    }

    async list(limit = 10, offset = 0): Promise<ListUserResponse> {
        const [data, total] = await this.usersRepository.list(limit, offset);

        return { data, total };
    }

    async findById(id: string): Promise<User> {
        const user = await this.usersRepository.findById(id);

        if (!user) {
            throw new NotFoundException(`user with id ${id} not found`);
        }

        return user;
    }
}

import { Args, Int, Query, Resolver } from '@nestjs/graphql';
import { User } from './entities/user.entity';
import { UsersService } from './users.service';
import { ListUserResponse } from './dto/list-user-response';

@Resolver(() => User)
export class UsersResolver {
    constructor(private readonly usersService: UsersService) {}

    @Query(() => ListUserResponse, { name: 'userList' })
    async list(
        @Args('limit', { type: () => Int, defaultValue: 10 }) limit: number,
        @Args('offset', { type: () => Int, defaultValue: 0 }) offset: number,
    ): Promise<ListUserResponse> {
        return this.usersService.list(limit, offset);
    }

    @Query(() => User, { name: 'user' })
    async findById(
        @Args('id', { type: () => String }) id: string,
    ): Promise<User> {
        return this.usersService.findById(id);
    }
}

import { Args, Int, Query, Resolver } from '@nestjs/graphql';
import { ResidentLocation } from './entities/resident-location.entity';
import { ResidentLocationCountResponse } from './dto/resident-location-count-response';
import { ResidentLocationService } from './resident-location.service';
import { ListResidentLocationResponse } from './dto/list-resident-location-response';

@Resolver(() => ResidentLocation)
export class ResidentLocationResolver {
    constructor(private readonly residentLocationService: ResidentLocationService) {}

    @Query(() => ResidentLocationCountResponse)
    async count(): Promise<ResidentLocationCountResponse> {
        const count = await this.residentLocationService.count();
        
        return {
            total: count,
        };
    }

    @Query(() => ListResidentLocationResponse)
    async list(
        @Args('limit', { type: () => Int, defaultValue: 10 }) limit: number,
        @Args('offset', { type: () => Int, defaultValue: 0 }) offset: number,
    ): Promise<ListResidentLocationResponse> {
        return await this.residentLocationService.list(limit, offset);
    }
}

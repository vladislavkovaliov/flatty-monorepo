import { Query, Resolver } from '@nestjs/graphql';
import { ResidentLocation } from './entities/resident-location.entity';
import { ResidentLocationCountResponse } from './dto/resident-location-count-response';
import { ResidentLocationService } from './resident-location.service';

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
}

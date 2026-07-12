import { Args, Int, Mutation, Query, Resolver } from '@nestjs/graphql';
import { ResidentLocation } from './entities/resident-location.entity';
import { ResidentLocationCountResponse } from './dto/resident-location-count-response';
import { ResidentLocationService } from './resident-location.service';
import { ListResidentLocationResponse } from './dto/list-resident-location-response';
import { ResidentLocationInput } from './entities/resident-location-input.entity';
import { DeleteResidentLocationResponse } from './dto/delete-resident-location-response';

@Resolver(() => ResidentLocation)
export class ResidentLocationResolver {
    constructor(private readonly residentLocationService: ResidentLocationService) {}

    @Query(() => ResidentLocationCountResponse, { name: 'residentLocationCount' })
    async count(): Promise<ResidentLocationCountResponse> {
        const count = await this.residentLocationService.count();
        
        return {
            total: count,
        };
    }

    @Query(() => ListResidentLocationResponse, { name: 'residentLocationList' })
    async list(
        @Args('limit', { type: () => Int, defaultValue: 10 }) limit: number,
        @Args('offset', { type: () => Int, defaultValue: 0 }) offset: number,
    ): Promise<ListResidentLocationResponse> {
        return await this.residentLocationService.list(limit, offset);
    }

    @Mutation(() => ResidentLocation, { name: 'createResidentLocation' })
    async create(
        @Args('residentLocatoinData') residentLocatoinData: ResidentLocationInput,
    ): Promise<ResidentLocation> {
        return await this.residentLocationService.create(residentLocatoinData);
    }

    @Mutation(() => ResidentLocation, { name: 'updateResidentLocation' })
    async update(
        @Args('id', { type: () => Int }) id: number,
        @Args('residentLocatoinData') residentLocatoinData: ResidentLocationInput,
    ): Promise<ResidentLocation> {
        return await this.residentLocationService.update(id, residentLocatoinData);
    }

    @Mutation(() => DeleteResidentLocationResponse, { name: 'deleteResidentLocation' })
    async delete(
        @Args('id', { type: () => Int }) id: number,
    ): Promise<DeleteResidentLocationResponse> {
        return this.residentLocationService.delete(id);
    }
}

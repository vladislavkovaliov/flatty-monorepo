import { Injectable, NotFoundException } from '@nestjs/common';
import { ResidentLocationRepository } from './resident-location.repository';
import { ListResidentLocationResponse } from './dto/list-resident-location-response';
import { ResidentLocation } from './entities/resident-location.entity';
import { ResidentLocationInput } from './entities/resident-location-input.entity';

@Injectable()
export class ResidentLocationService {
    constructor(private readonly residentLocationRepository: ResidentLocationRepository) {

    }

    async count(userId: string): Promise<number> {
        return await this.residentLocationRepository.count(userId);
    }

    async list(limit = 10, offset = 0, userId: string): Promise<ListResidentLocationResponse> {
        const [data, total] =  await this.residentLocationRepository.list(limit, offset, userId);

        return { data: data, total: total };
    }

    async create(residentLocatoinData: ResidentLocationInput, userId: string): Promise<ResidentLocation> {
        return await this.residentLocationRepository.create(residentLocatoinData, userId)
    }

    async update(id: number, residentLocatoinData: ResidentLocationInput, userId: string) {
        const entity = await this.residentLocationRepository.update(id, residentLocatoinData, userId);
        
        if (!entity) {
            throw new NotFoundException(`resident location with id ${id} not found`);
        }
        
        return entity;
    }

    async delete(id: number, userId: string): Promise<{data: number}> {
        const rows = await this.residentLocationRepository.delete(id, userId);

        if (!rows.affected) {
            throw new NotFoundException(`resident location with id ${id} not found`);
        }
        
        return { data: id }
    }
}

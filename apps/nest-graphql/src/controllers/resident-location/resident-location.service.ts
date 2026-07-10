import { Injectable } from '@nestjs/common';
import { ResidentLocationRepository } from './resident-location.repository';
import { ListResidentLocationResponse } from './dto/list-resident-location-response';
import { ResidentLocation } from './entities/resident-location.entity';
import { ResidentLocationInput } from './entities/resident-location-input.entity';

@Injectable()
export class ResidentLocationService {
    constructor(private readonly residentLocationRepository: ResidentLocationRepository) {

    }

    async count(): Promise<number> {
        return await this.residentLocationRepository.count();
    }

    async list(limit = 10, offset = 0): Promise<ListResidentLocationResponse> {
        const [data, total] =  await this.residentLocationRepository.list(limit, offset);

        return { data: data, total: total };
    }

    async create(residentLocatoinData: ResidentLocationInput): Promise<ResidentLocation> {
        return await this.residentLocationRepository.create(residentLocatoinData)
    }
}

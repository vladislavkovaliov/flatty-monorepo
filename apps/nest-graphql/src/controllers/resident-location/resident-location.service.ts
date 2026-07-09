import { Injectable } from '@nestjs/common';
import { ResidentLocationRepository } from './resident-location.repository';

@Injectable()
export class ResidentLocationService {
    constructor(private readonly residentLocationRepository: ResidentLocationRepository) {

    }

    async count(): Promise<number> {
        return await this.residentLocationRepository.count();
    }
}

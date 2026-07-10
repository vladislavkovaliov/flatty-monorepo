import { Injectable } from '@nestjs/common';
import { InjectDataSource, InjectRepository, TypeOrmDataSourceFactory } from '@nestjs/typeorm';
import { Repository } from 'typeorm'
import { ResidentLocation } from './entities/resident-location.entity';
import { ResidentLocationInput } from './entities/resident-location-input.entity';

@Injectable()
export class ResidentLocationRepository {
    constructor(
        @InjectRepository(ResidentLocation)
        private readonly residentLocationRepository: Repository<ResidentLocation>,
        
        @InjectDataSource()
        private readonly dataSource: TypeOrmDataSourceFactory,
    ) {}

    async count(): Promise<number> {
        return this.residentLocationRepository.count();
    }

    async list(limit = 10, offset = 0): Promise<[ResidentLocation[], number]> {
        return this.residentLocationRepository.findAndCount({
            skip: offset,
            take: limit,
        });
    }

    async create(residentLocatoinData: ResidentLocationInput): Promise<ResidentLocation> {
        const entity = this.residentLocationRepository.create({
            country: residentLocatoinData.country,
            city: residentLocatoinData.city,
            postalCode: residentLocatoinData.postalCode,
            street: residentLocatoinData.street,
            house: residentLocatoinData.house,
            apartment: residentLocatoinData.apartment,
        });

        return this.residentLocationRepository.save(entity); 
    }
}

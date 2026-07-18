import { Injectable } from '@nestjs/common';
import { InjectDataSource, InjectRepository, TypeOrmDataSourceFactory } from '@nestjs/typeorm';
import { Repository } from 'typeorm'
import { ResidentLocation } from './entities/resident-location.entity';
import { ResidentLocationInput } from './entities/resident-location-input.entity';
import { DeleteResult } from 'typeorm/browser';

@Injectable()
export class ResidentLocationRepository {
    constructor(
        @InjectRepository(ResidentLocation)
        private readonly residentLocationRepository: Repository<ResidentLocation>,
        
        @InjectDataSource()
        private readonly dataSource: TypeOrmDataSourceFactory,
    ) {}

    async count(userId: string): Promise<number> {
        return this.residentLocationRepository.count({ where: { userId: userId } });
    }

    async list(limit = 10, offset = 0, userId: string): Promise<[ResidentLocation[], number]> {
        return this.residentLocationRepository.findAndCount({
            where: { userId: userId },
            skip: offset,
            take: limit,
        });
    }

    async create(residentLocatoinData: ResidentLocationInput, userId: string): Promise<ResidentLocation> {
        const entity = this.residentLocationRepository.create({
            userId: userId,
            country: residentLocatoinData.country,
            city: residentLocatoinData.city,
            postalCode: residentLocatoinData.postalCode,
            street: residentLocatoinData.street,
            house: residentLocatoinData.house,
            apartment: residentLocatoinData.apartment,
        });

        return this.residentLocationRepository.save(entity); 
    }

    async update(id: number, residentLocatoinData: ResidentLocationInput, userId: string): Promise<ResidentLocation | undefined> {
        const entity = await this.residentLocationRepository.findOneBy({ id, userId });
        
        if (!entity) {
            return undefined
        }

        const merged = this.residentLocationRepository.merge(entity, {
            country: residentLocatoinData.country,
            city: residentLocatoinData.city,
            postalCode: residentLocatoinData.postalCode,
            street: residentLocatoinData.street,
            house: residentLocatoinData.house,
            apartment: residentLocatoinData.apartment,
        });

        return this.residentLocationRepository.save(merged);
    }

    async delete(id: number, userId: string): Promise<DeleteResult> {
        return this.residentLocationRepository.delete({
            id: id,
            userId: userId
        });
    }
}

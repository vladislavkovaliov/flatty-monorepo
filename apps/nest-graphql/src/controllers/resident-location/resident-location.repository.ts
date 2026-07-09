import { Injectable } from '@nestjs/common';
import { InjectDataSource, InjectRepository, TypeOrmDataSourceFactory } from '@nestjs/typeorm';
import { Repository } from 'typeorm'
import { ResidentLocation } from './entities/resident-location.entity';

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
}

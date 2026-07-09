import { Module } from '@nestjs/common';
import {ResidentLocationService} from './resident-location.service'
import { ResidentLocationResolver } from './resident-location.resolver';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ResidentLocation } from './entities/resident-location.entity';
import { ResidentLocationRepository } from './resident-location.repository'

@Module({
    imports: [TypeOrmModule.forFeature([ResidentLocation])],
    providers: [ResidentLocationRepository, ResidentLocationService, ResidentLocationResolver]
})
export class ResidentLocationModule {}

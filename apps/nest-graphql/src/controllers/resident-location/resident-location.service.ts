import { Injectable, NotFoundException, Inject, Logger } from '@nestjs/common';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { Cache } from 'cache-manager';
import { ResidentLocationRepository } from './resident-location.repository';
import { ListResidentLocationResponse } from './dto/list-resident-location-response';
import { ResidentLocation } from './entities/resident-location.entity';
import { ResidentLocationInput } from './entities/resident-location-input.entity';
import { ResidentLocationServiceKeysManager } from './resident-location-keys.manager';

@Injectable()
export class ResidentLocationService {
    private readonly logger = new Logger(ResidentLocationService.name, { timestamp: true });

    constructor(
        @Inject(CACHE_MANAGER) private cacheManager: Cache,
        private readonly residentLocationRepository: ResidentLocationRepository,
    ) {}

    async count(userId: string): Promise<number> {
        const version = await this.getCacheVersion(userId);

        const ttl = ResidentLocationServiceKeysManager.SECONDS_30;
        const key = ResidentLocationServiceKeysManager.getKeyCount(version, userId);

        const cached = await this.cacheManager.get<number>(key);

        if (cached !== undefined) {
            this.logger.log(`Return resident location count from cache = ${cached}`);

            return cached;
        }

        this.logger.log(`No resident location count in cache`);

        const count = await this.residentLocationRepository.count(userId);

        void this.cacheManager.set(key, count, ttl).catch((error) => {
            this.logger.error('Failed to write resident location count to cache', error);
        });

        this.logger.log(`Resident location count is written in cache = ${count}`);

        return count;
    }

    async list(limit = 10, offset = 0, userId: string): Promise<ListResidentLocationResponse> {
        const version = await this.getCacheVersion(userId);

        const key = ResidentLocationServiceKeysManager.getKeyList(limit, offset, version, userId);
        const ttl = ResidentLocationServiceKeysManager.SECONDS_30;

        type ReturnData = { data: ResidentLocation[]; total: number; limit: number; offset: number; };

        const cached = await this.cacheManager.get<ReturnData>(key);

        if (cached !== undefined) {
            return cached;
        }

        const [data, total] = await this.residentLocationRepository.list(limit, offset, userId);

        void this.cacheManager.set<ReturnData>(key, { data, total, limit, offset }, ttl).catch((error) => {
            this.logger.error(`Failed to write resident location list to cache`, error);
        });

        return { data, total };
    }

    async create(residentLocatoinData: ResidentLocationInput, userId: string): Promise<ResidentLocation> {
        const residentLocation = await this.residentLocationRepository.create(residentLocatoinData, userId);

        await this.incrementCacheVersion(userId);

        return residentLocation;
    }

    async update(id: number, residentLocatoinData: ResidentLocationInput, userId: string) {
        const entity = await this.residentLocationRepository.update(id, residentLocatoinData, userId);

        if (!entity) {
            throw new NotFoundException(`resident location with id ${id} not found`);
        }

        await this.incrementCacheVersion(userId);

        return entity;
    }

    async delete(id: number, userId: string): Promise<{ data: number }> {
        const rows = await this.residentLocationRepository.delete(id, userId);

        if (!rows.affected) {
            throw new NotFoundException(`resident location with id ${id} not found`);
        }

        await this.incrementCacheVersion(userId);

        return { data: id };
    }

    private async getCacheVersion(userId: string) {
        return await this.cacheManager.get<number>(ResidentLocationServiceKeysManager.getKeyVersion(userId)) || 0;
    }

    private async incrementCacheVersion(userId: string) {
        const cachedVersion = await this.getCacheVersion(userId);

        const key = ResidentLocationServiceKeysManager.getKeyVersion(userId);
        const ttl = ResidentLocationServiceKeysManager.WEEK;

        await this.cacheManager.set(key, cachedVersion + 1, ttl);

        this.logger.log(`Version is updated from ${cachedVersion} to ${cachedVersion + 1} for user ${userId}`);
    }
}

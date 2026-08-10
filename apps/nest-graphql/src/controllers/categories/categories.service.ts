import { Injectable, NotFoundException, Inject, Logger } from '@nestjs/common';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { Cache } from 'cache-manager';
import { CategoryRepository } from './categories.repository';
import { CategoryInput } from './entities/category-input.entity';
import { Category } from './entities/category.entity';
import { ListCategoryResponse } from './dto/list-category-response';
import { CategoryServiceKeysManager } from './categories-keys.manager';

@Injectable()
export class CategoryService {
    private readonly logger = new Logger(CategoryService.name, { timestamp: true });

    constructor(
        @Inject(CACHE_MANAGER) private cacheManager: Cache,
        private readonly categoryRepository: CategoryRepository,
    ) {}

    async count(): Promise<number> {
        const version = await this.getCacheVersion();

        const ttl = CategoryServiceKeysManager.SECONDS_30;
        const key = CategoryServiceKeysManager.getKeyCount(version);

        let cached = await this.cacheManager.get<number>(key);

        if (cached !== undefined) {
            this.logger.log(`Return category count from cache = ${cached}`);

            return cached;
        }

        this.logger.log(`No category count in cache`);

        const count = await this.categoryRepository.count();

        void this.cacheManager.set(key, count, ttl).catch((error) => {
            this.logger.error('Failed to write category count to cache', error);
        });

        this.logger.log(`Category count is written in cache = ${count}`);

        return count;
    }

    async list(limit = 10, offset = 0): Promise<ListCategoryResponse> {
        const version = await this.getCacheVersion();

        const key = CategoryServiceKeysManager.getKeyList(limit, offset, version);
        const ttl = CategoryServiceKeysManager.SECONDS_30;
        
        type ReturnData = { data: Category[]; total: number; limit: number; offset: number; };

        const cached = await this.cacheManager.get<ReturnData>(key);

        if (cached !== undefined) {
            return cached;
        }
        
        const [data, total] = await this.categoryRepository.list(limit, offset);

        void this.cacheManager.set<ReturnData>(key, { data, total, limit, offset }, ttl).catch((error) => {
            this.logger.error(`Failed to write category list to cache`, error);
        });
    

        return { data, total };
    }

    async create(categoryData: CategoryInput): Promise<Category> {
        const category = await this.categoryRepository.create(categoryData);
        
        await this.incrementCacheVersion();

        return category;
    }

    async update(id: number, categoryData: CategoryInput): Promise<Category> {
        const entity = await this.categoryRepository.update(id, categoryData);

        if (!entity) {
            throw new NotFoundException(`category with id ${id} not found`);
        }

        await this.incrementCacheVersion();

        return entity;
    }

    async delete(id: number): Promise<{ data: number }> {
        const rows = await this.categoryRepository.delete(id);

        if (!rows.affected) {
            throw new NotFoundException(`category with id ${id} not found`);
        }

        await this.incrementCacheVersion();

        return { data: id };
    }

    private async getCacheVersion() {
        return await this.cacheManager.get<number>(CategoryServiceKeysManager.getKeyVersion()) || 0;
    }

    private async incrementCacheVersion() {
        const cachedVersion = await this.getCacheVersion();
        
        const key = CategoryServiceKeysManager.getKeyVersion();
        const ttl = CategoryServiceKeysManager.WEEK;
        
        await this.cacheManager.set(key, cachedVersion + 1, ttl);

        this.logger.log(`Version is updated from ${cachedVersion} to ${cachedVersion + 1}`);
    }
}

import { Injectable, NotFoundException } from '@nestjs/common';
import { Inject, Logger } from "@nestjs/common"
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { Cache } from 'cache-manager';
import { CategoryRepository } from './categories.repository';
import { CategoryInput } from './entities/category-input.entity';
import { Category } from './entities/category.entity';
import { ListCategoryResponse } from './dto/list-category-response';

@Injectable()
export class CategoryService {
    private readonly logger = new Logger(CategoryService.name, { timestamp: true });

    constructor(
        @Inject(CACHE_MANAGER) private cacheManager: Cache,
        private readonly categoryRepository: CategoryRepository,
    ) {}

    async count(): Promise<number> {
        const key = `CategoryResolver.name:count`;

        const countFromCache = await this.cacheManager.get<number>(key);
        let count: number | undefined = undefined;

        if (countFromCache === undefined) {
            this.logger.log(`No category count in cache = ${countFromCache}`);

            count = await this.categoryRepository.count();

            this.cacheManager.set(`CategoryResolver.name:count`, count, 30000);

            this.logger.log(`Category count is written in cache = ${count}`);
        } else {
            count = countFromCache;

            this.logger.log(`Return category count from cacha = ${countFromCache}`);
        }

        return Promise.resolve(count);
    }

    async list(limit = 10, offset = 0): Promise<ListCategoryResponse> {
        const [data, total] = await this.categoryRepository.list(limit, offset);

        return { data, total };
    }

    async create(categoryData: CategoryInput): Promise<Category> {
        return await this.categoryRepository.create(categoryData);
    }

    async update(id: number, categoryData: CategoryInput): Promise<Category> {
        const entity = await this.categoryRepository.update(id, categoryData);

        if (!entity) {
            throw new NotFoundException(`category with id ${id} not found`);
        }

        return entity;
    }

    async delete(id: number): Promise<{ data: number }> {
        const rows = await this.categoryRepository.delete(id);

        if (!rows.affected) {
            throw new NotFoundException(`category with id ${id} not found`);
        }

        return { data: id };
    }
}

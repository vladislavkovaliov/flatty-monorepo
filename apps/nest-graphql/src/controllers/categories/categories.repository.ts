import { Injectable } from '@nestjs/common';
import { InjectDataSource, InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm'
import { Category } from './entities/category.entity';
import { CategoryInput } from './entities/category-input.entity';
import { DeleteResult } from 'typeorm/browser';

@Injectable()
export class CategoryRepository {
    constructor(
        @InjectRepository(Category)
        private readonly categoryRepository: Repository<Category>,

        @InjectDataSource()
        private readonly dataSource: any,
    ) {}

    async count(): Promise<number> {
        return this.categoryRepository.count();
    }

    async list(limit = 10, offset = 0): Promise<[Category[], number]> {
        return this.categoryRepository.findAndCount({
            skip: offset,
            take: limit,
        });
    }

    async create(categoryData: CategoryInput): Promise<Category> {
        const entity = this.categoryRepository.create({
            name: categoryData.name,
            description: categoryData.description,
        });

        return this.categoryRepository.save(entity);
    }

    async update(id: number, categoryData: CategoryInput): Promise<Category | undefined> {
        const entity = await this.categoryRepository.findOneBy({ id });

        if (!entity) {
            return undefined
        }

        const merged = this.categoryRepository.merge(entity, {
            name: categoryData.name,
            description: categoryData.description,
        });

        return this.categoryRepository.save(merged);
    }

    async delete(id: number): Promise<DeleteResult> {
        return this.categoryRepository.delete({
            id: id
        });
    }
}

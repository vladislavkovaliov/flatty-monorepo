import { Injectable, NotFoundException } from '@nestjs/common';
import { CategoryRepository } from './categories.repository';
import { CategoryInput } from './entities/category-input.entity';
import { Category } from './entities/category.entity';
import { ListCategoryResponse } from './dto/list-category-response';

@Injectable()
export class CategoryService {
    constructor(private readonly categoryRepository: CategoryRepository) {}

    async count(): Promise<number> {
        return await this.categoryRepository.count();
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

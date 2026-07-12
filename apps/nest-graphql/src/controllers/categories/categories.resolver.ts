import { Args, Int, Mutation, Query, Resolver } from '@nestjs/graphql';
import { Category } from './entities/category.entity';
import { CategoryCountResponse } from './dto/category-count-response';
import { CategoryService } from './categories.service';
import { ListCategoryResponse } from './dto/list-category-response';
import { CategoryInput } from './entities/category-input.entity';
import { DeleteCategoryResponse } from './dto/delete-category-response';

@Resolver(() => Category)
export class CategoryResolver {
    constructor(private readonly categoryService: CategoryService) {}

    @Query(() => CategoryCountResponse, { name: 'categoryCount' })
    async count(): Promise<CategoryCountResponse> {
        const count = await this.categoryService.count();

        return {
            total: count,
        };
    }

    @Query(() => ListCategoryResponse, { name: 'categoryList' })
    async list(
        @Args('limit', { type: () => Int, defaultValue: 10 }) limit: number,
        @Args('offset', { type: () => Int, defaultValue: 0 }) offset: number,
    ): Promise<ListCategoryResponse> {
        return await this.categoryService.list(limit, offset);
    }

    @Mutation(() => Category, { name: 'createCategory' })
    async create(
        @Args('categoryData') categoryData: CategoryInput,
    ): Promise<Category> {
        return await this.categoryService.create(categoryData);
    }

    @Mutation(() => Category, { name: 'updateCategory' })
    async update(
        @Args('id', { type: () => Int }) id: number,
        @Args('categoryData') categoryData: CategoryInput,
    ): Promise<Category> {
        return await this.categoryService.update(id, categoryData);
    }

    @Mutation(() => DeleteCategoryResponse, { name: 'deleteCategory' })
    async delete(
        @Args('id', { type: () => Int }) id: number,
    ): Promise<DeleteCategoryResponse> {
        return this.categoryService.delete(id);
    }
}

import { Module } from '@nestjs/common';
import { CategoryService } from './categories.service'
import { CategoryResolver } from './categories.resolver';
import { TypeOrmModule } from '@nestjs/typeorm';
import { Category } from './entities/category.entity';
import { CategoryRepository } from './categories.repository'

@Module({
    imports: [TypeOrmModule.forFeature([Category])],
    providers: [CategoryRepository, CategoryService, CategoryResolver]
})
export class CategoriesModule {}

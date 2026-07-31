import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { Category } from '../categories/entities/category.entity';
import { Expense } from './entities/expense.entity';
import { ExpenseRepository } from './expenses.repository';
import { ExpenseResolver } from './expenses.resolver';
import { ExpenseService } from './expenses.service';
import { ResidentLocation } from '../resident-location/entities/resident-location.entity';

@Module({
  imports: [TypeOrmModule.forFeature([Expense, Category, ResidentLocation])],
  providers: [ExpenseRepository, ExpenseService, ExpenseResolver],
})
export class ExpensesModule {}

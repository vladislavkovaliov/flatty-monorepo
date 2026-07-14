import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { Category } from '../categories/entities/category.entity';
import { Expense } from './entities/expense.entity';
import { ExpenseRepository } from './expenses.repository';
import { ExpenseResolver } from './expenses.resolver';
import { ExpenseService } from './expenses.service';

@Module({
  imports: [TypeOrmModule.forFeature([Expense, Category])],
  providers: [ExpenseRepository, ExpenseService, ExpenseResolver],
})
export class ExpensesModule {}

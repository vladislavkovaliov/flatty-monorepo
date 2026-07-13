import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { Expense } from './entities/expense.entity';
import { ExpenseRepository } from './expenses.repository';
import { ExpenseResolver } from './expenses.resolver';
import { ExpenseService } from './expenses.service';

@Module({
  imports: [TypeOrmModule.forFeature([Expense])],
  providers: [ExpenseRepository, ExpenseService, ExpenseResolver],
})
export class ExpensesModule {}

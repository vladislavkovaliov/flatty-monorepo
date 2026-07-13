import { Injectable } from '@nestjs/common';
import { InjectDataSource, InjectRepository } from '@nestjs/typeorm';
import { DeleteResult, Repository } from 'typeorm';
import { ExpenseInput } from './entities/expense-input.entity';
import { Expense } from './entities/expense.entity';

@Injectable()
export class ExpenseRepository {
  constructor(
    @InjectRepository(Expense)
    private readonly expenseRepository: Repository<Expense>,

    @InjectDataSource()
    private readonly dataSource: any,
  ) {}

  async count(): Promise<number> {
    return this.expenseRepository.count();
  }

  async list(limit = 10, offset = 0): Promise<[Expense[], number]> {
    return this.expenseRepository.findAndCount({
      skip: offset,
      take: limit,
    });
  }

  async create(expenseData: ExpenseInput): Promise<Expense> {
    const entity = this.expenseRepository.create({
      residentLocationId: expenseData.residentLocationId,
      categoryId: expenseData.categoryId,
      amount: expenseData.amount,
      month: expenseData.month,
      year: expenseData.year,
    });
    return this.expenseRepository.save(entity);
  }

  async update(id: number, expenseData: ExpenseInput): Promise<Expense | undefined> {
    const entity = await this.expenseRepository.findOneBy({ id });
    if (!entity) {
      return undefined;
    }
    const merged = this.expenseRepository.merge(entity, {
      residentLocationId: expenseData.residentLocationId,
      categoryId: expenseData.categoryId,
      amount: expenseData.amount,
      month: expenseData.month,
      year: expenseData.year,
    });
    return this.expenseRepository.save(merged);
  }

  async delete(id: number): Promise<DeleteResult> {
    return this.expenseRepository.delete({ id });
  }
}

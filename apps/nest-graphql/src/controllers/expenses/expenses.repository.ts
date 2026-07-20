import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Expense } from './entities/expense.entity';

@Injectable()
export class ExpenseRepository {
  constructor(
    @InjectRepository(Expense)
    private readonly expenseRepository: Repository<Expense>,
  ) {}

  async count(): Promise<number> {
    return this.expenseRepository.count();
  }

  async list(limit = 10, offset = 0): Promise<[Expense[], number]> {
    return this.expenseRepository.findAndCount({
      skip: offset,
      take: limit,
      order: {
        id: 'ASC'
      },
      relations: { category: true },
    });
  }
}


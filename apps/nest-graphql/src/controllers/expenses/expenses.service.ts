import { Injectable, NotFoundException } from '@nestjs/common';
import { ExpenseRepository } from './expenses.repository';
import { ExpenseInput } from './entities/expense-input.entity';
import { ListExpenseResponse } from './dto/list-expense-response';

@Injectable()
export class ExpenseService {
  constructor(private readonly expenseRepository: ExpenseRepository) {}

  async count(): Promise<number> {
    return this.expenseRepository.count();
  }

  async list(limit = 10, offset = 0): Promise<ListExpenseResponse> {
    const [data, total] = await this.expenseRepository.list(limit, offset);
    return { data, total };
  }

  async create(expenseData: ExpenseInput) {
    return this.expenseRepository.create(expenseData);
  }

  async update(id: number, expenseData: ExpenseInput) {
    const entity = await this.expenseRepository.update(id, expenseData);
    if (!entity) {
      throw new NotFoundException(`expense with id ${id} not found`);
    }
    return entity;
  }

  async delete(id: number): Promise<{ data: number }> {
    const rows = await this.expenseRepository.delete(id);
    if (!rows.affected) {
      throw new NotFoundException(`expense with id ${id} not found`);
    }
    return { data: id };
  }
}

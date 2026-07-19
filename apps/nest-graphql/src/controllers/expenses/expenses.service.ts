import { Injectable } from '@nestjs/common';
import { ExpenseRepository } from './expenses.repository';
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
}

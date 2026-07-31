import { Injectable } from '@nestjs/common';
import { ExpenseRepository } from './expenses.repository';
import { ListExpenseResponse } from './dto/list-expense-response';

@Injectable()
export class ExpenseService {
  constructor(private readonly expenseRepository: ExpenseRepository) {}

  async count(residentLocationId: number, userId: string): Promise<number> {
    return this.expenseRepository.count(residentLocationId, userId);
  }

  async list(residentLocationId: number, userId: string, limit = 10, offset = 0): Promise<ListExpenseResponse> {
    const [data, total] = await this.expenseRepository.list(residentLocationId, userId, limit, offset);
    return { data, total };
  }
}

import { Args, Int, Mutation, Query, Resolver } from '@nestjs/graphql';
import { ExpenseCountResponse } from './dto/expense-count-response';
import { DeleteExpenseResponse } from './dto/delete-expense-response';
import { ListExpenseResponse } from './dto/list-expense-response';
import { ExpenseInput } from './entities/expense-input.entity';
import { Expense } from './entities/expense.entity';
import { ExpenseService } from './expenses.service';

@Resolver(() => Expense)
export class ExpenseResolver {
  constructor(private readonly expenseService: ExpenseService) {}

  @Query(() => ExpenseCountResponse, { name: 'expenseCount' })
  async count(): Promise<ExpenseCountResponse> {
    const total = await this.expenseService.count();
    return { total };
  }

  @Query(() => ListExpenseResponse, { name: 'expenseList' })
  async list(
    @Args('limit', { type: () => Int, defaultValue: 10 }) limit: number,
    @Args('offset', { type: () => Int, defaultValue: 0 }) offset: number,
  ): Promise<ListExpenseResponse> {
    return this.expenseService.list(limit, offset);
  }

  @Mutation(() => Expense, { name: 'createExpense' })
  async create(
    @Args('expenseData') expenseData: ExpenseInput,
  ): Promise<Expense> {
    return this.expenseService.create(expenseData);
  }

  @Mutation(() => Expense, { name: 'updateExpense' })
  async update(
    @Args('id', { type: () => Int }) id: number,
    @Args('expenseData') expenseData: ExpenseInput,
  ): Promise<Expense> {
    return this.expenseService.update(id, expenseData);
  }

  @Mutation(() => DeleteExpenseResponse, { name: 'deleteExpense' })
  async delete(
    @Args('id', { type: () => Int }) id: number,
  ): Promise<DeleteExpenseResponse> {
    return this.expenseService.delete(id);
  }
}

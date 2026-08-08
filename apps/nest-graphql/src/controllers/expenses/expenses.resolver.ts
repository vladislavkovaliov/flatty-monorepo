import { Args, Int, Query, Resolver } from '@nestjs/graphql';
import { ExpenseCountResponse } from './dto/expense-count-response';
import { ListExpenseResponse } from './dto/list-expense-response';
import { Expense } from './entities/expense.entity';
import { ExpenseService } from './expenses.service';
import { CurrentUser } from '../../auth/current-user.decorator';

@Resolver(() => Expense)
export class ExpenseResolver {
  constructor(private readonly expenseService: ExpenseService) {}

  @Query(() => ExpenseCountResponse, { name: 'expenseCount' })
  async count(
    @CurrentUser() userId: string,
    @Args('residentLocationId', { type: () => Int }) residentLocationId: number,
  ): Promise<ExpenseCountResponse> {
    console.log({residentLocationId, userId})
    const total = await this.expenseService.count(residentLocationId, userId);
    return { total };
  }

  @Query(() => ListExpenseResponse, { name: 'expenseList' })
  async list(
    @CurrentUser() userId: string,
    @Args('residentLocationId', { type: () => Int }) residentLocationId: number,
    @Args('limit', { type: () => Int, defaultValue: 10 }) limit: number,
    @Args('offset', { type: () => Int, defaultValue: 0 }) offset: number,
  ): Promise<ListExpenseResponse> {
    return this.expenseService.list(residentLocationId, userId, limit, offset);
  }

}

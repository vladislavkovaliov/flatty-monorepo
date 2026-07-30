import { Args, Int, Query, Resolver } from '@nestjs/graphql';
import { CurrentUser } from '../../auth/current-user.decorator';
import { ExpenseStatsService } from './expense-stats.service';
import { ListMonthlyTotalsResponse } from './dto/list-monthly-totals-response';
import { ListMonthlyAveragesResponse } from './dto/list-monthly-averages-response';

@Resolver()
export class ExpenseStatsResolver {
  constructor(private readonly statsService: ExpenseStatsService) {}

  @Query(() => ListMonthlyTotalsResponse, { name: 'expenseMonthlyTotals' })
  async listTotals(
    @CurrentUser() userId: string,
    @Args('residentLocationId', { type: () => Int }) residentLocationId: number,
    @Args('month', { type: () => Int, nullable: true }) month?: number,
    @Args('year', { type: () => Int, nullable: true }) year?: number,
  ): Promise<ListMonthlyTotalsResponse> {
    return this.statsService.listTotals(residentLocationId, userId, month, year);
  }

  @Query(() => ListMonthlyAveragesResponse, { name: 'expenseMonthlyAverages' })
  async listAverages(
    @CurrentUser() userId: string,
    @Args('residentLocationId', { type: () => Int }) residentLocationId: number,
    @Args('month', { type: () => Int, nullable: true }) month?: number,
    @Args('year', { type: () => Int, nullable: true }) year?: number,
  ): Promise<ListMonthlyAveragesResponse> {
    return this.statsService.listAverages(residentLocationId, userId, month, year);
  }
}

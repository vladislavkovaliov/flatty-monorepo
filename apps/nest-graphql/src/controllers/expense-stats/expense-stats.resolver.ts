import { Args, Int, Query, Resolver } from '@nestjs/graphql';
import { ExpenseStatsService } from './expense-stats.service';
import { ListMonthlyTotalsResponse } from './dto/list-monthly-totals-response';
import { ListMonthlyAveragesResponse } from './dto/list-monthly-averages-response';

@Resolver()
export class ExpenseStatsResolver {
  constructor(private readonly statsService: ExpenseStatsService) {}

  @Query(() => ListMonthlyTotalsResponse, { name: 'expenseMonthlyTotals' })
  async listTotals(
    @Args('month', { type: () => Int, nullable: true }) month?: number,
    @Args('year', { type: () => Int, nullable: true }) year?: number,
  ): Promise<ListMonthlyTotalsResponse> {
    return this.statsService.listTotals(month, year);
  }

  @Query(() => ListMonthlyAveragesResponse, { name: 'expenseMonthlyAverages' })
  async listAverages(
    @Args('month', { type: () => Int, nullable: true }) month?: number,
    @Args('year', { type: () => Int, nullable: true }) year?: number,
  ): Promise<ListMonthlyAveragesResponse> {
    return this.statsService.listAverages(month, year);
  }
}

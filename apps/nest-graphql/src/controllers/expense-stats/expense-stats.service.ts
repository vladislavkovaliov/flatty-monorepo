import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { ExpenseMonthlyTotal } from './entities/expense-monthly-total.entity';
import { ExpenseMonthlyAverage } from './entities/expense-monthly-average.entity';
import { ListMonthlyTotalsResponse } from './dto/list-monthly-totals-response';
import { ListMonthlyAveragesResponse } from './dto/list-monthly-averages-response';

@Injectable()
export class ExpenseStatsService {
  constructor(
    @InjectRepository(ExpenseMonthlyTotal)
    private readonly totalsRepo: Repository<ExpenseMonthlyTotal>,
    @InjectRepository(ExpenseMonthlyAverage)
    private readonly averagesRepo: Repository<ExpenseMonthlyAverage>,
  ) {}

  async listTotals(month?: number, year?: number): Promise<ListMonthlyTotalsResponse> {
    const where: Record<string, number> = {};
    if (month !== undefined) where.month = month;
    if (year !== undefined) where.year = year;

    const data = await this.totalsRepo.find({
      where,
      order: { year: 'DESC', month: 'DESC' },
    });
    return { data };
  }

  async listAverages(month?: number, year?: number): Promise<ListMonthlyAveragesResponse> {
    const where: Record<string, number> = {};
    if (month !== undefined) where.month = month;
    if (year !== undefined) where.year = year;

    const data = await this.averagesRepo.find({
      where,
      order: { year: 'DESC', month: 'DESC' },
    });
    return { data };
  }
}

import { ForbiddenException, Inject, Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { ExpenseMonthlyTotal } from './entities/expense-monthly-total.entity';
import { ExpenseMonthlyAverage } from './entities/expense-monthly-average.entity';
import { ResidentLocation } from '../resident-location/entities/resident-location.entity';
import { ListMonthlyTotalsResponse } from './dto/list-monthly-totals-response';
import { ListMonthlyAveragesResponse } from './dto/list-monthly-averages-response';

@Injectable()
export class ExpenseStatsService {
  constructor(
    @InjectRepository(ExpenseMonthlyTotal)
    private readonly totalsRepo: Repository<ExpenseMonthlyTotal>,
    @InjectRepository(ExpenseMonthlyAverage)
    private readonly averagesRepo: Repository<ExpenseMonthlyAverage>,
    @InjectRepository(ResidentLocation)
    private readonly residentLocationRepo: Repository<ResidentLocation>,
  ) {}

  private async assertOwnership(residentLocationId: number, userId: string): Promise<void> {
    const owned = await this.residentLocationRepo.findOneBy({ id: residentLocationId, userId });
    if (!owned) {
      throw new ForbiddenException('resident location not found for current user');
    }
  }

  async listTotals(
    residentLocationId: number,
    userId: string,
    month?: number,
    year?: number,
  ): Promise<ListMonthlyTotalsResponse> {
    await this.assertOwnership(residentLocationId, userId);

    const where: Record<string, number> = { residentLocationId };

    if (month !== undefined) {
      where.month = month;
    }

    if (year !== undefined) {
      where.year = year;
    }

    const data = await this.totalsRepo.find({
      where,
      order: { year: 'DESC', month: 'DESC' },
    });

    return { data };
  }

  async listAverages(
    residentLocationId: number,
    userId: string,
    month?: number,
    year?: number,
  ): Promise<ListMonthlyAveragesResponse> {
    await this.assertOwnership(residentLocationId, userId);

    const where: Record<string, number> = { residentLocationId };
    if (month !== undefined) where.month = month;
    if (year !== undefined) where.year = year;

    const data = await this.averagesRepo.find({
      where,
      order: { year: 'DESC', month: 'DESC' },
    });
    return { data };
  }
}



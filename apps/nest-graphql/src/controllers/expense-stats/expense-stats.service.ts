import { ForbiddenException, Inject, Injectable, Logger } from '@nestjs/common';
import { CACHE_MANAGER } from '@nestjs/cache-manager';
import { Cache } from 'cache-manager';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { ExpenseMonthlyTotal } from './entities/expense-monthly-total.entity';
import { ExpenseMonthlyAverage } from './entities/expense-monthly-average.entity';
import { ResidentLocation } from '../resident-location/entities/resident-location.entity';
import { ListMonthlyTotalsResponse } from './dto/list-monthly-totals-response';
import { ListMonthlyAveragesResponse } from './dto/list-monthly-averages-response';
import { ExpenseKeyManager } from './expense-keys.manager';

@Injectable()
export class ExpenseStatsService {
  private readonly logger = new Logger(ExpenseStatsService.name, { timestamp: true });

  constructor(
    @Inject(CACHE_MANAGER) private readonly cacheManager: Cache,
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

    const key = ExpenseKeyManager.getKeyListTotal(userId, residentLocationId, month, year);
    const ttl = ExpenseKeyManager.SECONDS_30;

    const cached = await this.cacheManager.get<ListMonthlyTotalsResponse>(key);
    if (cached !== undefined) {
      return cached;
    }

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

    const response: ListMonthlyTotalsResponse = { data };

    void this.cacheManager.set(key, response, ttl).catch((error) => {
      this.logger.error('Failed to write expense monthly totals to cache', error);
    });

    return response;
  }

  async listAverages(
    residentLocationId: number,
    userId: string,
    month?: number,
    year?: number,
  ): Promise<ListMonthlyAveragesResponse> {
    await this.assertOwnership(residentLocationId, userId);

    const key = ExpenseKeyManager.getKeyListAverages(userId, residentLocationId, month, year);
    const ttl = ExpenseKeyManager.SECONDS_30;

    const cached = await this.cacheManager.get<ListMonthlyAveragesResponse>(key);
    if (cached !== undefined) {
      return cached;
    }

    const where: Record<string, number> = { residentLocationId };
    if (month !== undefined) where.month = month;
    if (year !== undefined) where.year = year;

    const data = await this.averagesRepo.find({
      where,
      order: { year: 'DESC', month: 'DESC' },
    });

    const response: ListMonthlyAveragesResponse = { data };

    void this.cacheManager.set(key, response, ttl).catch((error) => {
      this.logger.error('Failed to write expense monthly averages to cache', error);
    });

    return response;
  }
}
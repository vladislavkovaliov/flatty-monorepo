import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ExpenseMonthlyTotal } from './entities/expense-monthly-total.entity';
import { ExpenseMonthlyAverage } from './entities/expense-monthly-average.entity';
import { ResidentLocation } from '../resident-location/entities/resident-location.entity';
import { ExpenseStatsResolver } from './expense-stats.resolver';
import { ExpenseStatsService } from './expense-stats.service';

@Module({
  imports: [TypeOrmModule.forFeature([ExpenseMonthlyTotal, ExpenseMonthlyAverage, ResidentLocation])],
  providers: [ExpenseStatsService, ExpenseStatsResolver],
})
export class ExpenseStatsModule {}

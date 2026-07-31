import { ForbiddenException, Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Expense } from './entities/expense.entity';
import { ResidentLocation } from '../resident-location/entities/resident-location.entity';

@Injectable()
export class ExpenseRepository {
  constructor(
    @InjectRepository(Expense)
    private readonly expenseRepository: Repository<Expense>,
    @InjectRepository(ResidentLocation)
    private readonly residentLocationRepo: Repository<ResidentLocation>,
  ) {}

  private async assertOwnership(residentLocationId: number, userId: string): Promise<void> {
    const owned = await this.residentLocationRepo.findOneBy({ id: residentLocationId, userId });
    if (!owned) {
      throw new ForbiddenException('resident location not found for current user');
    }
  }

  async count(residentLocationId: number, userId: string): Promise<number> {
    await this.assertOwnership(residentLocationId, userId);

    return this.expenseRepository.count({
      where: {
        residentLocationId: residentLocationId
      }
    });
  }

  async list(residentLocationId: number, userId: string, limit = 10, offset = 0): Promise<[Expense[], number]> {
    await this.assertOwnership(residentLocationId, userId);

    return this.expenseRepository.findAndCount({
      skip: offset,
      take: limit,
      order: {
        year: 'DESC',
        month: 'DESC'
      },
      relations: { category: true },
      where: {
        residentLocationId: residentLocationId
      }
    });
  }
}


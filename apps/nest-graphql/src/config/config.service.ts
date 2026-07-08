import { Injectable } from '@nestjs/common';

@Injectable()
export class ConfigService {
  getConfig() {
    return {
      app: 'flatty-budget',
      version: '1.0.0',
      environment: process.env.NODE_ENV ?? 'development',
    };
  }
}

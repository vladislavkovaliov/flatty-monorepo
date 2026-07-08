import { Module } from '@nestjs/common';
import { ConfigController } from './config/config.controller';
import { ConfigService } from './config/config.service';

@Module({
  controllers: [ConfigController],
  providers: [ConfigService],
})
export class AppModule {}

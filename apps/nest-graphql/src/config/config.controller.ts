import { Controller, Get } from '@nestjs/common';
import { ConfigService } from './config.service';
import { CacheTTL } from '@nestjs/cache-manager';
import { Public } from '../auth/public.decorator';

@Public()
@Controller('config')
export class ConfigController {
  constructor(private readonly configService: ConfigService) {}

  @Get()
  @CacheTTL(30 * 1000)
  getConfig() {
    return this.configService.getConfig();
  }
}

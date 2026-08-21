import { Controller, Get, Header } from '@nestjs/common';
import { registry } from './metrics';
import { Public } from '../auth/public.decorator';

@Controller('metrics')
@Public()
export class MetricsController {
  @Get()
  @Header('Content-Type', registry.contentType)
  async index(): Promise<string> {
    return registry.metrics();
  }
}

import {
  CallHandler,
  ExecutionContext,
  Injectable,
  NestInterceptor,
} from '@nestjs/common';
import { GqlExecutionContext } from '@nestjs/graphql';
import { trace, SpanStatusCode } from '@opentelemetry/api';
import { Observable } from 'rxjs';
import { catchError, finalize } from 'rxjs/operators';

@Injectable()
export class GraphQLTracingInterceptor implements NestInterceptor {
  private readonly tracer = trace.getTracer('nest-graphql');

  intercept(context: ExecutionContext, next: CallHandler): Observable<unknown> {
    const gqlContext = GqlExecutionContext.create(context);

    const info = gqlContext.getInfo();

    const span = this.tracer.startSpan(info.fieldName);

    return next.handle().pipe(
      catchError((error) => {
        span.recordException(error);
        span.setStatus({
          code: SpanStatusCode.ERROR,
          message: error?.message,
        });

        throw error;
      }),
      finalize(() => {
        span.end();
      }),
    );
  }
}
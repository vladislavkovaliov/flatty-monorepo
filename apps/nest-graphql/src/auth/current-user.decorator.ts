import { createParamDecorator, ExecutionContext } from '@nestjs/common'
import { GqlContextType, GqlExecutionContext } from '@nestjs/graphql'
import type { Request } from 'express'

export const CurrentUser = createParamDecorator(
  (_data: unknown, ctx: ExecutionContext): string | undefined => {
    let request: Request

    if (ctx.getType<GqlContextType>() === 'graphql') {
      const gqlCtx = GqlExecutionContext.create(ctx)
      request = gqlCtx.getContext().req
    } else {
      request = ctx.switchToHttp().getRequest<Request>()
    }

    return (request as any).userID
  },
)

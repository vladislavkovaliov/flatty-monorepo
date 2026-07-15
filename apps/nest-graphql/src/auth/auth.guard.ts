import {
  Injectable,
  CanActivate,
  ExecutionContext,
  UnauthorizedException,
} from '@nestjs/common'
import { Reflector } from '@nestjs/core'
import { GqlContextType, GqlExecutionContext } from '@nestjs/graphql'
import type { Request } from 'express'
import { AuthService } from './auth.service'
import { IS_PUBLIC_KEY } from './public.decorator'

@Injectable()
export class AuthGuard implements CanActivate {
  constructor(
    private readonly authService: AuthService,
    private readonly reflector: Reflector,
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const isPublic = this.reflector.getAllAndOverride<boolean>(IS_PUBLIC_KEY, [
      context.getHandler(),
      context.getClass(),
    ])
    if (isPublic) return true

    let request: Request

    if (context.getType<GqlContextType>() === 'graphql') {
      const gqlCtx = GqlExecutionContext.create(context)
      request = gqlCtx.getContext().req
    } else {
      request = context.switchToHttp().getRequest<Request>()
    }

    const token = this.extractToken(request)
    if (!token) {
      throw new UnauthorizedException()
    }

    const userID = await this.authService.validateSession(token)
    if (!userID) {
      throw new UnauthorizedException()
    }

    ;(request as any).userID = userID
    return true
  }

  private extractToken(request: Request): string | null {
    const cookie = request.cookies?.['better-auth.session_token']
    if (cookie) return cookie

    const authHeader = request.headers?.authorization
    if (authHeader?.startsWith('Bearer ')) {
      return authHeader.slice(7)
    }

    return null
  }
}

import { Controller, Get, Post, Body, Req, Res } from '@nestjs/common'
import { ApiTags, ApiOperation, ApiBody } from '@nestjs/swagger'
import type { Request, Response } from 'express'
import { auth } from '../lib/auth'
import { Public } from './public.decorator'
import { SignInEmailDto } from './dto/sign-in-email.dto'
import { SignUpEmailDto } from './dto/sign-up-email.dto'

function headersFromRequest(req: Request): Headers {
  const headers = new Headers()
  for (const [key, value] of Object.entries(req.headers)) {
    if (value) {
      headers.set(key, Array.isArray(value) ? value.join(', ') : value)
    }
  }
  return headers
}

@ApiTags('Auth')
@Controller('auth')
export class AuthController {
  @Public()
  @Post('sign-in/email')
  @ApiOperation({ summary: 'Sign in with email and password' })
  @ApiBody({ type: SignInEmailDto })
  async signInEmail(
    @Body() body: SignInEmailDto,
    @Req() req: Request,
    @Res({ passthrough: true }) res: Response,
  ) {
    const authRes = await auth.api.signInEmail({
      body,
      headers: headersFromRequest(req),
      asResponse: true,
    })

    res.status(authRes.status)
    const cookies = authRes.headers.getSetCookie()
    if (cookies.length > 0) {
      res.setHeader('Set-Cookie', cookies)
    }

    return authRes.json()
  }

  @Public()
  @Post('sign-up/email')
  @ApiOperation({ summary: 'Create a new account with email and password' })
  @ApiBody({ type: SignUpEmailDto })
  async signUpEmail(
    @Body() body: SignUpEmailDto,
    @Req() req: Request,
    @Res({ passthrough: true }) res: Response,
  ) {
    const authRes = await auth.api.signUpEmail({
      body,
      headers: headersFromRequest(req),
      asResponse: true,
    })

    res.status(authRes.status)
    const cookies = authRes.headers.getSetCookie()
    if (cookies.length > 0) {
      res.setHeader('Set-Cookie', cookies)
    }

    return authRes.json()
  }

  @Public()
  @Post('sign-out')
  @ApiOperation({ summary: 'Sign out current session' })
  async signOut(
    @Req() req: Request,
    @Res({ passthrough: true }) res: Response,
  ) {
    const authRes = await auth.api.signOut({
      headers: headersFromRequest(req),
      asResponse: true,
    })

    res.status(authRes.status)
    const cookies = authRes.headers.getSetCookie()
    if (cookies.length > 0) {
      res.setHeader('Set-Cookie', cookies)
    }

    return authRes.json()
  }

  @Public()
  @Get('get-session')
  @ApiOperation({ summary: 'Get current session' })
  async getSession(@Req() req: Request) {
    const session = await auth.api.getSession({
      headers: headersFromRequest(req),
    })
    return session
  }
}

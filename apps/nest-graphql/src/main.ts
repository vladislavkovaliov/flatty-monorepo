import 'dotenv/config'

import cookieParser from 'cookie-parser'
import express from 'express'
import { NestFactory } from '@nestjs/core'
import { ExpressAdapter } from '@nestjs/platform-express'
import { SwaggerModule, DocumentBuilder } from '@nestjs/swagger'
import { AppModule } from './app.module'
import { auth, ensureMigrations } from './lib/auth'

async function bootstrap() {
  await ensureMigrations()

  const expressApp = express()

  // POST /api/auth/set-password — custom handler for magic link users to set a password
  // Must be before the better-auth catch-all to take routing precedence
  expressApp.post('/api/auth/set-password', async (req, res) => {
    try {
      // Read raw body (express.json() hasn't been applied yet)
      const chunks: Buffer[] = []
      for await (const chunk of req) {
        chunks.push(chunk)
      }
      const rawBody = Buffer.concat(chunks).toString()
      const { newPassword } = JSON.parse(rawBody)

      if (!newPassword || typeof newPassword !== 'string' || newPassword.length < 1) {
        res.status(400).json({ error: 'Password is required' })
        return
      }

      // Forward headers from the incoming request (includes session cookie)
      const headers = new Headers()
      for (const [key, value] of Object.entries(req.headers)) {
        if (value) {
          headers.set(key, Array.isArray(value) ? value.join(', ') : value)
        }
      }

      await auth.api.setPassword({
        body: { newPassword },
        headers,
      })

      res.json({ success: true })
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error)

      if (message.includes('PASSWORD_TOO_SHORT') || /password.*(short|weak|min)/i.test(message)) {
        res.status(400).json({ error: 'Password too weak' })
      } else if (message.includes('already has a password') || /already.*credential/i.test(message)) {
        res.status(400).json({ error: 'User already has a password' })
      } else if (message.includes('UNAUTHORIZED') || /not.?authenticated|session/i.test(message)) {
        res.status(401).json({ error: 'Not authenticated' })
      } else {
        res.status(500).json({ error: 'Failed to set password' })
      }
    }
  })

  // Mount better-auth handler for its own routes (magic-link verify, session, etc.)
  // Must be before NestFactory.create so better-auth catches its routes first
  expressApp.use('/api/auth', async (req, res, next) => {
    try {
      const url = new URL(req.originalUrl, `http://${req.headers.host ?? 'localhost:3000'}`)
      const headers = new Headers()
      for (const [key, value] of Object.entries(req.headers)) {
        if (value) {
          headers.set(key, Array.isArray(value) ? value.join(', ') : value)
        }
      }

      // Read raw body from the incoming request (express.json() hasn't been applied yet)
      const chunks: Buffer[] = []
      for await (const chunk of req) {
        chunks.push(chunk)
      }
      const rawBody = Buffer.concat(chunks).toString()

      const response = await auth.handler(
        new Request(url, {
          method: req.method,
          headers,
          body: req.method !== 'GET' && req.method !== 'HEAD' ? rawBody || undefined : undefined,
        }),
      )
      res.status(response.status)
      response.headers.forEach((value, key) => res.setHeader(key, value))
      const body = await response.text()
      res.send(body)
    } catch (error) {
      next(error)
    }
  })

  const app = await NestFactory.create(AppModule, new ExpressAdapter(expressApp))

  expressApp.use(express.json())
  app.use(cookieParser())

  app.setGlobalPrefix('api')

  const config = new DocumentBuilder()
    .setTitle('Flatty Budget API')
    .setVersion('1.0')
    .addServer(`http://localhost:${process.env.PORT ?? 3000}`)
    .build()

  const document = SwaggerModule.createDocument(app, config)
  SwaggerModule.setup('api/docs', app, document)

  await app.listen(process.env.PORT ?? 3000)
}

bootstrap()

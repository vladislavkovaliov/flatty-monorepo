import 'dotenv/config'

import cookieParser from 'cookie-parser'
import express from 'express'
import { NestFactory } from '@nestjs/core'
import { ExpressAdapter } from '@nestjs/platform-express'
import { SwaggerModule, DocumentBuilder } from '@nestjs/swagger'
import { AppModule } from './app.module'
import { ensureMigrations } from './lib/auth'

async function bootstrap() {
  await ensureMigrations()

  const expressApp = express()

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

import { Module } from '@nestjs/common';
import { GraphQLModule } from '@nestjs/graphql';
import { ApolloDriver, ApolloDriverConfig } from '@nestjs/apollo';
import { ConfigController } from './config/config.controller';
import { ConfigService } from './config/config.service';
import { ResidentLocationModule } from './controllers/resident-location/resident-location.module'
import { CategoriesModule } from './controllers/categories/categories.module'
import { ExpensesModule } from './controllers/expenses/expenses.module'
import { ExpenseStatsModule } from './controllers/expense-stats/expense-stats.module'
import { UsersModule } from './controllers/users/users.module'
import { TypeOrmModule } from '@nestjs/typeorm'
import { AuthModule } from './auth/auth.module'
import { InvitationModule } from './invitation/invitation.module'
import { EmailModule } from './email/email.module'
import type { Request } from 'express'
import { join } from 'path';
import { ApolloServerPluginLandingPageLocalDefault, ApolloServerPluginLandingPageProductionDefault } from '@apollo/server/plugin/landingPage/default';


@Module({
  controllers: [ConfigController],
  providers: [ConfigService],
  imports: [
    TypeOrmModule.forRootAsync({
      useFactory: () => {
        return {
          type: 'postgres',
          url: process.env.DATABASE_URL,
          ssl: false,
          entities: [join(__dirname, '/**/*/*.entity{.js,.ts}')]
        }
      }
    }),
    GraphQLModule.forRoot<ApolloDriverConfig>({
      driver: ApolloDriver,
      autoSchemaFile: join(process.cwd(), 'src/schema.gql'),
      playground: false,
      introspection: true,
      csrfPrevention: false,
      context: ({ req }: { req: Request }) => ({
        req,
        userID: (req as any).userID,
      }),
      plugins: [ApolloServerPluginLandingPageLocalDefault()]
    }),
    AuthModule,
    EmailModule,
    InvitationModule,
    ResidentLocationModule,
    CategoriesModule,
    ExpensesModule,
    ExpenseStatsModule,
    UsersModule,
  ]
})
export class AppModule {}

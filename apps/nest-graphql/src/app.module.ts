import { Module } from '@nestjs/common';
import { GraphQLModule } from '@nestjs/graphql';
import { ApolloDriver, ApolloDriverConfig } from '@nestjs/apollo';
import { ConfigController } from './config/config.controller';
import { ConfigService } from './config/config.service';
import { ResidentLocationModule } from './controllers/resident-location/resident-location.module'
import { CategoriesModule } from './controllers/categories/categories.module'
import { TypeOrmModule } from '@nestjs/typeorm'
import { join } from 'path';
import { ApolloServerPluginLandingPageLocalDefault, ApolloServerPluginLandingPageProductionDefault } from '@apollo/server/plugin/landingPage/default';


@Module({
  controllers: [ConfigController],
  providers: [ConfigService],
  imports: [
    TypeOrmModule.forRootAsync({
      useFactory: () => {
        console.log(process.env.DATABASE_URL)
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
      // context: ({ req, res }) => ({ req, res }),
      plugins: [ApolloServerPluginLandingPageLocalDefault()]
    }),
    ResidentLocationModule,
    CategoriesModule,
  ]
})
export class AppModule {}

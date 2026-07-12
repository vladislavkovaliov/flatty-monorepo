import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../nest-graphql/src/schema.gql',
  generates: {
    './src/lib/types/graphql.ts': {
      plugins: ['typescript'],
    },
  },
};

export default config;

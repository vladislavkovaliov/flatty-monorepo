import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: '../../apps/nest-graphql/src/schema.gql',
  generates: {
    './src/types/graphql.ts': {
      plugins: ['typescript'],
      config: {
        enumsAsTypes: true,
      },
    },
  },
};

export default config;

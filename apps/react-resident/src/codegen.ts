import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  schema: 'http://localhost:3000/graphql',
  generates: {
    './src/lib/types/graphql.ts': {
      plugins: ['typescript', 'typescript-operations'],
    },
  },
};

export default config;

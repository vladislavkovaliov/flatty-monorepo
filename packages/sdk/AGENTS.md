# @flatty-budget/sdk

Shared API client and generated types for frontend apps.

- **Stack:** TypeScript, TanStack Query
- **Role:** Shared package — REST (Swagger) + GraphQL types and hooks

## Commands

| Command | Description |
|---------|-------------|
| `npm run api` | Generate REST types from Swagger |
| `npm run graphql:codegen` | Generate GraphQL types |

## Usage

```ts
import { useResidentLocation } from '@flatty-budget/sdk';
import type { ResidentLocationListData } from '@flatty-budget/sdk';
```

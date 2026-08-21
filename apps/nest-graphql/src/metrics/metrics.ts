// Prometheus metrics registry for nest-graphql.
// Plain module singleton (no DI): the GraphQLTracingInterceptor is
// instantiated manually in main.ts, so a shared module is the simplest wiring.
import client from 'prom-client';

export const registry = new client.Registry();

client.collectDefaultMetrics({ register: registry });

export const httpRequestDuration = new client.Histogram({
  name: 'nest_graphql_http_request_duration_seconds',
  help: 'HTTP request duration in seconds, by route and status.',
  labelNames: ['route', 'method', 'status'],
  buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5],
  registers: [registry],
});

export const graphqlResolverDuration = new client.Histogram({
  name: 'nest_graphql_resolver_duration_seconds',
  help: 'GraphQL resolver execution duration in seconds, by field.',
  labelNames: ['field'],
  buckets: [0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1],
  registers: [registry],
});

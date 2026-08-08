import { NodeSDK } from '@opentelemetry/sdk-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-proto';
import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';
import { resourceFromAttributes } from '@opentelemetry/resources';
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions';
import { HttpInstrumentation } from '@opentelemetry/instrumentation-http';
import { ExpressInstrumentation } from '@opentelemetry/instrumentation-express';

const traceExporter = new OTLPTraceExporter({
  url:
    process.env.OTEL_EXPORTER_OTLP_TRACES_ENDPOINT ??
    'http://localhost:4318/v1/traces',
});

const resource = resourceFromAttributes({
  [ATTR_SERVICE_NAME]: 'nest-graphql',
});

const sdk = new NodeSDK({
  traceExporter,
  resource,
  instrumentations: [
    new HttpInstrumentation(),
    // new ExpressInstrumentation(),
    // getNodeAutoInstrumentations({
    //   '@opentelemetry/instrumentation-fs': {
    //     enabled: false,
    //   },
    // }),
  ],
});

sdk.start();

process.on('SIGTERM', () => {
  sdk
    .shutdown()
    .then(() => {
      console.log('OpenTelemetry terminated');
    })
    .catch((error) => {
      console.error('Error terminating OpenTelemetry', error);
    });
});
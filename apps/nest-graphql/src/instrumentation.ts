import { NodeSDK } from '@opentelemetry/sdk-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';
import { resourceFromAttributes } from '@opentelemetry/resources';
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions';
import { HttpInstrumentation } from '@opentelemetry/instrumentation-http';
import { ExpressInstrumentation } from '@opentelemetry/instrumentation-express';

// Prevent unhandled async errors from OTel protobuf serializer from crashing the process
process.on('uncaughtException', (error) => {
  console.error('[OTel] Uncaught exception (suppressed):', error.message);
});
process.on('unhandledRejection', (error) => {
  console.error('[OTel] Unhandled rejection (suppressed):', error);
});

const traceExporter = new OTLPTraceExporter();

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
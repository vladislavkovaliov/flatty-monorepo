// Shared k6 options and report summary for all scenarios.
// NOTE: instrumentHTTP (W3C trace propagation) was removed because it breaks
// k6's built-in cookie jar, causing 100% auth failures after the first request.
// Trace propagation is handled server-side via nginx → Jaeger OTLP.
export const baseOptions = {
  // All scenarios hit nginx (production-accurate path).
  // Override with K6_BASE_URL when running against dev servers.
  scenarios: {},
};

export function withThresholds(thresholds) {
  return { thresholds };
}

export function handleSummary(data) {
  const lines = Object.entries(data.metrics)
    .filter(([name]) => /^(http_req|iterations|vus)/.test(name))
    .map(([name, m]) => {
      const v = m.values;
      const parts = [];
      if (v.avg !== undefined) parts.push(`avg=${v.avg.toFixed(2)}`);
      if (v['p(95)'] !== undefined) parts.push(`p95=${v['p(95)'].toFixed(2)}`);
      if (v['p(99)'] !== undefined) parts.push(`p99=${v['p(99)'].toFixed(2)}`);
      if (v.rate !== undefined) parts.push(`rate=${v.rate.toFixed(2)}`);
      if (v.count !== undefined) parts.push(`count=${v.count}`);
      return `  ${name}: ${parts.join(' ')}`;
    })
    .join('\n');

  return {
    stdout: `\n=== k6 summary ===\n${lines}\n`,
    'reports/summary.json': JSON.stringify(data, null, 2),
  };
}
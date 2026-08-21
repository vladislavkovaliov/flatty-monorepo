// Stress test: ramp to 100 VUs to find the breaking point. On-demand.
import http from 'k6/http';
import { check } from 'k6';
import { baseOptions, withThresholds, handleSummary } from './common/options.js';
import { stressThresholds } from './common/thresholds.js';
import { ensureLoggedIn } from './common/auth.js';

export const options = {
  ...baseOptions,
  ...withThresholds(stressThresholds),
  scenarios: {
    stress: {
      executor: 'ramping-vus',
      stages: [
        { duration: '1m', target: 20 },
        { duration: '2m', target: 100 },
        { duration: '2m', target: 100 },
        { duration: '1m', target: 0 },
      ],
    },
  },
};

export { handleSummary };

const BASE_URL = __ENV.K6_BASE_URL ?? 'http://localhost';
const RESIDENT_LOCATION_ID = __ENV.K6_RESIDENT_LOCATION_ID ?? '9';

export default function () {
  const cookie = ensureLoggedIn();
  const headers = { cookie };

  const rest = http.get(`${BASE_URL}/api/expenses/count?residentLocationId=${RESIDENT_LOCATION_ID}`, { headers });
  check(rest, { 'expenses/count 200': (r) => r.status === 200 });

  const gql = http.post(
    `${BASE_URL}/graphql`,
    JSON.stringify({
      query: `query { expenseList(residentLocationId: ${RESIDENT_LOCATION_ID}, limit: 10) { total data { id amount } } }`,
    }),
    { headers: { ...headers, 'Content-Type': 'application/json' } },
  );
  check(gql, { 'expenseList 200': (r) => r.status === 200 });
}

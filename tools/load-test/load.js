// Load test: ramp to 50 VUs, hold, ramp down. Main nightly scenario.
import http from 'k6/http';
import { check } from 'k6';
import { baseOptions, withThresholds, handleSummary } from './common/options.js';
import { loadThresholds } from './common/thresholds.js';
import { ensureLoggedIn } from './common/auth.js';

export const options = {
  ...baseOptions,
  ...withThresholds(loadThresholds),
  scenarios: {
    load: {
      executor: 'ramping-vus',
      stages: [
        { duration: '1m', target: 10 },
        { duration: '3m', target: 50 },
        { duration: '2m', target: 50 },
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

  // REST: expenses count + categories (go-api via nginx /api/)
  const rest = http.get(`${BASE_URL}/api/expenses/count?residentLocationId=${RESIDENT_LOCATION_ID}`, { headers });
  check(rest, { 'expenses/count 200': (r) => r.status === 200 });

  const categories = http.get(`${BASE_URL}/api/categories`, { headers });
  check(categories, { 'categories 200': (r) => r.status === 200 });

  // GraphQL: expenseMonthlyTotals (nest-graphql via nginx /graphql)
  const gql = http.post(
    `${BASE_URL}/graphql`,
    JSON.stringify({
      query: `query { expenseMonthlyTotals(residentLocationId: ${RESIDENT_LOCATION_ID}) { data { month year totalSpent } } }`,
    }),
    { headers: { ...headers, 'Content-Type': 'application/json' } },
  );
  check(gql, { 'expenseMonthlyTotals 200': (r) => r.status === 200 });
}

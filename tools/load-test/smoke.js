// Smoke test: 5 VUs for 1 minute. Fast CI sanity check.
// Verifies health, one REST call, one GraphQL call, and the auth flow.
import http from 'k6/http';
import { check } from 'k6';
import { baseOptions, withThresholds, handleSummary } from './common/options.js';
import { smokeThresholds } from './common/thresholds.js';
import { ensureLoggedIn } from './common/auth.js';

export const options = {
  ...baseOptions,
  ...withThresholds(smokeThresholds),
  scenarios: {
    smoke: {
      executor: 'constant-vus',
      vus: 5,
      duration: '1m',
    },
  },
};

export { handleSummary };

const BASE_URL = __ENV.K6_BASE_URL ?? 'http://localhost';
const RESIDENT_LOCATION_ID = __ENV.K6_RESIDENT_LOCATION_ID ?? '1';

export default function () {
  const cookie = ensureLoggedIn();
  const headers = { cookie };

  // REST: expenses count (go-api via nginx /api/)
  const rest = http.get(`${BASE_URL}/api/expenses/count?residentLocationId=${RESIDENT_LOCATION_ID}`, { headers });
  check(rest, {
    'expenses/count 200': (r) => r.status === 200,
  });

  // GraphQL: expenseList (nest-graphql via nginx /graphql)
  const gql = http.post(
    `${BASE_URL}/graphql`,
    JSON.stringify({
      query: `query { expenseList(residentLocationId: ${RESIDENT_LOCATION_ID}, limit: 10) { total data { id amount } } }`,
    }),
    { headers: { ...headers, 'Content-Type': 'application/json' } },
  );
  check(gql, {
    'expenseList 200': (r) => r.status === 200,
    'expenseList no errors': (r) => {
      const body = JSON.parse(r.body);
      return !body.errors;
    },
  });
}

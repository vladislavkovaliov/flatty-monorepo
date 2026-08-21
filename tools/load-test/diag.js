// Diagnostic: single VU, 1 iteration — prints exact status and body for each request.
import http from 'k6/http';
import { check } from 'k6';
import { ensureLoggedIn } from './common/auth.js';
import { baseOptions, handleSummary } from './common/options.js';

export const options = {
  ...baseOptions,
  scenarios: {
    diag: {
      executor: 'constant-vus',
      vus: 1,
      duration: '10s',
    },
  },
};

export { handleSummary };

const BASE_URL = __ENV.K6_BASE_URL ?? 'http://localhost';
const RESIDENT_LOCATION_ID = __ENV.K6_RESIDENT_LOCATION_ID ?? '1';

export default function () {
  const cookie = ensureLoggedIn();
  const headers = { cookie };

  // REST
  const rest = http.get(`${BASE_URL}/api/expenses/count?residentLocationId=${RESIDENT_LOCATION_ID}`, { headers });
  console.log(`REST status=${rest.status} body=${rest.body}`);

  // GraphQL
  const gql = http.post(
    `${BASE_URL}/graphql`,
    JSON.stringify({
      query: `query { expenseList(residentLocationId: ${RESIDENT_LOCATION_ID}, limit: 10) { total data { id amount } } }`,
    }),
    { headers: { ...headers, 'Content-Type': 'application/json' } },
  );
  console.log(`GQL status=${gql.status} body=${gql.body}`);

  check(rest, { 'rest 200': (r) => r.status === 200 });
  check(gql, { 'gql 200': (r) => r.status === 200 });
}

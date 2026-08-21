// better-auth login helper.
// POST /api/auth/sign-in/email returns Set-Cookie: better-auth.session_token=...
// We extract the cookie value and return it as a string so callers can pass it
// via the `cookie` header explicitly — k6's automatic cookie jar is unreliable
// in Docker networking.
import http from 'k6/http';
import exec from 'k6/execution';

const tokens = new Map();

/**
 * Log in once per VU and return the cookie header string.
 * Subsequent calls return the cached value.
 *
 * Usage:
 *   const cookie = ensureLoggedIn();
 *   http.get(url, { headers: { cookie } });
 */
export function ensureLoggedIn() {
  const id = exec.vu.idInTest;
  if (tokens.has(id)) return tokens.get(id);

  const baseURL = __ENV.K6_BASE_URL ?? 'http://localhost';
  const email = __ENV.K6_EMAIL;
  const password = __ENV.K6_PASSWORD;

  if (!email || !password) {
    throw new Error('K6_EMAIL and K6_PASSWORD env vars are required');
  }

  const res = http.post(
    `${baseURL}/api/auth/sign-in/email`,
    JSON.stringify({ email, password }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  if (res.status !== 200) {
    throw new Error(
      `login failed (${res.status}) — is the load-test user seeded? ` +
        `Seed via: curl -X POST ${baseURL}/api/auth/sign-up/email ` +
        `-H 'Content-Type: application/json' ` +
        `-d '{"email":"${email}","password":"${password}"}'`,
    );
  }

  // Extract cookie from Set-Cookie header
  const setCookie = res.headers['Set-Cookie'] || res.headers['set-cookie'] || '';
  // Parse: "better-auth.session_token=abc123.signature; Path=/; ..."
  const match = setCookie.match(/^better-auth\.session_token=([^;]+)/);
  if (!match) {
    throw new Error(
      `login succeeded (${res.status}) but no session cookie in response. ` +
        `Set-Cookie: ${setCookie.substring(0, 200)}`,
    );
  }

  const cookie = `better-auth.session_token=${match[1]}`;
  tokens.set(id, cookie);
  return cookie;
}

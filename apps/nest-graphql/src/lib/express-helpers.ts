import type { IncomingMessage } from 'http'

/** Convert Express IncomingMessage headers to a Web Headers object. */
export function toWebHeaders(req: IncomingMessage): Headers {
  const headers = new Headers()
  for (const [key, value] of Object.entries(req.headers)) {
    if (value) {
      headers.set(key, Array.isArray(value) ? value.join(', ') : value)
    }
  }
  return headers
}

/** Read the full raw body from a streaming IncomingMessage. */
export async function readRawBody(req: IncomingMessage): Promise<string> {
  const chunks: Buffer[] = []
  for await (const chunk of req) {
    chunks.push(chunk)
  }
  return Buffer.concat(chunks).toString()
}

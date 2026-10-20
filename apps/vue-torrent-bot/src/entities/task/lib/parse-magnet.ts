export interface MagnetInfo {
  raw: string
  infoHash: string
  displayName: string | null
  trackers: string[]
  sizeBytes: number | null
}

const MAGNET_PREFIX = 'magnet:?'
const BTIH_URN_PATTERN = /^urn:btih:([0-9a-z]{32,40})$/i

// Client-side mirror of the bot's magnet parser: `xt` with a btih hash is required,
// `dn`/`tr`/`xl` are optional. Returns null for anything that is not a usable magnet.
export function parseMagnet(raw: string): MagnetInfo | null {
  const trimmed = raw.trim()
  if (!trimmed.toLowerCase().startsWith(MAGNET_PREFIX)) return null

  const params = new URLSearchParams(trimmed.slice(MAGNET_PREFIX.length))
  const infoHash = params
    .getAll('xt')
    .find((value) => BTIH_URN_PATTERN.test(value))
    ?.split(':')[2]
  if (!infoHash) return null

  const size = Number(params.get('xl'))

  return {
    raw: trimmed,
    infoHash,
    displayName: params.get('dn')?.trim() || null,
    trackers: params.getAll('tr').filter(Boolean),
    sizeBytes: Number.isFinite(size) && size > 0 ? size : null,
  }
}

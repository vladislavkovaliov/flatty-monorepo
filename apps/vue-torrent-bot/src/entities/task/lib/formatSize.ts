const UNITS = ['B', 'KiB', 'MiB', 'GiB', 'TiB'] as const

export function formatSize(bytes: number | null): string {
  if (bytes === null) {
    return 'не определён'
  }

  let value = bytes
  let unitIndex = 0

  while (value >= 1024 && unitIndex < UNITS.length - 1) {
    value /= 1024
    unitIndex += 1
  }

  const unit = UNITS[unitIndex] ?? UNITS[0]
  return `${value.toFixed(unitIndex === 0 ? 0 : 1)} ${unit}`
}

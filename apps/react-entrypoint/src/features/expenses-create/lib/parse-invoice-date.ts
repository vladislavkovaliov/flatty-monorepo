export function parseInvoiceDate(
  value: string | undefined,
): { month: number; year: number } | null {
  if (!value) return null;

  const text = value.trim();
  const patterns: Array<
    [RegExp, (m: RegExpMatchArray) => { month: number; year: number }]
  > = [
    [
      /^(\d{4})-(\d{1,2})-(\d{1,2})$/,
      (m) => ({ year: Number(m[1]), month: Number(m[2]) }),
    ],
    [
      /^(\d{1,2})\.(\d{1,2})\.(\d{4})$/,
      (m) => ({ year: Number(m[3]), month: Number(m[2]) }),
    ],
    [
      /^(\d{1,2})\/(\d{1,2})\/(\d{4})$/,
      (m) => ({ year: Number(m[3]), month: Number(m[1]) }),
    ],
    [
      /^(\d{4})-(\d{1,2})$/,
      (m) => ({ year: Number(m[1]), month: Number(m[2]) }),
    ],
  ];

  for (const [pattern, map] of patterns) {
    const match = text.match(pattern);
    if (match) {
      const { month, year } = map(match);
      if (month >= 1 && month <= 12 && year >= 2000) {
        return { month, year };
      }
    }
  }

  return null;
}

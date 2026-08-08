export interface CategoryOption {
  value: string;
  label: string;
  name: string;
}

export function normalizeCategory(value: string): string {
  return value.toLowerCase().replace(/\s+/g, " ").trim();
}

export function categoryStrings(option: CategoryOption): string[] {
  return [option.label, option.name];
}

export function findCategoryId(
  category: string,
  options: CategoryOption[],
): number | undefined {
  const normalized = normalizeCategory(category);
  if (!normalized) return undefined;

  const exact = options.find((option) =>
    categoryStrings(option).some(
      (text) => normalizeCategory(text) === normalized,
    ),
  );
  if (exact) return Number(exact.value);

  // Fuzzy fallback: partial containment, prefer the most specific (longest)
  // label so "Коммунальные" resolves to "Коммунальные платежи" over "Платежи".
  let best: string | undefined;
  let bestLength = 0;
  for (const option of options) {
    const matchLength = categoryStrings(option).reduce((longest, text) => {
      const label = normalizeCategory(text);
      return label.length >= 3 &&
        normalized.length >= 3 &&
        (label.includes(normalized) || normalized.includes(label))
        ? Math.max(longest, label.length)
        : longest;
    }, 0);
    if (matchLength > bestLength) {
      best = option.value;
      bestLength = matchLength;
    }
  }
  return best === undefined ? undefined : Number(best);
}

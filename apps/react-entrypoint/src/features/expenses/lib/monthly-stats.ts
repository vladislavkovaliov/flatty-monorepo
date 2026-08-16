export interface MonthlyStatRow {
  month: number;
  year: number;
  totalSpent: number | null;
  averageAmount: number | null;
  expenseCount: number | null;
}

export interface MonthlyTotalInput {
  month: number;
  year: number;
  totalSpent: number;
}

export interface MonthlyAverageInput {
  month: number;
  year: number;
  averageAmount: number;
  expenseCount: number;
}

export interface YearGroup {
  year: number;
  rows: MonthlyStatRow[];
  totalSpent: number;
  expenseCount: number;
}

function yearMonthKey(year: number, month: number): string {
  return `${year}-${String(month).padStart(2, "0")}`;
}

export function mergeMonthlyStats(
  totals: MonthlyTotalInput[],
  averages: MonthlyAverageInput[],
): MonthlyStatRow[] {
  const rowsByKey = new Map<string, MonthlyStatRow>();

  for (const total of totals) {
    rowsByKey.set(yearMonthKey(total.year, total.month), {
      month: total.month,
      year: total.year,
      totalSpent: total.totalSpent,
      averageAmount: null,
      expenseCount: null,
    });
  }

  for (const average of averages) {
    const key = yearMonthKey(average.year, average.month);
    const existing = rowsByKey.get(key);
    if (existing) {
      existing.averageAmount = average.averageAmount;
      existing.expenseCount = average.expenseCount;
    } else {
      rowsByKey.set(key, {
        month: average.month,
        year: average.year,
        totalSpent: null,
        averageAmount: average.averageAmount,
        expenseCount: average.expenseCount,
      });
    }
  }

  return [...rowsByKey.values()].sort(
    (a, b) => b.year - a.year || b.month - a.month,
  );
}

export function groupByYear(rows: MonthlyStatRow[]): YearGroup[] {
  const groupsByYear = new Map<number, YearGroup>();

  for (const row of rows) {
    let group = groupsByYear.get(row.year);
    if (!group) {
      group = { year: row.year, rows: [], totalSpent: 0, expenseCount: 0 };
      groupsByYear.set(row.year, group);
    }
    group.rows.push(row);
    group.totalSpent += row.totalSpent ?? 0;
    group.expenseCount += row.expenseCount ?? 0;
  }

  return [...groupsByYear.values()].sort((a, b) => b.year - a.year);
}

export interface YearAndMonthInput {
  year?: number;
  month: number;
  expenses: number;
}

export interface MonthSummary {
  month: number;
  expenseCount: number;
}

export interface YearMonthGroup {
  year: number;
  months: MonthSummary[];
}

export function groupYearsAndMonths(items: YearAndMonthInput[]): YearMonthGroup[] {
  const groupsByYear = new Map<number, YearMonthGroup>();

  for (const item of items) {
    if (item.year == null) continue;
    let group = groupsByYear.get(item.year);
    if (!group) {
      group = { year: item.year, months: [] };
      groupsByYear.set(item.year, group);
    }
    group.months.push({ month: item.month, expenseCount: item.expenses });
  }

  return [...groupsByYear.values()]
    .sort((a, b) => b.year - a.year)
    .map((group) => ({
      ...group,
      months: [...group.months].sort((a, b) => a.month - b.month),
    }));
}

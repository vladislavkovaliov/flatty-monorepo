import { describe, expect, it } from "vitest";

import { groupByYear, groupYearsAndMonths, mergeMonthlyStats } from "../monthly-stats";

describe("mergeMonthlyStats", () => {
  it("merges totals and averages by year-month key", () => {
    const totals = [
      { month: 3, year: 2025, totalSpent: 100 },
      { month: 1, year: 2025, totalSpent: 50 },
    ];
    const averages = [
      { month: 3, year: 2025, averageAmount: 25, expenseCount: 4 },
      { month: 1, year: 2025, averageAmount: 50, expenseCount: 1 },
    ];

    expect(mergeMonthlyStats(totals, averages)).toEqual([
      {
        month: 3,
        year: 2025,
        totalSpent: 100,
        averageAmount: 25,
        expenseCount: 4,
      },
      {
        month: 1,
        year: 2025,
        totalSpent: 50,
        averageAmount: 50,
        expenseCount: 1,
      },
    ]);
  });

  it("keeps nulls when a month exists only in totals", () => {
    const totals = [{ month: 6, year: 2024, totalSpent: 200 }];

    expect(mergeMonthlyStats(totals, [])).toEqual([
      {
        month: 6,
        year: 2024,
        totalSpent: 200,
        averageAmount: null,
        expenseCount: null,
      },
    ]);
  });

  it("keeps nulls when a month exists only in averages", () => {
    const averages = [
      { month: 11, year: 2023, averageAmount: 10, expenseCount: 2 },
    ];

    expect(mergeMonthlyStats([], averages)).toEqual([
      {
        month: 11,
        year: 2023,
        totalSpent: null,
        averageAmount: 10,
        expenseCount: 2,
      },
    ]);
  });

  it("sorts rows by year desc, then month desc", () => {
    const totals = [
      { month: 1, year: 2024, totalSpent: 1 },
      { month: 12, year: 2025, totalSpent: 2 },
      { month: 6, year: 2024, totalSpent: 3 },
    ];

    const result = mergeMonthlyStats(totals, []);

    expect(result.map((row) => `${row.year}-${row.month}`)).toEqual([
      "2025-12",
      "2024-6",
      "2024-1",
    ]);
  });

  it("returns an empty array for empty inputs", () => {
    expect(mergeMonthlyStats([], [])).toEqual([]);
  });
});

describe("groupByYear", () => {
  it("groups rows by year and computes subtotals", () => {
    const rows = [
      {
        month: 1,
        year: 2025,
        totalSpent: 100,
        averageAmount: 100,
        expenseCount: 1,
      },
      {
        month: 2,
        year: 2025,
        totalSpent: 50,
        averageAmount: 25,
        expenseCount: 2,
      },
      {
        month: 3,
        year: 2024,
        totalSpent: 30,
        averageAmount: 30,
        expenseCount: 1,
      },
    ];

    expect(groupByYear(rows)).toEqual([
      {
        year: 2025,
        rows: [
          {
            month: 1,
            year: 2025,
            totalSpent: 100,
            averageAmount: 100,
            expenseCount: 1,
          },
          {
            month: 2,
            year: 2025,
            totalSpent: 50,
            averageAmount: 25,
            expenseCount: 2,
          },
        ],
        totalSpent: 150,
        expenseCount: 3,
      },
      {
        year: 2024,
        rows: [
          {
            month: 3,
            year: 2024,
            totalSpent: 30,
            averageAmount: 30,
            expenseCount: 1,
          },
        ],
        totalSpent: 30,
        expenseCount: 1,
      },
    ]);
  });

  it("sorts year groups descending", () => {
    const rows = [
      {
        month: 1,
        year: 2023,
        totalSpent: 1,
        averageAmount: 1,
        expenseCount: 1,
      },
      {
        month: 1,
        year: 2025,
        totalSpent: 2,
        averageAmount: 2,
        expenseCount: 1,
      },
      {
        month: 1,
        year: 2024,
        totalSpent: 3,
        averageAmount: 3,
        expenseCount: 1,
      },
    ];

    expect(groupByYear(rows).map((group) => group.year)).toEqual([
      2025, 2024, 2023,
    ]);
  });

  it("treats null totals and counts as zero in subtotals", () => {
    const rows = [
      {
        month: 1,
        year: 2025,
        totalSpent: null,
        averageAmount: 10,
        expenseCount: null,
      },
      {
        month: 2,
        year: 2025,
        totalSpent: 40,
        averageAmount: 40,
        expenseCount: 2,
      },
    ];

    expect(groupByYear(rows)).toEqual([
      {
        year: 2025,
        rows: [
          {
            month: 1,
            year: 2025,
            totalSpent: null,
            averageAmount: 10,
            expenseCount: null,
          },
          {
            month: 2,
            year: 2025,
            totalSpent: 40,
            averageAmount: 40,
            expenseCount: 2,
          },
        ],
        totalSpent: 40,
        expenseCount: 2,
      },
    ]);
  });

  it("returns an empty array for empty input", () => {
    expect(groupByYear([])).toEqual([]);
  });
});

describe("groupYearsAndMonths", () => {
  it("groups flat items by year with month summaries", () => {
    const items = [
      { year: 2025, month: 3, expenses: 4 },
      { year: 2025, month: 1, expenses: 1 },
      { year: 2024, month: 11, expenses: 2 },
    ];

    expect(groupYearsAndMonths(items)).toEqual([
      {
        year: 2025,
        months: [
          { month: 1, expenseCount: 1 },
          { month: 3, expenseCount: 4 },
        ],
      },
      {
        year: 2024,
        months: [{ month: 11, expenseCount: 2 }],
      },
    ]);
  });

  it("sorts years descending and months ascending", () => {
    const items = [
      { year: 2023, month: 6, expenses: 1 },
      { year: 2025, month: 2, expenses: 2 },
      { year: 2024, month: 12, expenses: 3 },
      { year: 2025, month: 1, expenses: 4 },
    ];

    const result = groupYearsAndMonths(items);

    expect(result.map((group) => group.year)).toEqual([2025, 2024, 2023]);
    expect(result[0].months.map((m) => m.month)).toEqual([1, 2]);
  });

  it("skips items without a year", () => {
    const items = [
      { month: 5, expenses: 3 },
      { year: 2025, month: 1, expenses: 1 },
    ];

    expect(groupYearsAndMonths(items)).toEqual([
      { year: 2025, months: [{ month: 1, expenseCount: 1 }] },
    ]);
  });

  it("returns an empty array for empty input", () => {
    expect(groupYearsAndMonths([])).toEqual([]);
  });
});

import { useState, useMemo } from "react";
import { Box, Title } from "@mantine/core";
import { useExpenseMonthlyTotalsGraphql, useExpenseMonthlyAveragesGraphql } from "@flatty-budget/sdk";
import { SpendingFilters } from "./spending-filters";
import { SpendingChart } from "./spending-chart";
import { SpendingTable } from "./spending-table";

export interface SpendingRow {
  month: number;
  year: number;
  totalSpent: number | null;
  averageAmount: number | null;
  expenseCount: number | null;
}

export function SpendingPage() {
  const { data: totalsData } = useExpenseMonthlyTotalsGraphql();
  const { data: averagesData } = useExpenseMonthlyAveragesGraphql();

  const [fromDate, setFromDate] = useState<Date | null>(null);
  const [toDate, setToDate] = useState<Date | null>(null);
  const [appliedFrom, setAppliedFrom] = useState<Date | null>(null);
  const [appliedTo, setAppliedTo] = useState<Date | null>(null);

  const handleApply = () => {
    setAppliedFrom(fromDate);
    setAppliedTo(toDate);
  };

  const handleReset = () => {
    setFromDate(null);
    setToDate(null);
    setAppliedFrom(null);
    setAppliedTo(null);
  };

  const totals = useMemo(() => totalsData?.expenseMonthlyTotals?.data ?? [], [totalsData?.expenseMonthlyTotals?.data]);
  const averages = useMemo(() => averagesData?.expenseMonthlyAverages?.data ?? [], [averagesData?.expenseMonthlyAverages?.data]);

  const merged: SpendingRow[] = useMemo(() => {
    const map = new Map<string, SpendingRow>();

    for (const t of totals) {
      const key = `${t.year}-${String(t.month).padStart(2, "0")}`;
      map.set(key, {
        month: t.month,
        year: t.year,
        totalSpent: t.totalSpent,
        averageAmount: null,
        expenseCount: null,
      });
    }

    for (const a of averages) {
      const key = `${a.year}-${String(a.month).padStart(2, "0")}`;
      const existing = map.get(key);
      if (existing) {
        existing.averageAmount = a.averageAmount;
        existing.expenseCount = a.expenseCount;
      } else {
        map.set(key, {
          month: a.month,
          year: a.year,
          totalSpent: null,
          averageAmount: a.averageAmount,
          expenseCount: a.expenseCount,
        });
      }
    }

    return Array.from(map.values())
      .sort((a, b) => b.year - a.year || b.month - a.month);
  }, [totals, averages]);

  const minMax = useMemo(() => {
    if (merged.length === 0) return { minDate: undefined, maxDate: undefined };

    let minYear = Infinity;
    let minMonth = 1;
    let maxYear = -Infinity;
    let maxMonth = 1;

    for (const row of merged) {
      if (row.year < minYear || (row.year === minYear && row.month < minMonth)) {
        minYear = row.year;
        minMonth = row.month;
      }
      if (row.year > maxYear || (row.year === maxYear && row.month > maxMonth)) {
        maxYear = row.year;
        maxMonth = row.month;
      }
    }

    return {
      minDate: new Date(minYear, minMonth - 1, 1),
      maxDate: new Date(maxYear, maxMonth - 1, 1),
    };
  }, [merged]);

  const filtered = useMemo(() => {
    let result = merged;

    if (appliedFrom) {
      const fromYear = appliedFrom.getFullYear();
      const fromMonth = appliedFrom.getMonth() + 1;
      result = result.filter(
        (r) => r.year > fromYear || (r.year === fromYear && r.month >= fromMonth),
      );
    }

    if (appliedTo) {
      const toYear = appliedTo.getFullYear();
      const toMonth = appliedTo.getMonth() + 1;
      result = result.filter(
        (r) => r.year < toYear || (r.year === toYear && r.month <= toMonth),
      );
    }

    return result;
  }, [merged, appliedFrom, appliedTo]);

  return (
    <>
      <Box py="md">
        <Title order={2} px="md">Spending Statistics</Title>
      </Box>

      <SpendingFilters
        fromDate={fromDate}
        toDate={toDate}
        onFromDateChange={setFromDate}
        onToDateChange={setToDate}
        onApply={handleApply}
        onReset={handleReset}
        minDate={minMax.minDate}
        maxDate={minMax.maxDate}
      />

      <SpendingChart data={filtered} />
      <SpendingTable data={filtered} />
    </>
  );
}

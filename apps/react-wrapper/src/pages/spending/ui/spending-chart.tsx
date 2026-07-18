import { CompositeChart } from "@mantine/charts";
import { Box, Container } from "@mantine/core";
import type { SpendingRow } from "./spending-page";

const MONTH_LABELS = [
  "Jan", "Feb", "Mar", "Apr", "May", "Jun",
  "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
];

interface ChartDataPoint {
  month: string;
  totalSpent: number;
  averageAmount: number;
}

interface SpendingChartProps {
  data: SpendingRow[];
}

export function SpendingChart({ data }: SpendingChartProps) {
  const chartData: ChartDataPoint[] = data
    .map((row) => ({
      month: `${MONTH_LABELS[row.month - 1]} '${String(row.year).slice(2)}`,
      totalSpent: row.totalSpent ?? 0,
      averageAmount: row.averageAmount ?? 0,
    }))
    .reverse(); // data is sorted desc, chart reads left-to-right, so reverse

  if (chartData.length === 0) {
    return null;
  }

  return (
    <Box py="md">
      <Container fluid>
        <CompositeChart
          h={300}
          data={chartData}
          dataKey="month"
          series={[
            { name: "totalSpent", color: "cyan.6", type: "bar", label: "Total Spent" },
            { name: "averageAmount", color: "orange.6", type: "line", label: "Average Amount" },
          ]}
          tickLine="xy"
          gridAxis="xy"
          withLegend
          legendProps={{ verticalAlign: "bottom" }}
        />
      </Container>
    </Box>
  );
}

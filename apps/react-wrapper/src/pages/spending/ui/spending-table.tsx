import { useState, useMemo } from "react";
import { Table, Box, Container, Pagination, Text } from "@mantine/core";
import { useSearchParams } from "react-router-dom";
import type { SpendingRow } from "./spending-page";

const PAGE_SIZE = 10;

type SortField = "month" | "year" | "totalSpent" | "averageAmount" | "expenseCount";
type SortDir = "asc" | "desc";

interface SpendingTableProps {
  data: SpendingRow[];
}

interface SortConfig {
  field: SortField;
  dir: SortDir;
}

function sortRows(rows: SpendingRow[], { field, dir }: SortConfig): SpendingRow[] {
  return [...rows].sort((a, b) => {
    const aVal = a[field] ?? 0;
    const bVal = b[field] ?? 0;
    const cmp = aVal < bVal ? -1 : aVal > bVal ? 1 : 0;
    return dir === "asc" ? cmp : -cmp;
  });
}

function SortIcon({ field, config }: { field: SortField; config: SortConfig | null }) {
  if (config?.field !== field) {
    return <Text component="span" c="dimmed" ml={4}>↕</Text>;
  }
  return <Text component="span" ml={4}>{config.dir === "asc" ? "↑" : "↓"}</Text>;
}

export function SpendingTable({ data }: SpendingTableProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get("page") || "1");

  const [sort, setSort] = useState<SortConfig | null>(null);

  const handleSort = (field: SortField) => {
    setSort((prev) => {
      if (prev?.field === field) {
        return { field, dir: prev.dir === "asc" ? "desc" : "asc" };
      }
      return { field, dir: "asc" };
    });
  };

  const sorted = useMemo(
    () => (sort ? sortRows(data, sort) : data),
    [data, sort],
  );

  const totalPages = Math.max(1, Math.ceil(sorted.length / PAGE_SIZE));
  const safePage = Math.min(page, totalPages);
  const start = (safePage - 1) * PAGE_SIZE;
  const pageRows = sorted.slice(start, start + PAGE_SIZE);

  const handlePageChange = (newPage: number) => {
    setSearchParams({ page: String(newPage) });
  };

  const rows = pageRows.map((row) => (
    <Table.Tr key={`${row.year}-${row.month}`}>
      <Table.Td>{row.month}</Table.Td>
      <Table.Td>{row.year}</Table.Td>
      <Table.Td>{row.totalSpent?.toFixed(2) ?? "\u2014"}</Table.Td>
      <Table.Td>{row.averageAmount?.toFixed(2) ?? "\u2014"}</Table.Td>
      <Table.Td>{row.expenseCount ?? "\u2014"}</Table.Td>
    </Table.Tr>
  ));

  return (
    <Box py="md">
      <Container fluid>
        <Table>
          <Table.Thead>
            <Table.Tr>
              <Table.Th style={{ cursor: "pointer" }} onClick={() => handleSort("month")}>
                Month<SortIcon field="month" config={sort} />
              </Table.Th>
              <Table.Th style={{ cursor: "pointer" }} onClick={() => handleSort("year")}>
                Year<SortIcon field="year" config={sort} />
              </Table.Th>
              <Table.Th style={{ cursor: "pointer" }} onClick={() => handleSort("totalSpent")}>
                Total Spent<SortIcon field="totalSpent" config={sort} />
              </Table.Th>
              <Table.Th style={{ cursor: "pointer" }} onClick={() => handleSort("averageAmount")}>
                Average Amount<SortIcon field="averageAmount" config={sort} />
              </Table.Th>
              <Table.Th style={{ cursor: "pointer" }} onClick={() => handleSort("expenseCount")}>
                Expenses<SortIcon field="expenseCount" config={sort} />
              </Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>{rows}</Table.Tbody>
        </Table>
      </Container>

      {totalPages > 1 && (
        <Box py="xl">
          <Container fluid>
            <Pagination total={totalPages} value={safePage} onChange={handlePageChange} />
          </Container>
        </Box>
      )}
    </Box>
  );
}

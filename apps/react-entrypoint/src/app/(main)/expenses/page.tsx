"use client";

import { Box, Button, Group, SegmentedControl } from "@mantine/core";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import { ExpensesGroupedView } from "@/features/expenses/ui/expenses-grouped-view";
import { ExpensesTable } from "@/features/expenses/ui/expenses-table";

type ExpensesView = "table" | "grouped";

const VIEW_OPTIONS: { label: string; value: string }[] = [
  { label: "Table", value: "table" },
  { label: "Grouped", value: "grouped" },
];

export default function ExpensesPage() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams() ?? new URLSearchParams();
  const view: ExpensesView =
    searchParams.get("view") === "grouped" ? "grouped" : "table";

  const handleViewChange = (value: string) => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("view", value);
    router.replace(`${pathname}?${params.toString()}`);
  };

  return (
    <>
      <Box py="md" />
      <Group justify="space-between" px="md">
        <SegmentedControl
          data={VIEW_OPTIONS}
          value={view}
          onChange={handleViewChange}
        />
        <Group>
          <Button variant="outline" onClick={() => router.push("/pdfs/upload")}>
            Upload PDF
          </Button>
          <Button onClick={() => router.push("/expenses/create")}>
            Create Expense
          </Button>
        </Group>
      </Group>
      {view === "grouped" ? <ExpensesGroupedView /> : <ExpensesTable />}
    </>
  );
}

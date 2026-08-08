"use client";

import { Box, Button, Group } from "@mantine/core";
import { useRouter } from "next/navigation";
import { ExpensesTable } from "@/features/expenses/ui/expenses-table";

export default function ExpensesPage() {
  const router = useRouter();

  return (
    <>
      <Box py="md" />
      <Group justify="flex-end" px="md">
        <Button variant="outline" onClick={() => router.push("/pdfs/upload")}>
          Upload PDF
        </Button>
        <Button onClick={() => router.push("/expenses/create")}>
          Create Expense
        </Button>
      </Group>
      <ExpensesTable />
    </>
  );
}

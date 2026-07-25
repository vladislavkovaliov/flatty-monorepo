"use client";

import { Box } from "@mantine/core";
import { UsersTable } from "@/features/users/ui/users-table";

export default function UsersPage() {
  return (
    <>
      <Box py="md" />
      <UsersTable />
    </>
  );
}

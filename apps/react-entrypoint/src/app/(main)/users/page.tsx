"use client";

import { Box, TextInput } from "@mantine/core";
import { useRef } from "react";
import { logMessage } from "@/actions/log";
import { UsersTable } from "@/features/users/ui/users-table";

export default function UsersPage() {
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(null);

  const handleSearch = (value: string) => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    debounceRef.current = setTimeout(async () => {
      if (value) {
        await logMessage(value);
      }
    }, 300);
  };

  return (
    <>
      <Box py="md">
        <TextInput
          placeholder="Search users..."
          onChange={(event) => handleSearch(event.currentTarget.value)}
        />
      </Box>
      <UsersTable />
    </>
  );
}

"use client";

import type { ListUserResponse } from "@flatty-budget/sdk";
import { useUsersGraphql } from "@flatty-budget/sdk";
import { Badge, Box, Container, Pagination, Table } from "@mantine/core";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useEffect } from "react";

const LIMIT = 10;

type UserRow = ListUserResponse["data"][number];

export function UsersTable() {
  const searchParams = useSearchParams() ?? new URLSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const page = Number(searchParams.get("page") || "1");
  const offset = (page - 1) * LIMIT;

  const { data } = useUsersGraphql(LIMIT, offset);

  const responseData = data?.userList;
  const total = responseData?.total ?? 0;
  const totalPages = Math.ceil(total / LIMIT);

  useEffect(() => {
    if (page > 1 && totalPages > 0 && page > totalPages) {
      const params = new URLSearchParams(searchParams.toString());
      params.set("page", "1");
      router.push(`${pathname}?${params.toString()}`);
    }
  });

  const handlePageChange = (newPage: number) => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("page", String(newPage));
    router.push(`${pathname}?${params.toString()}`);
  };

  const rows = ((responseData?.data ?? []) as UserRow[]).map((element) => (
    <Table.Tr key={element.id}>
      <Table.Td>
        <Link href={`/users/${element.id}`}>{element.id.slice(0, 8)}</Link>
      </Table.Td>
      <Table.Td>{element.name}</Table.Td>
      <Table.Td>{element.email}</Table.Td>
      <Table.Td>
        <Badge color={element.emailVerified ? "green" : "gray"} variant="light">
          {element.emailVerified ? "Verified" : "Not verified"}
        </Badge>
      </Table.Td>
      <Table.Td>{String(element.createdAt)}</Table.Td>
    </Table.Tr>
  ));

  return (
    <>
      <Box py="md">
        <Container fluid>
          <Table>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>ID</Table.Th>
                <Table.Th>Name</Table.Th>
                <Table.Th>Email</Table.Th>
                <Table.Th>Email Verified</Table.Th>
                <Table.Th>Created At</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>{rows}</Table.Tbody>
          </Table>
        </Container>
      </Box>

      {totalPages > 1 ? (
        <Box py="xl">
          <Container fluid>
            <Pagination
              total={totalPages}
              value={page}
              onChange={handlePageChange}
            />
          </Container>
        </Box>
      ) : null}
    </>
  );
}

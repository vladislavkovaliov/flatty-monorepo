"use client";

import { useCategoriesGraphql, useDeleteCategory } from "@flatty-budget/sdk";
import {
  Box,
  Button,
  Container,
  Group,
  Pagination,
  Table,
} from "@mantine/core";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useEffect } from "react";

const LIMIT = 5;

export function CategoriesTable() {
  const searchParams = useSearchParams() ?? new URLSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const page = Number(searchParams.get("page") || "1");
  const offset = (page - 1) * LIMIT;

  // const { data } = useCategories(LIMIT, offset);
  const { data } = useCategoriesGraphql(LIMIT, offset);
  const deleteMutation = useDeleteCategory();

  const total = data?.categoryList.total ?? 0;
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

  const rows = (data?.categoryList.data ?? []).map((element) => (
    <Table.Tr key={element.id}>
      <Table.Td>{element.id}</Table.Td>
      <Table.Td>{element.name}</Table.Td>
      <Table.Td>{element.description}</Table.Td>
      <Table.Td>
        <Group gap="xs">
          <Button
            size="xs"
            variant="light"
            color="red"
            loading={deleteMutation.isPending}
            onClick={() => deleteMutation.mutate(element.id)}
          >
            Delete
          </Button>
        </Group>
      </Table.Td>
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
                <Table.Th>Description</Table.Th>
                <Table.Th>Actions</Table.Th>
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

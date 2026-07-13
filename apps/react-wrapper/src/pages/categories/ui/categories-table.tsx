import { useEffect } from 'react';
import { useCategories, useDeleteCategory } from "@flatty-budget/sdk";
import { Table, Group, Button, Box, Container, Pagination } from "@mantine/core";
import { useSearchParams } from "react-router-dom";

const LIMIT = 5;

export function CategoriesTable() {
    const [searchParams, setSearchParams] = useSearchParams();
    const page = Number(searchParams.get('page') || '1');
    const offset = (page - 1) * LIMIT;

    const { data } = useCategories(LIMIT, offset);
    const deleteMutation = useDeleteCategory()

    const total = data?.total ?? 0;
    const totalPages = Math.ceil(total / LIMIT);

    useEffect(() => {
      if (page > 1 && totalPages > 0 && page > totalPages) {
        setSearchParams({ page: '1' });
      }
    });

    const handlePageChange = (newPage: number) => {
        setSearchParams({ page: String(newPage) });
    };

    const rows = (data?.data || []).map((element) => (
        <Table.Tr key={element.id}>
            <Table.Td>{element.id}</Table.Td>
            <Table.Td>{element.name}</Table.Td>
            <Table.Td>{element.description}</Table.Td>
            <Table.Td>
                <Group gap="xs">
                    <Button size="xs" variant="light" color="red" loading={deleteMutation.isPending} onClick={() => deleteMutation.mutate(element.id)}>
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
                    <Pagination total={totalPages} value={page} onChange={handlePageChange} />
                </Container>
                </Box>
            ) : null}
      </>
    )
}
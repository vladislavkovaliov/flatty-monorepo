import { useState } from 'react';
import { Alert, Box, Button, Center, Container, Loader, Pagination, Table, Text } from "@mantine/core";
import { useResidentLocation, useDeleteResidentLocation } from "@flatty-budget/sdk";

const PAGE_SIZE = 10;

export function ResidentTableTab() {
    const [page, setPage] = useState(1);
    const { data, isLoading, isError, error } = useResidentLocation();
    const deleteMutation = useDeleteResidentLocation();
    const [deletingId, setDeletingId] = useState<number | null>(null);

    const handleDelete = (id: number) => {
        setDeletingId(id);
        deleteMutation.mutate(id, {
            onSettled: () => setDeletingId(null),
        });
    };

    if (isLoading && !data) {
        return (
            <Center py="xl">
                <Loader />
                <Text ml="sm">Loading residents...</Text>
            </Center>
        );
    }

    if (isError) {
        return (
            <Box py="md">
                <Container fluid>
                    <Alert color="red" title="Failed to load residents">
                        {error instanceof Error ? error.message : 'An unexpected error occurred'}
                    </Alert>
                </Container>
            </Box>
        );
    }

    const rows = (data?.data || []).map((element) => (
        <Table.Tr key={element.id}>
            <Table.Td>{element.id}</Table.Td>
            <Table.Td>{element.country}</Table.Td>
            <Table.Td>{element.city}</Table.Td>
            <Table.Td>
                <Button
                    size="xs"
                    variant="light"
                    color="red"
                    loading={deletingId === element.id}
                    onClick={() => handleDelete(element.id)}
                >
                    Delete
                </Button>
            </Table.Td>
        </Table.Tr>
    ));

    const totalPages = data?.total ? Math.ceil(data.total / PAGE_SIZE) : 0;

    return (
        <>
            <Box py="md">
                <Container fluid>
                    {(rows.length === 0) ? (
                        <Text c="dimmed" ta="center" py="xl">No residents found</Text>
                    ) : (
                        <Table>
                            <Table.Thead>
                                <Table.Tr>
                                    <Table.Th>ID</Table.Th>
                                    <Table.Th>Country</Table.Th>
                                    <Table.Th>City</Table.Th>
                                    <Table.Th>Actions</Table.Th>
                                </Table.Tr>
                            </Table.Thead>
                            <Table.Tbody>{rows}</Table.Tbody>
                        </Table>
                    )}
                </Container>
            </Box>

            {totalPages > 1 ? (
                <Box py="xl">
                    <Container fluid>
                        <Pagination total={totalPages} value={page} onChange={setPage} />
                    </Container>
                </Box>
            ) : null}
        </>
    );
}

import { useEffect } from 'react';
import { useExpensesGraphql, useDeleteExpensesGraphql } from "@flatty-budget/sdk";
import { Table, Button, Box, Container, Pagination } from "@mantine/core";
import { useSearchParams } from "react-router-dom";

const LIMIT = 5;

export function ExpensesTable() {
    const [searchParams, setSearchParams] = useSearchParams();
    const page = Number(searchParams.get('page') || '1');
    const offset = (page - 1) * LIMIT;

    const { data } = useExpensesGraphql(LIMIT, offset);
    const deleteMutation = useDeleteExpensesGraphql()

    const total = data?.expenseList?.total ?? 0;
    const totalPages = Math.ceil(total / LIMIT);

    useEffect(() => {
      if (page > 1 && totalPages > 0 && page > totalPages) {
        setSearchParams({ page: '1' });
      }
    });

    const handlePageChange = (newPage: number) => {
        setSearchParams({ page: String(newPage) });
    };

    const rows = (data?.expenseList?.data || []).map((element) => (
        <Table.Tr key={element.id}>
            <Table.Td>{element.id}</Table.Td>
            <Table.Td>{element.amount}</Table.Td>
            <Table.Td>{element.category?.description}</Table.Td>
            <Table.Td>{element.month}</Table.Td>
            <Table.Td>{element.year}</Table.Td>
            <Table.Td>
                <Button size="xs" variant="light" color="red" loading={deleteMutation.isPending} onClick={() => deleteMutation.mutate(element.id)}>
                    Delete
                </Button>
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
                        <Table.Th>Amount</Table.Th>
                        <Table.Th>Category</Table.Th>
                        <Table.Th>Month</Table.Th>
                        <Table.Th>Year</Table.Th>
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

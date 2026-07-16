import { useEffect } from 'react';
import { useUsersGraphql } from "@flatty-budget/sdk";
import { Table, Box, Container, Pagination, Badge } from "@mantine/core";
import { useSearchParams } from "react-router-dom";

const LIMIT = 10;

export function UsersTable() {
    const [searchParams, setSearchParams] = useSearchParams();
    const page = Number(searchParams.get('page') || '1');
    const offset = (page - 1) * LIMIT;

    const { data } = useUsersGraphql(LIMIT, offset);

    const responseData = data?.userList;
    const total = responseData?.total ?? 0;
    const totalPages = Math.ceil(total / LIMIT);

    useEffect(() => {
        if (page > 1 && totalPages > 0 && page > totalPages) {
            setSearchParams({ page: '1' });
        }
    });

    const handlePageChange = (newPage: number) => {
        setSearchParams({ page: String(newPage) });
    };

    const rows = (responseData?.data || []).map((element) => (
        <Table.Tr key={element.id}>
            <Table.Td>{element.id}</Table.Td>
            <Table.Td>{element.name}</Table.Td>
            <Table.Td>{element.email}</Table.Td>
            <Table.Td>
                <Badge color={element.emailVerified ? 'green' : 'gray'} variant="light">
                    {element.emailVerified ? 'Verified' : 'Not verified'}
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
                        <Pagination total={totalPages} value={page} onChange={handlePageChange} />
                    </Container>
                </Box>
            ) : null}
        </>
    )
}

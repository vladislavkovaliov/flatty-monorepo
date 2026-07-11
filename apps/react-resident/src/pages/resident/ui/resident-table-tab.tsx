import { Box, Button, Container, Group, Pagination, Table } from "@mantine/core";
import { useResidentLocation, useDeleteResidentLocation } from "../api/resident-location.queries";

export function ResidentTableTab() {
    const {data} = useResidentLocation()
    const deleteMutation = useDeleteResidentLocation()

    const rows = (data?.data || []).map((element) => (
        <Table.Tr key={element.id}>
            <Table.Td>{element.id}</Table.Td>
            <Table.Td>{element.country}</Table.Td>
            <Table.Td>{element.city}</Table.Td>
            <Table.Td>
                <Group gap="xs">
                    <Button size="xs" variant="light" color="red" loading={deleteMutation.isPending} onClick={() => deleteMutation.mutate(element.id)}>
                        Delete
                    </Button>
                </Group>
            </Table.Td>
        </Table.Tr>
    ));
    
    const total = data?.total ?? 0;

    return (
        <>
            <Box py="md">
            <Container fluid>
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
            </Container>
            </Box>

            {total !== 0 ? (
                <Box py="xl">
                <Container fluid>
                    <Pagination total={total} />
                </Container>
                </Box>
            ) : null}
      </>
    )
}
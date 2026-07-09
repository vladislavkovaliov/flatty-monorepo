import { Box, Container, Pagination, Table } from "@mantine/core";
import { useResidentLocation } from "../api/resident-location.queries";
import { useResidentLocationGraphql } from "../api/resident-location.graphql";

export function ResidentTableTab() {
    const {data} = useResidentLocation()
    const grap = useResidentLocationGraphql();

    console.log(grap)
    
    const rows = (data?.data || []).map((element) => (
        <Table.Tr key={element.id}>
            <Table.Td>{element.id}</Table.Td>
            <Table.Td>{element.country}</Table.Td>
            <Table.Td>{element.city}</Table.Td>
        </Table.Tr>
    ));
    
    const total = data?.total ?? 0;

    return (
        <>
            <Box py="xl">
            <Container fluid>
                <Table>
                    <Table.Thead>
                    <Table.Tr>
                        <Table.Th>ID</Table.Th>
                        <Table.Th>Country</Table.Th>
                        <Table.Th>City</Table.Th>
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
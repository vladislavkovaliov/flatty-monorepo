import { Box, Button, Group } from "@mantine/core";
import { ExpensesTable } from "./expenses-table";
import { useNavigate } from "react-router-dom";

export function ExpensesPage() {
    const navigate = useNavigate();

    return (
        <>
            <Box py="md" />
            <Group justify="flex-end" px="md">
                <Button onClick={() => navigate('/expenses/create')}>
                    Create Expense
                </Button>
            </Group>
            <ExpensesTable />
        </>
    )
}

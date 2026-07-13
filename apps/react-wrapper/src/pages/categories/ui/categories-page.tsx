import { Box, Button, Container } from "@mantine/core";
import { CategoriesTable } from "./categories-table";
import { useNavigate } from "react-router-dom";

export function CategoriesPage() {
    const navigate = useNavigate()

    const handleCreateCategoryRedirect = () => {
        navigate("/categories/create");
    }

    return (
        <>
            <Container fluid>
                <Button onClick={handleCreateCategoryRedirect}>Create category</Button>
            </Container>

            <Box py="md" />
            
            <CategoriesTable />
        </>
    )
}
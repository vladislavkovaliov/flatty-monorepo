"use client";

import { Box, Button, Container } from "@mantine/core";
import { useRouter } from "next/navigation";
import { CategoriesTable } from "@/features/categories/ui/categories-table";

export default function CategoriesPage() {
  const router = useRouter();

  return (
    <>
      <Container fluid>
        <Button onClick={() => router.push("/categories/create")}>
          Create category
        </Button>
      </Container>

      <Box py="md" />

      <CategoriesTable />
    </>
  );
}

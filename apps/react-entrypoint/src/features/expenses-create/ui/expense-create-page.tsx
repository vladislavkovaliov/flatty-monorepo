"use client";

import { Container, Grid, Title } from "@mantine/core";
import { useState } from "react";

import { ExpenseForm } from "./expense-form";
import { PdfPreviewCard } from "./pdf-preview-card";

export function ExpenseCreatePage() {
  const [pdfUrl, setPdfUrl] = useState<string | null>(null);

  return (
    <Container size="xl" py="xl">
      <Title order={3} mb="lg">
        Create Expense
      </Title>

      <Grid>
        <Grid.Col span={{ base: 12, md: 7 }}>
          <ExpenseForm onPdfUrlChange={setPdfUrl} />
        </Grid.Col>

        <Grid.Col span={{ base: 12, md: 5 }}>
          <PdfPreviewCard url={pdfUrl} />
        </Grid.Col>
      </Grid>
    </Container>
  );
}

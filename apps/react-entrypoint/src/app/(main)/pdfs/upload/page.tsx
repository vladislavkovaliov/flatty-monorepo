"use client";

import {
  Button,
  Card,
  Container,
  FileInput,
  Group,
  Stack,
  Table,
  Text,
  Title,
} from "@mantine/core";
import { useState } from "react";

import type { ParsedInvoice } from "@/lib/parseInvoice";

export default function PdfUploadPage() {
  const [file, setFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [result, setResult] = useState<string | null>(null);
  const [parsed, setParsed] = useState<ParsedInvoice[] | null>(null);
  const [parseError, setParseError] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleUpload = async () => {
    setError(null);
    setResult(null);
    setParsed(null);
    setParseError(null);

    if (!file) {
      setError("Please select a PDF file");
      return;
    }

    const formData = new FormData();
    formData.append("file", file);

    setUploading(true);
    try {
      const response = await fetch("/api/uploads", {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        const body = (await response.json().catch(() => null)) as {
          error?: string;
        } | null;
        throw new Error(body?.error ?? "Failed to upload the PDF");
      }

      const body = (await response.json()) as {
        filename: string;
        parsed: ParsedInvoice[] | null;
        parseError: string | null;
      };
      setResult(body.filename);
      setParsed(body.parsed);
      setParseError(body.parseError);
    } catch (uploadError) {
      setError(
        uploadError instanceof Error
          ? uploadError.message
          : "Failed to upload the PDF",
      );
    } finally {
      setUploading(false);
    }
  };

  return (
    <Container size="lg" py="xl">
      <Title order={3} mb="lg">
        Upload PDF
      </Title>

      <Stack gap="md">
        <FileInput
          label="PDF file"
          placeholder="Choose a PDF"
          accept="application/pdf,.pdf"
          clearable
          error={error}
          onChange={(selected) => {
            setFile(selected);
            setError(null);
          }}
        />

        <Group justify="flex-end">
          <Button onClick={handleUpload} loading={uploading}>
            Upload
          </Button>
        </Group>

        {result ? <Text c="green">Uploaded: {result}</Text> : null}

        {parseError ? <Text c="red">Parse failed: {parseError}</Text> : null}

        {parsed && parsed.length > 0 ? (
          <Card withBorder padding="md">
            <Table striped highlightOnHover>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Vendor</Table.Th>
                  <Table.Th>Amount</Table.Th>
                  <Table.Th>Date</Table.Th>
                  <Table.Th>Description</Table.Th>
                  <Table.Th>Category</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {parsed.map((invoice) => (
                  <Table.Tr
                    key={`${invoice.vendor}-${invoice.date}-${invoice.amount}`}
                  >
                    <Table.Td>{invoice.vendor ?? "-"}</Table.Td>
                    <Table.Td>{invoice.amount ?? "-"}</Table.Td>
                    <Table.Td>{invoice.date ?? "-"}</Table.Td>
                    <Table.Td>{invoice.description ?? "-"}</Table.Td>
                    <Table.Td>{invoice.category ?? "-"}</Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          </Card>
        ) : null}
      </Stack>
    </Container>
  );
}

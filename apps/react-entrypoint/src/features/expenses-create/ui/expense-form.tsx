"use client";

import { useCategories, useCreateExpense } from "@flatty-budget/sdk";
import {
  Button,
  FileInput,
  Group,
  NumberInput,
  Select,
  Stack,
  Text,
  TextInput,
} from "@mantine/core";
import { useForm } from "@mantine/form";
import { useRouter } from "next/navigation";

import type { ParsedInvoice } from "@/lib/parseInvoice";

import { buildFormPatch } from "../lib/apply-parsed-invoice";
import type { CategoryOption } from "../lib/category-matching";
import {
  type CreateExpenseForm,
  createExpenseInitialValues,
  createExpenseValidate,
} from "../model/create-expense-form";
import { useInvoiceImport } from "../model/use-invoice-import";

interface ExpenseFormProps {
  onPdfUrlChange: (url: string | null) => void;
}

export function ExpenseForm({ onPdfUrlChange }: ExpenseFormProps) {
  const router = useRouter();
  const createMutation = useCreateExpense();
  const { data: categoriesData } = useCategories();

  const categoryOptions: CategoryOption[] = (categoriesData?.data ?? []).map(
    (c) => ({
      value: String(c.id),
      label: c.description,
      name: c.name,
    }),
  );

  const form = useForm<CreateExpenseForm>({
    initialValues: createExpenseInitialValues(),
    validate: createExpenseValidate,
  });

  const applyParsed = (parsed: ParsedInvoice[]): boolean => {
    const invoice = parsed[0];
    if (!invoice) return false;

    const next = buildFormPatch(invoice, categoryOptions);

    if (Object.keys(next).length > 0) {
      form.setValues(next);
      return true;
    }
    return false;
  };

  const { uploading, uploadError, parsedApplied, handleFileChange } =
    useInvoiceImport({ onParsed: applyParsed, onPdfUrlChange });

  const handleSubmit = (values: CreateExpenseForm) => {
    createMutation.mutate(values, {
      onSuccess: () => router.push("/expenses"),
    });
  };

  return (
    <form onSubmit={form.onSubmit(handleSubmit)}>
      <Stack gap="md">
        <NumberInput
          label="Resident Location ID"
          placeholder="e.g. 1"
          withAsterisk
          min={1}
          key={form.key("resident_location_id")}
          {...form.getInputProps("resident_location_id")}
        />

        <Select
          label="Category"
          placeholder="Select a category"
          withAsterisk
          data={categoryOptions}
          searchable
          key={form.key("category_id")}
          value={String(form.values.category_id)}
          onChange={(value) => form.setFieldValue("category_id", Number(value))}
        />

        <NumberInput
          label="Amount"
          placeholder="e.g. 150.50"
          withAsterisk
          min={0.01}
          decimalScale={2}
          fixedDecimalScale
          key={form.key("amount")}
          {...form.getInputProps("amount")}
        />

        <NumberInput
          label="Month"
          placeholder="e.g. 7"
          withAsterisk
          min={1}
          max={12}
          key={form.key("month")}
          {...form.getInputProps("month")}
        />

        <NumberInput
          label="Year"
          placeholder="e.g. 2026"
          withAsterisk
          min={2000}
          key={form.key("year")}
          {...form.getInputProps("year")}
        />

        <TextInput
          label="Description"
          placeholder="Description"
          key={form.key("description")}
          {...form.getInputProps("description")}
        />

        <FileInput
          label="Receipt (PDF)"
          description="Select a PDF to auto-fill the form from the invoice"
          placeholder="Upload a PDF"
          accept="application/pdf,.pdf"
          clearable
          error={uploadError}
          onChange={handleFileChange}
        />

        {uploading ? (
          <Text size="sm" c="dimmed">
            Parsing invoice… this can take a few seconds
          </Text>
        ) : null}

        {parsedApplied ? (
          <Text size="sm" c="green">
            Invoice data applied to the form. Review and fill in any missing
            fields.
          </Text>
        ) : null}

        <Group justify="flex-end" mt="md">
          <Button variant="default" onClick={() => router.push("/expenses")}>
            Cancel
          </Button>
          <Button type="submit" loading={createMutation.isPending || uploading}>
            Create
          </Button>
        </Group>
      </Stack>
    </form>
  );
}

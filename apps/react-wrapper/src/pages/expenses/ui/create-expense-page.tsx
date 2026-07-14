import { Select, NumberInput, Button, Stack, Container, Title, Group } from '@mantine/core';
import { useForm } from '@mantine/form';
import { useNavigate } from 'react-router-dom';
import { useCreateExpense, useCategories } from '@flatty-budget/sdk';

interface CreateExpenseForm {
  resident_location_id: number;
  category_id: number;
  amount: number;
  month: number;
  year: number;
}

export function CreateExpensePage() {
  const navigate = useNavigate();
  const createMutation = useCreateExpense();
  const { data: categoriesData } = useCategories();

  const categoryOptions = (categoriesData?.data ?? []).map((c) => ({
    value: String(c.id),
    label: c.description,
  }));

  const form = useForm<CreateExpenseForm>({
    initialValues: {
      resident_location_id: 1,
      category_id: 1,
      amount: 0,
      month: new Date().getMonth() + 1,
      year: new Date().getFullYear(),
    },
    validate: {
      amount: (value) => (value <= 0 ? 'Amount must be positive' : null),
      month: (value) => (value < 1 || value > 12 ? 'Month must be between 1 and 12' : null),
      year: (value) => (value < 2000 ? 'Year must be 2000 or later' : null),
    },
  });

  const handleSubmit = (values: CreateExpenseForm) => {
    createMutation.mutate(values, {
      onSuccess: () => navigate('/expenses'),
    });
  };

  return (
    <Container size="sm" py="xl">
      <Title order={3} mb="lg">Create Expense</Title>

      <form onSubmit={form.onSubmit(handleSubmit)}>
        <Stack gap="md">
          <NumberInput
            label="Resident Location ID"
            placeholder="e.g. 1"
            withAsterisk
            min={1}
            key={form.key('resident_location_id')}
            {...form.getInputProps('resident_location_id')}
          />

          <Select
            label="Category"
            placeholder="Select a category"
            withAsterisk
            data={categoryOptions}
            searchable
            key={form.key('category_id')}
            value={String(form.values.category_id)}
            onChange={(value) => form.setFieldValue('category_id', Number(value))}
          />

          <NumberInput
            label="Amount"
            placeholder="e.g. 150.50"
            withAsterisk
            min={0.01}
            decimalScale={2}
            fixedDecimalScale
            key={form.key('amount')}
            {...form.getInputProps('amount')}
          />

          <NumberInput
            label="Month"
            placeholder="e.g. 7"
            withAsterisk
            min={1}
            max={12}
            key={form.key('month')}
            {...form.getInputProps('month')}
          />

          <NumberInput
            label="Year"
            placeholder="e.g. 2026"
            withAsterisk
            min={2000}
            key={form.key('year')}
            {...form.getInputProps('year')}
          />

          <Group justify="flex-end" mt="md">
            <Button variant="default" onClick={() => navigate('/expenses')}>
              Cancel
            </Button>
            <Button type="submit" loading={createMutation.isPending}>
              Create
            </Button>
          </Group>
        </Stack>
      </form>
    </Container>
  );
}

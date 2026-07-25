"use client";

import type { CategoriesForm } from "@flatty-budget/sdk";
import { useCreateCategories } from "@flatty-budget/sdk";
import {
  Button,
  Container,
  Group,
  Stack,
  TextInput,
  Title,
} from "@mantine/core";
import { useForm } from "@mantine/form";
import { useRouter } from "next/navigation";

export default function CreateCategoryPage() {
  const router = useRouter();
  const createMutation = useCreateCategories();

  const form = useForm<CategoriesForm>({
    initialValues: {
      name: "",
      description: "",
    },
    validate: {
      name: (value) =>
        value.trim().length < 2 ? "Name is required (min 2 characters)" : null,
    },
  });

  const handleSubmit = (values: CategoriesForm) => {
    createMutation.mutate(values, {
      onSuccess: () => router.push("/categories"),
    });
  };

  return (
    <Container size="sm" py="xl">
      <Title order={3} mb="lg">
        Create Category
      </Title>

      <form onSubmit={form.onSubmit(handleSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="e.g. utilities"
            withAsterisk
            key={form.key("name")}
            {...form.getInputProps("name")}
          />

          <TextInput
            label="Description"
            placeholder="e.g. Utility bills"
            key={form.key("description")}
            {...form.getInputProps("description")}
          />

          <Group justify="flex-end" mt="md">
            <Button
              variant="default"
              onClick={() => router.push("/categories")}
            >
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

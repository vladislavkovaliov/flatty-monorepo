import { Box, Button, Group, SimpleGrid, Stack, TextInput, Title } from '@mantine/core';
import { useForm } from '@mantine/form';
import { useCreateResidentLocation } from '@flatty-budget/sdk';

interface ResidentCreateForm {
  country: string;
  city: string;
  apartment: string;
  house: string;
  street: string;
  postalCode: string;
}

export function ResidentCreateTab() {
  const createMutation = useCreateResidentLocation();

  const form = useForm<ResidentCreateForm>({
    initialValues: {
      country: '',
      city: '',
      apartment: '',
      house: '',
      street: '',
      postalCode: '',
    },
  });

  const handleSubmit = (values: ResidentCreateForm) => {
    createMutation.mutate(values, {
      onSuccess: () => form.reset(),
    });
  };

  return (
    <Box maw={800} mx="auto">
      <Box mb="sm">
        <Title order={3}>
          Create Resident
        </Title>
      </Box>

      <form onSubmit={form.onSubmit(handleSubmit)}>
        <Stack gap="md">
          <SimpleGrid cols={2}>
            <TextInput
              label="Country"
              placeholder="Poland"
              key={form.key('country')}
              {...form.getInputProps('country')}
            />
            <TextInput
              label="City"
              placeholder="Warsaw"
              key={form.key('city')}
              {...form.getInputProps('city')}
            />
          </SimpleGrid>

          <SimpleGrid cols={2}>
            <TextInput
              label="Street"
              placeholder="Bobr"
              key={form.key('street')}
              {...form.getInputProps('street')}
            />
            <TextInput
              label="House"
              placeholder="1"
              key={form.key('house')}
              {...form.getInputProps('house')}
            />
          </SimpleGrid>

          <SimpleGrid cols={2}>
            <TextInput
              label="Apartment"
              placeholder="2"
              key={form.key('apartment')}
              {...form.getInputProps('apartment')}
            />
            <TextInput
              label="Postal code"
              placeholder="00-945"
              key={form.key('postalCode')}
              {...form.getInputProps('postalCode')}
            />
          </SimpleGrid>

          <Box mt="md">
            <Group justify="flex-end">
              <Button variant="default" onClick={() => form.reset()}>
                Reset
              </Button>
              <Button type="submit" loading={createMutation.isPending}>Create</Button>
            </Group>
          </Box>
        </Stack>
      </form>
    </Box>
  );
}

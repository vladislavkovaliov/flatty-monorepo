import { Box, Button, Group, SimpleGrid, Stack, TextInput, Title } from '@mantine/core';
import { useForm } from '@mantine/form';
import type { IResidentCreate } from '../model/types';

export function ResidentCreateTab() {
  const form = useForm<IResidentCreate>({
    initialValues: {
      fullName: '',
      email: '',
      phone: '',
      dateOfBirth: '',
      country: '',
      city: '',
      address: '',
    },
    validate: {
      fullName: (value) => (value.trim().length < 2 ? 'Full name is required' : null),
      email: (value) => (/^\S+@\S+$/.test(value) ? null : 'Invalid email'),
    },
  });

  const handleSubmit = (values: IResidentCreate) => {
    console.log('Resident created:', values);
  };

  return (
    <Box maw={600} mx="auto">
      <Box mb="lg">
        <Title order={3}>
          Create Resident
        </Title>
      </Box>

      <form onSubmit={form.onSubmit(handleSubmit)}>
        <Stack gap="md">
          <SimpleGrid cols={2}>
            <TextInput
              label="Full name"
              placeholder="John Doe"
              withAsterisk
              key={form.key('fullName')}
              {...form.getInputProps('fullName')}
            />
            <TextInput
              label="Email"
              placeholder="john@example.com"
              withAsterisk
              key={form.key('email')}
              {...form.getInputProps('email')}
            />
          </SimpleGrid>

          <SimpleGrid cols={2}>
            <TextInput
              label="Phone"
              placeholder="+1 234 567 890"
              key={form.key('phone')}
              {...form.getInputProps('phone')}
            />
            <TextInput
              label="Date of birth"
              placeholder="YYYY-MM-DD"
              key={form.key('dateOfBirth')}
              {...form.getInputProps('dateOfBirth')}
            />
          </SimpleGrid>

          <SimpleGrid cols={2}>
            <TextInput
              label="Country"
              placeholder="Country"
              key={form.key('country')}
              {...form.getInputProps('country')}
            />
            <TextInput
              label="City"
              placeholder="City"
              key={form.key('city')}
              {...form.getInputProps('city')}
            />
          </SimpleGrid>

          <TextInput
            label="Address"
            placeholder="123 Main St"
            key={form.key('address')}
            {...form.getInputProps('address')}
          />

          <Box mt="md">
            <Group justify="flex-end">
            <Button variant="default" onClick={() => form.reset()}>
              Reset
            </Button>
            <Button type="submit">Create</Button>
          </Group>
          </Box>
        </Stack>
      </form>
    </Box>
  );
}

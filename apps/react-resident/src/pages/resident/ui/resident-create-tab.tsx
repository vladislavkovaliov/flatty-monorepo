import { Box, Button, Group, SimpleGrid, Stack, TextInput, Title } from '@mantine/core';
import { DatePickerInput } from '@mantine/dates';
import { useForm } from '@mantine/form';
import type { ResidentCreate } from '../model/types';

export function ResidentCreateTab() {
  const form = useForm<ResidentCreate>({
    initialValues: {
      fullName: '',
      email: '',
      phone: '',
      dateOfBirth: null,
      country: '',
      city: '',
      apartment: '',
      house: '',
      street: '',
      postalCode: '',
      address: '',
    },
    validate: {
      fullName: (value) => (value.trim().length < 2 ? 'Full name is required' : null),
      email: (value) => (/^\S+@\S+$/.test(value) ? null : 'Invalid email'),
    },
  });

  const handleSubmit = (values: ResidentCreate) => {
    console.log('Resident created:', values);
  };

  return (
    <Box maw={800} mx="auto">
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
            <DatePickerInput
              label="Date of birth"
              placeholder="Pick date"
              clearable
              value={form.values.dateOfBirth}
              onChange={(value) => form.setFieldValue('dateOfBirth', value)}
              error={form.errors.dateOfBirth}
            />
          </SimpleGrid>

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

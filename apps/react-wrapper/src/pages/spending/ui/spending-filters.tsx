import { Group, Button } from "@mantine/core";
import { MonthPickerInput } from "@mantine/dates";

interface SpendingFiltersProps {
  fromDate: Date | null;
  toDate: Date | null;
  onFromDateChange: (date: Date | null) => void;
  onToDateChange: (date: Date | null) => void;
  onApply: () => void;
  onReset: () => void;
  minDate?: Date;
  maxDate?: Date;
}

export function SpendingFilters({
  fromDate,
  toDate,
  onFromDateChange,
  onToDateChange,
  onApply,
  onReset,
  minDate,
  maxDate,
}: SpendingFiltersProps) {
  return (
    <Group px="md" gap="sm" align="flex-end">
      <MonthPickerInput
        label="From"
        placeholder="Pick start month"
        value={fromDate}
        onChange={(v) => onFromDateChange(v ? new Date(v) : null)}
        minDate={minDate}
        maxDate={maxDate}
        clearable
      />
      <MonthPickerInput
        label="To"
        placeholder="Pick end month"
        value={toDate}
        onChange={(v) => onToDateChange(v ? new Date(v) : null)}
        minDate={minDate}
        maxDate={maxDate}
        clearable
      />
      <Button onClick={onApply}>Apply</Button>
      <Button variant="light" onClick={onReset}>Reset</Button>
    </Group>
  );
}

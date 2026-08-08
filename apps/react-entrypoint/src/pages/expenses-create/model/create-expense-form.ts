export interface CreateExpenseForm {
  resident_location_id: number;
  category_id: number;
  amount: number;
  description: string;
  month: number;
  year: number;
}

export function createExpenseInitialValues(): CreateExpenseForm {
  return {
    resident_location_id: 1,
    category_id: 1,
    amount: 0,
    description: "",
    month: new Date().getMonth() + 1,
    year: new Date().getFullYear(),
  };
}

export function createExpenseValidate(values: CreateExpenseForm): {
  amount: string | null;
  month: string | null;
  year: string | null;
} {
  return {
    amount: values.amount <= 0 ? "Amount must be positive" : null,
    month:
      values.month < 1 || values.month > 12
        ? "Month must be between 1 and 12"
        : null,
    year: values.year < 2000 ? "Year must be 2000 or later" : null,
  };
}

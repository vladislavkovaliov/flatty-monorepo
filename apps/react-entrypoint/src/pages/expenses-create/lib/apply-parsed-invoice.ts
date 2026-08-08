import type { ParsedInvoice } from "@/lib/parseInvoice";

import type { CreateExpenseForm } from "../model/create-expense-form";
import { type CategoryOption, findCategoryId } from "./category-matching";
import { parseInvoiceDate } from "./parse-invoice-date";

export function buildFormPatch(
  invoice: ParsedInvoice,
  options: CategoryOption[],
): Partial<CreateExpenseForm> {
  const next: Partial<CreateExpenseForm> = {};

  if (typeof invoice.amount === "number" && invoice.amount > 0) {
    next.amount = invoice.amount;
  }

  const description = invoice.description?.trim();
  const vendor = invoice.vendor?.trim();
  if (description) {
    next.description = description;
  } else if (vendor) {
    next.description = vendor;
  }

  const date = parseInvoiceDate(invoice.date);
  if (date) {
    next.month = date.month;
    next.year = date.year;
  }

  if (invoice.category?.trim()) {
    const categoryId = findCategoryId(invoice.category, options);
    if (categoryId !== undefined) {
      next.category_id = categoryId;
    }
  }

  return next;
}

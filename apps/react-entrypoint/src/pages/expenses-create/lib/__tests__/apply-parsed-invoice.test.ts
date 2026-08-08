import { describe, expect, it } from "vitest";

import type { ParsedInvoice } from "../../../../lib/parseInvoice";

import { buildFormPatch } from "../apply-parsed-invoice";
import type { CategoryOption } from "../category-matching";

const OPTIONS: CategoryOption[] = [
  { value: "1", label: "Коммунальные платежи", name: "utilities" },
  { value: "3", label: "Электроэнергия", name: "electricity" },
  { value: "5", label: "МТС (292015787)", name: "mts_1" },
];

describe("buildFormPatch", () => {
  it("maps a positive amount", () => {
    const invoice: ParsedInvoice = { amount: 150.5 };
    expect(buildFormPatch(invoice, OPTIONS)).toEqual({ amount: 150.5 });
  });

  it("ignores a non-positive amount", () => {
    expect(buildFormPatch({ amount: 0 }, OPTIONS)).toEqual({});
  });

  it("uses the description when present", () => {
    expect(
      buildFormPatch({ description: "  Rent  ", vendor: "ACME" }, OPTIONS),
    ).toEqual({ description: "Rent" });
  });

  it("falls back to the vendor when description is missing", () => {
    expect(buildFormPatch({ vendor: "  ACME  " }, OPTIONS)).toEqual({
      description: "ACME",
    });
  });

  it("maps the invoice date", () => {
    expect(buildFormPatch({ date: "2026-07-15" }, OPTIONS)).toEqual({
      month: 7,
      year: 2026,
    });
  });

  it("maps the category via findCategoryId", () => {
    expect(buildFormPatch({ category: "МТС" }, OPTIONS)).toEqual({
      category_id: 5,
    });
  });

  it("returns an empty patch for an empty invoice", () => {
    expect(buildFormPatch({}, OPTIONS)).toEqual({});
  });
});

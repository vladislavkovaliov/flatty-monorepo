import { describe, expect, it } from "vitest";

import { parseInvoiceDate } from "../parse-invoice-date";

describe("parseInvoiceDate", () => {
  it("parses ISO date YYYY-MM-DD", () => {
    expect(parseInvoiceDate("2026-07-15")).toEqual({ month: 7, year: 2026 });
  });

  it("parses DD.MM.YYYY", () => {
    expect(parseInvoiceDate("15.07.2026")).toEqual({ month: 7, year: 2026 });
  });

  it("parses MM/DD/YYYY", () => {
    expect(parseInvoiceDate("07/15/2026")).toEqual({ month: 7, year: 2026 });
  });

  it("parses YYYY-MM", () => {
    expect(parseInvoiceDate("2026-07")).toEqual({ month: 7, year: 2026 });
  });

  it("returns null for an invalid month", () => {
    expect(parseInvoiceDate("2026-13-01")).toBeNull();
  });

  it("returns null for an invalid year", () => {
    expect(parseInvoiceDate("1999-07-01")).toBeNull();
  });

  it("returns null for an empty input", () => {
    expect(parseInvoiceDate("")).toBeNull();
    expect(parseInvoiceDate(undefined)).toBeNull();
  });

  it("returns null for an unparseable string", () => {
    expect(parseInvoiceDate("not a date")).toBeNull();
  });
});

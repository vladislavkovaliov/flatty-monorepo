import { describe, expect, it } from "vitest";

import {
  type CategoryOption,
  categoryStrings,
  findCategoryId,
  normalizeCategory,
} from "../category-matching";

const OPTIONS: CategoryOption[] = [
  { value: "1", label: "Коммунальные платежи", name: "utilities" },
  { value: "3", label: "Электроэнергия", name: "electricity" },
  { value: "5", label: "МТС (292015787)", name: "mts_1" },
];

describe("normalizeCategory", () => {
  it("lowercases, collapses whitespace and trims", () => {
    expect(normalizeCategory("  Коммунальные   Платежи  ")).toBe(
      "коммунальные платежи",
    );
  });
});

describe("categoryStrings", () => {
  it("returns the label and name of an option", () => {
    expect(categoryStrings(OPTIONS[0])).toEqual([
      "Коммунальные платежи",
      "utilities",
    ]);
  });
});

describe("findCategoryId", () => {
  it("matches exactly by name", () => {
    expect(findCategoryId("utilities", OPTIONS)).toBe(1);
  });

  it("matches exactly by label", () => {
    expect(findCategoryId("Коммунальные платежи", OPTIONS)).toBe(1);
  });

  it("matches a label with different casing", () => {
    expect(findCategoryId("Электроэнергия", OPTIONS)).toBe(3);
  });

  it("fuzzy-matches a partial label", () => {
    expect(findCategoryId("МТС", OPTIONS)).toBe(5);
  });

  it("returns undefined when nothing matches", () => {
    expect(findCategoryId("unknown", OPTIONS)).toBeUndefined();
  });

  it("returns undefined for an empty input", () => {
    expect(findCategoryId("", OPTIONS)).toBeUndefined();
  });
});

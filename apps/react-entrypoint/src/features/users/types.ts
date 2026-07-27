import type { InferSelectModel } from "drizzle-orm";
import type {
  expenseMonthlyAverages,
  expenseMonthlyTotals,
  residentLocations,
  user,
} from "@/drizzle/schema";

export type UserDetailUser = InferSelectModel<typeof user>;

export type UserDetailLocation = Omit<
  InferSelectModel<typeof residentLocations>,
  "userId"
>;

export type UserDetailExpenseTotal = InferSelectModel<
  typeof expenseMonthlyTotals
>;

export type UserDetailExpenseAverage = InferSelectModel<
  typeof expenseMonthlyAverages
>;

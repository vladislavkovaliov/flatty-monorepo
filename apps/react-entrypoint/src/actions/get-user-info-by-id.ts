"use server";

import { desc, eq } from "drizzle-orm";
import * as schema from "@/drizzle/schema";
import { db } from "@/lib/db";

export async function getUserInfoById(userId: string): Promise<any> {
  const [resultUserInfo, resultExpenseMonthTotal, resultExpenseMonthAvg] =
    await Promise.all([
      db
        .select()
        .from(schema.user)
        .where(eq(schema.user.id, userId))
        .rightJoin(
          schema.residentLocations,
          eq(schema.residentLocations.userId, userId),
        ),
      db
        .select()
        .from(schema.expenseMonthlyTotals)
        .orderBy(desc(schema.expenseMonthlyTotals.createdAt))
        .limit(1),
      db
        .select()
        .from(schema.expenseMonthlyAverages)
        .orderBy(desc(schema.expenseMonthlyAverages.createdAt))
        .limit(1),
    ]);

  return {
    userInfo: resultUserInfo,
    expense: {
      total: resultExpenseMonthTotal,
      average: resultExpenseMonthAvg,
    },
  };
}

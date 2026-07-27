import { desc, eq } from "drizzle-orm";
import * as schema from "@/drizzle/schema";
import { db } from "@/lib/db";

export async function getUserInfoById(userId: string) {
  const [userRows, totalRows, avgRows] = await Promise.all([
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
    user: userRows[0] ?? null,
    expense: {
      total: totalRows[0] ?? null,
      average: avgRows[0] ?? null,
    },
  };
}

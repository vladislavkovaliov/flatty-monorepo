"use server";

import type { ListUserResponse } from "@flatty-budget/sdk";
import { ilike } from "drizzle-orm";
import * as schema from "@/drizzle/schema";
import { db } from "@/lib/db";

export async function logMessage(email: string): Promise<ListUserResponse> {
  const users = await db
    .select()
    .from(schema.user)
    .where(ilike(schema.user.email, `%${email}%`));

  return { data: users, total: users.length };
}

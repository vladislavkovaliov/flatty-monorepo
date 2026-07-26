import { drizzle } from "drizzle-orm/node-postgres";
import { Pool } from "pg";
import * as schema from "@/drizzle/schema";

function getPool(): Pool {
  const globalPool = globalThis as typeof globalThis & { _pgPool?: Pool };
  if (!globalPool._pgPool) {
    globalPool._pgPool = new Pool({
      connectionString: process.env.DATABASE_URL,
    });
  }
  return globalPool._pgPool;
}

const pool = getPool();

export const db = drizzle(pool, { schema });

async function closeDb(): Promise<void> {
  await pool.end();
}

process.on("SIGTERM", () => void closeDb());
process.on("SIGINT", () => void closeDb());

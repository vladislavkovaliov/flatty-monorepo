import { config } from "dotenv";
import { defineConfig } from "drizzle-kit";

// Explicitly load the Next.js local env file
config({ path: ".env.local" });

export default defineConfig({
  out: "./src/drizzle",
  schema: "./src/drizzle/schema.ts",
  dialect: "postgresql",
  dbCredentials: {
    url: process.env.DATABASE_URL!,
  },
});

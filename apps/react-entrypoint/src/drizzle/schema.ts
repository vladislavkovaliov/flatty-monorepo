import { sql } from "drizzle-orm";
import {
  bigint,
  bigserial,
  boolean,
  check,
  foreignKey,
  index,
  integer,
  numeric,
  pgEnum,
  pgTable,
  primaryKey,
  text,
  timestamp,
  unique,
  uuid,
  varchar,
} from "drizzle-orm/pg-core";

export const invitationStatus = pgEnum("invitation_status", [
  "pending",
  "accepted",
  "expired",
]);

export const invitation = pgTable("invitation", {
  id: uuid().defaultRandom().primaryKey().notNull(),
  email: varchar({ length: 255 }).notNull(),
  invitedBy: varchar({ length: 255 }).notNull(),
  status: invitationStatus().default("pending").notNull(),
  createdAt: timestamp({ withTimezone: true, mode: "string" })
    .defaultNow()
    .notNull(),
  acceptedAt: timestamp({ withTimezone: true, mode: "string" }),
});

export const residentLocations = pgTable(
  "resident_locations",
  {
    id: bigserial({ mode: "bigint" }).primaryKey().notNull(),
    country: varchar({ length: 100 }).notNull(),
    city: varchar({ length: 100 }).notNull(),
    postalCode: varchar("postal_code", { length: 20 }).notNull(),
    street: varchar({ length: 150 }).notNull(),
    house: varchar({ length: 20 }).notNull(),
    apartment: varchar({ length: 20 }),
    createdAt: timestamp("created_at", { mode: "string" }).default(
      sql`CURRENT_TIMESTAMP`,
    ),
    updatedAt: timestamp("updated_at", { mode: "string" }).default(
      sql`CURRENT_TIMESTAMP`,
    ),
    userId: text("user_id").notNull(),
  },
  (table) => [
    foreignKey({
      columns: [table.userId],
      foreignColumns: [user.id],
      name: "resident_locations_user_id_fkey",
    }).onDelete("cascade"),
  ],
);

export const categories = pgTable(
  "categories",
  {
    id: bigserial({ mode: "bigint" }).primaryKey().notNull(),
    name: varchar({ length: 200 }).notNull(),
    description: text(),
    createdAt: timestamp("created_at", { mode: "string" }).default(
      sql`CURRENT_TIMESTAMP`,
    ),
    updatedAt: timestamp("updated_at", { mode: "string" }).default(
      sql`CURRENT_TIMESTAMP`,
    ),
  },
  (table) => [unique("categories_name_key").on(table.name)],
);

export const expenses = pgTable(
  "expenses",
  {
    id: bigserial({ mode: "bigint" }).primaryKey().notNull(),
    // You can use { mode: "bigint" } if numbers are exceeding js number limitations
    residentLocationId: bigint("resident_location_id", {
      mode: "number",
    }).notNull(),
    // You can use { mode: "bigint" } if numbers are exceeding js number limitations
    categoryId: bigint("category_id", { mode: "number" }).notNull(),
    amount: numeric({ precision: 12, scale: 2 }).notNull(),
    month: integer().notNull(),
    year: integer().notNull(),
    createdAt: timestamp("created_at", { mode: "string" }).default(
      sql`CURRENT_TIMESTAMP`,
    ),
    updatedAt: timestamp("updated_at", { mode: "string" }).default(
      sql`CURRENT_TIMESTAMP`,
    ),
    description: text().default(""),
    userId: text("user_id").notNull(),
  },
  (table) => [
    foreignKey({
      columns: [table.residentLocationId],
      foreignColumns: [residentLocations.id],
      name: "expenses_resident_location_id_fkey",
    }).onDelete("cascade"),
    foreignKey({
      columns: [table.categoryId],
      foreignColumns: [categories.id],
      name: "expenses_category_id_fkey",
    }).onDelete("cascade"),
    foreignKey({
      columns: [table.userId],
      foreignColumns: [user.id],
      name: "expenses_user_id_fkey",
    }).onDelete("cascade"),
    index("expenses_user_id_idx").using(
      "btree",
      table.userId.asc().nullsLast().op("text_ops"),
    ),
    unique("expenses_resident_location_id_category_id_month_year_key").on(
      table.year,
      table.residentLocationId,
      table.month,
      table.categoryId,
    ),
    check("expenses_month_check", sql`(month >= 1) AND (month <= 12)`),
    check("expenses_year_check", sql`year >= 2000`),
  ],
);

export const user = pgTable(
  "user",
  {
    id: text().primaryKey().notNull(),
    name: text().notNull(),
    email: text().notNull(),
    emailVerified: boolean().notNull(),
    image: text(),
    createdAt: timestamp({ withTimezone: true, mode: "string" })
      .default(sql`CURRENT_TIMESTAMP`)
      .notNull(),
    updatedAt: timestamp({ withTimezone: true, mode: "string" })
      .default(sql`CURRENT_TIMESTAMP`)
      .notNull(),
  },
  (table) => [unique("user_email_key").on(table.email)],
);

export const session = pgTable(
  "session",
  {
    id: text().primaryKey().notNull(),
    expiresAt: timestamp({ withTimezone: true, mode: "string" }).notNull(),
    token: text().notNull(),
    createdAt: timestamp({ withTimezone: true, mode: "string" })
      .default(sql`CURRENT_TIMESTAMP`)
      .notNull(),
    updatedAt: timestamp({ withTimezone: true, mode: "string" }).notNull(),
    ipAddress: text(),
    userAgent: text(),
    userId: text().notNull(),
  },
  (table) => [
    index("session_userId_idx").using(
      "btree",
      table.userId.asc().nullsLast().op("text_ops"),
    ),
    foreignKey({
      columns: [table.userId],
      foreignColumns: [user.id],
      name: "session_userId_fkey",
    }).onDelete("cascade"),
    unique("session_token_key").on(table.token),
  ],
);

export const account = pgTable(
  "account",
  {
    id: text().primaryKey().notNull(),
    accountId: text().notNull(),
    providerId: text().notNull(),
    userId: text().notNull(),
    accessToken: text(),
    refreshToken: text(),
    idToken: text(),
    accessTokenExpiresAt: timestamp({ withTimezone: true, mode: "string" }),
    refreshTokenExpiresAt: timestamp({ withTimezone: true, mode: "string" }),
    scope: text(),
    password: text(),
    createdAt: timestamp({ withTimezone: true, mode: "string" })
      .default(sql`CURRENT_TIMESTAMP`)
      .notNull(),
    updatedAt: timestamp({ withTimezone: true, mode: "string" }).notNull(),
  },
  (table) => [
    index("account_userId_idx").using(
      "btree",
      table.userId.asc().nullsLast().op("text_ops"),
    ),
    foreignKey({
      columns: [table.userId],
      foreignColumns: [user.id],
      name: "account_userId_fkey",
    }).onDelete("cascade"),
  ],
);

export const verification = pgTable(
  "verification",
  {
    id: text().primaryKey().notNull(),
    identifier: text().notNull(),
    value: text().notNull(),
    expiresAt: timestamp({ withTimezone: true, mode: "string" }).notNull(),
    createdAt: timestamp({ withTimezone: true, mode: "string" })
      .default(sql`CURRENT_TIMESTAMP`)
      .notNull(),
    updatedAt: timestamp({ withTimezone: true, mode: "string" })
      .default(sql`CURRENT_TIMESTAMP`)
      .notNull(),
  },
  (table) => [
    index("verification_identifier_idx").using(
      "btree",
      table.identifier.asc().nullsLast().op("text_ops"),
    ),
  ],
);

export const roles = pgTable(
  "roles",
  {
    id: bigserial({ mode: "bigint" }).primaryKey().notNull(),
    name: varchar({ length: 50 }).notNull(),
  },
  (table) => [unique("roles_name_key").on(table.name)],
);

export const userSettings = pgTable(
  "user_settings",
  {
    userId: text("user_id").primaryKey().notNull(),
    language: text().default("en").notNull(),
    theme: text().default("system").notNull(),
    timezone: text().default("UTC").notNull(),
    dateFormat: text("date_format").default("YYYY-MM-DD").notNull(),
    createdAt: timestamp("created_at", { withTimezone: true, mode: "string" })
      .defaultNow()
      .notNull(),
    updatedAt: timestamp("updated_at", { withTimezone: true, mode: "string" })
      .defaultNow()
      .notNull(),
  },
  (table) => [
    foreignKey({
      columns: [table.userId],
      foreignColumns: [user.id],
      name: "user_settings_user_id_fkey",
    }).onDelete("cascade"),
  ],
);

export const userRoles = pgTable(
  "user_roles",
  {
    userId: varchar("user_id", { length: 255 }).notNull(),
    // You can use { mode: "bigint" } if numbers are exceeding js number limitations
    roleId: bigint("role_id", { mode: "number" }).notNull(),
  },
  (table) => [
    foreignKey({
      columns: [table.userId],
      foreignColumns: [user.id],
      name: "user_roles_user_id_fkey",
    }).onDelete("cascade"),
    foreignKey({
      columns: [table.roleId],
      foreignColumns: [roles.id],
      name: "user_roles_role_id_fkey",
    }).onDelete("cascade"),
    primaryKey({
      columns: [table.userId, table.roleId],
      name: "user_roles_pkey",
    }),
  ],
);

export const expenseMonthlyTotals = pgTable(
  "expense_monthly_totals",
  {
    month: integer().notNull(),
    year: integer().notNull(),
    totalSpent: numeric("total_spent", { precision: 12, scale: 2 })
      .default("0")
      .notNull(),
    updatedAt: timestamp("updated_at", { mode: "string" }).default(
      sql`CURRENT_TIMESTAMP`,
    ),
    createdAt: timestamp("created_at", { mode: "string" }).default(
      sql`CURRENT_TIMESTAMP`,
    ),
  },
  (table) => [
    primaryKey({
      columns: [table.year, table.month],
      name: "expense_monthly_totals_pkey",
    }),
    check(
      "expense_monthly_totals_month_check",
      sql`(month >= 1) AND (month <= 12)`,
    ),
    check("expense_monthly_totals_year_check", sql`year >= 2000`),
  ],
);

export const expenseMonthlyAverages = pgTable(
  "expense_monthly_averages",
  {
    month: integer().notNull(),
    year: integer().notNull(),
    averageAmount: numeric("average_amount", { precision: 12, scale: 2 })
      .default("0")
      .notNull(),
    expenseCount: integer("expense_count").default(0).notNull(),
    updatedAt: timestamp("updated_at", { mode: "string" }).default(
      sql`CURRENT_TIMESTAMP`,
    ),
    createdAt: timestamp("created_at", { mode: "string" }).default(
      sql`CURRENT_TIMESTAMP`,
    ),
  },
  (table) => [
    primaryKey({
      columns: [table.year, table.month],
      name: "expense_monthly_averages_pkey",
    }),
    check(
      "expense_monthly_averages_month_check",
      sql`(month >= 1) AND (month <= 12)`,
    ),
    check("expense_monthly_averages_year_check", sql`year >= 2000`),
  ],
);

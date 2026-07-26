import { relations } from "drizzle-orm/relations";
import {
  account,
  categories,
  expenses,
  residentLocations,
  roles,
  session,
  user,
  userRoles,
  userSettings,
} from "./schema";

export const residentLocationsRelations = relations(
  residentLocations,
  ({ one, many }) => ({
    user: one(user, {
      fields: [residentLocations.userId],
      references: [user.id],
    }),
    expenses: many(expenses),
  }),
);

export const userRelations = relations(user, ({ many }) => ({
  residentLocations: many(residentLocations),
  sessions: many(session),
  accounts: many(account),
  userSettings: many(userSettings),
  userRoles: many(userRoles),
}));

export const expensesRelations = relations(expenses, ({ one }) => ({
  residentLocation: one(residentLocations, {
    fields: [expenses.residentLocationId],
    references: [residentLocations.id],
  }),
  category: one(categories, {
    fields: [expenses.categoryId],
    references: [categories.id],
  }),
}));

export const categoriesRelations = relations(categories, ({ many }) => ({
  expenses: many(expenses),
}));

export const sessionRelations = relations(session, ({ one }) => ({
  user: one(user, {
    fields: [session.userId],
    references: [user.id],
  }),
}));

export const accountRelations = relations(account, ({ one }) => ({
  user: one(user, {
    fields: [account.userId],
    references: [user.id],
  }),
}));

export const userSettingsRelations = relations(userSettings, ({ one }) => ({
  user: one(user, {
    fields: [userSettings.userId],
    references: [user.id],
  }),
}));

export const userRolesRelations = relations(userRoles, ({ one }) => ({
  user: one(user, {
    fields: [userRoles.userId],
    references: [user.id],
  }),
  role: one(roles, {
    fields: [userRoles.roleId],
    references: [roles.id],
  }),
}));

export const rolesRelations = relations(roles, ({ many }) => ({
  userRoles: many(userRoles),
}));

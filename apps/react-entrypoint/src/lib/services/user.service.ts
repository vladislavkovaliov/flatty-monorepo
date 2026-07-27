import { getUserInfoById } from "@/lib/queries/get-user-info-by-id";

export async function loadUserProfile(userId: string) {
  const data = await getUserInfoById(userId);

  return {
    ...data,
    hasExpenseData: Boolean(data.expense.total || data.expense.average),
  };
}

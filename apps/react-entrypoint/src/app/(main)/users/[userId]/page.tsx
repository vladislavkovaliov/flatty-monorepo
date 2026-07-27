import { UserDetail } from "@/features/users/ui/user-detail";
import { loadUserProfile } from "@/lib/services/user.service";

export default async function UserPage({
  params,
}: {
  params: Promise<{ userId: string }>;
}) {
  const { userId } = await params;
  const {
    user: userRow,
    expense,
    hasExpenseData,
  } = await loadUserProfile(userId);

  const user = userRow?.user ?? null;
  const location = userRow?.resident_locations ?? null;
  const expenseTotal = expense.total ?? null;
  const expenseAverage = expense.average ?? null;

  return (
    <UserDetail
      user={user}
      locations={location}
      expenseTotal={expenseTotal}
      expenseAverage={expenseAverage}
      hasExpenseData={hasExpenseData}
    />
  );
}

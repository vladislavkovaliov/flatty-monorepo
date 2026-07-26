import { getUserInfoById } from "@/actions/get-user-info-by-id";
import { UserDetail } from "@/features/users/ui/user-detail";

export default async function UserPage({
  params,
}: {
  params: Promise<{ userId: string }>;
}) {
  const { userId } = await params;
  const {
    userInfo,
    expense: { total, average },
  } = await getUserInfoById(userId);

  const user = userInfo.length > 0 ? userInfo[0].user : null;
  const locations = userInfo.map(
    (row: { resident_locations: unknown }) => row.resident_locations,
  );

  return (
    <UserDetail
      user={user}
      locations={locations}
      expenseTotal={total}
      expenseAverage={average}
    />
  );
}

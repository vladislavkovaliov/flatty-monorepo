export class ExpenseKeyManager {
  static getKeyListTotal = (userId: string, residentLocationId: number, month?: number, year?: number) =>
    `expense:listTotal:${userId}:${residentLocationId}:${month ?? '*'}:${year ?? '*'}`;
  static getKeyListAverages = (userId: string, residentLocationId: number, month?: number, year?: number) =>
    `expense:listAverages:${userId}:${residentLocationId}:${month ?? '*'}:${year ?? '*'}`;

  static SECONDS_30 = 30_000;
}
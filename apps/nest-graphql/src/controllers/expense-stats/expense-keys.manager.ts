
export class ExpenseKeyManager {
    static getKeyVersion = () => `expense:version`;
    static getKeyListTotal = (version: number) => `expense:listTotal:v${version}`
    static getKeyListAverages = (version: number) => `expense:listAverages:v${version}`

    static SECONDS_30 = 30_000;
    static WEEK = 7 * 24 * 60 * 60 * 1000;
} 
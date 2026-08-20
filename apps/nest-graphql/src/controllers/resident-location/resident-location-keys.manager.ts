export class ResidentLocationServiceKeysManager {
    static getKeyVersion = (userId: string) => `resident-location:version:${userId}`;
    static getKeyCount = (version: number, userId: string) => `resident-location:count:v${version}:${userId}`;
    static getKeyList = (limit: number, offset: number, version: number, userId: string) => `${limit}:${offset}:resident-location:list:v${version}:${userId}`;

    static SECONDS_30 = 30_000;
    static WEEK = 7 * 24 * 60 * 60 * 1000;
}

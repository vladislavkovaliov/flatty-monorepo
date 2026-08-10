export class CategoryServiceKeysManager {
    static getKeyVersion = () => `category:version`;
    static getKeyCount = (version: number) => `category:count:v${version}`;
    static getKeyList = (limit: number, offset: number, version: number) => `${limit}:${offset}:category:list:v${version}`;

    static SECONDS_30 = 30_000;
    static WEEK = 7 * 24 * 60 * 60 * 1000;
}
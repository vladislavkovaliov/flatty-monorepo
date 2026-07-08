export const reactApp = () => {
    return {
        bundleName: "app",
        cssBundleName: "styles",
        remoteOrigin: "http://localhost:8080",
        proxyBasePath: "/external-app",
        basePath: "/",
    };
}
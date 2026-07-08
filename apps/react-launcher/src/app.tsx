import { MicrofrontendHost } from './components/MicrofrontendHost';
import { reactSettings } from './applications/react-app';
import { angularApp } from './applications/angular-app';

const settingsConfig = reactSettings();
const angularConfig = angularApp();

export function App() {
  return (
    <div>
      <h1>Flatty Budget — Launcher</h1>
      <div style={{ display: 'flex', gap: 16 }}>
        <MicrofrontendHost
          bundleName={settingsConfig.bundleName}
          remoteOrigin={settingsConfig.remoteOrigin}
          proxyBasePath={settingsConfig.proxyBasePath}
        />
        <MicrofrontendHost
          bundleName={angularConfig.bundleName}
          remoteOrigin={angularConfig.remoteOrigin}
          proxyBasePath={angularConfig.proxyBasePath}
        />
      </div>
    </div>
  );
}

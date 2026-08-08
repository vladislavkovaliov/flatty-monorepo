import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { useEffect } from 'react';
import { useApplications, type UseApplicationsResult } from './applications';

function Harness({ onResult }: { onResult: (r: UseApplicationsResult) => void }) {
  const result = useApplications();
  useEffect(() => {
    onResult(result);
  }, [result, onResult]);
  return null;
}

let root: Root;
let fetchMock: ReturnType<typeof vi.fn>;

beforeEach(() => {
  document.body.innerHTML = '<div id="root"></div>';
  root = createRoot(document.getElementById('root')!);
  fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
});

afterEach(async () => {
  await act(async () => {
    root.unmount();
  });
  vi.unstubAllGlobals();
});

function renderHarness(): () => UseApplicationsResult | null {
  let latest: UseApplicationsResult | null = null;
  void act(() => {
    root.render(<Harness onResult={(r) => { latest = r; }} />);
  });
  return () => latest;
}

/** Flush fetch promise chains + React state updates. */
async function flush() {
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 0));
  });
}

describe('useApplications', () => {
  it('starts loading and maps registry rows to ApplicationConfig on success', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      json: async () => ({
        data: [
          {
            id: 1,
            name: 'settings',
            env: 'development',
            bundle_js: 'settings',
            style_url: 'styles',
            remote_origin: 'http://localhost:8081',
            proxy_base_path: '/external-settings',
            base_path: '/',
          },
        ],
      }),
    });

    const getResult = renderHarness();
    expect(getResult()?.loading).toBe(true);

    await flush();

    expect(getResult()?.loading).toBe(false);
    expect(getResult()?.error).toBeNull();
    expect(getResult()?.data).toEqual([
      {
        name: 'settings',
        bundleJs: 'settings',
        styleUrl: 'styles',
        remoteOrigin: 'http://localhost:8081',
        proxyBasePath: '/external-settings',
        basePath: '/',
      },
    ]);
  });

  it('sets error and keeps data null when the request fails', async () => {
    fetchMock.mockResolvedValue({ ok: false, status: 500, json: async () => ({}) });

    const getResult = renderHarness();

    await flush();

    expect(getResult()?.loading).toBe(false);
    expect(getResult()?.data).toBeNull();
    expect(getResult()?.error?.message).toContain('HTTP 500');
  });

  it('sets error on network failure', async () => {
    fetchMock.mockRejectedValue(new Error('network down'));

    const getResult = renderHarness();

    await flush();

    expect(getResult()?.loading).toBe(false);
    expect(getResult()?.data).toBeNull();
    expect(getResult()?.error?.message).toBe('network down');
  });
});

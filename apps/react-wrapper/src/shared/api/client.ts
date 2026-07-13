function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function getItem<T>(key: string): Promise<T | null> {
  await delay(300);
  try {
    const raw = localStorage.getItem(key);
    return raw ? (JSON.parse(raw) as T) : null;
  } catch {
    return null;
  }
}

async function setItem<T>(key: string, value: T): Promise<T> {
  await delay(300);
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch {
    console.warn('Failed to write to localStorage:', key);
  }
  return value;
}

export const apiClient = {
  get: <T>(key: string): Promise<T | null> => getItem<T>(key),
  set: <T>(key: string, value: T): Promise<T> => setItem<T>(key, value),
};

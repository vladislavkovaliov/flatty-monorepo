export interface ILoadScriptProps {
  src: string;
  params?: Record<string, unknown> & {
    defer?: boolean;
    async?: boolean;
  };
}

export class LoaderUtils {
  static isScriptLoaded(src: string): boolean {
    const scripts = Array.from(document.head.querySelectorAll('script'));

    return scripts.some((script) => script.src === src);
  }

  static isStyleLoaded(href: string): boolean {
    const links = Array.from(document.head.querySelectorAll<HTMLLinkElement>('link[rel="stylesheet"]'));

    return links.some((link) => link.href === href);
  }

  static loadScript({ src, params: { defer = false, async = true, ...otherProps } = {} }: ILoadScriptProps): Promise<void> {
    if (LoaderUtils.isScriptLoaded(src)) {
      return Promise.resolve();
    }

    return new Promise((resolve, reject) => {
      const script = document.createElement('script');

      script.src = src;
      script.defer = defer;
      script.async = async;

      Object.assign(script, otherProps);

      if (!script.type) {
        script.type = 'text/javascript';
      }

      script.addEventListener('load', () => resolve());

      script.addEventListener('error', (event) => {
        const ev = event as ErrorEvent;

        const cause = ev.error ?? ev.message;

        reject(
          new Error(`Failed to load script: ${src}${cause ? ` (${String(cause)})` : ''}`, {
            cause: event,
          }),
        );
      });

      document.head.appendChild(script);
    });
  }

  static loadStyle(href: string): Promise<void> {
    if (LoaderUtils.isStyleLoaded(href)) {
      return Promise.resolve();
    }

    return new Promise((resolve, reject) => {
      const link = document.createElement('link');

      link.rel = 'stylesheet';
      link.href = href;
      link.type = 'text/css';
      link.media = 'all';

      link.addEventListener('load', () => {
        resolve();
      });

      link.addEventListener('error', (event) => {
        reject(new Error(`Failed to load stylesheet: ${href}`, { cause: event }));
      });

      document.head.appendChild(link);
    });
  }
}

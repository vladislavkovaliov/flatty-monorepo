import type { ComponentType, FunctionComponent } from "react";

import { type IAppConfig, type IAppComponent } from "../contracts/app-config.type";
import { createRoot, type Root } from "react-dom/client";
import { flushSync } from "react-dom";
import { sanitizeConfig } from "./sanitize-config";

export class AppComponent implements IAppComponent {
    AppComponent: ComponentType<IAppConfig> | FunctionComponent<IAppConfig>;

    root: Root | null = null;

    constructor(
        AppComponent: ComponentType<IAppConfig> | FunctionComponent<IAppConfig>,
    ) {
        this.AppComponent = AppComponent;
    }

    initialize(element: Element, config: IAppConfig) {
        element.classList.add('react-app-root');
        this.root = createRoot(element);

        const safeConfig = sanitizeConfig(config);

        flushSync(() => {
            this.root!.render(<this.AppComponent {...safeConfig} />);
        });
    }

    destroy(): void {
        if (this.root) {
            flushSync(() => {
                this.root!.unmount();
            });
        }
    }
}

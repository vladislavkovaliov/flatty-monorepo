import type {ComponentType, FunctionComponent} from "react";

import {type IAppConfig as Config, type IAppComponent} from "../../types/external-app-config.type"
import {createRoot, type Root} from "react-dom/client";
import {sanitizeConfig} from "../../lib/utils";

export interface IAppConfig extends Config {
    countryCode?: string;
}

export class AppComponent implements IAppComponent {
    AppComponent: ComponentType<IAppConfig> | FunctionComponent<IAppConfig>;

    root: Root | null = null;

    constructor(
        AppComponent: ComponentType<IAppConfig> | FunctionComponent<IAppConfig>,
    ) {
        // super();

        this.AppComponent = AppComponent;
    }

    initialize(element: Element, config: IAppConfig) {
        element.classList.add('react-app-root');
        this.root = createRoot(element);

        const safeConfig = sanitizeConfig(config);

        this.root.render(<this.AppComponent {...safeConfig} />);
    }

    destroy(): void {
        if (this.root) {
            this.root.unmount();
        }
    }
}
import type { DtoUserSettingsResponse } from "@flatty-budget/sdk";
import { createContext } from "react";

export type UserSettings = {
    dateFormat?: DtoUserSettingsResponse["date_format"];
    language?: DtoUserSettingsResponse["language"];
    theme?: DtoUserSettingsResponse["theme"];
    timezone?: DtoUserSettingsResponse["timezone"];
};

interface UserSettingsContextValue {
    settings?: UserSettings;
    isLoading: boolean;
    isError: boolean;
    isPendingMutate: boolean;
    isSuccessMutate: boolean;
    isErrorMutate: boolean;
    error: Error | null;
    updateTimezone: (dateFormat: string, timezone: string) => void;
    updateLanguage: (language: string) => void;
    updateTheme: (theme: string) => void;
}

export const UserSettingsContext = createContext<UserSettingsContextValue | null>(null);

import { createContext } from "react";

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

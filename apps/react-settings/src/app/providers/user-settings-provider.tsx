import { createContext, useContext, useMemo } from "react";
import { type ReactNode } from 'react';

import { useUserSettings, type DtoUserSettingsResponse, useUpdateUserSettings } from "@flatty-budget/sdk"

type UserSettings = {
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

const UserSettingsContext = createContext<UserSettingsContextValue | null>(null);

interface UserSettingsProviderProps {
      children: ReactNode;
}

export function UserSettingsProvider({ children }: UserSettingsProviderProps) {
    const { data, isError, isLoading } = useUserSettings();
    const { 
        mutate, 
        isPending: isPendingMutate,
        isSuccess: isSuccessMutate,
        isError: isErrorMutate,
        error, 
    } = useUpdateUserSettings()

    const updateTimezone = (dateFormat: UserSettings["dateFormat"], timezone: UserSettings['timezone']) => {
        mutate({ data: { dateFormat: dateFormat, timezone: timezone } })
    };

    const updateLanguage = (lang: UserSettings['language']) => {
        mutate({ data: { language: lang } })
    };

    const updateTheme = (theme: UserSettings['theme']) => {
        mutate({ data: { theme: theme } })
    };

    const value = useMemo(
        () => {
            return {
                settings: data
                    ? {
                        dateFormat: data.date_format,
                        language: data.language,
                        theme: data.theme,
                        timezone: data.timezone,
                    }
                    : undefined,
                isLoading,
                isError,
                isPendingMutate,
                isSuccessMutate,
                isErrorMutate,
                error,
                updateTimezone,
                updateLanguage,
                updateTheme,
            };
        },
        [data, isLoading, isError]
    );

    

    return (
        <UserSettingsContext value={value}>
            {children}
        </UserSettingsContext>
    )
}

export function useUserSettingsContext() {
    const context = useContext(UserSettingsContext);

    if (!context) {
        throw new Error("useUserSettingsContext must be used within UserSettingsProvider");
    }

    return context;
}
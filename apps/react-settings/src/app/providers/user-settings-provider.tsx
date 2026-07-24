import { useCallback, useMemo } from "react";
import { type ReactNode } from 'react';

import { useUserSettings, useUpdateUserSettings } from "@flatty-budget/sdk"
import { UserSettingsContext, type UserSettings } from "./user-settings-context";

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

    const updateTimezone = useCallback((dateFormat: UserSettings["dateFormat"], timezone: UserSettings['timezone']) => {
        mutate({ data: { dateFormat: dateFormat, timezone: timezone } })
    }, [mutate]);

    const updateLanguage = useCallback((lang: UserSettings['language']) => {
        mutate({ data: { language: lang } })
    }, [mutate]);

    const updateTheme = useCallback((theme: UserSettings['theme']) => {
        mutate({ data: { theme: theme } })
    }, [mutate]);

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
        [data, isLoading, isError, isErrorMutate, isPendingMutate, isSuccessMutate, error, updateLanguage, updateTheme, updateTimezone]
    );

    

    return (
        <UserSettingsContext value={value}>
            {children}
        </UserSettingsContext>
    )
}

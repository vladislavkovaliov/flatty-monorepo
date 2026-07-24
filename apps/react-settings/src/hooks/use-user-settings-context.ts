import { useContext } from "react";
import { UserSettingsContext } from "../app/providers/user-settings-provider";

export function useUserSettingsContext() {
    const context = useContext(UserSettingsContext);

    if (!context) {
        throw new Error("useUserSettingsContext must be used within UserSettingsProvider");
    }

    return context;
}
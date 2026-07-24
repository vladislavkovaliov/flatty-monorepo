import { useContext } from "react";
import { UserSettingsContext } from "../app/providers/user-settings-context";

export function useUserSettingsContext() {
    const context = useContext(UserSettingsContext);

    if (!context) {
        throw new Error("useUserSettingsContext must be used within UserSettingsProvider");
    }

    return context;
}
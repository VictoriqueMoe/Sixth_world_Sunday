import { createContext } from "react";

interface ThemeContextValue {
    wideLayout: boolean;
    setWideLayout: (enabled: boolean) => void;
}

export const ThemeContext = createContext<ThemeContextValue | null>(null);

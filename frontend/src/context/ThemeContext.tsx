import { type PropsWithChildren, useCallback, useEffect, useLayoutEffect, useState } from "react";
import { useAuth } from "../hooks/useAuth";
import { useUpdateAppearance } from "../api/mutations/auth";
import { ThemeContext } from "./themeContextValue";

const WIDE_LAYOUT_KEY = "ut-wide-layout";

function getStoredWideLayout(): boolean {
    try {
        const stored = localStorage.getItem(WIDE_LAYOUT_KEY);
        if (stored !== null) {
            return stored === "true";
        }
    } catch {}
    return false;
}

export function ThemeProvider({ children }: PropsWithChildren) {
    const { user } = useAuth();
    const [override, setOverride] = useState<{ userId: string | null; wideLayout: boolean | null }>(() => ({
        userId: null,
        wideLayout: getStoredWideLayout(),
    }));

    const activeUserId = user?.id ?? null;
    const activeOverride = override.userId === activeUserId ? override : null;

    const userWideLayout = typeof user?.private?.wide_layout === "boolean" ? user.private.wide_layout : null;

    const wideLayout: boolean = activeOverride?.wideLayout ?? userWideLayout ?? getStoredWideLayout();

    useEffect(() => {
        try {
            localStorage.setItem(WIDE_LAYOUT_KEY, String(wideLayout));
        } catch {}
    }, [wideLayout]);

    useLayoutEffect(() => {
        if (wideLayout) {
            document.documentElement.setAttribute("data-width", "wide");
        } else {
            document.documentElement.removeAttribute("data-width");
        }
    }, [wideLayout]);

    const updateAppearanceMutation = useUpdateAppearance();

    const setWideLayout = useCallback(
        (enabled: boolean) => {
            setOverride({ userId: activeUserId, wideLayout: enabled });

            if (user) {
                updateAppearanceMutation.mutate({ wideLayout: enabled });
            }
        },
        [activeUserId, user, updateAppearanceMutation],
    );

    return <ThemeContext.Provider value={{ wideLayout, setWideLayout }}>{children}</ThemeContext.Provider>;
}

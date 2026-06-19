import { useEffect, useState } from "react";

const MOBILE_QUERY = "(max-width: 960px)";

function readIsMobile(): boolean {
    if (typeof window === "undefined") {
        return false;
    }

    return window.matchMedia(MOBILE_QUERY).matches;
}

export function useIsMobile(): boolean {
    const [isMobile, setIsMobile] = useState(readIsMobile);

    useEffect(() => {
        const mql = window.matchMedia(MOBILE_QUERY);

        function onChange() {
            setIsMobile(mql.matches);
        }

        mql.addEventListener("change", onChange);
        return () => {
            mql.removeEventListener("change", onChange);
        };
    }, []);

    return isMobile;
}

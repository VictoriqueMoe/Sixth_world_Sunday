import { createContext, useContext } from "react";

interface MobileNavValue {
    openNav: () => void;
}

export const MobileNavContext = createContext<MobileNavValue>({ openNav: () => {} });

export function useMobileNav(): MobileNavValue {
    return useContext(MobileNavContext);
}

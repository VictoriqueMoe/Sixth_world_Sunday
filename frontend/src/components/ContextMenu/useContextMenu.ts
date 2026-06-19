import { useCallback, useState } from "react";
import type { ContextMenuItem, ContextMenuState } from "./ContextMenu";

interface ContextMenuTrigger {
    preventDefault: () => void;
    clientX: number;
    clientY: number;
}

export function useContextMenu() {
    const [state, setState] = useState<ContextMenuState | null>(null);

    const open = useCallback((event: ContextMenuTrigger, items: ContextMenuItem[]) => {
        event.preventDefault();
        setState({ x: event.clientX, y: event.clientY, items });
    }, []);

    const close = useCallback(() => {
        setState(null);
    }, []);

    return { state, open, close };
}

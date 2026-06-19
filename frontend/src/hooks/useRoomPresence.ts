import { useCallback, useState } from "react";
import type { ChatRoomMember } from "../types/api";
import { usePresenceReporter } from "./usePresenceReporter";

interface UseRoomPresenceArgs {
    roomId: string | undefined;
    baseMembers: ChatRoomMember[];
    sendWSMessage: (msg: object) => void;
    wsEpoch: number;
}

export function useRoomPresence({ roomId, baseMembers, sendWSMessage, wsEpoch }: UseRoomPresenceArgs) {
    const [presenceState, setPresenceState] = useState<{
        roomId: string | null;
        map: Record<string, "active" | "idle">;
    }>({
        roomId: null,
        map: {},
    });

    const presenceMap = presenceState.roomId === roomId ? presenceState.map : {};
    const setPresenceMap = useCallback(
        (updater: (prev: Record<string, "active" | "idle">) => Record<string, "active" | "idle">) => {
            setPresenceState(prev => {
                const base = prev.roomId === roomId ? prev.map : {};
                return { roomId: roomId ?? null, map: updater(base) };
            });
        },
        [roomId],
    );

    const [onlineOverrides, setOnlineOverrides] = useState<Record<string, boolean>>({});
    const setOnline = useCallback((id: string, online: boolean) => {
        setOnlineOverrides(prev => {
            if (prev[id] === online) {
                return prev;
            }
            return { ...prev, [id]: online };
        });
    }, []);

    usePresenceReporter({ roomId, sendWSMessage, wsEpoch });

    const presenceSeed: Record<string, "active" | "idle"> = {};
    const onlineSeed: Record<string, boolean> = {};
    for (const m of baseMembers) {
        if (m.presence === "active" || m.presence === "idle") {
            presenceSeed[m.user.id] = m.presence;
        }
        if (m.online) {
            onlineSeed[m.user.id] = true;
        }
    }
    const presenceMapMerged = { ...presenceSeed, ...presenceMap };

    const isOnline = (id: string): boolean => {
        const override = onlineOverrides[id];
        if (override !== undefined) {
            return override;
        }
        const p = presenceMapMerged[id];
        if (p === "active" || p === "idle") {
            return true;
        }
        return !!onlineSeed[id];
    };

    const memberOnlineWeight = (id: string) => (isOnline(id) ? 0 : 1);

    return { presenceMap, setPresenceMap, presenceMapMerged, memberOnlineWeight, isOnline, setOnline };
}

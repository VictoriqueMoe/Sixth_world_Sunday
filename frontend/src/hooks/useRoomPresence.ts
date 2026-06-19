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

    usePresenceReporter({ roomId, sendWSMessage, wsEpoch });

    const presenceSeed: Record<string, "active" | "idle"> = {};
    for (const m of baseMembers) {
        if (m.presence === "active" || m.presence === "idle") {
            presenceSeed[m.user.id] = m.presence;
        }
    }
    const presenceMapMerged = { ...presenceSeed, ...presenceMap };

    const memberOnlineWeight = (id: string) => {
        const p = presenceMapMerged[id];
        if (p === "active" || p === "idle") {
            return 0;
        }
        return 1;
    };

    return { presenceMap, setPresenceMap, presenceMapMerged, memberOnlineWeight };
}

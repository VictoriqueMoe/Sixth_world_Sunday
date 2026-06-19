import { createContext, useContext } from "react";
import type { Room } from "livekit-client";

export type VoiceStatus = "idle" | "connecting" | "connected";

export interface VoiceContextValue {
    status: VoiceStatus;
    activeRoomId: string | null;
    activeRoomName: string;
    room: Room | null;
    presence: Record<string, string[]>;
    join: (roomId: string, roomName: string) => void;
    leave: () => void;
}

export const VoiceContext = createContext<VoiceContextValue | null>(null);

export function useVoice(): VoiceContextValue {
    const ctx = useContext(VoiceContext);
    if (!ctx) {
        throw new Error("useVoice must be used within a VoiceProvider");
    }
    return ctx;
}

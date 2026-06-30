import type { ChatRoom } from "../types/api";
import { useVoice } from "../context/voiceContextValue";

interface UseRoomVoiceArgs {
    roomId: string | undefined;
    room: ChatRoom | null;
}

export function useRoomVoice({ roomId, room }: UseRoomVoiceArgs) {
    const voiceCtx = useVoice();
    const voiceActiveHere = !!roomId && voiceCtx.activeRoomId === roomId;
    const voiceParticipantIds = roomId ? (voiceCtx.presence[roomId] ?? room?.voice_participants ?? []) : [];
    const voice = {
        status: voiceActiveHere ? voiceCtx.status : ("idle" as const),
        room: voiceActiveHere ? voiceCtx.room : null,
        participantIds: voiceParticipantIds,
        presenceCount: voiceParticipantIds.length,
        join: () => voiceCtx.join(roomId ?? "", room?.name ?? ""),
        leave: voiceCtx.leave,
    };
    const voiceIdSet = new Set(voice.participantIds);

    const isVoiceChannel = room?.channel_kind === "voice";

    return { voice, isVoiceChannel, voiceIdSet };
}

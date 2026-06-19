import { type Dispatch, type SetStateAction, useCallback, useState } from "react";
import type { ChatRoom, ChatRoomMember } from "../types/api";
import { useChatRoomMembers, useChannels } from "../api/queries/chat";

export function useRoomData(roomId: string | undefined) {
    const [roomOverride, setRoomOverride] = useState<{ roomId: string | null; room: ChatRoom | null }>({
        roomId: null,
        room: null,
    });
    const [membersOverride, setMembersOverride] = useState<{ roomId: string | null; members: ChatRoomMember[] | null }>(
        {
            roomId: null,
            members: null,
        },
    );

    const userRoomsQuery = useChannels();
    const userRoomsLoading = userRoomsQuery.loading;
    const userRoomsList = userRoomsQuery.rooms;
    const baseRoom = roomId ? (userRoomsList.find(r => r.id === roomId) ?? null) : null;
    const room = roomOverride.roomId === roomId && roomOverride.room ? roomOverride.room : baseRoom;

    const loading = !!roomId && userRoomsLoading;

    const setRoom: Dispatch<SetStateAction<ChatRoom | null>> = useCallback(
        updater => {
            setRoomOverride(prev => {
                const baseValue = prev.roomId === roomId && prev.room ? prev.room : null;
                const next =
                    typeof updater === "function"
                        ? (updater as (p: ChatRoom | null) => ChatRoom | null)(baseValue)
                        : updater;
                return { roomId: roomId ?? null, room: next };
            });
        },
        [roomId],
    );

    const membersQuery = useChatRoomMembers(roomId ?? "", !!roomId && !!room);
    const membersRefresh = membersQuery.refresh;
    const baseMembers = membersQuery.members;
    const members =
        membersOverride.roomId === roomId && membersOverride.members ? membersOverride.members : baseMembers;
    const setMembers: Dispatch<SetStateAction<ChatRoomMember[]>> = useCallback(
        updater => {
            setMembersOverride(prev => {
                const baseValue = prev.roomId === roomId && prev.members ? prev.members : baseMembers;
                const next =
                    typeof updater === "function"
                        ? (updater as (p: ChatRoomMember[]) => ChatRoomMember[])(baseValue)
                        : updater;
                return { roomId: roomId ?? null, members: next };
            });
        },
        [roomId, baseMembers],
    );

    const loadMembers = useCallback(() => {
        membersRefresh();
    }, [membersRefresh]);

    return { room, setRoom, members, setMembers, baseMembers, loadMembers, loading };
}

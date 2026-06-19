import { useQuery } from "@tanstack/react-query";
import {
    getChatCategories,
    getChatRoomMembers,
    getChatRoomPinnedMessages,
    getRoomMessages,
    getRoomMessagesBefore,
    getUserRooms,
    listChatRoomBans,
    listChatRoomBannedWords,
} from "../endpoints";
import { queryKeys } from "../queryKeys";

export function fetchRoomMessages(roomId: string, limit?: number, offset?: number) {
    return getRoomMessages(roomId, limit, offset);
}

export function fetchRoomMessagesBefore(roomId: string, beforeCursor: string, limit?: number) {
    return getRoomMessagesBefore(roomId, beforeCursor, limit);
}

export function fetchUserRooms() {
    return getUserRooms();
}

export function useChannels() {
    const query = useQuery({
        queryKey: ["channels"],
        queryFn: () => getUserRooms(),
    });
    return { rooms: query.data?.rooms ?? [], loading: query.isLoading, refresh: query.refetch };
}

export function useChatCategories() {
    const query = useQuery({
        queryKey: ["chat-categories"],
        queryFn: () => getChatCategories(),
    });
    return { categories: query.data?.categories ?? [], loading: query.isLoading };
}

export function useChatRoomMembers(roomId: string, enabled = true) {
    const query = useQuery({
        queryKey: queryKeys.chat.roomMembers(roomId),
        queryFn: () => getChatRoomMembers(roomId),
        enabled: enabled && !!roomId,
    });
    return { members: query.data?.members ?? [], loading: query.isLoading, refresh: query.refetch };
}

export function useChatRoomBans(roomId: string, enabled = true) {
    const query = useQuery({
        queryKey: ["chat", "rooms", roomId, "bans"],
        queryFn: () => listChatRoomBans(roomId),
        enabled: enabled && !!roomId,
    });
    return { bans: query.data?.bans ?? [], loading: query.isLoading, refresh: query.refetch };
}

export function useChatRoomBannedWords(roomId: string, enabled = true) {
    const query = useQuery({
        queryKey: ["chat", "rooms", roomId, "banned-words"],
        queryFn: () => listChatRoomBannedWords(roomId),
        enabled: enabled && !!roomId,
    });
    return { rules: query.data?.rules ?? [], loading: query.isLoading, refresh: query.refetch };
}

export function useChatRoomPinnedMessages(roomId: string, enabled = true) {
    const query = useQuery({
        queryKey: queryKeys.chat.pinned(roomId),
        queryFn: () => getChatRoomPinnedMessages(roomId),
        enabled: enabled && !!roomId,
    });
    return { messages: query.data?.messages ?? [], loading: query.isLoading, refresh: query.refetch };
}

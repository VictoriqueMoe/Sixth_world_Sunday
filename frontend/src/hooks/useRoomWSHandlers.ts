import { type Dispatch, type SetStateAction, useEffect, useRef } from "react";
import { type NavigateFunction } from "react-router";
import type { ChatMessage, ChatRoom, ChatRoomMember, User, UserProfile, WSMessage } from "../types/api";
import { type SiteRole } from "../utils/permissions";
import { useMarkChatRoomRead } from "../api/mutations/chat";
import {
    applyChatMemberUpdate,
    applyChatMessagePinned,
    applyChatMessageUnpinned,
    applyReactionAdded,
    applyReactionRemoved,
    applySharedChatWSBranch,
    ChatMemberUpdatedPayload,
    ChatMessagePinnedPayload,
    ChatMessageUnpinnedPayload,
    ChatReactionPayload,
    handleIncomingChatMessage,
    maybePlayChatMessageSound,
} from "../utils/chatStream";

interface UseRoomWSHandlersArgs {
    user: UserProfile | null;
    roomId: string | undefined;
    room: ChatRoom | null;
    addWSListener: (cb: (msg: WSMessage) => void) => () => void;
    sendWSMessage: (msg: object) => void;
    wsEpoch: number;
    setRoom: Dispatch<SetStateAction<ChatRoom | null>>;
    setMembers: Dispatch<SetStateAction<ChatRoomMember[]>>;
    loadMembers: () => void;
    setMessages: Dispatch<SetStateAction<ChatMessage[]>>;
    scrollToBottom: (opts?: { force?: boolean }) => void;
    setPresenceMap: (updater: (prev: Record<string, "active" | "idle">) => Record<string, "active" | "idle">) => void;
    setOnline: (userId: string, online: boolean) => void;
    noteTyping: (userId: string) => void;
    clearTypingUser: (userId: string) => void;
    navigate: NavigateFunction;
    setToast: (msg: string | null) => void;
    setPinnedRefreshKey: Dispatch<SetStateAction<number>>;
}

export function useRoomWSHandlers({
    user,
    roomId,
    room,
    addWSListener,
    sendWSMessage,
    wsEpoch,
    setRoom,
    setMembers,
    loadMembers,
    setMessages,
    scrollToBottom,
    setPresenceMap,
    setOnline,
    noteTyping,
    clearTypingUser,
    navigate,
    setToast,
    setPinnedRefreshKey,
}: UseRoomWSHandlersArgs) {
    const roomIdRef = useRef(roomId);
    const roomMutedRef = useRef(false);

    useEffect(() => {
        roomIdRef.current = roomId;
    }, [roomId]);

    useEffect(() => {
        roomMutedRef.current = room?.viewer_muted ?? false;
    }, [room?.viewer_muted]);

    const markReadMutation = useMarkChatRoomRead();
    const markRead = markReadMutation.mutate;

    useEffect(() => {
        if (!roomId || !room) {
            return;
        }
        markRead(roomId);
    }, [roomId, room, markRead]);

    useEffect(() => {
        if (!roomId) {
            return;
        }
        sendWSMessage({ type: "join_room", data: { room_id: roomId } });
        return () => {
            sendWSMessage({ type: "leave_room", data: { room_id: roomId } });
        };
    }, [roomId, sendWSMessage, wsEpoch]);

    useEffect(() => {
        if (!user) {
            return;
        }
        return addWSListener((msg: WSMessage) => {
            if (msg.type === "chat_message") {
                const chatMsg = msg.data as ChatMessage;
                const added = handleIncomingChatMessage(
                    chatMsg,
                    roomIdRef.current ?? null,
                    setMessages,
                    scrollToBottom,
                );
                if (added && user) {
                    maybePlayChatMessageSound({
                        senderId: chatMsg.sender.id,
                        currentUserId: user.id,
                        roomMuted: roomMutedRef.current,
                        enabled: user.private?.play_message_sound ?? true,
                    });
                }
                return;
            }
            if (msg.type === "chat_member_joined") {
                const data = msg.data as { room_id: string; user: User };
                if (data.room_id !== roomIdRef.current) {
                    return;
                }
                loadMembers();
                setRoom(prev => {
                    if (!prev) {
                        return prev;
                    }
                    return {
                        ...prev,
                        member_count: (prev.member_count ?? prev.members.length) + 1,
                    };
                });
                return;
            }
            if (msg.type === "chat_member_left") {
                const data = msg.data as { room_id: string; user_id: string };
                if (data.room_id !== roomIdRef.current) {
                    return;
                }
                setMembers(prev => prev.filter(m => m.user.id !== data.user_id));
                setRoom(prev => {
                    if (!prev) {
                        return prev;
                    }
                    return {
                        ...prev,
                        member_count: Math.max(0, (prev.member_count ?? prev.members.length) - 1),
                    };
                });
                return;
            }
            if (msg.type === "chat_kicked") {
                const data = msg.data as { room_id: string; reason?: string };
                if (data.room_id !== roomIdRef.current) {
                    return;
                }
                const baseMessage = "You were removed from this room";
                setToast(data.reason ? `${baseMessage}: ${data.reason}` : baseMessage);
                setTimeout(() => navigate("/channels"), 1500);
                return;
            }
            if (msg.type === "chat_room_deleted" || msg.type === "channel_deleted") {
                const data = msg.data as { room_id: string };
                if (data.room_id !== roomIdRef.current) {
                    return;
                }
                setToast("This channel was deleted");
                setTimeout(() => navigate("/channels"), 1500);
                return;
            }
            if (msg.type === "chat_truncated") {
                const data = msg.data as { room_id: string };
                if (data.room_id === roomIdRef.current) {
                    setMessages([]);
                }
                return;
            }
            if (msg.type === "channel_updated") {
                const data = msg.data as ChatRoom;
                if (data.id !== roomIdRef.current) {
                    return;
                }
                setRoom(prev => {
                    if (!prev) {
                        return prev;
                    }
                    return {
                        ...prev,
                        name: data.name,
                        description: data.description,
                        category_id: data.category_id,
                        position: data.position,
                    };
                });
                return;
            }
            if (msg.type === "chat_member_updated") {
                const data = msg.data as ChatMemberUpdatedPayload;
                if (data.room_id !== roomIdRef.current) {
                    return;
                }
                applyChatMemberUpdate(data, setMembers, setMessages);
                return;
            }
            if (msg.type === "chat_message_pinned") {
                const data = msg.data as ChatMessagePinnedPayload;
                if (data.room_id !== roomIdRef.current) {
                    return;
                }
                applyChatMessagePinned(data, setMessages);
                setPinnedRefreshKey(k => k + 1);
                return;
            }
            if (msg.type === "chat_message_unpinned") {
                const data = msg.data as ChatMessageUnpinnedPayload;
                if (data.room_id !== roomIdRef.current) {
                    return;
                }
                applyChatMessageUnpinned(data, setMessages);
                setPinnedRefreshKey(k => k + 1);
                return;
            }
            if (msg.type === "chat_reaction_added") {
                const data = msg.data as ChatReactionPayload;
                if (data.room_id !== roomIdRef.current) {
                    return;
                }
                applyReactionAdded(data, user.id, setMessages);
                return;
            }
            if (msg.type === "chat_reaction_removed") {
                const data = msg.data as ChatReactionPayload;
                if (data.room_id !== roomIdRef.current) {
                    return;
                }
                applyReactionRemoved(data, user.id, setMessages);
                return;
            }
            if (msg.type === "chat_presence_changed") {
                const data = msg.data as { room_id: string; user_id: string; state: string };
                if (data.room_id !== roomIdRef.current) {
                    return;
                }
                setPresenceMap(prev => {
                    const next = { ...prev };
                    if (data.state === "active" || data.state === "idle") {
                        next[data.user_id] = data.state;
                    } else {
                        delete next[data.user_id];
                    }
                    return next;
                });
                return;
            }
            if (msg.type === "presence_online") {
                const data = msg.data as { user_id: string };
                setOnline(data.user_id, true);
                return;
            }
            if (msg.type === "presence_offline") {
                const data = msg.data as { user_id: string };
                setOnline(data.user_id, false);
                return;
            }
            if (
                applySharedChatWSBranch(msg, {
                    activeRoomId: roomIdRef.current ?? null,
                    setMessages,
                    noteTyping,
                })
            ) {
                return;
            }
            if (msg.type === "chat_message") {
                const chatMsg = msg.data as ChatMessage;
                if (chatMsg.room_id === roomIdRef.current) {
                    clearTypingUser(chatMsg.sender.id);
                }
            }
            if (msg.type === "role_changed") {
                const data = msg.data as { user_id?: string; role?: string };
                if (!data.user_id) {
                    return;
                }
                const newRole = (data.role ?? "") as SiteRole;
                setMembers(prev =>
                    prev.map(m => {
                        if (m.user.id !== data.user_id) {
                            return m;
                        }
                        return { ...m, user: { ...m.user, role: newRole || undefined } };
                    }),
                );
                setMessages(prev =>
                    prev.map(m => {
                        if (m.sender.id !== data.user_id) {
                            return m;
                        }
                        return { ...m, sender: { ...m.sender, role: newRole || undefined } };
                    }),
                );
            }
        });
    }, [
        user,
        addWSListener,
        scrollToBottom,
        setMessages,
        navigate,
        loadMembers,
        noteTyping,
        clearTypingUser,
        setMembers,
        setPresenceMap,
        setOnline,
        setRoom,
        setToast,
        setPinnedRefreshKey,
    ]);
}

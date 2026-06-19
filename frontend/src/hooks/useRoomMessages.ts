import { useEffect, useRef, useState } from "react";
import { useLocation } from "react-router";
import type { ChatMessage, ChatRoom, UserProfile } from "../types/api";
import { useMessageHistory } from "./useMessageHistory";
import { useChatMessageHandlers } from "./useChatMessageHandlers";
import {
    useAddChatMessageReaction,
    usePinChatMessage,
    useRemoveChatMessageReaction,
    useUnpinChatMessage,
} from "../api/mutations/chat";

interface UseRoomMessagesArgs {
    roomId: string | undefined;
    room: ChatRoom | null;
    user: UserProfile | null;
    wsEpoch: number;
    viewerTimedOut: boolean;
    setToast: (msg: string | null) => void;
}

export function useRoomMessages({ roomId, room, user, wsEpoch, viewerTimedOut, setToast }: UseRoomMessagesArgs) {
    const location = useLocation();

    const {
        messages,
        setMessages,
        hasMore,
        loadingMore,
        containerRef: messagesContainerRef,
        contentRef: messagesContentRef,
        endRef: messagesEndRef,
        scrollToBottom,
        handleScroll: handleMessagesScroll,
        addMessage,
        loadUntilMessage,
        resync,
    } = useMessageHistory(room ? roomId : undefined);

    const [editingMessageId, setEditingMessageId] = useState<string | null>(null);
    const [highlightedMsgId, setHighlightedMsgId] = useState<string | null>(null);

    const didResyncMountRef = useRef(false);
    useEffect(() => {
        if (!didResyncMountRef.current) {
            didResyncMountRef.current = true;
            return;
        }

        resync().catch(() => {});
    }, [wsEpoch, resync]);

    const targetMsgId = location.hash.startsWith("#msg-") ? location.hash.slice(5) : null;
    const targetMsgCreatedAt = new URLSearchParams(location.search).get("at") ?? undefined;
    const [handledHash, setHandledHash] = useState<string | null>(null);
    const hashLoadRef = useRef<string | null>(null);
    const pendingTargetMsgId = targetMsgId && handledHash !== targetMsgId ? targetMsgId : null;

    useEffect(() => {
        if (!pendingTargetMsgId || messages.length === 0) {
            return;
        }
        if (!messages.some(m => m.id === pendingTargetMsgId)) {
            if (hashLoadRef.current !== pendingTargetMsgId) {
                hashLoadRef.current = pendingTargetMsgId;
                loadUntilMessage(pendingTargetMsgId, targetMsgCreatedAt).then(found => {
                    if (!found) {
                        setHandledHash(pendingTargetMsgId);
                    }
                });
            }
            return;
        }
        const t = setTimeout(() => {
            const el = document.getElementById(`chat-msg-${pendingTargetMsgId}`);
            if (el) {
                el.scrollIntoView({ behavior: "smooth", block: "center" });
                setHighlightedMsgId(pendingTargetMsgId);
                setHandledHash(pendingTargetMsgId);
            }
        }, 300);
        return () => clearTimeout(t);
    }, [pendingTargetMsgId, targetMsgCreatedAt, messages, loadUntilMessage]);

    useEffect(() => {
        if (!highlightedMsgId) {
            return;
        }
        const t = setTimeout(() => setHighlightedMsgId(null), 3000);
        return () => clearTimeout(t);
    }, [highlightedMsgId]);

    const { handleDeleteMessage, handleEditMessage, handleEditLast } = useChatMessageHandlers({
        user,
        messages,
        setMessages,
        setEditingMessageId,
        onError: (msg: string) => setToast(msg),
        editLastBlocked: viewerTimedOut,
    });

    const pinMutation = usePinChatMessage(roomId ?? undefined);
    const unpinMutation = useUnpinChatMessage(roomId ?? undefined);
    const addReactionMutation = useAddChatMessageReaction();
    const removeReactionMutation = useRemoveChatMessageReaction();

    function handleSentMessage(message: ChatMessage) {
        addMessage(message);
        scrollToBottom({ force: true });
    }

    async function handleReactionToggle(message: ChatMessage, emoji: string) {
        const existing = (message.reactions ?? []).find(r => r.emoji === emoji);
        try {
            if (existing && existing.viewer_reacted) {
                await removeReactionMutation.mutateAsync({ messageId: message.id, emoji });
            } else {
                await addReactionMutation.mutateAsync({ messageId: message.id, emoji });
            }
        } catch (err) {
            setToast(err instanceof Error ? err.message : "Failed to update reaction");
        }
    }

    async function handlePinToggle(message: ChatMessage) {
        try {
            if (message.pinned) {
                await unpinMutation.mutateAsync(message.id);
            } else {
                await pinMutation.mutateAsync(message.id);
            }
        } catch (err) {
            setToast(err instanceof Error ? err.message : "Failed to update pin");
        }
    }

    async function handleJumpToMessage(messageId: string, targetCreatedAt?: string) {
        const scrollToEl = (smooth: boolean) => {
            const el = document.getElementById(`chat-msg-${messageId}`);
            if (el) {
                el.scrollIntoView({ behavior: smooth ? "smooth" : "auto", block: "center" });
                setHighlightedMsgId(messageId);
            }
        };
        if (messages.some(m => m.id === messageId)) {
            scrollToEl(true);
            return;
        }
        const found = await loadUntilMessage(messageId, targetCreatedAt);
        if (!found) {
            setToast("Couldn't locate that message.");
            return;
        }
        requestAnimationFrame(() => scrollToEl(false));
        setTimeout(() => scrollToEl(false), 300);
        setTimeout(() => scrollToEl(true), 600);
    }

    return {
        messages,
        setMessages,
        hasMore,
        loadingMore,
        messagesContainerRef,
        messagesContentRef,
        messagesEndRef,
        scrollToBottom,
        handleMessagesScroll,
        addMessage,
        loadUntilMessage,
        resync,
        editingMessageId,
        setEditingMessageId,
        highlightedMsgId,
        handleSentMessage,
        handleReactionToggle,
        handlePinToggle,
        handleJumpToMessage,
        handleDeleteMessage,
        handleEditMessage,
        handleEditLast,
    };
}

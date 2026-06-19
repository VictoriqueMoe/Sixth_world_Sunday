import { type Dispatch, type SetStateAction, useState } from "react";
import { type NavigateFunction } from "react-router";
import type { ChatMessage, ChatRoom, ChatRoomMember } from "../types/api";
import { parseServerDate } from "../utils/time";
import { applyLocalMemberChange } from "../utils/chatStream";
import {
    useBanChatRoomMember,
    useClearChatRoomMemberTimeout,
    useDeleteChannel,
    useKickChatRoomMember,
    useSetChatRoomMemberNickname,
    useSetChatRoomMemberTimeout,
    useSetChatRoomMuted,
    useUnlockChatRoomMemberNickname,
} from "../api/mutations/chat";

interface UseRoomModerationArgs {
    roomId: string | undefined;
    room: ChatRoom | null;
    setRoom: Dispatch<SetStateAction<ChatRoom | null>>;
    setMembers: Dispatch<SetStateAction<ChatRoomMember[]>>;
    setMessages: Dispatch<SetStateAction<ChatMessage[]>>;
    setToast: (msg: string | null) => void;
    navigate: NavigateFunction;
}

export function useRoomModeration({
    roomId,
    room,
    setRoom,
    setMembers,
    setMessages,
    setToast,
    navigate,
}: UseRoomModerationArgs) {
    const [busy, setBusy] = useState<string | null>(null);
    const [openMemberMenu, setOpenMemberMenu] = useState<string | null>(null);
    const [nicknameDialogTarget, setNicknameDialogTarget] = useState<ChatRoomMember | null>(null);
    const [nicknameDialogValue, setNicknameDialogValue] = useState("");
    const [nicknameDialogError, setNicknameDialogError] = useState<string>("");
    const [nicknameDialogSaving, setNicknameDialogSaving] = useState(false);
    const [timeoutDialogTarget, setTimeoutDialogTarget] = useState<ChatRoomMember | null>(null);
    const [timeoutDialogAmount, setTimeoutDialogAmount] = useState("1");
    const [timeoutDialogUnit, setTimeoutDialogUnit] = useState("hours");
    const [timeoutDialogError, setTimeoutDialogError] = useState("");
    const [timeoutDialogSaving, setTimeoutDialogSaving] = useState(false);

    const deleteRoomMutation = useDeleteChannel();
    const setMutedMutation = useSetChatRoomMuted();
    const kickMutation = useKickChatRoomMember(roomId ?? "");
    const banMutation = useBanChatRoomMember(roomId ?? "");
    const setNicknameMutation = useSetChatRoomMemberNickname(roomId ?? "");
    const unlockNicknameMutation = useUnlockChatRoomMemberNickname(roomId ?? "");
    const setTimeoutMutation = useSetChatRoomMemberTimeout(roomId ?? "");
    const clearTimeoutMutation = useClearChatRoomMemberTimeout(roomId ?? "");

    function openNicknameDialog(member: ChatRoomMember) {
        setNicknameDialogTarget(member);
        setNicknameDialogValue(member.nickname ?? "");
        setNicknameDialogError("");
        setOpenMemberMenu(null);
    }

    function openTimeoutDialog(member: ChatRoomMember) {
        setTimeoutDialogTarget(member);
        setTimeoutDialogAmount("1");
        setTimeoutDialogUnit("hours");
        setTimeoutDialogError("");
        setOpenMemberMenu(null);
    }

    function formatTimeoutUntil(value?: string): string {
        if (!value) {
            return "";
        }
        const parsed = parseServerDate(value);
        if (!parsed) {
            return value;
        }
        return parsed.toLocaleString();
    }

    async function handleModSetNickname() {
        if (!roomId || !nicknameDialogTarget) {
            return;
        }
        setNicknameDialogSaving(true);
        setNicknameDialogError("");
        try {
            const updated = await setNicknameMutation.mutateAsync({
                userId: nicknameDialogTarget.user.id,
                nickname: nicknameDialogValue.trim(),
            });
            applyLocalMemberChange(updated, setMembers, setMessages);
            setNicknameDialogTarget(null);
        } catch (err) {
            setNicknameDialogError(err instanceof Error ? err.message : "Failed to set nickname");
        } finally {
            setNicknameDialogSaving(false);
        }
    }

    async function handleModUnlockNickname(targetId: string) {
        if (!roomId) {
            return;
        }
        setBusy(targetId);
        setOpenMemberMenu(null);
        try {
            const updated = await unlockNicknameMutation.mutateAsync(targetId);
            applyLocalMemberChange(updated, setMembers, setMessages);
        } catch (err) {
            setToast(err instanceof Error ? err.message : "Failed to unlock nickname");
        } finally {
            setBusy(null);
        }
    }

    async function handleSetTimeout() {
        if (!roomId || !timeoutDialogTarget) {
            return;
        }
        const amount = Number(timeoutDialogAmount);
        if (!Number.isInteger(amount) || amount <= 0) {
            setTimeoutDialogError("Enter a whole number greater than zero");
            return;
        }

        setTimeoutDialogSaving(true);
        setTimeoutDialogError("");
        try {
            const updated = await setTimeoutMutation.mutateAsync({
                userId: timeoutDialogTarget.user.id,
                amount,
                unit: timeoutDialogUnit,
            });
            setMembers(prev => prev.map(m => (m.user.id === updated.user.id ? updated : m)));
            setTimeoutDialogTarget(null);
        } catch (err) {
            setTimeoutDialogError(err instanceof Error ? err.message : "Failed to set timeout");
        } finally {
            setTimeoutDialogSaving(false);
        }
    }

    async function handleClearTimeout(targetId: string) {
        if (!roomId) {
            return;
        }
        setBusy(`timeout:${targetId}`);
        setOpenMemberMenu(null);
        try {
            const updated = await clearTimeoutMutation.mutateAsync(targetId);
            setMembers(prev => prev.map(m => (m.user.id === updated.user.id ? updated : m)));
        } catch (err) {
            setToast(err instanceof Error ? err.message : "Failed to clear timeout");
        } finally {
            setBusy(null);
        }
    }

    async function handleKick(targetId: string) {
        if (!roomId || !window.confirm("Kick this member from the room?")) {
            return;
        }
        setBusy(targetId);
        try {
            await kickMutation.mutateAsync(targetId);
            setMembers(prev => prev.filter(m => m.user.id !== targetId));
        } catch (err) {
            setToast(err instanceof Error ? err.message : "Failed to kick");
        } finally {
            setBusy(null);
        }
    }

    async function handleBan(targetId: string) {
        if (!roomId) {
            return;
        }
        const reason = window.prompt(
            "Ban this member from the room? They will not be able to rejoin or see the room. Optional reason:",
            "",
        );
        if (reason === null) {
            return;
        }
        setBusy(targetId);
        try {
            await banMutation.mutateAsync({ userId: targetId, reason });
            setMembers(prev => prev.filter(m => m.user.id !== targetId));
            setToast("Member banned from the room.");
        } catch (err) {
            setToast(err instanceof Error ? err.message : "Failed to ban");
        } finally {
            setBusy(null);
        }
    }

    async function handleToggleMute() {
        if (!roomId || !room) {
            return;
        }
        setBusy("mute");
        const next = !room.viewer_muted;
        try {
            await setMutedMutation.mutateAsync({ roomId, muted: next });
            setRoom(prev => {
                if (!prev) {
                    return prev;
                }
                return { ...prev, viewer_muted: next };
            });
            setToast(next ? "Notifications muted" : "Notifications unmuted");
        } catch (err) {
            setToast(err instanceof Error ? err.message : "Failed to update mute");
        } finally {
            setBusy(null);
        }
    }

    async function handleDelete() {
        if (!roomId || !window.confirm("Delete this room? Everyone will be removed and the messages will be lost.")) {
            return;
        }
        setBusy("delete");
        try {
            await deleteRoomMutation.mutateAsync(roomId);
            navigate("/channels");
        } catch (err) {
            setToast(err instanceof Error ? err.message : "Failed to delete");
            setBusy(null);
        }
    }

    return {
        busy,
        openMemberMenu,
        setOpenMemberMenu,
        nicknameDialogTarget,
        setNicknameDialogTarget,
        nicknameDialogValue,
        setNicknameDialogValue,
        nicknameDialogError,
        nicknameDialogSaving,
        timeoutDialogTarget,
        setTimeoutDialogTarget,
        timeoutDialogAmount,
        setTimeoutDialogAmount,
        timeoutDialogUnit,
        setTimeoutDialogUnit,
        timeoutDialogError,
        timeoutDialogSaving,
        formatTimeoutUntil,
        openNicknameDialog,
        openTimeoutDialog,
        handleModSetNickname,
        handleModUnlockNickname,
        handleSetTimeout,
        handleClearTimeout,
        handleKick,
        handleBan,
        handleToggleMute,
        handleDelete,
    };
}

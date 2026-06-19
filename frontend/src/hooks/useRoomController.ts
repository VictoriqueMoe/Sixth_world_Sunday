import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router";
import { useAuth } from "./useAuth";
import { useNotifications } from "./useNotifications";
import { usePageTitle } from "./usePageTitle";
import { buildMentionMatcher } from "../utils/mentions";
import { parseServerDate } from "../utils/time";
import { buildMemberGroups } from "../utils/memberGroups";
import { useTypingIndicator } from "./useTypingIndicator";
import { useSiteInfo } from "./useSiteInfo";
import { useRoomData } from "./useRoomData";
import { useRoomVoice } from "./useRoomVoice";
import { useRoomPresence } from "./useRoomPresence";
import { useRoomMessages } from "./useRoomMessages";
import { useRoomPanels } from "./useRoomPanels";
import { useRoomModeration } from "./useRoomModeration";
import { useRoomWSHandlers } from "./useRoomWSHandlers";

export function useRoomController() {
    const { roomId } = useParams<{ roomId: string }>();
    const navigate = useNavigate();
    const { user } = useAuth();
    const matchesViewerMention = useMemo(() => buildMentionMatcher(user?.username), [user?.username]);
    const { addWSListener, sendWSMessage, wsEpoch } = useNotifications();
    const voiceEnabled = useSiteInfo()?.voice_enabled ?? false;

    const [toast, setToast] = useState<string | null>(null);
    useEffect(() => {
        if (!toast) {
            return;
        }
        const t = setTimeout(() => setToast(null), 4000);
        return () => clearTimeout(t);
    }, [toast]);

    useEffect(() => {
        document.body.dataset.chatPage = "true";
        return () => {
            delete document.body.dataset.chatPage;
        };
    }, []);

    const { room, setRoom, members, setMembers, baseMembers, loadMembers, loading } = useRoomData(roomId);

    usePageTitle(room?.name ?? "Chat Room");

    const { voice, voiceIdSet } = useRoomVoice({ roomId, room, voiceEnabled });

    const { setPresenceMap, presenceMapMerged, memberOnlineWeight } = useRoomPresence({
        roomId,
        baseMembers,
        sendWSMessage,
        wsEpoch,
    });

    const { typingUserIds, noteTyping, clearUser: clearTypingUser } = useTypingIndicator(roomId);

    const [nowTick, setNowTick] = useState(() => Date.now());
    useEffect(() => {
        const t = setInterval(() => setNowTick(Date.now()), 30_000);
        return () => clearInterval(t);
    }, []);
    const currentMember = members.find(m => m.user.id === user?.id) ?? null;
    const viewerTimeoutUntil = currentMember?.timeout_until ?? undefined;
    const viewerTimedOutDate = parseServerDate(viewerTimeoutUntil);
    const viewerTimedOut = viewerTimedOutDate ? viewerTimedOutDate.getTime() > nowTick : false;

    const messages = useRoomMessages({ roomId, room, user, wsEpoch, viewerTimedOut, setToast });

    const panels = useRoomPanels(roomId);

    const moderation = useRoomModeration({
        roomId,
        room,
        setRoom,
        setMembers,
        setMessages: messages.setMessages,
        setToast,
        navigate,
    });

    const memberGroups = buildMemberGroups(members, memberOnlineWeight, voiceIdSet);

    useRoomWSHandlers({
        user,
        roomId,
        room,
        addWSListener,
        sendWSMessage,
        wsEpoch,
        setRoom,
        setMembers,
        loadMembers,
        setMessages: messages.setMessages,
        scrollToBottom: messages.scrollToBottom,
        setPresenceMap,
        noteTyping,
        clearTypingUser,
        navigate,
        setToast,
        setPinnedRefreshKey: panels.setPinnedRefreshKey,
    });

    const typingNames = typingUserIds
        .filter(id => id !== user?.id)
        .map(id => {
            const m = members.find(mem => mem.user.id === id);
            if (!m) {
                return "Someone";
            }

            if (m.nickname && m.nickname.trim() !== "") {
                return m.nickname;
            }

            if (m.user.display_name && m.user.display_name.trim() !== "") {
                return m.user.display_name;
            }

            return m.user.username;
        });

    return {
        user,
        navigate,
        loading,
        room,
        roomId,
        members,
        memberGroups,
        presenceMapMerged,
        memberOnlineWeight,
        currentMember,
        mobileView: panels.mobileView,
        setMobileView: panels.setMobileView,
        sidebarCollapsed: panels.sidebarCollapsed,
        toggleSidebar: panels.toggleSidebar,
        descExpanded: panels.descExpanded,
        toggleDescExpanded: panels.toggleDescExpanded,
        messages: messages.messages,
        hasMore: messages.hasMore,
        loadingMore: messages.loadingMore,
        messagesContainerRef: messages.messagesContainerRef,
        messagesContentRef: messages.messagesContentRef,
        messagesEndRef: messages.messagesEndRef,
        handleMessagesScroll: messages.handleMessagesScroll,
        scrollToBottom: messages.scrollToBottom,
        highlightedMsgId: messages.highlightedMsgId,
        matchesViewerMention,
        typingNames,
        voice,
        voiceEnabled,
        replyingTo: panels.replyingTo,
        setReplyingTo: panels.setReplyingTo,
        editingMessageId: messages.editingMessageId,
        setEditingMessageId: messages.setEditingMessageId,
        viewerTimeoutUntil,
        viewerTimedOut,
        lightboxSrc: panels.lightboxSrc,
        setLightboxSrc: panels.setLightboxSrc,
        toast,
        setToast,
        busy: moderation.busy,
        sendWSMessage,
        pinnedOpen: panels.pinnedOpen,
        setPinnedOpen: panels.setPinnedOpen,
        searchOpen: panels.searchOpen,
        setSearchOpen: panels.setSearchOpen,
        pinnedRefreshKey: panels.pinnedRefreshKey,
        editProfileOpen: panels.editProfileOpen,
        setEditProfileOpen: panels.setEditProfileOpen,
        inviteModalOpen: panels.inviteModalOpen,
        setInviteModalOpen: panels.setInviteModalOpen,
        moderationDialogOpen: panels.moderationDialogOpen,
        setModerationDialogOpen: panels.setModerationDialogOpen,
        openMemberMenu: moderation.openMemberMenu,
        setOpenMemberMenu: moderation.setOpenMemberMenu,
        setMembers,
        nicknameDialogTarget: moderation.nicknameDialogTarget,
        setNicknameDialogTarget: moderation.setNicknameDialogTarget,
        nicknameDialogValue: moderation.nicknameDialogValue,
        setNicknameDialogValue: moderation.setNicknameDialogValue,
        nicknameDialogError: moderation.nicknameDialogError,
        nicknameDialogSaving: moderation.nicknameDialogSaving,
        timeoutDialogTarget: moderation.timeoutDialogTarget,
        setTimeoutDialogTarget: moderation.setTimeoutDialogTarget,
        timeoutDialogAmount: moderation.timeoutDialogAmount,
        setTimeoutDialogAmount: moderation.setTimeoutDialogAmount,
        timeoutDialogUnit: moderation.timeoutDialogUnit,
        setTimeoutDialogUnit: moderation.setTimeoutDialogUnit,
        timeoutDialogError: moderation.timeoutDialogError,
        timeoutDialogSaving: moderation.timeoutDialogSaving,
        formatTimeoutUntil: moderation.formatTimeoutUntil,
        openNicknameDialog: moderation.openNicknameDialog,
        openTimeoutDialog: moderation.openTimeoutDialog,
        handleSentMessage: messages.handleSentMessage,
        handleModSetNickname: moderation.handleModSetNickname,
        handleModUnlockNickname: moderation.handleModUnlockNickname,
        handleSetTimeout: moderation.handleSetTimeout,
        handleClearTimeout: moderation.handleClearTimeout,
        handleKick: moderation.handleKick,
        handleBan: moderation.handleBan,
        handleToggleMute: moderation.handleToggleMute,
        handleDelete: moderation.handleDelete,
        handleReactionToggle: messages.handleReactionToggle,
        handlePinToggle: messages.handlePinToggle,
        handleJumpToMessage: messages.handleJumpToMessage,
        handleDeleteMessage: messages.handleDeleteMessage,
        handleEditMessage: messages.handleEditMessage,
        handleEditLast: messages.handleEditLast,
    };
}

export type RoomController = ReturnType<typeof useRoomController>;

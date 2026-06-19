import { useState } from "react";
import { type ReplyTarget } from "../components/chat/ChatComposer/ChatComposer";

export function useRoomPanels(roomId: string | undefined) {
    const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);
    const [mobileView, setMobileView] = useState<"members" | "chat">("chat");
    const [replyingTo, setReplyingTo] = useState<ReplyTarget | null>(null);
    const [pinnedOpen, setPinnedOpen] = useState(false);
    const [searchOpen, setSearchOpen] = useState(false);
    const [pinnedRefreshKey, setPinnedRefreshKey] = useState(0);
    const [editProfileOpen, setEditProfileOpen] = useState(false);
    const [inviteModalOpen, setInviteModalOpen] = useState(false);
    const [moderationDialogOpen, setModerationDialogOpen] = useState(false);

    const [sidebarCollapsedOverride, setSidebarCollapsedOverride] = useState<{
        roomId: string | null;
        value: boolean | null;
    }>({
        roomId: null,
        value: null,
    });
    const sidebarCollapsed = (() => {
        if (!roomId) {
            return false;
        }
        if (sidebarCollapsedOverride.roomId === roomId && sidebarCollapsedOverride.value !== null) {
            return sidebarCollapsedOverride.value;
        }
        return localStorage.getItem(`ut-room-sidebar-collapsed-${roomId}`) === "1";
    })();

    function toggleSidebar() {
        if (!roomId) {
            return;
        }
        const next = !sidebarCollapsed;
        try {
            if (next) {
                localStorage.setItem(`ut-room-sidebar-collapsed-${roomId}`, "1");
            } else {
                localStorage.removeItem(`ut-room-sidebar-collapsed-${roomId}`);
            }
        } catch {
            setSidebarCollapsedOverride({ roomId, value: next });
            return;
        }
        setSidebarCollapsedOverride({ roomId, value: next });
    }

    const roomInfoStorageKey = roomId ? `roomInfoExpanded:${roomId}` : null;
    const [descExpandedOverride, setDescExpandedOverride] = useState<{ key: string | null; value: boolean | null }>({
        key: null,
        value: null,
    });
    const descExpanded = (() => {
        if (typeof window === "undefined") {
            return true;
        }
        if (descExpandedOverride.key === roomInfoStorageKey && descExpandedOverride.value !== null) {
            return descExpandedOverride.value;
        }
        if (roomInfoStorageKey) {
            const stored = window.localStorage.getItem(roomInfoStorageKey);
            if (stored !== null) {
                return stored === "true";
            }
        }
        return window.matchMedia("(min-width: 769px)").matches;
    })();

    function toggleDescExpanded() {
        const next = !descExpanded;
        if (roomInfoStorageKey) {
            window.localStorage.setItem(roomInfoStorageKey, next ? "true" : "false");
        }
        setDescExpandedOverride({ key: roomInfoStorageKey, value: next });
    }

    return {
        lightboxSrc,
        setLightboxSrc,
        mobileView,
        setMobileView,
        replyingTo,
        setReplyingTo,
        pinnedOpen,
        setPinnedOpen,
        searchOpen,
        setSearchOpen,
        pinnedRefreshKey,
        setPinnedRefreshKey,
        editProfileOpen,
        setEditProfileOpen,
        inviteModalOpen,
        setInviteModalOpen,
        moderationDialogOpen,
        setModerationDialogOpen,
        sidebarCollapsed,
        toggleSidebar,
        descExpanded,
        toggleDescExpanded,
    };
}

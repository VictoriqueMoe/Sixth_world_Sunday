import type { Notification, NotificationType } from "../types/api";

export type NotificationCategory = "social" | "moderation";

interface NotificationConfig {
    text: string;
    category: NotificationCategory | "dynamic";
    route: (notif: Notification) => string;
}

const roleDisplayNames: Record<string, string> = {
    super_admin: "Sysop",
    admin: "Admin",
    moderator: "Moderator",
};

const categoryLabels: Record<NotificationCategory, string> = {
    social: "Social",
    moderation: "Moderation",
};

const categoryOrder: NotificationCategory[] = ["social", "moderation"];

function routeByReferenceType(notif: Notification): string {
    const refType = notif.reference_type;
    if (refType.startsWith("chat_message:")) {
        const msgId = refType.split(":")[1];
        return `/channels/${notif.reference_id}#msg-${msgId}`;
    }
    return "/notifications";
}

const notificationConfigs: Partial<Record<NotificationType, NotificationConfig>> = {
    chat_room_message: {
        text: "sent a message in a chat room",
        category: "social",
        route: chatMessageRoute,
    },
    report: {
        text: "reported content",
        category: "moderation",
        route: () => "/admin/reports",
    },
    report_resolved: {
        text: "resolved your report",
        category: "moderation",
        route: routeByReferenceType,
    },
    chat_mention: {
        text: "mentioned you in a chat room",
        category: "social",
        route: chatMessageRoute,
    },
    chat_room_invite: {
        text: "added you to a chat room",
        category: "social",
        route: notif => `/channels/${notif.reference_id}`,
    },
    chat_reply: {
        text: "replied to your message",
        category: "social",
        route: chatMessageRoute,
    },
    chat_reaction: {
        text: "reacted to your message",
        category: "social",
        route: chatMessageRoute,
    },
    chat_room_banned: {
        text: "banned you from a chat room",
        category: "moderation",
        route: notif => `/channels/${notif.reference_id}`,
    },
    chat_room_kicked: {
        text: "kicked you from a chat room",
        category: "moderation",
        route: notif => `/channels/${notif.reference_id}`,
    },
    chat_room_unbanned: {
        text: "unbanned you from a chat room",
        category: "moderation",
        route: notif => `/channels/${notif.reference_id}`,
    },
};

function chatMessageRoute(notif: Notification): string {
    const refType = notif.reference_type;
    if (refType.startsWith("chat_message:")) {
        const msgId = refType.split(":")[1];
        return `/channels/${notif.reference_id}#msg-${msgId}`;
    }
    return `/channels/${notif.reference_id}`;
}

export function getNotificationText(notif: Notification): string {
    if (notif.message && notif.type !== "content_edited") {
        return notif.message;
    }
    return notificationConfigs[notif.type]?.text ?? "";
}

export function showDesktopNotification(notif: Notification): void {
    if (typeof window === "undefined" || !("Notification" in window)) {
        return;
    }
    if (window.Notification.permission !== "granted") {
        return;
    }
    if (document.visibilityState === "visible" && document.hasFocus()) {
        return;
    }
    const actorName = notif.actor?.display_name || "";
    const body = getNotificationText(notif);
    const title = actorName ? `${actorName} ${body}` : body;
    const route = getNotificationRoute(notif);
    const osNotif = new window.Notification(title, {
        body: notif.message || "",
        icon: notif.actor?.avatar_url || "/favicon/favicon.svg",
        badge: "/favicon/favicon.svg",
        tag: `notif-${notif.id}`,
    });
    osNotif.onclick = () => {
        window.focus();
        window.location.href = route;
        osNotif.close();
    };
}

export async function ensureNotificationPermission(): Promise<boolean> {
    if (typeof window === "undefined" || !("Notification" in window)) {
        return false;
    }
    if (window.Notification.permission === "granted") {
        return true;
    }
    if (window.Notification.permission === "denied") {
        return false;
    }
    try {
        const result = await window.Notification.requestPermission();
        return result === "granted";
    } catch {
        return false;
    }
}

export function getNotificationRoute(notif: Notification): string {
    const config = notificationConfigs[notif.type];
    if (config) {
        return config.route(notif);
    }
    return "/notifications";
}

export function getNotificationCategory(notif: Notification): NotificationCategory {
    const config = notificationConfigs[notif.type];
    if (!config) {
        return "social";
    }
    if (config.category === "dynamic") {
        return "social";
    }
    return config.category;
}

export function getCategoryLabel(category: NotificationCategory): string {
    return categoryLabels[category];
}

export function getCategoryOrder(): NotificationCategory[] {
    return categoryOrder;
}

export function groupByCategory(notifications: Notification[]): Map<NotificationCategory, Notification[]> {
    const groups = new Map<NotificationCategory, Notification[]>();
    for (const notif of notifications) {
        const cat = getNotificationCategory(notif);
        const list = groups.get(cat);
        if (list) {
            list.push(notif);
        } else {
            groups.set(cat, [notif]);
        }
    }
    return groups;
}

export function isContentEditedNotification(notif: Notification): boolean {
    return notif.type === "content_edited";
}

export function formatContentEditedText(notif: Notification): { message: string; role: string; actorName: string } {
    const message = notif.message || "your content has been edited";
    const role = notif.actor.role ? (roleDisplayNames[notif.actor.role] ?? "") : "";
    return { message, role, actorName: notif.actor.display_name };
}

export { relativeTime } from "./time";

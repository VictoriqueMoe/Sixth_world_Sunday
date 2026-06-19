export const queryKeys = {
    chat: {
        all: ["chat"] as const,
        room: (id: string) => ["chat", "room", id] as const,
        roomMembers: (id: string) => ["chat", "room", id, "members"] as const,
        roomMessages: (id: string) => ["chat", "room", id, "messages"] as const,
        roomList: (params: Record<string, unknown> = {}) => ["chat", "rooms", params] as const,
        systemRooms: () => ["chat", "system-rooms"] as const,
        pinned: (id: string) => ["chat", "room", id, "pinned"] as const,
    },
    profile: {
        all: ["profile"] as const,
        byUsername: (username: string) => ["profile", "username", username] as const,
        byID: (id: string) => ["profile", "id", id] as const,
        blockedUsers: (userID: string) => ["profile", id(userID), "blocked"] as const,
    },
    notifications: {
        all: ["notifications"] as const,
        list: (params: Record<string, unknown> = {}) => ["notifications", "list", params] as const,
        unreadCount: () => ["notifications", "unread-count"] as const,
    },
    admin: {
        all: ["admin"] as const,
        users: (params: Record<string, unknown> = {}) => ["admin", "users", params] as const,
        invites: () => ["admin", "invites"] as const,
        reports: (params: Record<string, unknown> = {}) => ["admin", "reports", params] as const,
        auditLog: (params: Record<string, unknown> = {}) => ["admin", "audit-log", params] as const,
        bannedWords: (scope: string) => ["admin", "banned-words", scope] as const,
        vanityRoles: () => ["admin", "vanity-roles"] as const,
    },
    siteInfo: () => ["site-info"] as const,
    settings: () => ["settings"] as const,
    theme: () => ["theme"] as const,
} as const;

function id(value: string): string {
    return value;
}

import type { SearchEntityType } from "../../../types/api";

export type SearchTypeGroup = "chats" | "users";

interface SearchTypeMeta {
    type: SearchEntityType;
    label: string;
    short: string;
    color: string;
    group: SearchTypeGroup;
}

interface SearchTypeGroupDef {
    id: SearchTypeGroup;
    label: string;
}

const SEARCH_REGISTRY: SearchTypeMeta[] = [
    { type: "chat_message", label: "Chat message", short: "Chat", color: "var(--cyan)", group: "chats" },
    { type: "user", label: "User", short: "User", color: "var(--magenta)", group: "users" },
];

const SEARCH_GROUP_DEFS: SearchTypeGroupDef[] = [
    { id: "chats", label: "Chats" },
    { id: "users", label: "Users" },
];

export const SEARCH_TYPE_META: Record<SearchEntityType, SearchTypeMeta> = Object.fromEntries(
    SEARCH_REGISTRY.map(entry => [entry.type, entry]),
) as Record<SearchEntityType, SearchTypeMeta>;

export const SEARCH_GROUP_LABEL: Record<SearchTypeGroup, string> = Object.fromEntries(
    SEARCH_GROUP_DEFS.map(g => [g.id, g.label]),
) as Record<SearchTypeGroup, string>;

export const SEARCH_GROUP_ORDER: SearchTypeGroup[] = SEARCH_GROUP_DEFS.map(g => g.id);

interface SearchFilterOption {
    value: string;
    label: string;
}

export const SEARCH_FILTER_OPTIONS: SearchFilterOption[] = [
    { value: "", label: "All" },
    ...SEARCH_GROUP_DEFS.map(group => ({
        value: SEARCH_REGISTRY.filter(r => r.group === group.id)
            .map(r => r.type)
            .join(","),
        label: group.label,
    })),
];

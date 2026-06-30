import { lazy, Suspense, useEffect, useRef, useState } from "react";
import { useLocation, useMatch, useNavigate } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import { useChannels, useChatCategories } from "../../../api/queries/chat";
import { useDeleteChannel, useReorderChannels, useTruncateChannel } from "../../../api/mutations/chat";
import { useAuth } from "../../../hooks/useAuth";
import { useNotifications } from "../../../hooks/useNotifications";
import { can } from "../../../utils/permissions";
import type { ChatCategory, ChatRoom, WSMessage } from "../../../types/api";
import { useVoice } from "../../../context/voiceContextValue";
import { CreateChannelModal } from "../../chat/CreateChannelModal/CreateChannelModal";
import { EditChannelModal } from "../../chat/EditChannelModal/EditChannelModal";
import { ContextMenu } from "../../ContextMenu/ContextMenu";
import { useContextMenu } from "../../ContextMenu/useContextMenu";
import { Modal } from "../../Modal/Modal";
import { Button } from "../../Button/Button";
import styles from "./ChannelRail.module.css";

const VoiceDock = lazy(() => import("../../chat/Voice/VoiceDock").then(m => ({ default: m.VoiceDock })));

interface ChannelGroup {
    id: string;
    name: string;
    kind?: string;
    reorderable: boolean;
    channels: ChatRoom[];
}

function buildChannelGroups(rooms: ChatRoom[], categories: ChatCategory[]): ChannelGroup[] {
    const byPosition = (a: ChatRoom, b: ChatRoom) => a.position - b.position;

    if (categories.length === 0) {
        const text: ChatRoom[] = [];
        const voice: ChatRoom[] = [];
        for (const c of rooms) {
            if (c.channel_kind === "voice") {
                voice.push(c);
            } else {
                text.push(c);
            }
        }
        text.sort(byPosition);
        voice.sort(byPosition);
        return [
            { id: "kind:text", name: "text channels", kind: "text", reorderable: false, channels: text },
            { id: "kind:voice", name: "voice channels", kind: "voice", reorderable: false, channels: voice },
        ];
    }

    const knownCategoryIds = new Set(categories.map(c => c.id));
    const builtinByKind = new Map<string, ChatCategory>();
    for (const c of categories) {
        if (c.is_builtin && c.kind) {
            builtinByKind.set(c.kind, c);
        }
    }

    const channelsByCategory = new Map<string, ChatRoom[]>();
    const uncategorized: ChatRoom[] = [];
    for (const room of rooms) {
        let categoryId: string | undefined;
        if (room.category_id && knownCategoryIds.has(room.category_id)) {
            categoryId = room.category_id;
        } else {
            categoryId = builtinByKind.get(room.channel_kind)?.id;
        }
        if (!categoryId) {
            uncategorized.push(room);
            continue;
        }
        const arr = channelsByCategory.get(categoryId) ?? [];
        arr.push(room);
        channelsByCategory.set(categoryId, arr);
    }

    const ordered = [...categories].sort((a, b) => a.position - b.position);
    const groups: ChannelGroup[] = [];
    for (const category of ordered) {
        const channels = (channelsByCategory.get(category.id) ?? []).sort(byPosition);
        groups.push({ id: category.id, name: category.name, kind: category.kind, reorderable: true, channels });
    }

    if (uncategorized.length > 0) {
        uncategorized.sort(byPosition);
        groups.push({ id: "uncategorized", name: "uncategorized", reorderable: false, channels: uncategorized });
    }

    return groups;
}

export function ChannelRail({ onNavigate }: { onNavigate?: () => void } = {}) {
    const { user } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const activeId = useMatch("/channels/:roomId")?.params.roomId;

    const go = (path: string) => {
        navigate(path);
        onNavigate?.();
    };
    const qc = useQueryClient();
    const { addWSListener } = useNotifications();
    const { rooms } = useChannels();
    const { categories } = useChatCategories();
    const voice = useVoice();

    const [createOpen, setCreateOpen] = useState(false);
    const [createKind, setCreateKind] = useState<"text" | "voice">("text");
    const [editChannel, setEditChannel] = useState<ChatRoom | null>(null);
    const [pendingDelete, setPendingDelete] = useState<ChatRoom | null>(null);
    const [pendingTruncate, setPendingTruncate] = useState<ChatRoom | null>(null);
    const [dragOverId, setDragOverId] = useState<string | null>(null);

    const dragRef = useRef<{ id: string; categoryId: string } | null>(null);

    const { state: menuState, open: openMenu, close: closeMenu } = useContextMenu();
    const reorderMutation = useReorderChannels();
    const deleteMutation = useDeleteChannel();
    const truncateMutation = useTruncateChannel();

    useEffect(() => {
        return addWSListener((msg: WSMessage) => {
            if (
                msg.type === "channel_created" ||
                msg.type === "channel_deleted" ||
                msg.type === "channel_updated" ||
                msg.type === "channels_reordered" ||
                msg.type === "chat_room_invited" ||
                msg.type === "chat_kicked"
            ) {
                qc.invalidateQueries({ queryKey: ["channels"] });
                return;
            }
            if (msg.type === "chat_read") {
                const data = msg.data as { room_id: string };
                qc.setQueryData<{ rooms: ChatRoom[] }>(["channels"], prev => {
                    if (!prev) {
                        return prev;
                    }
                    return {
                        ...prev,
                        rooms: prev.rooms.map(r => (r.id === data.room_id ? { ...r, unread: false } : r)),
                    };
                });
            }
        });
    }, [addWSListener, qc]);

    const canManage = can(user?.role, "manage_channels");

    const groups = buildChannelGroups(rooms, categories);

    function openCreate(kind: "text" | "voice") {
        setCreateKind(kind);
        setCreateOpen(true);
    }

    function openChannelMenu(e: React.MouseEvent, channel: ChatRoom) {
        openMenu(e, [
            { id: "edit", label: "Edit Channel", icon: "✎", onClick: () => setEditChannel(channel) },
            {
                id: "truncate",
                label: "Truncate Channel",
                icon: "🧹",
                variant: "danger",
                onClick: () => setPendingTruncate(channel),
            },
            {
                id: "delete",
                label: "Delete Channel",
                icon: "✕",
                variant: "danger",
                onClick: () => setPendingDelete(channel),
            },
        ]);
    }

    function reorderWithin(group: ChannelGroup, orderedIds: string[]) {
        reorderMutation.mutate({ categoryId: group.id, roomIds: orderedIds });
    }

    function handleDropOnChannel(group: ChannelGroup, targetId: string) {
        const drag = dragRef.current;
        dragRef.current = null;
        setDragOverId(null);

        if (!drag || drag.categoryId !== group.id || drag.id === targetId) {
            return;
        }

        const remaining = group.channels.map(c => c.id).filter(id => id !== drag.id);
        const targetIndex = remaining.indexOf(targetId);
        if (targetIndex === -1) {
            return;
        }

        remaining.splice(targetIndex, 0, drag.id);
        reorderWithin(group, remaining);
    }

    function handleDropAtEnd(group: ChannelGroup) {
        const drag = dragRef.current;
        dragRef.current = null;
        setDragOverId(null);

        if (!drag || drag.categoryId !== group.id) {
            return;
        }

        const remaining = group.channels.map(c => c.id).filter(id => id !== drag.id);
        remaining.push(drag.id);
        reorderWithin(group, remaining);
    }

    function confirmDelete() {
        if (!pendingDelete) {
            return;
        }

        const id = pendingDelete.id;
        setPendingDelete(null);
        deleteMutation.mutate(id);
    }

    function confirmTruncate() {
        if (!pendingTruncate) {
            return;
        }

        const id = pendingTruncate.id;
        setPendingTruncate(null);
        truncateMutation.mutate(id);
    }

    function renderRow(c: ChatRoom, group: ChannelGroup) {
        const active = c.id === activeId;
        const hasUnread = c.unread && !c.viewer_muted;
        const glyph = c.channel_kind === "voice" ? "◊" : "#";
        const voiceCount = voice.presence[c.id]?.length ?? c.voice_count ?? 0;
        const draggable = canManage && group.reorderable;
        const classes = [styles.channel];
        if (active) {
            classes.push(styles.active);
        }
        if (hasUnread) {
            classes.push(styles.unreadChannel);
        }
        if (dragOverId === c.id) {
            classes.push(styles.dragOver);
        }
        return (
            <button
                key={c.id}
                type="button"
                className={classes.join(" ")}
                onClick={() => go(`/channels/${c.id}`)}
                onContextMenu={canManage ? e => openChannelMenu(e, c) : undefined}
                draggable={draggable}
                onDragStart={
                    draggable
                        ? () => {
                              dragRef.current = { id: c.id, categoryId: group.id };
                          }
                        : undefined
                }
                onDragOver={
                    draggable
                        ? e => {
                              if (dragRef.current && dragRef.current.categoryId === group.id) {
                                  e.preventDefault();
                                  setDragOverId(c.id);
                              }
                          }
                        : undefined
                }
                onDragLeave={
                    draggable
                        ? () => {
                              setDragOverId(prev => (prev === c.id ? null : prev));
                          }
                        : undefined
                }
                onDrop={
                    draggable
                        ? e => {
                              e.preventDefault();
                              e.stopPropagation();
                              handleDropOnChannel(group, c.id);
                          }
                        : undefined
                }
                onDragEnd={
                    draggable
                        ? () => {
                              dragRef.current = null;
                              setDragOverId(null);
                          }
                        : undefined
                }
            >
                <span className={styles.glyph}>{glyph}</span>
                <span className={styles.channelName}>{c.name}</span>
                {c.channel_kind === "voice" && voiceCount > 0 && (
                    <span className={styles.voiceCount}>{voiceCount}</span>
                )}
                {hasUnread && <span className={styles.unreadDot} aria-label="Unread messages" />}
            </button>
        );
    }

    return (
        <aside className={styles.rail}>
            <div className={styles.serverHead}>
                <span className={styles.serverName}>
                    SIXTH WORLD <b>SUNDAY</b>
                </span>
            </div>

            <div className={styles.groups}>
                <button
                    type="button"
                    className={`${styles.channel}${location.pathname === "/weather" ? ` ${styles.active}` : ""}`}
                    onClick={() => go("/weather")}
                    title="Weather"
                >
                    <span className={styles.glyph}>{"☀"}</span>
                    <span className={styles.channelName}>Weather</span>
                </button>

                <button
                    type="button"
                    className={`${styles.channel}${location.pathname === "/events" ? ` ${styles.active}` : ""}`}
                    onClick={() => go("/events")}
                    title="Events"
                >
                    <span className={styles.glyph}>{"▦"}</span>
                    <span className={styles.channelName}>Events</span>
                </button>

                <button
                    type="button"
                    className={`${styles.channel}${location.pathname === "/maps" ? ` ${styles.active}` : ""}`}
                    onClick={() => go("/maps")}
                    title="Maps"
                >
                    <span className={styles.glyph}>{"❖"}</span>
                    <span className={styles.channelName}>Maps</span>
                </button>

                <button
                    type="button"
                    className={`${styles.channel}${location.pathname === "/files" ? ` ${styles.active}` : ""}`}
                    onClick={() => go("/files")}
                    title="File Vault"
                >
                    <span className={styles.glyph}>{"⛁"}</span>
                    <span className={styles.channelName}>Data Vault</span>
                </button>

                {groups.map(group => (
                    <div className={styles.group} key={group.id}>
                        <div className={styles.groupHeader}>
                            <span className={styles.groupLabel}>{group.name}</span>
                            {canManage && group.kind && (
                                <button
                                    type="button"
                                    className={styles.addBtn}
                                    onClick={() => openCreate(group.kind === "voice" ? "voice" : "text")}
                                    aria-label="Create channel"
                                    title="Create channel"
                                >
                                    {"+"}
                                </button>
                            )}
                        </div>
                        <div
                            className={styles.groupList}
                            onDragOver={
                                canManage && group.reorderable
                                    ? e => {
                                          if (dragRef.current && dragRef.current.categoryId === group.id) {
                                              e.preventDefault();
                                          }
                                      }
                                    : undefined
                            }
                            onDrop={
                                canManage && group.reorderable
                                    ? e => {
                                          e.preventDefault();
                                          handleDropAtEnd(group);
                                      }
                                    : undefined
                            }
                        >
                            {group.channels.length === 0 && (
                                <div className={styles.emptyGroup}>{`no ${group.name}`}</div>
                            )}
                            {group.channels.map(c => renderRow(c, group))}
                        </div>
                    </div>
                ))}
            </div>

            {voice.status === "connected" && voice.activeRoomId !== activeId && (
                <Suspense fallback={null}>
                    <VoiceDock />
                </Suspense>
            )}

            {user && (
                <div className={styles.meCard}>
                    <button
                        type="button"
                        className={styles.meInfo}
                        onClick={() => go(`/user/${user.username}`)}
                        title="Your profile"
                    >
                        {user.avatar_url ? (
                            <img className={styles.meAvatar} src={user.avatar_url} alt="" />
                        ) : (
                            <span className={styles.meAvatarPlaceholder}>{user.display_name[0]}</span>
                        )}
                        <span className={styles.meName}>{user.display_name}</span>
                    </button>
                    <button
                        type="button"
                        className={styles.meAction}
                        onClick={() => go("/settings")}
                        aria-label="Settings"
                        title="Settings"
                    >
                        {"⚙"}
                    </button>
                </div>
            )}

            <CreateChannelModal
                isOpen={createOpen}
                initialKind={createKind}
                onClose={() => setCreateOpen(false)}
                onCreated={c => {
                    setCreateOpen(false);
                    go(`/channels/${c.id}`);
                }}
            />

            <EditChannelModal channel={editChannel} onClose={() => setEditChannel(null)} />

            <Modal isOpen={pendingDelete !== null} onClose={() => setPendingDelete(null)} title="Delete Channel">
                <div className={styles.confirmBody}>
                    <p className={styles.confirmText}>
                        Delete <strong>{pendingDelete?.name}</strong>? This removes the channel and all of its messages.
                        This can&apos;t be undone.
                    </p>
                    <div className={styles.confirmActions}>
                        <Button variant="ghost" size="small" onClick={() => setPendingDelete(null)}>
                            Cancel
                        </Button>
                        <Button variant="danger" size="small" onClick={confirmDelete}>
                            Delete channel
                        </Button>
                    </div>
                </div>
            </Modal>

            <Modal isOpen={pendingTruncate !== null} onClose={() => setPendingTruncate(null)} title="Truncate Channel">
                <div className={styles.confirmBody}>
                    <p className={styles.confirmText}>
                        Truncate <strong>{pendingTruncate?.name}</strong>? This permanently deletes all messages and
                        media in the channel, but keeps the channel itself. This can&apos;t be undone.
                    </p>
                    <div className={styles.confirmActions}>
                        <Button variant="ghost" size="small" onClick={() => setPendingTruncate(null)}>
                            Cancel
                        </Button>
                        <Button variant="danger" size="small" onClick={confirmTruncate}>
                            Truncate channel
                        </Button>
                    </div>
                </div>
            </Modal>

            <ContextMenu state={menuState} onClose={closeMenu} />
        </aside>
    );
}

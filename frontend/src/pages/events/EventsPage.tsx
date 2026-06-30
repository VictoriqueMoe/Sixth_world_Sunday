import { useEffect, useState } from "react";
import { useAuth } from "../../hooks/useAuth";
import { useIsMobile } from "../../hooks/useIsMobile";
import { canManageEvents } from "../../utils/permissions";
import { parseServerDate } from "../../utils/time";
import { frequencyLabel } from "../../utils/recurrence";
import { apiUrl } from "../../api/client";
import { useEvents } from "../../api/queries/events";
import { useCancelEvent, useDeleteEvent, useRsvpEvent } from "../../api/mutations/events";
import type { EventItem } from "../../types/api";
import { ChannelRail } from "../../components/layout/ChannelRail/ChannelRail";
import { Button } from "../../components/Button/Button";
import { Modal } from "../../components/Modal/Modal";
import { ContextMenu } from "../../components/ContextMenu/ContextMenu";
import { useContextMenu } from "../../components/ContextMenu/useContextMenu";
import { CreateEventModal } from "../../components/events/CreateEventModal/CreateEventModal";
import styles from "./EventsPage.module.css";

function formatWhen(iso: string): string {
    const d = parseServerDate(iso);
    if (!d) {
        return "";
    }
    const datePart = d.toLocaleDateString(undefined, { weekday: "short", month: "short", day: "numeric" });
    const timePart = d.toLocaleTimeString(undefined, { hour: "numeric", minute: "2-digit" });
    return `${datePart} · ${timePart}`;
}

interface ConfirmTarget {
    action: "cancel" | "delete";
    event: EventItem;
}

interface EventCardProps {
    event: EventItem;
    onInterested: (event: EventItem) => void;
    onShare: (event: EventItem) => void;
    onManage: (e: React.MouseEvent, event: EventItem) => void;
}

function EventCard({ event, onInterested, onShare, onManage }: EventCardProps) {
    const [seriesOpen, setSeriesOpen] = useState(false);
    const recurrence = frequencyLabel(event.frequency, event.start_at);
    const series = event.next_occurrences ?? [];
    const avatars = event.rsvp_avatars ?? [];

    return (
        <article className={styles.card}>
            {event.cover_url && (
                <div className={styles.cover}>
                    <img src={event.cover_url} alt="" />
                </div>
            )}
            <div className={styles.cardBody}>
                <div className={styles.cardTop}>
                    <div className={styles.when}>
                        <span className={styles.whenGlyph}>{"▦"}</span>
                        <div>
                            <div className={styles.whenMain}>{formatWhen(event.next_start_at)}</div>
                            {recurrence && <div className={styles.recurrence}>{recurrence}</div>}
                        </div>
                    </div>
                    <div className={styles.rsvpStack}>
                        {avatars.map((a, i) => (
                            <img key={i} className={styles.rsvpAvatar} src={apiUrl(a)} alt="" />
                        ))}
                        {event.rsvp_count > 0 && (
                            <span className={styles.rsvpCount}>
                                {"\u{1F465}"} {event.rsvp_count}
                            </span>
                        )}
                    </div>
                </div>

                <h3 className={styles.title}>{event.title}</h3>
                {event.description && <p className={styles.description}>{event.description}</p>}

                <div className={styles.location}>
                    {event.location_type === "voice" ? (
                        <span className={styles.locationVoice}>
                            {"\u{1F50A}"} {event.voice_room_name || "Voice channel"}
                        </span>
                    ) : (
                        <a className={styles.locationLink} href={event.external_url} target="_blank" rel="noreferrer">
                            {"\u{1F517}"} {event.external_url}
                        </a>
                    )}
                </div>

                <div className={styles.cardActions}>
                    <Button
                        variant={event.viewer_interested ? "primary" : "secondary"}
                        size="small"
                        onClick={() => onInterested(event)}
                    >
                        {event.viewer_interested ? "\u{1F514} Interested ✓" : "\u{1F514} Interested"}
                    </Button>
                    <Button variant="secondary" size="small" onClick={() => onShare(event)}>
                        {"\u{1F517} Share"}
                    </Button>
                    {event.can_manage && (
                        <Button
                            variant="secondary"
                            size="icon"
                            onClick={e => onManage(e, event)}
                            aria-label="Manage event"
                        >
                            {"⋯"}
                        </Button>
                    )}
                </div>

                {series.length > 1 && (
                    <div className={styles.series}>
                        <button type="button" className={styles.seriesToggle} onClick={() => setSeriesOpen(o => !o)}>
                            {seriesOpen ? "Hide series ▲" : `Events in series (${series.length}) ▼`}
                        </button>
                        {seriesOpen && (
                            <ul className={styles.seriesList}>
                                {series.map(occ => (
                                    <li key={occ} className={styles.seriesItem}>
                                        <span className={styles.seriesGlyph}>{"▦"}</span>
                                        {formatWhen(occ)}
                                    </li>
                                ))}
                            </ul>
                        )}
                    </div>
                )}
            </div>
        </article>
    );
}

export function EventsPage() {
    const { user } = useAuth();
    const isMobile = useIsMobile();
    const { events, loading } = useEvents();
    const canManage = canManageEvents(user?.role);

    const rsvpMutation = useRsvpEvent();
    const cancelMutation = useCancelEvent();
    const deleteMutation = useDeleteEvent();
    const { state: menuState, open: openMenu, close: closeMenu } = useContextMenu();

    const [modalOpen, setModalOpen] = useState(false);
    const [editing, setEditing] = useState<EventItem | null>(null);
    const [confirm, setConfirm] = useState<ConfirmTarget | null>(null);
    const [toast, setToast] = useState("");

    useEffect(() => {
        document.body.setAttribute("data-chat-page", "true");
        return () => {
            document.body.removeAttribute("data-chat-page");
        };
    }, []);

    function openCreate() {
        setEditing(null);
        setModalOpen(true);
    }

    function handleInterested(event: EventItem) {
        rsvpMutation.mutate({ id: event.id, interested: !event.viewer_interested });
    }

    function handleShare(event: EventItem) {
        const link = `${window.location.origin}/events`;
        navigator.clipboard
            .writeText(link)
            .then(() => {
                setToast("Link copied");
                window.setTimeout(() => setToast(""), 2500);
            })
            .catch(() => {
                setToast(event.title);
            });
    }

    function handleManage(e: React.MouseEvent, event: EventItem) {
        openMenu(e, [
            {
                id: "edit",
                label: "Edit event",
                icon: "✎",
                onClick: () => {
                    setEditing(event);
                    setModalOpen(true);
                },
            },
            {
                id: "cancel",
                label: "Cancel event",
                icon: "⏸",
                onClick: () => setConfirm({ action: "cancel", event }),
            },
            {
                id: "delete",
                label: "Delete event",
                icon: "✕",
                variant: "danger",
                onClick: () => setConfirm({ action: "delete", event }),
            },
        ]);
    }

    function confirmAction() {
        if (!confirm) {
            return;
        }
        const id = confirm.event.id;
        if (confirm.action === "delete") {
            deleteMutation.mutate(id);
        } else {
            cancelMutation.mutate(id);
        }
        setConfirm(null);
    }

    return (
        <div className={styles.shell}>
            {!isMobile && <ChannelRail />}
            <div className={styles.main}>
                <div className={styles.page}>
                    <div className={styles.head}>
                        <h2 className={styles.headTitle}>
                            <span className={styles.headGlyph}>{"▦"}</span>
                            {events.length} {events.length === 1 ? "Event" : "Events"}
                        </h2>
                        {canManage && (
                            <Button variant="primary" size="small" onClick={openCreate}>
                                Create Event
                            </Button>
                        )}
                    </div>

                    {loading ? (
                        <div className={styles.empty}>Loading events…</div>
                    ) : events.length === 0 ? (
                        <div className={styles.empty}>No events scheduled yet.</div>
                    ) : (
                        <div className={styles.list}>
                            {events.map(event => (
                                <EventCard
                                    key={event.id}
                                    event={event}
                                    onInterested={handleInterested}
                                    onShare={handleShare}
                                    onManage={handleManage}
                                />
                            ))}
                        </div>
                    )}
                </div>
            </div>

            <CreateEventModal isOpen={modalOpen} editing={editing} onClose={() => setModalOpen(false)} />

            <Modal
                isOpen={confirm !== null}
                onClose={() => setConfirm(null)}
                title={confirm?.action === "delete" ? "Delete Event" : "Cancel Event"}
            >
                <div className={styles.confirmBody}>
                    <p>
                        {confirm?.action === "delete" ? "Delete" : "Cancel"} <strong>{confirm?.event.title}</strong>?
                        {confirm?.action === "delete"
                            ? " This permanently removes the event and all of its occurrences."
                            : " This removes the event from the schedule."}
                    </p>
                    <div className={styles.confirmActions}>
                        <Button variant="ghost" size="small" onClick={() => setConfirm(null)}>
                            Keep event
                        </Button>
                        <Button variant="danger" size="small" onClick={confirmAction}>
                            {confirm?.action === "delete" ? "Delete event" : "Cancel event"}
                        </Button>
                    </div>
                </div>
            </Modal>

            <ContextMenu state={menuState} onClose={closeMenu} />

            {toast && <div className={styles.toast}>{toast}</div>}
        </div>
    );
}

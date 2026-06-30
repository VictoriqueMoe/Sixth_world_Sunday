import React, { useEffect, useState } from "react";
import { useAuth } from "../../hooks/useAuth";
import { useIsMobile } from "../../hooks/useIsMobile";
import { canManageMaps } from "../../utils/permissions";
import { useMaps } from "../../api/queries/maps";
import { useDeleteMap } from "../../api/mutations/maps";
import type { MapItem } from "../../types/api";
import { ChannelRail } from "../../components/layout/ChannelRail/ChannelRail";
import { Button } from "../../components/Button/Button";
import { Modal } from "../../components/Modal/Modal";
import { ContextMenu } from "../../components/ContextMenu/ContextMenu";
import { useContextMenu } from "../../components/ContextMenu/useContextMenu";
import { MapFormModal } from "../../components/maps/MapFormModal";
import styles from "./MapsPage.module.css";

export function MapsPage() {
    const { user } = useAuth();
    const isMobile = useIsMobile();
    const { maps, loading } = useMaps();
    const canManage = canManageMaps(user?.role);

    const deleteMutation = useDeleteMap();
    const { state: menuState, open: openMenu, close: closeMenu } = useContextMenu();

    const [modalOpen, setModalOpen] = useState(false);
    const [editing, setEditing] = useState<MapItem | null>(null);
    const [confirmDelete, setConfirmDelete] = useState<MapItem | null>(null);
    const [deleteError, setDeleteError] = useState("");

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

    function handleManage(e: React.MouseEvent, item: MapItem) {
        openMenu(e, [
            {
                id: "edit",
                label: "Edit map",
                icon: "✎",
                onClick: () => {
                    setEditing(item);
                    setModalOpen(true);
                },
            },
            {
                id: "delete",
                label: "Delete map",
                icon: "✕",
                variant: "danger",
                onClick: () => {
                    setDeleteError("");
                    setConfirmDelete(item);
                },
            },
        ]);
    }

    function closeDeleteConfirm() {
        setConfirmDelete(null);
        setDeleteError("");
    }

    async function confirmDeletion() {
        if (!confirmDelete) {
            return;
        }
        setDeleteError("");
        try {
            await deleteMutation.mutateAsync(confirmDelete.id);
            setConfirmDelete(null);
        } catch {
            setDeleteError("Couldn't delete this map. Please try again.");
        }
    }

    return (
        <div className={styles.shell}>
            {!isMobile && <ChannelRail />}
            <div className={styles.main}>
                <div className={styles.page}>
                    <div className={styles.head}>
                        <h2 className={styles.title}>
                            <span className={styles.headGlyph}>{"❖"}</span>
                            Maps
                        </h2>
                        {canManage && (
                            <Button variant="primary" size="small" onClick={openCreate}>
                                Add Map
                            </Button>
                        )}
                    </div>

                    {loading ? (
                        <div className={styles.empty}>Loading maps…</div>
                    ) : maps.length === 0 ? (
                        <div className={styles.empty}>No maps yet.</div>
                    ) : (
                        <div className={styles.list}>
                            {maps.map(item => (
                                <article key={item.id} className={styles.card}>
                                    <div className={styles.cardHead}>
                                        <div className={styles.cardInfo}>
                                            <h3 className={styles.cardTitle}>{item.title || "Untitled map"}</h3>
                                            {item.description && <p className={styles.cardDesc}>{item.description}</p>}
                                        </div>
                                        {item.can_manage && (
                                            <Button
                                                variant="secondary"
                                                size="icon"
                                                onClick={e => handleManage(e, item)}
                                                aria-label="Manage map"
                                            >
                                                {"⋯"}
                                            </Button>
                                        )}
                                    </div>
                                    <div className={styles.frameWrap}>
                                        <iframe
                                            className={styles.frame}
                                            src={item.embed_url}
                                            title={item.title || "Map"}
                                            loading="lazy"
                                            allowFullScreen
                                        />
                                    </div>
                                </article>
                            ))}
                        </div>
                    )}
                </div>
            </div>

            <MapFormModal isOpen={modalOpen} editing={editing} onClose={() => setModalOpen(false)} />

            <Modal isOpen={confirmDelete !== null} onClose={closeDeleteConfirm} title="Delete Map">
                <div className={styles.confirmBody}>
                    {deleteError && <div className={styles.error}>{deleteError}</div>}
                    <p>
                        Delete <strong>{confirmDelete?.title || "this map"}</strong>? This removes the embedded map for
                        everyone.
                    </p>
                    <div className={styles.confirmActions}>
                        <Button variant="ghost" size="small" onClick={closeDeleteConfirm}>
                            Cancel
                        </Button>
                        <Button
                            variant="danger"
                            size="small"
                            onClick={confirmDeletion}
                            disabled={deleteMutation.isPending}
                        >
                            {deleteMutation.isPending ? "Deleting…" : "Delete map"}
                        </Button>
                    </div>
                </div>
            </Modal>

            <ContextMenu state={menuState} onClose={closeMenu} />
        </div>
    );
}

import { useState } from "react";
import type { MapItem, SaveMapPayload } from "../../types/api";
import { useCreateMap, useUpdateMap } from "../../api/mutations/maps";
import { Modal } from "../Modal/Modal";
import { Input } from "../Input/Input";
import { TextArea } from "../TextArea/TextArea";
import { Button } from "../Button/Button";
import styles from "./MapFormModal.module.css";

interface MapFormModalProps {
    isOpen: boolean;
    onClose: () => void;
    editing?: MapItem | null;
}

export function MapFormModal({ isOpen, onClose, editing }: MapFormModalProps) {
    const createMutation = useCreateMap();
    const updateMutation = useUpdateMap();

    const [sourceUrl, setSourceUrl] = useState("");
    const [title, setTitle] = useState("");
    const [description, setDescription] = useState("");
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState("");

    const [prevOpen, setPrevOpen] = useState(isOpen);
    if (isOpen !== prevOpen) {
        setPrevOpen(isOpen);
        if (isOpen) {
            setSourceUrl(editing?.source_url ?? "");
            setTitle(editing?.title ?? "");
            setDescription(editing?.description ?? "");
            setSubmitting(false);
            setError("");
        }
    }

    async function handleSubmit() {
        if (!sourceUrl.trim() || submitting) {
            return;
        }

        setSubmitting(true);
        setError("");

        const payload: SaveMapPayload = {
            title: title.trim(),
            description: description.trim(),
            source_url: sourceUrl.trim(),
        };

        try {
            if (editing) {
                await updateMutation.mutateAsync({ id: editing.id, payload });
            } else {
                await createMutation.mutateAsync(payload);
            }
            onClose();
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to save map");
        } finally {
            setSubmitting(false);
        }
    }

    return (
        <Modal isOpen={isOpen} onClose={onClose} title={editing ? "Edit Map" : "Add Map"}>
            <div className={styles.body}>
                {error && <div className={styles.error}>{error}</div>}

                <div className={styles.field}>
                    <label className={styles.label}>Google My Maps link</label>
                    <Input
                        fullWidth
                        type="url"
                        value={sourceUrl}
                        onChange={e => setSourceUrl(e.target.value)}
                        placeholder="https://www.google.com/maps/d/edit?mid=…"
                    />
                    <span className={styles.hint}>Paste the share or edit link of a Google My Maps map.</span>
                </div>

                <div className={styles.field}>
                    <label className={styles.label}>Title</label>
                    <Input
                        fullWidth
                        type="text"
                        value={title}
                        onChange={e => setTitle(e.target.value)}
                        placeholder="Optional name"
                        maxLength={120}
                    />
                </div>

                <div className={styles.field}>
                    <label className={styles.label}>Description</label>
                    <TextArea
                        rows={3}
                        value={description}
                        onChange={e => setDescription(e.target.value)}
                        placeholder="Optional caption"
                    />
                </div>

                <div className={styles.actions}>
                    <Button variant="ghost" size="small" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button
                        variant="primary"
                        size="small"
                        onClick={handleSubmit}
                        disabled={submitting || !sourceUrl.trim()}
                    >
                        {submitting ? "Saving…" : editing ? "Save changes" : "Add map"}
                    </Button>
                </div>
            </div>
        </Modal>
    );
}

import { useState } from "react";
import type { ChatRoom } from "../../../types/api";
import { useCreateChannel } from "../../../api/mutations/chat";
import { Modal } from "../../Modal/Modal";
import { Input } from "../../Input/Input";
import { Button } from "../../Button/Button";
import styles from "./CreateChannelModal.module.css";

interface CreateChannelModalProps {
    isOpen: boolean;
    onClose: () => void;
    onCreated: (channel: ChatRoom) => void;
    initialKind?: "text" | "voice";
}

export function CreateChannelModal({ isOpen, onClose, onCreated, initialKind = "text" }: CreateChannelModalProps) {
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");
    const [kind, setKind] = useState<"text" | "voice">("text");
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState("");
    const createMutation = useCreateChannel();

    const [prevOpen, setPrevOpen] = useState(isOpen);
    if (isOpen !== prevOpen) {
        setPrevOpen(isOpen);
        if (isOpen) {
            setName("");
            setDescription("");
            setKind(initialKind);
            setError("");
            setSubmitting(false);
        }
    }

    async function handleSubmit() {
        if (!name.trim() || submitting) {
            return;
        }

        setSubmitting(true);
        setError("");
        try {
            const channel = await createMutation.mutateAsync({
                name: name.trim(),
                description: description.trim(),
                channel_kind: kind,
            });
            onCreated(channel);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to create channel");
        } finally {
            setSubmitting(false);
        }
    }

    return (
        <Modal isOpen={isOpen} onClose={onClose} title="Spin Up Channel">
            <div className={styles.body}>
                {error && <div className={styles.error}>{error}</div>}

                <div className={styles.field}>
                    <label className={styles.label}>Channel type</label>
                    <div className={styles.typePicker}>
                        <button
                            type="button"
                            className={`${styles.typeOption}${kind === "text" ? ` ${styles.typeActive}` : ""}`}
                            onClick={() => setKind("text")}
                        >
                            <span className={styles.typeGlyph}>{"#"}</span>
                            <span className={styles.typeLabel}>Text</span>
                            <span className={styles.typeDesc}>messages, media, replies</span>
                        </button>
                        <button
                            type="button"
                            className={`${styles.typeOption}${kind === "voice" ? ` ${styles.typeActive}` : ""}`}
                            onClick={() => setKind("voice")}
                        >
                            <span className={styles.typeGlyph}>{"◊"}</span>
                            <span className={styles.typeLabel}>Voice</span>
                            <span className={styles.typeDesc}>VC, screenshare + chat</span>
                        </button>
                    </div>
                </div>

                <div className={styles.field}>
                    <label className={styles.label}>Name</label>
                    <Input
                        fullWidth
                        type="text"
                        value={name}
                        onChange={e => setName(e.target.value)}
                        placeholder="e.g. sunday-run"
                        maxLength={80}
                    />
                </div>

                <div className={styles.field}>
                    <label className={styles.label}>Topic (optional)</label>
                    <Input
                        fullWidth
                        type="text"
                        value={description}
                        onChange={e => setDescription(e.target.value)}
                        placeholder="What's this channel about?"
                        maxLength={500}
                    />
                </div>

                <div className={styles.actions}>
                    <Button variant="ghost" size="small" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button variant="primary" size="small" onClick={handleSubmit} disabled={submitting || !name.trim()}>
                        {submitting ? "Creating..." : "Create channel"}
                    </Button>
                </div>
            </div>
        </Modal>
    );
}

import { useState } from "react";
import type { ChatRoom } from "../../../types/api";
import { useUpdateChannel } from "../../../api/mutations/chat";
import { Modal } from "../../Modal/Modal";
import { Input } from "../../Input/Input";
import { Button } from "../../Button/Button";
import styles from "./EditChannelModal.module.css";

interface EditChannelModalProps {
    channel: ChatRoom | null;
    onClose: () => void;
    onSaved?: (channel: ChatRoom) => void;
}

export function EditChannelModal({ channel, onClose, onSaved }: EditChannelModalProps) {
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState("");
    const updateMutation = useUpdateChannel();

    const [prevId, setPrevId] = useState<string | null>(null);
    const currentId = channel?.id ?? null;
    if (currentId !== prevId) {
        setPrevId(currentId);
        setName(channel?.name ?? "");
        setDescription(channel?.description ?? "");
        setError("");
        setSubmitting(false);
    }

    async function handleSubmit() {
        if (!channel || !name.trim() || submitting) {
            return;
        }

        setSubmitting(true);
        setError("");
        try {
            const updated = await updateMutation.mutateAsync({
                roomId: channel.id,
                name: name.trim(),
                description: description.trim(),
            });
            if (onSaved) {
                onSaved(updated);
            }
            onClose();
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to update channel");
        } finally {
            setSubmitting(false);
        }
    }

    return (
        <Modal isOpen={channel !== null} onClose={onClose} title="Edit Channel">
            <div className={styles.body}>
                {error && <div className={styles.error}>{error}</div>}

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
                    <label className={styles.label}>Topic</label>
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
                        {submitting ? "Saving..." : "Save changes"}
                    </Button>
                </div>
            </div>
        </Modal>
    );
}

import { useState } from "react";
import type { EventFormPayload, EventFrequency, EventItem, EventLocationType } from "../../../types/api";
import { useChannels } from "../../../api/queries/chat";
import { useCreateEvent, useUpdateEvent, useUploadEventCover } from "../../../api/mutations/events";
import { frequencyOptions } from "../../../utils/recurrence";
import { parseServerDate } from "../../../utils/time";
import { Modal } from "../../Modal/Modal";
import { Input } from "../../Input/Input";
import { Select } from "../../Select/Select";
import { TextArea } from "../../TextArea/TextArea";
import { Button } from "../../Button/Button";
import { MediaPickerButton } from "../../MediaPicker/MediaPicker";
import styles from "./CreateEventModal.module.css";

interface CreateEventModalProps {
    isOpen: boolean;
    onClose: () => void;
    editing?: EventItem | null;
}

const STEPS = ["Location", "Event Info", "Review"];

function toDateInput(d: Date): string {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${y}-${m}-${day}`;
}

function toTimeInput(d: Date): string {
    const hh = String(d.getHours()).padStart(2, "0");
    const mm = String(d.getMinutes()).padStart(2, "0");
    return `${hh}:${mm}`;
}

function combinedDate(dateStr: string, timeStr: string): Date | null {
    if (!dateStr) {
        return null;
    }
    const [y, m, d] = dateStr.split("-").map(Number);
    const [hh, mm] = (timeStr || "00:00").split(":").map(Number);
    const dt = new Date(y, (m || 1) - 1, d || 1, hh || 0, mm || 0);
    return Number.isNaN(dt.getTime()) ? null : dt;
}

export function CreateEventModal({ isOpen, onClose, editing }: CreateEventModalProps) {
    const { rooms } = useChannels();
    const voiceChannels = rooms.filter(r => r.channel_kind === "voice");

    const createMutation = useCreateEvent();
    const updateMutation = useUpdateEvent();
    const uploadCover = useUploadEventCover();

    const [step, setStep] = useState(0);
    const [locationType, setLocationType] = useState<EventLocationType>("voice");
    const [voiceRoomId, setVoiceRoomId] = useState("");
    const [externalUrl, setExternalUrl] = useState("");
    const [title, setTitle] = useState("");
    const [startDate, setStartDate] = useState("");
    const [startTime, setStartTime] = useState("");
    const [frequency, setFrequency] = useState<EventFrequency>("none");
    const [description, setDescription] = useState("");
    const [coverUrl, setCoverUrl] = useState("");
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState("");

    const [prevOpen, setPrevOpen] = useState(isOpen);
    if (isOpen !== prevOpen) {
        setPrevOpen(isOpen);
        if (isOpen) {
            setStep(0);
            setSubmitting(false);
            setError("");
            if (editing) {
                const start = parseServerDate(editing.start_at);
                setLocationType(editing.location_type);
                setVoiceRoomId(editing.voice_room_id ?? "");
                setExternalUrl(editing.external_url);
                setTitle(editing.title);
                setStartDate(start ? toDateInput(start) : "");
                setStartTime(start ? toTimeInput(start) : "");
                setFrequency(editing.frequency);
                setDescription(editing.description);
                setCoverUrl(editing.cover_url);
            } else {
                setLocationType("voice");
                setVoiceRoomId("");
                setExternalUrl("");
                setTitle("");
                setStartDate("");
                setStartTime("");
                setFrequency("none");
                setDescription("");
                setCoverUrl("");
            }
        }
    }

    const freqOptions = frequencyOptions(combinedDate(startDate, startTime));

    const locationValid = locationType === "voice" ? voiceRoomId !== "" : externalUrl.trim() !== "";
    const infoValid = title.trim() !== "" && startDate !== "" && startTime !== "";

    async function handleCover(files: File[]) {
        const file = files[0];
        if (!file) {
            return;
        }
        setError("");
        try {
            const res = await uploadCover.mutateAsync(file);
            setCoverUrl(res.cover_url);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to upload cover image");
        }
    }

    async function handleSubmit() {
        const start = combinedDate(startDate, startTime);
        if (!start || submitting) {
            return;
        }

        setSubmitting(true);
        setError("");

        const payload: EventFormPayload = {
            title: title.trim(),
            description: description.trim(),
            cover_url: coverUrl,
            location_type: locationType,
            voice_room_id: locationType === "voice" ? voiceRoomId : null,
            external_url: locationType === "external" ? externalUrl.trim() : "",
            start_at: start.toISOString(),
            frequency,
        };

        try {
            if (editing) {
                await updateMutation.mutateAsync({ id: editing.id, payload });
            } else {
                await createMutation.mutateAsync(payload);
            }
            onClose();
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to save event");
        } finally {
            setSubmitting(false);
        }
    }

    return (
        <Modal isOpen={isOpen} onClose={onClose} title={editing ? "Edit Event" : "Create Event"}>
            <div className={styles.body}>
                <div className={styles.stepper}>
                    {STEPS.map((label, i) => (
                        <div
                            key={label}
                            className={`${styles.step}${i === step ? ` ${styles.stepActive}` : ""}${
                                i < step ? ` ${styles.stepDone}` : ""
                            }`}
                        >
                            <span className={styles.stepBar} />
                            <span className={styles.stepLabel}>{label}</span>
                        </div>
                    ))}
                </div>

                {error && <div className={styles.error}>{error}</div>}

                {step === 0 && (
                    <div className={styles.section}>
                        <h4 className={styles.heading}>Where is it happening?</h4>
                        <div className={styles.typePicker}>
                            <button
                                type="button"
                                className={`${styles.typeOption}${locationType === "voice" ? ` ${styles.typeActive}` : ""}`}
                                onClick={() => setLocationType("voice")}
                            >
                                <span className={styles.typeGlyph}>{"\u{1F50A}"}</span>
                                <span className={styles.typeLabel}>Voice Channel</span>
                                <span className={styles.typeDesc}>Meet in a server VC</span>
                            </button>
                            <button
                                type="button"
                                className={`${styles.typeOption}${locationType === "external" ? ` ${styles.typeActive}` : ""}`}
                                onClick={() => setLocationType("external")}
                            >
                                <span className={styles.typeGlyph}>{"\u{1F517}"}</span>
                                <span className={styles.typeLabel}>Somewhere Else</span>
                                <span className={styles.typeDesc}>An external link</span>
                            </button>
                        </div>

                        {locationType === "voice" ? (
                            <div className={styles.field}>
                                <label className={styles.label}>Voice channel</label>
                                {voiceChannels.length === 0 ? (
                                    <div className={styles.note}>No voice channels exist yet.</div>
                                ) : (
                                    <Select value={voiceRoomId} onChange={e => setVoiceRoomId(e.target.value)}>
                                        <option value="">Select a voice channel…</option>
                                        {voiceChannels.map(c => (
                                            <option key={c.id} value={c.id}>
                                                {c.name}
                                            </option>
                                        ))}
                                    </Select>
                                )}
                            </div>
                        ) : (
                            <div className={styles.field}>
                                <label className={styles.label}>External link</label>
                                <Input
                                    fullWidth
                                    type="url"
                                    value={externalUrl}
                                    onChange={e => setExternalUrl(e.target.value)}
                                    placeholder="https://example.com/where"
                                />
                            </div>
                        )}
                    </div>
                )}

                {step === 1 && (
                    <div className={styles.section}>
                        <h4 className={styles.heading}>What&apos;s your event about?</h4>

                        <div className={styles.field}>
                            <label className={styles.label}>
                                Event Topic <span className={styles.req}>*</span>
                            </label>
                            <Input
                                fullWidth
                                type="text"
                                value={title}
                                onChange={e => setTitle(e.target.value)}
                                placeholder="What's your event?"
                                maxLength={120}
                            />
                        </div>

                        <div className={styles.row}>
                            <div className={styles.field}>
                                <label className={styles.label}>
                                    Start Date <span className={styles.req}>*</span>
                                </label>
                                <Input
                                    fullWidth
                                    type="date"
                                    value={startDate}
                                    onChange={e => setStartDate(e.target.value)}
                                />
                            </div>
                            <div className={styles.field}>
                                <label className={styles.label}>
                                    Start Time <span className={styles.req}>*</span>
                                </label>
                                <Input
                                    fullWidth
                                    type="time"
                                    value={startTime}
                                    onChange={e => setStartTime(e.target.value)}
                                />
                            </div>
                        </div>

                        <div className={styles.field}>
                            <label className={styles.label}>
                                Event Frequency <span className={styles.req}>*</span>
                            </label>
                            <Select value={frequency} onChange={e => setFrequency(e.target.value as EventFrequency)}>
                                {freqOptions.map(o => (
                                    <option key={o.value} value={o.value}>
                                        {o.label}
                                    </option>
                                ))}
                            </Select>
                        </div>

                        <div className={styles.field}>
                            <label className={styles.label}>Description</label>
                            <TextArea
                                rows={4}
                                value={description}
                                onChange={e => setDescription(e.target.value)}
                                placeholder="Tell people a little more about your event. Markdown, new lines and links are supported."
                            />
                        </div>

                        <div className={styles.field}>
                            <label className={styles.label}>Cover Image</label>
                            {coverUrl && <img className={styles.coverPreview} src={coverUrl} alt="" />}
                            <MediaPickerButton
                                multiple={false}
                                label={uploadCover.isPending ? "Uploading…" : "Upload Cover Image"}
                                onFiles={handleCover}
                                onError={setError}
                            />
                        </div>
                    </div>
                )}

                {step === 2 && (
                    <div className={styles.section}>
                        <h4 className={styles.heading}>Review</h4>
                        {coverUrl && <img className={styles.coverPreview} src={coverUrl} alt="" />}
                        <dl className={styles.review}>
                            <dt>Topic</dt>
                            <dd>{title.trim() || "—"}</dd>
                            <dt>When</dt>
                            <dd>
                                {combinedDate(startDate, startTime)?.toLocaleString() ?? "—"}
                                {frequency !== "none" && (
                                    <span className={styles.reviewFreq}>
                                        {" · "}
                                        {freqOptions.find(o => o.value === frequency)?.label}
                                    </span>
                                )}
                            </dd>
                            <dt>Location</dt>
                            <dd>
                                {locationType === "voice"
                                    ? (voiceChannels.find(c => c.id === voiceRoomId)?.name ?? "—")
                                    : externalUrl.trim() || "—"}
                            </dd>
                            {description.trim() && (
                                <>
                                    <dt>Details</dt>
                                    <dd className={styles.reviewDesc}>{description.trim()}</dd>
                                </>
                            )}
                        </dl>
                    </div>
                )}

                <div className={styles.actions}>
                    {step > 0 ? (
                        <Button variant="ghost" size="small" onClick={() => setStep(s => s - 1)}>
                            Back
                        </Button>
                    ) : (
                        <span />
                    )}
                    <div className={styles.actionsRight}>
                        <Button variant="secondary" size="small" onClick={onClose}>
                            Cancel
                        </Button>
                        {step < 2 ? (
                            <Button
                                variant="primary"
                                size="small"
                                disabled={step === 0 ? !locationValid : !infoValid}
                                onClick={() => setStep(s => s + 1)}
                            >
                                Next
                            </Button>
                        ) : (
                            <Button
                                variant="primary"
                                size="small"
                                disabled={submitting || !infoValid || !locationValid}
                                onClick={handleSubmit}
                            >
                                {submitting ? "Saving…" : editing ? "Save changes" : "Create event"}
                            </Button>
                        )}
                    </div>
                </div>
            </div>
        </Modal>
    );
}

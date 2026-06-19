import { RoomAudioRenderer, RoomContext, useLocalParticipant } from "@livekit/components-react";
import type { Room } from "livekit-client";

import { VoiceParticipantList } from "./VoiceParticipants";
import { Button } from "../../Button/Button";
import styles from "./Voice.module.css";

interface VoiceBarProps {
    room: Room;
    onLeave: () => void;
    canModerate?: boolean;
    onForceMute?: (identity: string, muted: boolean) => void;
}

export function VoiceBar({ room, onLeave, canModerate = false, onForceMute }: VoiceBarProps) {
    return (
        <RoomContext.Provider value={room}>
            <RoomAudioRenderer />
            <VoiceBarInner onLeave={onLeave} canModerate={canModerate} onForceMute={onForceMute} />
        </RoomContext.Provider>
    );
}

function VoiceBarInner({
    onLeave,
    canModerate,
    onForceMute,
}: {
    onLeave: () => void;
    canModerate: boolean;
    onForceMute?: (identity: string, muted: boolean) => void;
}) {
    const { localParticipant, isMicrophoneEnabled } = useLocalParticipant();
    const sharingScreen = localParticipant.isScreenShareEnabled;

    const toggleMute = () => {
        localParticipant.setMicrophoneEnabled(!isMicrophoneEnabled).catch(() => {});
    };

    const toggleScreenShare = () => {
        localParticipant.setScreenShareEnabled(!sharingScreen).catch(() => {});
    };

    return (
        <div className={styles.bar}>
            <div className={styles.presence}>
                <span className={styles.icon}>{"\u{1F50A}"}</span>
                <VoiceParticipantList canModerate={canModerate} onForceMute={onForceMute} />
            </div>

            <div className={styles.controls}>
                <Button variant="control" tone={isMicrophoneEnabled ? "default" : "danger"} onClick={toggleMute}>
                    {isMicrophoneEnabled ? "Mute" : "Unmute"}
                </Button>
                <Button variant="control" active={sharingScreen} onClick={toggleScreenShare}>
                    {sharingScreen ? "Stop share" : "Share"}
                </Button>
                <Button variant="control" tone="danger" onClick={onLeave}>
                    Leave
                </Button>
            </div>
        </div>
    );
}

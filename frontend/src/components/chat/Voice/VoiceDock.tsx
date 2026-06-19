import { useAuth } from "../../../hooks/useAuth";
import { isSiteStaff } from "../../../utils/permissions";
import { forceMuteVoiceParticipant } from "../../../api/endpoints";
import { useVoice } from "../../../context/voiceContextValue";
import { VoiceBar } from "./VoiceBar";
import styles from "./Voice.module.css";

export function VoiceDock() {
    const { user } = useAuth();
    const { status, room, activeRoomId, activeRoomName, leave } = useVoice();

    if (status !== "connected" || !room) {
        return null;
    }

    const canModerate = isSiteStaff(user?.role);

    return (
        <div className={styles.dock}>
            <div className={styles.dockHead}>
                <span className={styles.dockGlyph}>{"◊"}</span>
                <span className={styles.dockName}>{activeRoomName}</span>
            </div>
            <VoiceBar
                room={room}
                onLeave={leave}
                canModerate={canModerate}
                onForceMute={(id, muted) => {
                    forceMuteVoiceParticipant(activeRoomId ?? "", id, muted).catch(() => {});
                }}
            />
        </div>
    );
}

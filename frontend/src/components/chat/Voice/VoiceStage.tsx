import { useEffect, useRef, useState } from "react";
import {
    RoomAudioRenderer,
    RoomContext,
    VideoTrack,
    useIsSpeaking,
    useLocalParticipant,
    useParticipants,
    useTracks,
} from "@livekit/components-react";
import { AudioPresets, RemoteParticipant, Track } from "livekit-client";
import type { Participant, Room } from "livekit-client";

import type { ChatRoomMember } from "../../../types/api";
import { effectiveMemberUser } from "../../../utils/chatMembers";
import { Button } from "../../Button/Button";
import styles from "./VoiceStage.module.css";

type ScreenShareMode = "gaming" | "screenshare";

interface ScreenSharePreset {
    contentHint: "motion" | "detail";
    resolution: { width: number; height: number; frameRate: number };
    videoCodec: "vp9";
    degradationPreference: "maintain-framerate" | "maintain-resolution";
    maxBitrate: number;
}

const SCREEN_SHARE_PRESETS: Record<ScreenShareMode, ScreenSharePreset> = {
    gaming: {
        contentHint: "motion",
        resolution: { width: 1920, height: 1080, frameRate: 60 },
        videoCodec: "vp9",
        degradationPreference: "maintain-framerate",
        maxBitrate: 6_000_000,
    },
    screenshare: {
        contentHint: "detail",
        resolution: { width: 1920, height: 1080, frameRate: 15 },
        videoCodec: "vp9",
        degradationPreference: "maintain-resolution",
        maxBitrate: 2_500_000,
    },
};

interface VoiceStageProps {
    room: Room;
    members: ChatRoomMember[];
    canModerate: boolean;
    onLeave: () => void;
    onForceMute: (identity: string, muted: boolean) => void;
}

export function VoiceStage({ room, members, canModerate, onLeave, onForceMute }: VoiceStageProps) {
    return (
        <RoomContext.Provider value={room}>
            <RoomAudioRenderer />
            <VoiceStageInner
                room={room}
                members={members}
                canModerate={canModerate}
                onLeave={onLeave}
                onForceMute={onForceMute}
            />
        </RoomContext.Provider>
    );
}

function VoiceStageInner({ members, canModerate, onLeave, onForceMute }: VoiceStageProps) {
    const participants = useParticipants();
    const screenShares = useTracks([Track.Source.ScreenShare]);
    const { localParticipant, isMicrophoneEnabled } = useLocalParticipant();
    const [deafened, setDeafened] = useState(false);
    const [shareMenuOpen, setShareMenuOpen] = useState(false);
    const wasMicOnRef = useRef(false);
    const participantsRef = useRef(participants);
    const shareWrapRef = useRef<HTMLDivElement | null>(null);

    const sharingScreen = localParticipant.isScreenShareEnabled;

    const memberByIdentity = new Map<string, ChatRoomMember>();
    for (const m of members) {
        memberByIdentity.set(m.user.id, m);
    }

    useEffect(() => {
        participantsRef.current = participants;
    });

    useEffect(() => {
        for (const p of participants) {
            if (p instanceof RemoteParticipant) {
                p.setVolume(deafened ? 0 : 1);
            }
        }
    }, [participants, deafened]);

    useEffect(() => {
        return () => {
            for (const p of participantsRef.current) {
                if (p instanceof RemoteParticipant) {
                    p.setVolume(1);
                }
            }
        };
    }, []);

    useEffect(() => {
        if (!shareMenuOpen) {
            return;
        }
        function onDown(e: MouseEvent) {
            if (shareWrapRef.current && !shareWrapRef.current.contains(e.target as Node)) {
                setShareMenuOpen(false);
            }
        }
        document.addEventListener("mousedown", onDown);
        return () => {
            document.removeEventListener("mousedown", onDown);
        };
    }, [shareMenuOpen]);

    const toggleMute = () => {
        localParticipant.setMicrophoneEnabled(!isMicrophoneEnabled).catch(() => {});
    };

    const toggleDeafen = () => {
        const next = !deafened;
        setDeafened(next);

        if (next) {
            wasMicOnRef.current = isMicrophoneEnabled;
            if (isMicrophoneEnabled) {
                localParticipant.setMicrophoneEnabled(false).catch(() => {});
            }
        } else if (wasMicOnRef.current) {
            localParticipant.setMicrophoneEnabled(true).catch(() => {});
        }
    };

    const startShare = (mode: ScreenShareMode) => {
        setShareMenuOpen(false);
        const preset = SCREEN_SHARE_PRESETS[mode];

        localParticipant
            .setScreenShareEnabled(
                true,
                {
                    audio: {
                        echoCancellation: false,
                        noiseSuppression: false,
                        autoGainControl: false,
                    },
                    contentHint: preset.contentHint,
                    resolution: preset.resolution,
                },
                {
                    audioPreset: AudioPresets.musicHighQualityStereo,
                    dtx: false,
                    red: false,
                    forceStereo: true,
                    videoCodec: preset.videoCodec,
                    degradationPreference: preset.degradationPreference,
                    screenShareEncoding: {
                        maxBitrate: preset.maxBitrate,
                        maxFramerate: preset.resolution.frameRate,
                    },
                },
            )
            .catch(() => {});
    };

    const stopShare = () => {
        localParticipant.setScreenShareEnabled(false).catch(() => {});
    };

    return (
        <div className={styles.stage} data-sharing={screenShares.length > 0 ? "true" : "false"}>
            {screenShares.length > 0 && (
                <div className={styles.screenGrid}>
                    {screenShares.map(track => {
                        const isLocalShare = track.participant.isLocal;
                        const member = memberByIdentity.get(track.participant.identity);
                        const eu = member ? effectiveMemberUser(member) : undefined;
                        const label = isLocalShare ? "You" : eu?.display_name || track.participant.name || "Unknown";
                        return (
                            <div
                                key={track.participant.sid + track.publication.trackSid}
                                className={styles.screenTile}
                                onDoubleClick={e => {
                                    e.currentTarget.requestFullscreen().catch(() => {});
                                }}
                            >
                                <VideoTrack trackRef={track} className={styles.screenVideo} />
                                <span className={styles.screenLabel}>
                                    {"\u{1F5A5} "}
                                    {label}
                                </span>
                                <Button
                                    variant="secondary"
                                    size="icon"
                                    className={styles.fullscreenBtn}
                                    onClick={e => {
                                        e.stopPropagation();
                                        const tile = e.currentTarget.parentElement;
                                        if (tile) {
                                            tile.requestFullscreen().catch(() => {});
                                        }
                                    }}
                                    title="Fullscreen (or double-click)"
                                    aria-label="Fullscreen"
                                >
                                    {"⛶"}
                                </Button>
                            </div>
                        );
                    })}
                </div>
            )}

            <div className={styles.tileGrid}>
                {participants.map(p => (
                    <ParticipantTile
                        key={p.identity}
                        participant={p}
                        member={memberByIdentity.get(p.identity)}
                        canModerate={canModerate}
                        onForceMute={onForceMute}
                    />
                ))}
            </div>

            <div className={styles.controlBar}>
                <Button variant="control" tone={isMicrophoneEnabled ? "default" : "danger"} onClick={toggleMute}>
                    {isMicrophoneEnabled ? "\u{1F399} Mute" : "\u{1F507} Unmute"}
                </Button>
                <Button variant="control" tone={deafened ? "danger" : "default"} onClick={toggleDeafen}>
                    {deafened ? "\u{1F507} Undeafen" : "\u{1F3A7} Deafen"}
                </Button>
                <div className={styles.shareWrap} ref={shareWrapRef}>
                    <Button
                        variant="control"
                        active={sharingScreen}
                        onClick={() => (sharingScreen ? stopShare() : setShareMenuOpen(o => !o))}
                    >
                        {sharingScreen ? "\u{1F5A5} Stop share" : "\u{1F5A5} Share screen"}
                    </Button>
                    {shareMenuOpen && !sharingScreen && (
                        <div className={styles.shareMenu}>
                            <div className={styles.shareMenuHead}>What are you sharing?</div>
                            <button
                                type="button"
                                className={styles.shareOption}
                                onClick={() => startShare("screenshare")}
                            >
                                <span className={styles.shareOptTitle}>{"\u{1F5A5} Screenshare"}</span>
                                <span className={styles.shareOptDesc}>1080p · 15fps · sharp text &amp; detail</span>
                            </button>
                            <button type="button" className={styles.shareOption} onClick={() => startShare("gaming")}>
                                <span className={styles.shareOptTitle}>{"\u{1F3AE} Gaming / Video"}</span>
                                <span className={styles.shareOptDesc}>1080p · 60fps · smooth motion</span>
                            </button>
                        </div>
                    )}
                </div>
                <Button variant="control" tone="danger" onClick={onLeave}>
                    {"⏏ Disconnect"}
                </Button>
            </div>
        </div>
    );
}

interface ParticipantTileProps {
    participant: Participant;
    member?: ChatRoomMember;
    canModerate: boolean;
    onForceMute: (identity: string, muted: boolean) => void;
}

function ParticipantTile({ participant, member, canModerate, onForceMute }: ParticipantTileProps) {
    const isSpeaking = useIsSpeaking(participant);
    const micOn = participant.isMicrophoneEnabled;
    const eu = member ? effectiveMemberUser(member) : undefined;
    const name = eu?.display_name || participant.name || "Unknown";
    const avatarUrl = eu?.avatar_url;

    return (
        <div className={`${styles.tile} ${isSpeaking ? styles.tileSpeaking : ""}`} title={name}>
            <div className={styles.tileAvatar}>
                {avatarUrl ? (
                    <img src={avatarUrl} alt="" />
                ) : (
                    <span className={styles.tileInitial}>{name.charAt(0).toUpperCase()}</span>
                )}
                {!micOn && <span className={styles.tileMuted}>{"\u{1F507}"}</span>}
            </div>
            <span className={styles.tileName}>{name}</span>
            {!participant.isLocal && canModerate && (
                <Button
                    variant="control"
                    tone="danger"
                    size="small"
                    onClick={() => onForceMute(participant.identity, micOn)}
                    title={micOn ? "Mute for everyone" : "Unmute for everyone"}
                >
                    {micOn ? "Mute" : "Unmute"}
                </Button>
            )}
        </div>
    );
}

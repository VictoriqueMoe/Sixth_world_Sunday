import { Suspense, lazy, useEffect } from "react";
import { Navigate, useParams } from "react-router";
import { ChannelRail } from "../../components/layout/ChannelRail/ChannelRail";
import { useChannels } from "../../api/queries/chat";
import { useIsMobile } from "../../hooks/useIsMobile";
import styles from "./ChannelsLayout.module.css";

const RoomPage = lazy(() => import("../rooms/RoomPage").then(m => ({ default: m.RoomPage })));

export function ChannelsLayout() {
    const { roomId } = useParams<{ roomId: string }>();
    const { rooms, loading } = useChannels();
    const isMobile = useIsMobile();

    const firstChannel = rooms.find(c => c.channel_kind !== "voice") ?? rooms[0] ?? null;

    useEffect(() => {
        document.body.setAttribute("data-chat-page", "true");
        return () => {
            document.body.removeAttribute("data-chat-page");
        };
    }, []);

    function renderMain() {
        if (roomId) {
            return (
                <Suspense
                    fallback={
                        <div className={styles.empty}>
                            <p className={styles.emptyText}>Loading channel…</p>
                        </div>
                    }
                >
                    <RoomPage />
                </Suspense>
            );
        }

        if (loading) {
            return (
                <div className={styles.empty}>
                    <p className={styles.emptyText}>Loading channels…</p>
                </div>
            );
        }

        if (firstChannel) {
            return <Navigate to={`/channels/${firstChannel.id}`} replace />;
        }

        return (
            <div className={styles.empty}>
                <span className={styles.emptyGlyph}>{"#"}</span>
                <p className={styles.emptyText}>No channels yet.</p>
            </div>
        );
    }

    return (
        <div className={styles.shell} data-has-channel={roomId ? "true" : "false"}>
            {!isMobile && <ChannelRail />}
            <div className={styles.main}>{renderMain()}</div>
        </div>
    );
}

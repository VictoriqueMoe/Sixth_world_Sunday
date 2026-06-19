import { useEffect, useRef } from "react";

const IDLE_AFTER_MS = 60_000;

interface Options {
    roomId: string | undefined;
    sendWSMessage: (msg: object) => void;
    wsEpoch: number;
}

export function usePresenceReporter({ roomId, sendWSMessage, wsEpoch }: Options): void {
    const lastSentRef = useRef<"active" | "idle" | null>(null);
    const idleTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    useEffect(() => {
        if (!roomId) {
            return;
        }
        lastSentRef.current = null;

        const report = (state: "active" | "idle") => {
            if (lastSentRef.current === state) {
                return;
            }
            lastSentRef.current = state;
            sendWSMessage({ type: "viewer_state", data: { room_id: roomId, state } });
        };

        const clearIdleTimer = () => {
            if (idleTimerRef.current) {
                clearTimeout(idleTimerRef.current);
                idleTimerRef.current = null;
            }
        };

        const armIdleTimer = () => {
            clearIdleTimer();
            idleTimerRef.current = setTimeout(() => report("idle"), IDLE_AFTER_MS);
        };

        const onActivity = () => {
            if (document.visibilityState !== "visible" || !document.hasFocus()) {
                return;
            }
            report("active");
            armIdleTimer();
        };

        const goActive = () => {
            report("active");
            armIdleTimer();
        };

        const goIdle = () => {
            clearIdleTimer();
            report("idle");
        };

        const onVisibilityChange = () => {
            if (document.visibilityState === "visible" && document.hasFocus()) {
                goActive();
            } else {
                goIdle();
            }
        };

        const onFocus = () => {
            if (document.visibilityState === "visible") {
                goActive();
            }
        };

        const onBlur = () => {
            goIdle();
        };

        if (document.visibilityState === "visible" && document.hasFocus()) {
            goActive();
        } else {
            report("idle");
        }

        window.addEventListener("mousemove", onActivity);
        window.addEventListener("keydown", onActivity);
        window.addEventListener("click", onActivity);
        window.addEventListener("scroll", onActivity, true);
        window.addEventListener("touchstart", onActivity);
        window.addEventListener("focus", onFocus);
        window.addEventListener("blur", onBlur);
        document.addEventListener("visibilitychange", onVisibilityChange);

        return () => {
            clearIdleTimer();
            lastSentRef.current = null;
            window.removeEventListener("mousemove", onActivity);
            window.removeEventListener("keydown", onActivity);
            window.removeEventListener("click", onActivity);
            window.removeEventListener("scroll", onActivity, true);
            window.removeEventListener("touchstart", onActivity);
            window.removeEventListener("focus", onFocus);
            window.removeEventListener("blur", onBlur);
            document.removeEventListener("visibilitychange", onVisibilityChange);
        };
    }, [roomId, sendWSMessage, wsEpoch]);
}

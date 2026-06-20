import { type ReactNode, useEffect, useRef } from "react";
import { createPortal } from "react-dom";

interface MobileNavDrawerProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    children: ReactNode;
}

const EDGE_ZONE = 28;
const DRAG_SLOP = 8;

export function MobileNavDrawer({ open, onOpenChange, children }: MobileNavDrawerProps) {
    const drawerRef = useRef<HTMLDivElement>(null);
    const backdropRef = useRef<HTMLDivElement>(null);
    const openRef = useRef(open);

    useEffect(() => {
        openRef.current = open;
    }, [open]);

    useEffect(() => {
        let mode: "open" | "close" | null = null;
        let startX = 0;
        let startY = 0;
        let width = 0;
        let currentX = 0;
        let dragging = false;
        let snapTimer = 0;

        function paint(x: number) {
            const drawer = drawerRef.current;
            if (!drawer) {
                return;
            }

            drawer.style.transition = "none";
            drawer.style.transform = `translateX(${x}px)`;

            const backdrop = backdropRef.current;
            if (backdrop) {
                backdrop.style.transition = "none";
                backdrop.style.opacity = String(Math.max(0, Math.min(1, 1 + x / width)));
            }
        }

        function settle() {
            const drawer = drawerRef.current;
            if (drawer) {
                drawer.style.transition = "";
                drawer.style.transform = "";
            }

            const backdrop = backdropRef.current;
            if (backdrop) {
                backdrop.style.transition = "";
                backdrop.style.opacity = "";
            }
        }

        function snap(target: boolean) {
            const drawer = drawerRef.current;
            if (drawer) {
                drawer.style.transition = "transform 0.2s ease";
                drawer.style.transform = target ? "translateX(0)" : `translateX(-${width}px)`;
            }

            const backdrop = backdropRef.current;
            if (backdrop) {
                backdrop.style.transition = "opacity 0.2s ease";
                backdrop.style.opacity = target ? "1" : "0";
            }

            onOpenChange(target);
            snapTimer = window.setTimeout(settle, 220);
        }

        function onStart(event: TouchEvent) {
            const drawer = drawerRef.current;
            if (!drawer) {
                return;
            }

            window.clearTimeout(snapTimer);
            const touch = event.touches[0];
            width = drawer.offsetWidth || 268;
            dragging = false;

            if (openRef.current) {
                mode = "close";
            } else if (touch.clientX <= EDGE_ZONE) {
                mode = "open";
            } else {
                mode = null;
                return;
            }

            startX = touch.clientX;
            startY = touch.clientY;
            currentX = openRef.current ? 0 : -width;
        }

        function onMove(event: TouchEvent) {
            if (!mode) {
                return;
            }

            const touch = event.touches[0];
            const dx = touch.clientX - startX;
            const dy = touch.clientY - startY;

            if (!dragging) {
                if (Math.abs(dy) > Math.abs(dx) && Math.abs(dy) > DRAG_SLOP) {
                    mode = null;
                    return;
                }
                if (Math.abs(dx) < DRAG_SLOP) {
                    return;
                }
                dragging = true;
            }

            event.preventDefault();
            const base = mode === "open" ? -width : 0;
            currentX = Math.max(-width, Math.min(0, base + dx));
            paint(currentX);
        }

        function onEnd() {
            if (!mode) {
                return;
            }

            const didDrag = dragging;
            mode = null;
            dragging = false;

            if (!didDrag) {
                return;
            }

            snap(currentX > -width / 2);
        }

        document.addEventListener("touchstart", onStart, { passive: true });
        document.addEventListener("touchmove", onMove, { passive: false });
        document.addEventListener("touchend", onEnd, { passive: true });
        document.addEventListener("touchcancel", onEnd, { passive: true });

        return () => {
            window.clearTimeout(snapTimer);
            document.removeEventListener("touchstart", onStart);
            document.removeEventListener("touchmove", onMove);
            document.removeEventListener("touchend", onEnd);
            document.removeEventListener("touchcancel", onEnd);
        };
    }, [onOpenChange]);

    return createPortal(
        <>
            <div
                ref={backdropRef}
                className={`mobile-nav-backdrop${open ? " open" : ""}`}
                onClick={() => onOpenChange(false)}
            />
            <div ref={drawerRef} className={`mobile-nav-drawer${open ? " open" : ""}`}>
                {children}
            </div>
        </>,
        document.body,
    );
}

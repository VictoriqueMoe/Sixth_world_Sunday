import { Fragment, type ReactNode, useEffect, useRef } from "react";
import { createPortal } from "react-dom";
import styles from "./ContextMenu.module.css";

export interface ContextMenuItem {
    id: string;
    label: string;
    icon?: ReactNode;
    onClick: () => void;
    disabled?: boolean;
    variant?: "default" | "danger";
    separator?: boolean;
}

export interface ContextMenuState {
    x: number;
    y: number;
    items: ContextMenuItem[];
}

interface ContextMenuProps {
    state: ContextMenuState | null;
    onClose: () => void;
}

export function ContextMenu({ state, onClose }: ContextMenuProps) {
    const menuRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (!state) {
            return;
        }

        const handlePointerDown = (event: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
                onClose();
            }
        };

        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                onClose();
            }
        };

        document.addEventListener("mousedown", handlePointerDown);
        document.addEventListener("keydown", handleKeyDown);
        document.addEventListener("scroll", onClose, true);
        window.addEventListener("resize", onClose);

        return () => {
            document.removeEventListener("mousedown", handlePointerDown);
            document.removeEventListener("keydown", handleKeyDown);
            document.removeEventListener("scroll", onClose, true);
            window.removeEventListener("resize", onClose);
        };
    }, [state, onClose]);

    useEffect(() => {
        if (!state || !menuRef.current) {
            return;
        }

        const menu = menuRef.current;
        const rect = menu.getBoundingClientRect();

        let x = state.x;
        let y = state.y;
        if (x + rect.width > window.innerWidth) {
            x = state.x - rect.width;
        }
        if (y + rect.height > window.innerHeight) {
            y = state.y - rect.height;
        }

        menu.style.left = `${Math.max(8, x)}px`;
        menu.style.top = `${Math.max(8, y)}px`;
    }, [state]);

    if (!state || typeof document === "undefined") {
        return null;
    }

    const handleItemClick = (item: ContextMenuItem) => {
        if (item.disabled) {
            return;
        }
        item.onClick();
        onClose();
    };

    return createPortal(
        <div
            ref={menuRef}
            className={styles.menu}
            role="menu"
            aria-label="Context menu"
            style={{ left: state.x, top: state.y }}
        >
            {state.items.map((item, index) => (
                <Fragment key={item.id}>
                    {item.separator && index > 0 && <div className={styles.separator} />}
                    <button
                        type="button"
                        className={`${styles.item}${item.variant === "danger" ? ` ${styles.danger}` : ""}`}
                        onClick={() => handleItemClick(item)}
                        disabled={item.disabled}
                        role="menuitem"
                    >
                        {item.icon && <span className={styles.icon}>{item.icon}</span>}
                        <span className={styles.label}>{item.label}</span>
                    </button>
                </Fragment>
            ))}
        </div>,
        document.body,
    );
}

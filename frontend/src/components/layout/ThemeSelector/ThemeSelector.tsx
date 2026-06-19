import { useCallback, useRef, useState } from "react";
import { useTheme } from "../../../hooks/useTheme";
import { useClickOutside } from "../../../hooks/useClickOutside";
import { ToggleSwitch } from "../../ToggleSwitch/ToggleSwitch";
import styles from "./ThemeSelector.module.css";

export function ThemeSelector() {
    const { wideLayout, setWideLayout } = useTheme();
    const [isOpen, setIsOpen] = useState(false);
    const dropdownRef = useRef<HTMLDivElement>(null);

    useClickOutside(
        dropdownRef,
        useCallback(() => setIsOpen(false), []),
    );

    return (
        <div className={styles.selector} ref={dropdownRef}>
            <button
                className={styles.trigger}
                onClick={() => setIsOpen(!isOpen)}
                aria-expanded={isOpen}
                aria-haspopup="listbox"
            >
                <span className={styles.triggerLabel}>Appearance</span>
                <span className={`${styles.chevron}${isOpen ? ` ${styles.chevronOpen}` : ""}`}>{"▼"}</span>
            </button>

            {isOpen && (
                <div className={styles.dropdown} role="listbox">
                    <ToggleSwitch
                        enabled={wideLayout}
                        onChange={setWideLayout}
                        label="Wide layout"
                        description="Use the full width of the screen"
                    />
                </div>
            )}
        </div>
    );
}

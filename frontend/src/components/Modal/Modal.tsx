import type { PropsWithChildren } from "react";
import { Button } from "../Button/Button";
import styles from "./Modal.module.css";

interface ModalProps {
    isOpen: boolean;
    onClose: () => void;
    title: string;
}

export function Modal({ isOpen, onClose, title, children }: PropsWithChildren<ModalProps>) {
    if (!isOpen) {
        return null;
    }

    return (
        <div className={styles.overlay} onClick={onClose}>
            <div className={styles.modal} onClick={e => e.stopPropagation()}>
                <div className={styles.header}>
                    <h3>{title}</h3>
                    <Button variant="danger" size="icon" onClick={onClose} aria-label="Close">
                        {"\u2715"}
                    </Button>
                </div>
                <div className={styles.body}>{children}</div>
            </div>
        </div>
    );
}

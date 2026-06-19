import type { ButtonHTMLAttributes } from "react";
import styles from "./Button.module.css";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: "primary" | "secondary" | "danger" | "ghost" | "control";
    size?: "small" | "medium" | "icon";
    active?: boolean;
    tone?: "default" | "danger";
}

export function Button({
    variant = "secondary",
    size = "medium",
    active = false,
    tone = "default",
    className,
    children,
    ...rest
}: ButtonProps) {
    const classes = [
        styles.button,
        styles[variant],
        styles[size],
        active ? styles.active : "",
        tone === "danger" ? styles.toneDanger : "",
        className,
    ]
        .filter(Boolean)
        .join(" ");

    return (
        <button className={classes} {...rest}>
            {children}
        </button>
    );
}

import { NavLink } from "react-router";
import { useAuth } from "../../../hooks/useAuth";
import { ThemeSelector } from "../ThemeSelector/ThemeSelector";
import { NotificationBell } from "../NotificationBell/NotificationBell";
import { LoginButton } from "../../auth/LoginButton/LoginButton";
import { UserMenu } from "../../auth/UserMenu/UserMenu";
import { GlobalSearch } from "../GlobalSearch/GlobalSearch";
import styles from "./Header.module.css";

export function Header() {
    const { user, loading } = useAuth();

    return (
        <header className={styles.header}>
            <NavLink to="/" className={styles.brand} aria-label="Sixth World Sunday home">
                SIXTH WORLD <b>SUNDAY</b>
            </NavLink>

            <span className={styles.sep} aria-hidden="true" />

            <span className={styles.stat}>
                <span className={`${styles.dot} ${styles.dotGreen}`} /> node online
            </span>

            <span className={styles.spacer} />

            <GlobalSearch />

            <span className={styles.sep} aria-hidden="true" />

            <div className={styles.actions}>
                {!loading && user && <NotificationBell />}
                <ThemeSelector />
                {!loading && (user ? <UserMenu /> : <LoginButton />)}
            </div>
        </header>
    );
}

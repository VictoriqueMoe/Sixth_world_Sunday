import { usePageTitle } from "../../hooks/usePageTitle";
import styles from "./MaintenancePage.module.css";

interface MaintenancePageProps {
    title?: string;
    message?: string;
}

export function MaintenancePage({ title, message }: MaintenancePageProps) {
    usePageTitle("Maintenance");
    return (
        <div className={styles.page}>
            <div className={styles.card}>
                <span className={styles.status}>
                    <span className={styles.statusDot} aria-hidden="true" />
                    node offline // maintenance
                </span>
                <h1 className={styles.title}>{title || "The node is down for a hardware swap"}</h1>
                <p className={styles.message}>
                    {message || "Decker's rerouting the grid. Jack back in shortly, chummer."}
                </p>
            </div>
        </div>
    );
}

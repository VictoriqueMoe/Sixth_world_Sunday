import { Link } from "react-router";
import { usePageTitle } from "../../hooks/usePageTitle";
import styles from "./NotFoundPage.module.css";

export function NotFoundPage() {
    usePageTitle("Node Not Found");

    return (
        <div className={styles.page}>
            <div className={styles.code}>404</div>
            <h1 className={styles.title}>This node is off the grid</h1>
            <p className={styles.blurb}>
                Dead address, chummer. The host you pinged doesn't answer — maybe a broken link, maybe paydata that was
                scrubbed, maybe ICE took it down. Jack back to the main grid and try a fresh route.
            </p>
            <Link to="/" className={styles.cta}>
                Back to the Grid
            </Link>
        </div>
    );
}

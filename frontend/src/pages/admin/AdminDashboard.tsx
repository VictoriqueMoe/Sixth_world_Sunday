import { useAdminStats } from "../../api/queries/admin";
import { usePageTitle } from "../../hooks/usePageTitle";
import styles from "./AdminDashboard.module.css";

export function AdminDashboard() {
    usePageTitle("Admin");
    const { stats, loading } = useAdminStats();

    if (loading) {
        return <div className={styles.loading}>Reading node telemetry...</div>;
    }

    if (!stats) {
        return null;
    }

    return (
        <div className={styles.page}>
            <h1 className={styles.title}>Node Dashboard</h1>

            <div className={styles.statCards}>
                <div className={styles.statCard}>
                    <div className={styles.statLabel}>Runners</div>
                    <div className={styles.statValue}>{stats.total_users.toLocaleString()}</div>
                </div>
                <div className={styles.statCard}>
                    <div className={styles.statLabel}>Transmissions</div>
                    <div className={styles.statValue}>{stats.total_messages.toLocaleString()}</div>
                </div>
                <div className={styles.statCard}>
                    <div className={styles.statLabel}>Channels</div>
                    <div className={styles.statValue}>{stats.total_rooms.toLocaleString()}</div>
                </div>
            </div>

            <h2 className={styles.sectionTitle}>activity overview</h2>
            <table className={styles.table}>
                <thead>
                    <tr>
                        <th>Period</th>
                        <th>New Runners</th>
                        <th>New Transmissions</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>Last 24 hours</td>
                        <td>{stats.new_users_24h}</td>
                        <td>{stats.new_messages_24h}</td>
                    </tr>
                    <tr>
                        <td>Last 7 days</td>
                        <td>{stats.new_users_7d}</td>
                        <td>{stats.new_messages_7d}</td>
                    </tr>
                    <tr>
                        <td>Last 30 days</td>
                        <td>{stats.new_users_30d}</td>
                        <td>{stats.new_messages_30d}</td>
                    </tr>
                </tbody>
            </table>

            <h2 className={styles.sectionTitle}>top runners</h2>
            <div className={styles.activeUsersCard}>
                {stats.most_active_users.map(u => (
                    <div key={u.id} className={styles.activeUserRow}>
                        <div className={styles.activeUserInfo}>
                            {u.avatar_url ? (
                                <img className={styles.avatar} src={u.avatar_url} alt="" />
                            ) : (
                                <span className={styles.avatarPlaceholder}>{u.display_name[0]}</span>
                            )}
                            <span>{u.display_name}</span>
                        </div>
                        <span className={styles.actionCount}>{u.action_count} transmissions</span>
                    </div>
                ))}
                {stats.most_active_users.length === 0 && <div className={styles.loading}>No active runners yet</div>}
            </div>
        </div>
    );
}

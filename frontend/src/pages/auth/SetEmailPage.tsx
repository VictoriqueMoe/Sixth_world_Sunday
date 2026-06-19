import React, { useState } from "react";
import { Navigate, useNavigate } from "react-router";
import { useSetEmail } from "../../api/mutations/auth";
import { useAuth } from "../../hooks/useAuth";
import { usePageTitle } from "../../hooks/usePageTitle";
import { Input } from "../../components/Input/Input";
import styles from "./LoginPage.module.css";

export function SetEmailPage() {
    usePageTitle("Add your email");
    const navigate = useNavigate();
    const { user, loading } = useAuth();
    const setEmailMutation = useSetEmail();
    const [email, setEmail] = useState("");
    const [error, setError] = useState("");
    const [submitting, setSubmitting] = useState(false);

    if (loading) {
        return (
            <div className={styles.page}>
                <div className={styles.statusLine}>
                    syncing node<span className={styles.cursor}>_</span>
                </div>
            </div>
        );
    }
    if (!user) {
        return <Navigate to="/login" replace />;
    }
    if (user.email) {
        return <Navigate to="/" replace />;
    }

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault();
        setError("");
        setSubmitting(true);

        try {
            await setEmailMutation.mutateAsync(email);
            navigate("/");
        } catch (err) {
            setError(err instanceof Error ? err.message : "Something went wrong.");
        } finally {
            setSubmitting(false);
        }
    }

    return (
        <div className={styles.page}>
            <div className={styles.cardSolo}>
                <div className={styles.soloHead}>
                    <div className={styles.soloBadge}>NODE 6WS // IDENTITY</div>
                </div>

                <div className={`${styles.form} ${styles.formSolo}`}>
                    <h2 className={styles.title}>Link Comm Address</h2>
                    <p className={styles.sub}>register a recovery address for your handle</p>
                    <p className={styles.hint}>
                        A comm address is now required so you can recover your handle and stay jacked in. Enter one to
                        continue; we will transmit a confirmation link to verify it.
                    </p>

                    {error && <div className={styles.error}>{error}</div>}

                    <form onSubmit={handleSubmit}>
                        <div className={styles.fieldStack}>
                            <div className={styles.field}>
                                <label className={styles.label}>Comm Address</label>
                                <Input
                                    className={styles.input}
                                    type="email"
                                    fullWidth
                                    placeholder="you@thegrid.net"
                                    value={email}
                                    onChange={e => setEmail(e.target.value)}
                                    autoComplete="email"
                                />
                            </div>

                            <button className={styles.btnJack} type="submit" disabled={!email || submitting}>
                                {submitting ? "..." : "Link & Continue ▸"}
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    );
}

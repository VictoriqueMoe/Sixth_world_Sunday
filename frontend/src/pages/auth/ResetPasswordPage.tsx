import React, { useState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { useResetPassword } from "../../api/mutations/auth";
import { usePageTitle } from "../../hooks/usePageTitle";
import { Input } from "../../components/Input/Input";
import styles from "./LoginPage.module.css";

export function ResetPasswordPage() {
    usePageTitle("Reset Password");
    const navigate = useNavigate();
    const [params] = useSearchParams();
    const token = params.get("token") ?? "";
    const resetPasswordMutation = useResetPassword();
    const [password, setPassword] = useState("");
    const [confirm, setConfirm] = useState("");
    const [error, setError] = useState("");
    const [success, setSuccess] = useState(false);
    const [loading, setLoading] = useState(false);

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault();
        setError("");

        if (password !== confirm) {
            setError("Passwords do not match.");
            return;
        }

        setLoading(true);

        try {
            await resetPasswordMutation.mutateAsync({ token, newPassword: password });
            setSuccess(true);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Something went wrong.");
        } finally {
            setLoading(false);
        }
    }

    if (!token) {
        return (
            <div className={styles.page}>
                <div className={styles.cardSolo}>
                    <div className={styles.soloHead}>
                        <div className={styles.soloBadge}>NODE 6WS // RECOVERY</div>
                    </div>
                    <div className={`${styles.form} ${styles.formSolo}`}>
                        <h2 className={styles.title}>New Passkey</h2>
                        <p className={styles.sub}>set a new passkey for your handle</p>
                        <div className={styles.error}>This reset link is corrupted or incomplete.</div>
                        <button className={styles.btnGhost} type="button" onClick={() => navigate("/forgot-password")}>
                            Request a new link
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className={styles.page}>
            <div className={styles.cardSolo}>
                <div className={styles.soloHead}>
                    <div className={styles.soloBadge}>NODE 6WS // RECOVERY</div>
                </div>

                <div className={`${styles.form} ${styles.formSolo}`}>
                    <h2 className={styles.title}>New Passkey</h2>
                    <p className={styles.sub}>set a new passkey for your handle</p>

                    {error && <div className={styles.error}>{error}</div>}

                    {success ? (
                        <>
                            <div className={styles.success}>
                                Passkey rotated. You can now jack in with your new credentials.
                            </div>
                            <button className={styles.btnJack} type="button" onClick={() => navigate("/login")}>
                                Go to jack in ▸
                            </button>
                        </>
                    ) : (
                        <form onSubmit={handleSubmit}>
                            <div className={styles.fieldStack}>
                                <div className={styles.field}>
                                    <label className={styles.label}>New Passkey</label>
                                    <Input
                                        className={styles.input}
                                        type="password"
                                        fullWidth
                                        placeholder="••••••••••••"
                                        value={password}
                                        onChange={e => setPassword(e.target.value)}
                                        autoComplete="new-password"
                                    />
                                </div>

                                <div className={styles.field}>
                                    <label className={styles.label}>Confirm Passkey</label>
                                    <Input
                                        className={styles.input}
                                        type="password"
                                        fullWidth
                                        placeholder="••••••••••••"
                                        value={confirm}
                                        onChange={e => setConfirm(e.target.value)}
                                        autoComplete="new-password"
                                    />
                                </div>

                                <button
                                    className={styles.btnJack}
                                    type="submit"
                                    disabled={!password || !confirm || loading}
                                >
                                    {loading ? "..." : "Rotate Passkey ▸"}
                                </button>
                            </div>
                        </form>
                    )}
                </div>
            </div>
        </div>
    );
}

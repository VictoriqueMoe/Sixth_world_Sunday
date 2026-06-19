import React, { useRef, useState } from "react";
import { useNavigate } from "react-router";
import { Turnstile, type TurnstileInstance } from "@marsidev/react-turnstile";
import { useForgotPassword } from "../../api/mutations/auth";
import { useStaff } from "../../api/queries/auth";
import { usePageTitle } from "../../hooks/usePageTitle";
import { useSiteInfo } from "../../hooks/useSiteInfo";
import { Input } from "../../components/Input/Input";
import { ProfileLink } from "../../components/ProfileLink/ProfileLink";
import { ROLE_GROUPS } from "../../utils/permissions";
import styles from "./LoginPage.module.css";

export function ForgotPasswordPage() {
    usePageTitle("Forgot Password");
    const navigate = useNavigate();
    const siteInfo = useSiteInfo();
    const { staff } = useStaff();
    const forgotPasswordMutation = useForgotPassword();
    const [username, setUsername] = useState("");
    const [error, setError] = useState("");
    const [success, setSuccess] = useState("");
    const [loading, setLoading] = useState(false);
    const [turnstileToken, setTurnstileToken] = useState("");
    const turnstileRef = useRef<TurnstileInstance>(null);

    const turnstileEnabled = siteInfo.turnstile_enabled;
    const turnstileSiteKey = siteInfo.turnstile_site_key;

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault();
        setError("");
        setSuccess("");

        if (turnstileEnabled && !turnstileToken) {
            setError("Please complete the verification.");
            return;
        }

        setLoading(true);

        try {
            await forgotPasswordMutation.mutateAsync({
                username,
                turnstileToken: turnstileEnabled ? turnstileToken : undefined,
            });
            setSuccess("A password reset link has been sent to the email on your account.");
        } catch (err) {
            setError(err instanceof Error ? err.message : "Something went wrong.");
            setTurnstileToken("");
            turnstileRef.current?.reset();
        } finally {
            setLoading(false);
        }
    }

    return (
        <div className={styles.page}>
            <div className={styles.cardSolo}>
                <div className={styles.soloHead}>
                    <div className={styles.soloBadge}>NODE 6WS // RECOVERY</div>
                </div>

                <div className={`${styles.form} ${styles.formSolo}`}>
                    <h2 className={styles.title}>Recover Access</h2>
                    <p className={styles.sub}>request a new passkey for your handle</p>

                    {error && <div className={styles.error}>{error}</div>}
                    {success && <div className={styles.success}>{success}</div>}

                    {!success && (
                        <form onSubmit={handleSubmit}>
                            <div className={styles.fieldStack}>
                                <p className={styles.hint}>
                                    Enter your handle and we will transmit a reset link to the comm address on file.
                                </p>

                                <div className={styles.field}>
                                    <label className={styles.label}>Handle / SIN</label>
                                    <Input
                                        className={styles.input}
                                        type="text"
                                        fullWidth
                                        placeholder="ghost_in_the_grid"
                                        value={username}
                                        onChange={e => setUsername(e.target.value)}
                                        autoComplete="username"
                                    />
                                </div>

                                {turnstileEnabled && turnstileSiteKey && (
                                    <div className={styles.turnstile}>
                                        <Turnstile
                                            ref={turnstileRef}
                                            siteKey={turnstileSiteKey}
                                            onSuccess={setTurnstileToken}
                                            onExpire={() => setTurnstileToken("")}
                                            options={{
                                                refreshExpired: "auto",
                                                theme: "dark",
                                            }}
                                        />
                                    </div>
                                )}

                                <button
                                    className={styles.btnJack}
                                    type="submit"
                                    disabled={!username || loading || (turnstileEnabled && !turnstileToken)}
                                >
                                    {loading ? "..." : "Transmit Reset ▸"}
                                </button>
                            </div>
                        </form>
                    )}

                    {staff.length > 0 && (
                        <div className={styles.staffContact}>
                            <p className={styles.hint}>
                                No comm address on file? You cannot self-recover. Reach out to a node admin:
                            </p>
                            {ROLE_GROUPS.map(group => {
                                const members = staff.filter(member => member.role === group.role);
                                if (members.length === 0) {
                                    return null;
                                }
                                return (
                                    <div key={group.role} className={styles.staffGroup}>
                                        <span className={styles.staffGroupLabel}>{group.label}</span>
                                        <ul className={styles.staffList}>
                                            {members.map(member => (
                                                <li key={member.id}>
                                                    <ProfileLink user={member} size="small" showRoles={false} />
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                );
                            })}
                        </div>
                    )}

                    <button className={styles.btnGhost} type="button" onClick={() => navigate("/login")}>
                        Back to jack in
                    </button>
                </div>
            </div>
        </div>
    );
}

import { useEffect, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { useVerifyEmail } from "../../api/mutations/auth";
import { usePageTitle } from "../../hooks/usePageTitle";
import styles from "./LoginPage.module.css";

type Status = "verifying" | "ok" | "error";

export function VerifyEmailPage() {
    usePageTitle("Verify Email");
    const navigate = useNavigate();
    const [params] = useSearchParams();
    const token = params.get("token") ?? "";
    const verifyMutation = useVerifyEmail();
    const [status, setStatus] = useState<Status>(token ? "verifying" : "error");
    const started = useRef(false);

    useEffect(() => {
        if (started.current || !token) {
            return;
        }
        started.current = true;
        verifyMutation
            .mutateAsync(token)
            .then(() => setStatus("ok"))
            .catch(() => setStatus("error"));
    }, [token, verifyMutation]);

    return (
        <div className={styles.page}>
            <div className={styles.cardSolo}>
                <div className={styles.soloHead}>
                    <div className={styles.soloBadge}>NODE 6WS // IDENTITY</div>
                </div>

                <div className={`${styles.form} ${styles.formSolo}`}>
                    <h2 className={styles.title}>Verify Comm Address</h2>
                    <p className={styles.sub}>confirming your recovery address</p>

                    {status === "verifying" && (
                        <p className={styles.statusLine}>
                            verifying handshake<span className={styles.cursor}>_</span>
                        </p>
                    )}

                    {status === "ok" && (
                        <>
                            <div className={styles.success}>Comm address verified. Handshake complete.</div>
                            <button className={styles.btnJack} type="button" onClick={() => navigate("/")}>
                                Enter the Node ▸
                            </button>
                        </>
                    )}

                    {status === "error" && (
                        <>
                            <div className={styles.error}>This verification link is corrupted or expired.</div>
                            <button className={styles.btnGhost} type="button" onClick={() => navigate("/")}>
                                Return to grid
                            </button>
                        </>
                    )}
                </div>
            </div>
        </div>
    );
}
